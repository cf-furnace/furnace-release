package controller

import (
	"encoding/json"
	"time"

	"code.cloudfoundry.org/lager"

	"github.com/cf-furnace/route-emitter/cfroutes"
	"github.com/cf-furnace/route-emitter/models"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/cache"
	v1core "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3/typed/core/v1"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"
)

const PROCESS_GUID_LABEL = "cloudfoundry.org/process-guid"
const LOG_GUID_ANNO = "cloudfoundry.org/log-guid"
const ROUTING_INFO_ANNO = "cloudfoundry.org/routing-info"

type RCRouteController struct {
	logger     lager.Logger
	kubeClient v1core.CoreInterface
	events     chan<- models.Event

	store      cache.Store
	controller *framework.Controller
}

func NewRCRouteController(logger lager.Logger, coreClient v1core.CoreInterface, resyncPeriod time.Duration, events chan<- models.Event) *RCRouteController {
	rcrc := &RCRouteController{
		logger:     logger.Session("rc-route-controller"),
		kubeClient: coreClient,
		events:     events,
	}

	pguidSelector, err := labels.Parse(PROCESS_GUID_LABEL)
	if err != nil {
		return nil
	}

	rcrc.store, rcrc.controller = framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				options.LabelSelector = pguidSelector
				return coreClient.ReplicationControllers(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				options.LabelSelector = pguidSelector
				return coreClient.ReplicationControllers(api.NamespaceAll).Watch(options)
			},
		},
		&v1.ReplicationController{},
		resyncPeriod,
		rcrc,
	)

	return rcrc
}

func (rcrc *RCRouteController) Run(stopCh <-chan struct{}) {
	go rcrc.controller.Run(stopCh)
	<-stopCh
}

func (rcrc *RCRouteController) HasSynced() bool {
	return rcrc.HasSynced()
}

func (rcrc *RCRouteController) OnAdd(obj interface{}) {
	if d := asDesiredLRP(rcrc.logger, asRC(obj)); d != nil {
		rcrc.events <- &models.DesiredLRPCreatedEvent{
			DesiredLRP: d,
		}
	}
}

func (rcrc *RCRouteController) OnUpdate(old, new interface{}) {
	b := asDesiredLRP(rcrc.logger, asRC(old))
	a := asDesiredLRP(rcrc.logger, asRC(new))
	if b != nil && a != nil {
		rcrc.events <- &models.DesiredLRPChangedEvent{
			Before: b,
			After:  a,
		}
	}
}

func (rcrc *RCRouteController) OnDelete(obj interface{}) {
	if d := asDesiredLRP(rcrc.logger, asRC(obj)); d != nil {
		rcrc.events <- &models.DesiredLRPRemovedEvent{
			DesiredLRP: d,
		}
	}
}

func asRC(obj interface{}) *v1.ReplicationController {
	r, ok := obj.(*v1.ReplicationController)
	if !ok {
		return nil
	}

	return r
}

func asDesiredLRP(logger lager.Logger, r *v1.ReplicationController) *models.DesiredLRP {
	if r == nil {
		return nil
	}

	guid := r.Labels[PROCESS_GUID_LABEL]

	var routes cfroutes.CFRoutes
	if info := r.Annotations[ROUTING_INFO_ANNO]; info != "" {
		var r cfroutes.Routes
		err := json.Unmarshal([]byte(info), &r)
		if err != nil {
			logger.Error("routing-info-unmarshal-failed", err, lager.Data{"process-guid": guid})
			return nil
		}
		routes, err = cfroutes.CFRoutesFromRoutingInfo(r)
		if err != nil {
			logger.Error("cfroutes-from-routing-info-failed", err, lager.Data{"routes": info})
			return nil
		}
	}

	var ports []uint32
	if r.Spec.Template != nil {
		for _, cnr := range r.Spec.Template.Spec.Containers {
			for _, port := range cnr.Ports {
				ports = append(ports, uint32(port.ContainerPort))
			}
		}
	}

	return &models.DesiredLRP{
		ProcessGuid: guid,
		LogGuid:     r.Annotations[LOG_GUID_ANNO],

		Routes:          routes,
		Ports:           ports,
		ModificationTag: &models.ModificationTag{Epoch: r.ResourceVersion},
	}
}
