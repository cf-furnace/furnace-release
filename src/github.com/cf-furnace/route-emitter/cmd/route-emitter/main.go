package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket"

	"github.com/cf-furnace/route-emitter/controller"
	"github.com/cf-furnace/route-emitter/models"
	"github.com/cf-furnace/route-emitter/nats_emitter"
	"github.com/cf-furnace/route-emitter/routing_table"
	"github.com/cf-furnace/route-emitter/syncer"
	"github.com/cf-furnace/route-emitter/watcher"
	uuid "github.com/nu7hatch/gouuid"

	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/gunk/diegonats"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"

	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	"k8s.io/kubernetes/pkg/client/restclient"
)

const RouteEmitterLockSchemaKey = "route_emitter_lock"

var sessionName = flag.String(
	"sessionName",
	"route-emitter",
	"consul session name",
)

var consulCluster = flag.String(
	"consulCluster",
	"",
	"comma-separated list of consul server URLs (scheme://ip:port)",
)

var lockTTL = flag.Duration(
	"lockTTL",
	locket.LockTTL,
	"TTL for service lock",
)

var lockRetryInterval = flag.Duration(
	"lockRetryInterval",
	locket.RetryInterval,
	"interval to wait before retrying a failed lock acquisition",
)

var natsAddresses = flag.String(
	"natsAddresses",
	"127.0.0.1:4222",
	"comma-separated list of NATS addresses (ip:port)",
)

var natsUsername = flag.String(
	"natsUsername",
	"nats",
	"Username to connect to nats",
)

var natsPassword = flag.String(
	"natsPassword",
	"nats",
	"Password for nats user",
)

var syncInterval = flag.Duration(
	"syncInterval",
	time.Minute,
	"the interval between syncs of the routing table from etcd",
)

var dropsondePort = flag.Int(
	"dropsondePort",
	3457,
	"port the local metron agent is listening on",
)

var routeEmittingWorkers = flag.Int(
	"routeEmittingWorkers",
	20,
	"Max concurrency for sending route messages",
)

var kubeCluster = flag.String(
	"kubeCluster",
	"",
	"kubernetes API server URL (scheme://ip:port)",
)

var kubeCACert = flag.String(
	"kubeCACert",
	"",
	"path to kubernetes API server CA certificate",
)

var kubeClientCert = flag.String(
	"kubeClientCert",
	"",
	"path to client certificate for authentication with the kubernetes API server",
)

var kubeClientKey = flag.String(
	"kubeClientKey",
	"",
	"path to client key for authentication with the kubernetes API server",
)

const (
	dropsondeOrigin = "route_emitter"
)

func main() {
	debugserver.AddFlags(flag.CommandLine)
	cflager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, reconfigurableSink := cflager.New(*sessionName)
	natsClient := diegonats.NewClient()
	natsClient.SetPingInterval(30 * time.Second)
	clock := clock.NewClock()
	syncer := syncer.NewSyncer(clock, *syncInterval, natsClient, logger)

	initializeDropsonde(logger)

	natsClientRunner := diegonats.NewClientRunner(*natsAddresses, *natsUsername, *natsPassword, logger, natsClient)

	k8sclient := initializeCoreClient(logger)
	coreClient := k8sclient.Core()

	events := make(chan models.Event, 10)
	podController := ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		prc := controller.NewPodRouteController(coreClient, *syncInterval, events)
		close(ready)
		stop := make(chan struct{})
		go prc.Run(stop)
		<-signals
		close(stop)
		return nil
	})
	rcController := ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		rcrc := controller.NewRCRouteController(logger, coreClient, *syncInterval, events)
		close(ready)
		stop := make(chan struct{})
		go rcrc.Run(stop)
		<-signals
		close(stop)
		return nil
	})
	table := initializeRoutingTable(logger)
	emitter := initializeNatsEmitter(natsClient, logger)
	watcher := ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		return watcher.NewWatcher(clock, table, emitter, syncer.Events(), events, logger).Run(signals, ready)
	})

	syncRunner := ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		return syncer.Run(signals, ready)
	})

	lockMaintainer := initializeLockMaintainer(logger, *consulCluster, *sessionName, *lockTTL, *lockRetryInterval, clock)

	members := grouper.Members{
		{"lock-maintainer", lockMaintainer},
		{"nats-client", natsClientRunner},
		{"podRouteController", podController},
		{"replicationControllerRouteController", rcController},
		{"watcher", watcher},
		{"syncer", syncRunner},
	}

	if dbgAddr := debugserver.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", debugserver.Runner(dbgAddr, reconfigurableSink)},
		}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err := <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

func initializeDropsonde(logger lager.Logger) {
	dropsondeDestination := fmt.Sprint("localhost:", *dropsondePort)
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed to initialize dropsonde: %v", err)
	}
}

func initializeNatsEmitter(natsClient diegonats.NATSClient, logger lager.Logger) nats_emitter.NATSEmitter {
	workPool, err := workpool.NewWorkPool(*routeEmittingWorkers)
	if err != nil {
		logger.Fatal("failed-to-construct-nats-emitter-workpool", err, lager.Data{"num-workers": *routeEmittingWorkers}) // should never happen
	}

	return nats_emitter.New(natsClient, workPool, logger)
}

func initializeLockMaintainer(
	logger lager.Logger,
	consulCluster, sessionName string,
	lockTTL, lockRetryInterval time.Duration,
	clock clock.Clock,
) ifrit.Runner {
	consulClient, err := consuladapter.NewClientFromUrl(consulCluster)
	if err != nil {
		logger.Fatal("new-client-failed", err)
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		logger.Fatal("Couldn't generate uuid", err)
	}

	return locket.NewLock(logger, consulClient, locket.LockSchemaPath(RouteEmitterLockSchemaKey),
		[]byte(uuid.String()), clock, lockRetryInterval, lockTTL)
}

func initializeRoutingTable(logger lager.Logger) routing_table.RoutingTable {
	return routing_table.NewTable(logger)
}

func initializeCoreClient(logger lager.Logger) clientset.Interface {
	k8sClient, err := clientset.NewForConfig(&restclient.Config{
		Host: *kubeCluster,
		TLSClientConfig: restclient.TLSClientConfig{
			CertFile: *kubeClientCert,
			KeyFile:  *kubeClientKey,
			CAFile:   *kubeCACert,
		},
	})
	if err != nil {
		logger.Fatal("k8s-client-failed", err)
	}
	return k8sClient
}
