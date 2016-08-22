package controller

import (
	"time"

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

const ROUTING_ENDPOINTS_LABEL = "cloudfoundry.org/routing-endpoints"

type PodRouteController struct {
	kubeClient v1core.CoreInterface
	events     chan<- models.Event

	store      cache.Store
	controller *framework.Controller
}

func NewPodRouteController(coreClient v1core.CoreInterface, resyncPeriod time.Duration, events chan<- models.Event) *PodRouteController {
	prc := &PodRouteController{
		kubeClient: coreClient,
		events:     events,
	}

	pguidSelector, err := labels.Parse(ROUTING_ENDPOINTS_LABEL)
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
		ActualLRP: asActualLRP(asPod(obj)),
	}
}

func (prc *PodRouteController) OnUpdate(old, new interface{}) {
	prc.events <- &models.ActualLRPChangedEvent{
		Before: asActualLRP(asPod(old)),
		After:  asActualLRP(asPod(new)),
	}
}

func (prc *PodRouteController) OnDelete(obj interface{}) {
	prc.events <- &models.ActualLRPRemovedEvent{
		ActualLRP: asActualLRP(asPod(obj)),
	}
}

func asPod(obj interface{}) *v1.Pod {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return nil
	}

	return pod
}

func asActualLRP(p *v1.Pod) *models.ActualLRP {
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

	return &models.ActualLRP{
		ProcessGuid:  p.Labels[PROCESS_GUID_LABEL],
		InstanceGuid: string(p.UID),
		Index:        0,

		State:   state,
		Address: p.Status.HostIP,
		// Ports:
	}
}
