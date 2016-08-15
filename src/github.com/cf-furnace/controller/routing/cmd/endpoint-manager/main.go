package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	"k8s.io/kubernetes/pkg/client/restclient"

	"code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"

	"github.com/cf-furnace/controller/routing"
	"github.com/cf-furnace/controller/routing/iptables"
	"github.com/cf-furnace/pkg/port_pool"
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/gunk/command_runner/linux_command_runner"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

var dropsondePort = flag.Int(
	"dropsondePort",
	3457,
	"port the local metron agent is listening on",
)

var resyncInterval = flag.Duration(
	"resyncInterval",
	30*time.Second,
	"interval between route data synchronizations",
)

var kubeNodeName = flag.String(
	"kubeNodeName",
	"",
	"current kubernetes node name",
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
	dropsondeOrigin = "endpoint_manager"
)

func main() {
	debugserver.AddFlags(flag.CommandLine)
	cflager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, reconfigurableSink := cflager.New("endpoint-manager")
	initializeDropsonde(logger)

	if *kubeNodeName == "" {
		fatal(logger, "missing-kube-node-name", errors.New("kubeNodeName is a required flag"))
	}

	kubeClient := initializeKubeClient(logger)
	portPool, err := port_pool.New(61000, 1000, port_pool.State{})
	if err != nil {
		logger.Fatal("port-pool-new-failed", err)
	}

	ipt := iptables.New("iptables", linux_command_runner.New())

	if runtime.GOOS == "linux" {
		err = iptables.SetupNAT(ipt)
		if err != nil {
			logger.Fatal("iptables-setup-nat-failed", err)
		}
	}

	nodeRouteController := routing.NewNodeRouteController(
		kubeClient.Core(),
		*resyncInterval,
		*kubeNodeName,
		portPool,
		ipt,
		&sync.Mutex{},
	)

	members := grouper.Members{{
		Name:   "node-route-controller",
		Runner: ifrit.RunFunc(nodeRouteController.IfritRun),
	}}

	if dbgAddr := debugserver.DebugAddress(flag.CommandLine); dbgAddr != "" {
		debugMember := grouper.Member{
			Name:   "debug-server",
			Runner: debugserver.Runner(dbgAddr, reconfigurableSink),
		}
		members = append(grouper.Members{debugMember}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err = <-monitor.Wait()
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

func initializeKubeClient(logger lager.Logger) clientset.Interface {
	if *kubeCluster == "" {
		fatal(logger, "missing-kube-cluster", errors.New("kubeCluster is a required flag"))
	}

	client, err := clientset.NewForConfig(&restclient.Config{
		Host: *kubeCluster,
		TLSClientConfig: restclient.TLSClientConfig{
			CertFile: *kubeClientCert,
			KeyFile:  *kubeClientKey,
			CAFile:   *kubeCACert,
		},
	})

	if err != nil {
		fatal(logger, "failed-to-create-clientset", err, lager.Data{"address": *kubeCluster})
	}

	return client
}

func fatal(logger lager.Logger, key string, err error, data ...lager.Data) {
	logger.Error(key, err, data...)
	os.Exit(1)
}
