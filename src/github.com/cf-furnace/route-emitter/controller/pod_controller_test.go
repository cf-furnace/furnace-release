package controller_test

import (
	"time"

	"code.cloudfoundry.org/lager/lagertest"

	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/watch"

	"github.com/cf-furnace/pkg/cloudfoundry"
	"github.com/cf-furnace/pkg/kube_fakes"
	"github.com/cf-furnace/route-emitter/controller"
	"github.com/cf-furnace/route-emitter/models"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PodRouteController", func() {
	var (
		coreClient *kube_fakes.FakeCoreClient
		podClient  *kube_fakes.FakePodClient

		resyncPeriod time.Duration

		podController *controller.PodRouteController

		pg  cloudfoundry.ProcessGuid
		pod *v1.Pod

		events chan models.Event
	)

	BeforeEach(func() {
		coreClient = &kube_fakes.FakeCoreClient{}
		podClient = &kube_fakes.FakePodClient{}
		coreClient.PodsReturns(podClient)

		resyncPeriod = time.Second

		guid, err := uuid.NewV4()
		Expect(err).NotTo(HaveOccurred())

		pg = cloudfoundry.ProcessGuid{AppGuid: guid, AppVersion: guid}
		pod = &v1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					controller.PROCESS_GUID_LABEL: pg.ShortenedGuid(),
				},
			},
		}

		logger := lagertest.NewTestLogger("")
		events = make(chan models.Event, 10)
		podController = controller.NewPodRouteController(logger, coreClient, resyncPeriod, events)
	})

	Describe("Run", func() {
		var (
			eventCh chan watch.Event
			stopCh  chan struct{}
		)

		BeforeEach(func() {
			podList := &v1.PodList{Items: []v1.Pod{*pod}}
			podClient.ListReturns(podList, nil)

			eventCh = make(chan watch.Event, 1)
			watch := &kube_fakes.FakeWatch{}
			watch.ResultChanReturns(eventCh)
			podClient.WatchReturns(watch, nil)

			stopCh = make(chan struct{})
		})

		AfterEach(func() {
			select {
			case _, ok := <-stopCh:
				if ok {
					close(stopCh)
				}
			default:
			}
		})

		It("runs until stopped", func() {
			doneCh := make(chan struct{})
			go func() { podController.Run(stopCh); close(doneCh) }()

			Consistently(doneCh).ShouldNot(BeClosed())
			close(stopCh)
			Eventually(doneCh).Should(BeClosed())
		})

		It("lists all pods on the specified node", func() {
			go podController.Run(stopCh)

			Eventually(podClient.ListCallCount).Should(Equal(1))
			Eventually(podClient.WatchCallCount()).Should(Equal(1))
		})
	})

	Describe("OnAdd", func() {
		It("emits an Add event", func() {
			podController.OnAdd(pod)

			var e models.Event
			Eventually(events).Should(Receive(&e))
			Expect(e.EventType()).To(Equal(models.EventTypeActualLRPCreated))
			Expect(e.Key()).To(Equal(pg.ShortenedGuid()))
		})
	})

	Describe("OnUpdate", func() {
		var updated *v1.Pod

		BeforeEach(func() {
			newPod := *pod
			newPod.ObjectMeta.ResourceVersion = "new-resource-version"
			updated = &newPod
		})

		It("emits a Changed event", func() {
			podController.OnUpdate(pod, updated)

			var e models.Event
			Eventually(events).Should(Receive(&e))
			Expect(e.EventType()).To(Equal(models.EventTypeActualLRPChanged))
			Expect(e.Key()).To(Equal(pg.ShortenedGuid()))
		})
	})

	Describe("OnDelete", func() {
		It("emits a Delete event", func() {
			podController.OnDelete(pod)

			var e models.Event
			Eventually(events).Should(Receive(&e))
			Expect(e.EventType()).To(Equal(models.EventTypeActualLRPRemoved))
			Expect(e.Key()).To(Equal(pg.ShortenedGuid()))
		})
	})
})
