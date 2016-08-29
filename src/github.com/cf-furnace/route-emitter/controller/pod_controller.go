package controller

import (
	"encoding/json"
	"time"

	"code.cloudfoundry.org/lager"

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

const (
	NODE_PORTS_ANNOTATION = "cloudfoundry.org/node-ports"
)

type PodRouteController struct {
	logger     lager.Logger
	kubeClient v1core.CoreInterface
	events     chan<- models.Event

	store      cache.Store
	controller *framework.Controller
}

func NewPodRouteController(logger lager.Logger, coreClient v1core.CoreInterface, resyncPeriod time.Duration, events chan<- models.Event) *PodRouteController {
	prc := &PodRouteController{
		logger:     logger.Session("pod-route-controller"),
		kubeClient: coreClient,
		events:     events,
	}

	pguidSelector, err := labels.Parse(PROCESS_GUID_LABEL)
	if err != nil {
		return nil
	}

	prc.store, prc.controller = framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				options.LabelSelector = pguidSelector
				return coreClient.Pods(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				options.LabelSelector = pguidSelector
				return coreClient.Pods(api.NamespaceAll).Watch(options)
			},
		},
		&v1.Pod{},
		resyncPeriod,
		prc,
	)
	return prc
}

func (prc *PodRouteController) Run(stopCh <-chan struct{}) {
	go prc.controller.Run(stopCh)
	<-stopCh
}

func (prc *PodRouteController) HasSynced() bool {
	return prc.HasSynced()
}

func (prc *PodRouteController) OnAdd(obj interface{}) {
	prc.events <- &models.ActualLRPCreatedEvent{
		ActualLRP: asActualLRP(prc.logger, asPod(obj)),
	}
}

func (prc *PodRouteController) OnUpdate(old, new interface{}) {
	prc.events <- &models.ActualLRPChangedEvent{
		Before: asActualLRP(prc.logger, asPod(old)),
		After:  asActualLRP(prc.logger, asPod(new)),
	}
}

func (prc *PodRouteController) OnDelete(obj interface{}) {
	prc.events <- &models.ActualLRPRemovedEvent{
		ActualLRP: asActualLRP(prc.logger, asPod(obj)),
	}
}

func asPod(obj interface{}) *v1.Pod {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return nil
	}

	return pod
}

func asActualLRP(logger lager.Logger, p *v1.Pod) *models.ActualLRP {
	if p == nil {
		return nil
	}

	state := models.ActualLRPStateUnclaimed
	switch p.Status.Phase {
	case v1.PodPending:
		state = models.ActualLRPStateClaimed
	case v1.PodRunning:
		state = models.ActualLRPStateRunning
	case v1.PodSucceeded:
	case v1.PodFailed:
		state = models.ActualLRPStateCrashed
	case v1.PodUnknown:
	}

	guid := p.Labels[PROCESS_GUID_LABEL]

	portMapping := []models.PortMapping{}
	ports := p.Annotations[NODE_PORTS_ANNOTATION]
	if ports != "" {
		err := json.Unmarshal([]byte(ports), &portMapping)
		if err != nil {
			logger.Error("node-ports-unmarshal-failed", err, lager.Data{"process-guid": guid})
			return nil
		}
	}

	return &models.ActualLRP{
		ProcessGuid:  guid,
		InstanceGuid: string(p.UID),
		Index:        0,

		State:   state,
		Address: p.Status.HostIP,
		Ports:   portMapping,
	}
}
