package controller_test

import (
	"time"

	"code.cloudfoundry.org/lager/lagertest"

	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/watch"

	"github.com/cf-furnace/pkg/cloudfoundry"
	"github.com/cf-furnace/pkg/kube_fakes"
	"github.com/cf-furnace/route-emitter/cfroutes"
	"github.com/cf-furnace/route-emitter/controller"
	"github.com/cf-furnace/route-emitter/models"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReplicationControllerRouterController", func() {
	var (
		coreClient *kube_fakes.FakeCoreClient
		rcClient   *kube_fakes.FakeReplicationControllerClient

		resyncPeriod time.Duration

		rcController *controller.RCRouteController

		pg cloudfoundry.ProcessGuid
		rc *v1.ReplicationController

		events chan models.Event
	)

	BeforeEach(func() {
		coreClient = &kube_fakes.FakeCoreClient{}
		rcClient = &kube_fakes.FakeReplicationControllerClient{}
		coreClient.ReplicationControllersReturns(rcClient)

		resyncPeriod = time.Second

		guid, err := uuid.NewV4()
		Expect(err).NotTo(HaveOccurred())

		pg = cloudfoundry.ProcessGuid{AppGuid: guid, AppVersion: guid}
		rc = &v1.ReplicationController{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					controller.PROCESS_GUID_LABEL: pg.ShortenedGuid(),
				},
			},
		}

		logger := lagertest.NewTestLogger("")
		events = make(chan models.Event, 10)
		rcController = controller.NewRCRouteController(logger, coreClient, resyncPeriod, events)
	})

	Describe("Run", func() {
		var (
			eventCh chan watch.Event
			stopCh  chan struct{}
		)

		BeforeEach(func() {
			rcList := &v1.ReplicationControllerList{Items: []v1.ReplicationController{*rc}}
			rcClient.ListReturns(rcList, nil)

			eventCh = make(chan watch.Event, 1)
			watch := &kube_fakes.FakeWatch{}
			watch.ResultChanReturns(eventCh)
			rcClient.WatchReturns(watch, nil)

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
			go func() { rcController.Run(stopCh); close(doneCh) }()

			Consistently(doneCh).ShouldNot(BeClosed())
			close(stopCh)
			Eventually(doneCh).Should(BeClosed())
		})

		It("lists all pods on the specified node", func() {
			go rcController.Run(stopCh)

			Eventually(rcClient.ListCallCount).Should(Equal(1))
			Eventually(rcClient.WatchCallCount()).Should(Equal(1))
		})
	})

	Describe("OnAdd", func() {
		It("emits an Add event", func() {
			rcController.OnAdd(rc)

			var e models.Event
			Eventually(events).Should(Receive(&e))
			Expect(e.EventType()).To(Equal(models.EventTypeDesiredLRPCreated))
			Expect(e.Key()).To(Equal(pg.ShortenedGuid()))
		})

		Context("with an invalid RC", func() {
			BeforeEach(func() {
				rc.Annotations = map[string]string{
					controller.ROUTING_INFO_ANNO: "g",
				}
			})

			It("sends no event", func() {
				rcController.OnAdd(rc)

				Consistently(events).ShouldNot(Receive())
			})
		})
	})

	Describe("OnUpdate", func() {
		var updated *v1.ReplicationController

		BeforeEach(func() {
			newRC := *rc
			newRC.ObjectMeta.ResourceVersion = "new-resource-version"
			updated = &newRC
		})

		It("emits a Changed event", func() {
			rcController.OnUpdate(rc, updated)

			var e models.Event
			Eventually(events).Should(Receive(&e))
			Expect(e.EventType()).To(Equal(models.EventTypeDesiredLRPChanged))
			Expect(e.Key()).To(Equal(pg.ShortenedGuid()))
		})

		Context("with an invalid before RC", func() {
			BeforeEach(func() {
				rc.Annotations = map[string]string{
					controller.ROUTING_INFO_ANNO: "g",
				}
			})

			It("sends no event", func() {
				rcController.OnUpdate(rc, updated)

				Consistently(events).ShouldNot(Receive())
			})
		})

		Context("with an invalid after RC", func() {
			BeforeEach(func() {
				updated.Annotations = map[string]string{
					controller.ROUTING_INFO_ANNO: "g",
				}
			})

			It("sends no event", func() {
				rcController.OnUpdate(rc, updated)

				Consistently(events).ShouldNot(Receive())
			})
		})
	})

	Describe("OnDelete", func() {
		It("emits a Delete event", func() {
			rcController.OnDelete(rc)

			var e models.Event
			Eventually(events).Should(Receive(&e))
			Expect(e.EventType()).To(Equal(models.EventTypeDesiredLRPRemoved))
			Expect(e.Key()).To(Equal(pg.ShortenedGuid()))
		})

		Context("with an invalid RC", func() {
			BeforeEach(func() {
				rc.Annotations = map[string]string{
					controller.ROUTING_INFO_ANNO: "g",
				}
			})

			It("sends no event", func() {
				rcController.OnDelete(rc)

				Consistently(events).ShouldNot(Receive())
			})
		})
	})

	Describe("DesiredLRP", func() {
		var desiredLRP *models.DesiredLRP
		var expected bool

		BeforeEach(func() {
			rc.ResourceVersion = "aversion"
			rc.Annotations = map[string]string{
				controller.LOG_GUID_ANNO:     "log-guid",
				controller.ROUTING_INFO_ANNO: "{\"cf-router\":[{\"hostnames\":[\"dora.com\"],\"port\":8080}],\"tcp-router\":[]}",
			}
			rc.Spec.Template = &v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Ports: []v1.ContainerPort{{ContainerPort: 8080}}},
					},
				},
			}

			desiredLRP = nil
			expected = true
		})

		JustBeforeEach(func() {
			rcController.OnAdd(rc)
			if expected {
				var e models.Event
				Eventually(events).Should(Receive(&e))
				desiredLRP = e.(*models.DesiredLRPCreatedEvent).DesiredLRP
			} else {
				Consistently(events).ShouldNot(Receive())
			}
		})

		Context("when everything is present", func() {
			It("contains the necessary information", func() {
				Expect(desiredLRP).To(Equal(&models.DesiredLRP{
					ProcessGuid:     pg.ShortenedGuid(),
					LogGuid:         "log-guid",
					Routes:          cfroutes.CFRoutes{{Hostnames: []string{"dora.com"}, Port: 8080}},
					Ports:           []uint32{8080},
					ModificationTag: &models.ModificationTag{Epoch: "aversion"},
				}))
			})
		})

		Context("when there is no cf-routes", func() {
			BeforeEach(func() {
				delete(rc.Annotations, controller.ROUTING_INFO_ANNO)
			})

			It("has no routes", func() {
				Expect(desiredLRP).To(Equal(&models.DesiredLRP{
					ProcessGuid:     pg.ShortenedGuid(),
					LogGuid:         "log-guid",
					Routes:          nil,
					Ports:           []uint32{8080},
					ModificationTag: &models.ModificationTag{Epoch: "aversion"},
				}))
			})
		})

		Context("when there are container ports", func() {
			BeforeEach(func() {
				rc.Spec.Template.Spec.Containers = nil
			})

			It("has no routes", func() {
				Expect(desiredLRP).To(Equal(&models.DesiredLRP{
					ProcessGuid:     pg.ShortenedGuid(),
					LogGuid:         "log-guid",
					Routes:          cfroutes.CFRoutes{{Hostnames: []string{"dora.com"}, Port: 8080}},
					Ports:           nil,
					ModificationTag: &models.ModificationTag{Epoch: "aversion"},
				}))
			})
		})

		Context("when routing info is invalid", func() {
			BeforeEach(func() {
				expected = false
			})

			Context("when the JSON is invalid", func() {
				BeforeEach(func() {
					rc.Annotations[controller.ROUTING_INFO_ANNO] = "g{}arbage}"
				})

				It("has no desiredLRP created", func() {
					Expect(desiredLRP).To(BeNil())
				})
			})

			Context("when the CF Route JSON is invalid", func() {
				BeforeEach(func() {
					rc.Annotations[controller.ROUTING_INFO_ANNO] = "{\"cf-router\":{}]}"
				})

				It("has no desiredLRP created", func() {
					Expect(desiredLRP).To(BeNil())
				})
			})
		})
	})
})
