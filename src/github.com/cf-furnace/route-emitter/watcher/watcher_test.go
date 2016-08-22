package watcher_test

import (
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cf-furnace/route-emitter/cfroutes"
	"github.com/cf-furnace/route-emitter/models"
	"github.com/cf-furnace/route-emitter/nats_emitter/fake_nats_emitter"
	"github.com/cf-furnace/route-emitter/routing_table"
	"github.com/cf-furnace/route-emitter/routing_table/fake_routing_table"
	"github.com/cf-furnace/route-emitter/syncer"
	"github.com/cf-furnace/route-emitter/watcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/tedsuo/ifrit"

	fake_metrics_sender "github.com/cloudfoundry/dropsonde/metric_sender/fake"
	"github.com/cloudfoundry/dropsonde/metrics"
)

const logGuid = "some-log-guid"

type EventHolder struct {
	event models.Event
}

var nilEventHolder = EventHolder{}

var _ = Describe("Watcher", func() {
	const (
		expectedDomain                  = "domain"
		expectedProcessGuid             = "process-guid"
		expectedInstanceGuid            = "instance-guid"
		expectedIndex                   = 0
		expectedHost                    = "1.1.1.1"
		expectedExternalPort            = 11000
		expectedAdditionalExternalPort  = 22000
		expectedContainerPort           = 11
		expectedAdditionalContainerPort = 22
		expectedRouteServiceUrl         = "https://so.good.com"
	)

	var (
		table      *fake_routing_table.FakeRoutingTable
		emitter    *fake_nats_emitter.FakeNATSEmitter
		syncEvents syncer.Events
		events     chan models.Event

		clock          *fakeclock.FakeClock
		watcherProcess *watcher.Watcher
		process        ifrit.Process

		expectedRoutes     []string
		expectedRoutingKey routing_table.RoutingKey
		expectedCFRoute    cfroutes.CFRoute

		expectedAdditionalRoutes     []string
		expectedAdditionalRoutingKey routing_table.RoutingKey
		expectedAdditionalCFRoute    cfroutes.CFRoute

		dummyMessagesToEmit routing_table.MessagesToEmit
		fakeMetricSender    *fake_metrics_sender.FakeMetricSender

		logger *lagertest.TestLogger
	)

	BeforeEach(func() {
		table = &fake_routing_table.FakeRoutingTable{}
		emitter = &fake_nats_emitter.FakeNATSEmitter{}
		syncEvents = syncer.Events{
			Emit: make(chan struct{}),
		}
		events = make(chan models.Event, 10)
		logger = lagertest.NewTestLogger("test")

		dummyEndpoint := routing_table.Endpoint{InstanceGuid: expectedInstanceGuid, Index: expectedIndex, Host: expectedHost, Port: expectedContainerPort}
		dummyMessageFoo := routing_table.RegistryMessageFor(dummyEndpoint, routing_table.Route{Hostname: "foo.com", LogGuid: logGuid})
		dummyMessageBar := routing_table.RegistryMessageFor(dummyEndpoint, routing_table.Route{Hostname: "bar.com", LogGuid: logGuid})
		dummyMessagesToEmit = routing_table.MessagesToEmit{
			RegistrationMessages: []routing_table.RegistryMessage{dummyMessageFoo, dummyMessageBar},
		}

		clock = fakeclock.NewFakeClock(time.Now())

		watcherProcess = watcher.NewWatcher(clock, table, emitter, syncEvents, events, logger)

		expectedRoutes = []string{"route-1", "route-2"}
		expectedCFRoute = cfroutes.CFRoute{Hostnames: expectedRoutes, Port: expectedContainerPort, RouteServiceUrl: expectedRouteServiceUrl}
		expectedRoutingKey = routing_table.RoutingKey{
			ProcessGuid:   expectedProcessGuid,
			ContainerPort: expectedContainerPort,
		}

		expectedAdditionalRoutes = []string{"additional-1", "additional-2"}
		expectedAdditionalCFRoute = cfroutes.CFRoute{Hostnames: expectedAdditionalRoutes, Port: expectedAdditionalContainerPort}
		expectedAdditionalRoutingKey = routing_table.RoutingKey{
			ProcessGuid:   expectedProcessGuid,
			ContainerPort: expectedAdditionalContainerPort,
		}
		fakeMetricSender = fake_metrics_sender.NewFakeMetricSender()
		metrics.Initialize(fakeMetricSender, nil)
	})

	JustBeforeEach(func() {
		process = ifrit.Invoke(watcherProcess)
	})

	AfterEach(func() {
		process.Signal(os.Interrupt)
		Eventually(process.Wait()).Should(Receive())
	})

	Describe("Desired LRP changes", func() {
		Context("when a create event occurs", func() {
			var desiredLRP *models.DesiredLRP

			BeforeEach(func() {
				routes := cfroutes.CFRoutes{expectedCFRoute}
				desiredLRP = &models.DesiredLRP{
					// Domain:      "tests",
					ProcessGuid: expectedProcessGuid,
					Ports:       []uint32{expectedContainerPort},
					Routes:      routes,
					LogGuid:     logGuid,
				}
			})

			JustBeforeEach(func() {
				table.SetRoutesReturns(dummyMessagesToEmit)

				events <- &models.DesiredLRPCreatedEvent{desiredLRP}
			})

			It("should set the routes on the table", func() {
				Eventually(table.SetRoutesCallCount).Should(Equal(1))

				key, routes, _ := table.SetRoutesArgsForCall(0)
				Expect(key).To(Equal(expectedRoutingKey))
				Expect(routes).To(ConsistOf(
					routing_table.Route{
						Hostname:        expectedRoutes[0],
						LogGuid:         logGuid,
						RouteServiceUrl: expectedRouteServiceUrl,
					},
					routing_table.Route{
						Hostname:        expectedRoutes[1],
						LogGuid:         logGuid,
						RouteServiceUrl: expectedRouteServiceUrl,
					},
				))
			})

			It("sends a 'routes registered' metric", func() {
				Eventually(func() uint64 {
					return fakeMetricSender.GetCounter("RoutesRegistered")
				}).Should(BeEquivalentTo(2))
			})

			It("sends a 'routes unregistered' metric", func() {
				Eventually(func() uint64 {
					return fakeMetricSender.GetCounter("RoutesUnregistered")
				}).Should(BeEquivalentTo(0))
			})

			It("should emit whatever the table tells it to emit", func() {
				Eventually(emitter.EmitCallCount).Should(Equal(1))
				messagesToEmit := emitter.EmitArgsForCall(0)
				Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
			})

			Context("when there is a route service binding to only one hostname for a route", func() {
				BeforeEach(func() {
					cfRoute1 := cfroutes.CFRoute{
						Hostnames:       []string{"route-1"},
						Port:            expectedContainerPort,
						RouteServiceUrl: expectedRouteServiceUrl,
					}
					cfRoute2 := cfroutes.CFRoute{
						Hostnames: []string{"route-2"},
						Port:      expectedContainerPort,
					}
					routes := cfroutes.CFRoutes{cfRoute1, cfRoute2}
					desiredLRP.Routes = routes
				})
				It("registers all of the routes on the table", func() {
					Eventually(table.SetRoutesCallCount).Should(Equal(1))

					key, routes, _ := table.SetRoutesArgsForCall(0)
					Expect(key).To(Equal(expectedRoutingKey))
					Expect(routes).To(ConsistOf(
						routing_table.Route{
							Hostname:        "route-1",
							LogGuid:         logGuid,
							RouteServiceUrl: expectedRouteServiceUrl,
						},
						routing_table.Route{
							Hostname: "route-2",
							LogGuid:  logGuid,
						},
					))
				})

				It("emits whatever the table tells it to emit", func() {
					Eventually(emitter.EmitCallCount).Should(Equal(1))

					messagesToEmit := emitter.EmitArgsForCall(0)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
				})
			})

			Context("when there are multiple CF routes", func() {
				BeforeEach(func() {
					routes := cfroutes.CFRoutes{expectedCFRoute, expectedAdditionalCFRoute}
					desiredLRP.Routes = routes
				})

				It("registers all of the routes on the table", func() {
					Eventually(table.SetRoutesCallCount).Should(Equal(2))

					key1, routes1, _ := table.SetRoutesArgsForCall(0)
					key2, routes2, _ := table.SetRoutesArgsForCall(1)
					var routes = []routing_table.Route{}
					routes = append(routes, routes1...)
					routes = append(routes, routes2...)

					Expect([]routing_table.RoutingKey{key1, key2}).To(ConsistOf(expectedRoutingKey, expectedAdditionalRoutingKey))
					Expect(routes).To(ConsistOf(
						routing_table.Route{
							Hostname:        expectedRoutes[0],
							LogGuid:         logGuid,
							RouteServiceUrl: expectedRouteServiceUrl,
						},
						routing_table.Route{
							Hostname:        expectedRoutes[1],
							LogGuid:         logGuid,
							RouteServiceUrl: expectedRouteServiceUrl,
						},
						routing_table.Route{
							Hostname: expectedAdditionalRoutes[0],
							LogGuid:  logGuid,
						},
						routing_table.Route{
							Hostname: expectedAdditionalRoutes[1],
							LogGuid:  logGuid,
						},
					))
				})

				It("emits whatever the table tells it to emit", func() {
					Eventually(emitter.EmitCallCount).Should(Equal(2))

					messagesToEmit := emitter.EmitArgsForCall(0)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))

					messagesToEmit = emitter.EmitArgsForCall(1)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
				})
			})
		})

		Context("when a change event occurs", func() {
			var originalDesiredLRP, changedDesiredLRP *models.DesiredLRP

			BeforeEach(func() {
				table.SetRoutesReturns(dummyMessagesToEmit)
				routes := cfroutes.CFRoutes{{Hostnames: expectedRoutes, Port: expectedContainerPort}}

				originalDesiredLRP = &models.DesiredLRP{
					// Domain:      "tests",
					ProcessGuid: expectedProcessGuid,
					LogGuid:     logGuid,
					Routes:      routes,
					Instances:   3,
				}
				changedDesiredLRP = &models.DesiredLRP{
					// Domain:          "tests",
					ProcessGuid:     expectedProcessGuid,
					LogGuid:         logGuid,
					Routes:          routes,
					ModificationTag: &models.ModificationTag{Epoch: "abcd", Index: 1},
					Instances:       3,
				}
			})

			JustBeforeEach(func() {
				events <- &models.DesiredLRPChangedEvent{originalDesiredLRP, changedDesiredLRP}
			})

			Context("when scaling down the number of LRP instances", func() {
				BeforeEach(func() {
					changedDesiredLRP.Instances = 1

					table.EndpointsForIndexStub = func(key routing_table.RoutingKey, index int32) []routing_table.Endpoint {
						endpoint := routing_table.Endpoint{
							InstanceGuid:  fmt.Sprintf("instance-guid-%d", index),
							Index:         index,
							Host:          fmt.Sprintf("1.1.1.%d", index),
							Domain:        "domain",
							Port:          expectedExternalPort,
							ContainerPort: expectedContainerPort,
							Evacuating:    false,
						}

						return []routing_table.Endpoint{endpoint}
					}
				})

				It("removes route endpoints for instances that are no longer desired", func() {
					Eventually(table.RemoveEndpointCallCount).Should(Equal(2))
				})
			})

			It("should set the routes on the table", func() {
				Eventually(table.SetRoutesCallCount).Should(Equal(1))
				key, routes, _ := table.SetRoutesArgsForCall(0)
				Expect(key).To(Equal(expectedRoutingKey))
				Expect(routes).To(ConsistOf(
					routing_table.Route{
						Hostname: expectedRoutes[0],
						LogGuid:  logGuid,
					},
					routing_table.Route{
						Hostname: expectedRoutes[1],
						LogGuid:  logGuid,
					},
				))
			})

			It("sends a 'routes registered' metric", func() {
				Eventually(func() uint64 {
					return fakeMetricSender.GetCounter("RoutesRegistered")
				}).Should(BeEquivalentTo(2))
			})

			It("sends a 'routes unregistered' metric", func() {
				Eventually(func() uint64 {
					return fakeMetricSender.GetCounter("RoutesUnregistered")
				}).Should(BeEquivalentTo(0))
			})

			It("should emit whatever the table tells it to emit", func() {
				Eventually(emitter.EmitCallCount).Should(Equal(1))
				messagesToEmit := emitter.EmitArgsForCall(0)
				Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
			})

			Context("when CF routes are added without an associated container port", func() {
				BeforeEach(func() {
					routes := cfroutes.CFRoutes{expectedCFRoute, expectedAdditionalCFRoute}
					changedDesiredLRP.Routes = routes
				})

				It("registers all of the routes associated with a port on the table", func() {
					Eventually(table.SetRoutesCallCount).Should(Equal(2))

					key1, routes1, _ := table.SetRoutesArgsForCall(0)
					key2, routes2, _ := table.SetRoutesArgsForCall(1)
					var routes = []routing_table.Route{}
					routes = append(routes, routes1...)
					routes = append(routes, routes2...)

					Expect([]routing_table.RoutingKey{key1, key2}).To(ConsistOf(expectedRoutingKey, expectedAdditionalRoutingKey))
					Expect(routes).To(ConsistOf(
						routing_table.Route{
							Hostname:        expectedRoutes[0],
							LogGuid:         logGuid,
							RouteServiceUrl: expectedRouteServiceUrl,
						},
						routing_table.Route{
							Hostname:        expectedRoutes[1],
							LogGuid:         logGuid,
							RouteServiceUrl: expectedRouteServiceUrl,
						},
						routing_table.Route{
							Hostname: expectedAdditionalRoutes[0],
							LogGuid:  logGuid,
						},
						routing_table.Route{
							Hostname: expectedAdditionalRoutes[1],
							LogGuid:  logGuid,
						},
					))
				})

				It("emits whatever the table tells it to emit", func() {
					Eventually(emitter.EmitCallCount).Should(Equal(2))

					messagesToEmit := emitter.EmitArgsForCall(1)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
				})
			})

			Context("when CF routes and container ports are added", func() {
				BeforeEach(func() {
					routes := cfroutes.CFRoutes{expectedCFRoute, expectedAdditionalCFRoute}
					changedDesiredLRP.Routes = routes
				})

				It("registers all of the routes on the table", func() {
					Eventually(table.SetRoutesCallCount).Should(Equal(2))

					key1, routes1, _ := table.SetRoutesArgsForCall(0)
					key2, routes2, _ := table.SetRoutesArgsForCall(1)
					var routes = []routing_table.Route{}
					routes = append(routes, routes1...)
					routes = append(routes, routes2...)

					Expect([]routing_table.RoutingKey{key1, key2}).To(ConsistOf(expectedRoutingKey, expectedAdditionalRoutingKey))
					Expect(routes).To(ConsistOf(
						routing_table.Route{
							Hostname:        expectedRoutes[0],
							LogGuid:         logGuid,
							RouteServiceUrl: expectedRouteServiceUrl,
						},
						routing_table.Route{
							Hostname:        expectedRoutes[1],
							LogGuid:         logGuid,
							RouteServiceUrl: expectedRouteServiceUrl,
						},
						routing_table.Route{
							Hostname: expectedAdditionalRoutes[0],
							LogGuid:  logGuid,
						},
						routing_table.Route{
							Hostname: expectedAdditionalRoutes[1],
							LogGuid:  logGuid,
						},
					))
				})

				It("emits whatever the table tells it to emit", func() {
					Eventually(emitter.EmitCallCount).Should(Equal(2))

					messagesToEmit := emitter.EmitArgsForCall(0)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))

					messagesToEmit = emitter.EmitArgsForCall(1)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
				})
			})

			Context("when CF routes are removed", func() {
				BeforeEach(func() {
					routes := cfroutes.CFRoutes{}
					changedDesiredLRP.Routes = routes

					table.SetRoutesReturns(routing_table.MessagesToEmit{})
					table.RemoveRoutesReturns(dummyMessagesToEmit)
				})

				It("deletes the routes for the missng key", func() {
					Eventually(table.RemoveRoutesCallCount).Should(Equal(1))

					key, modTag := table.RemoveRoutesArgsForCall(0)
					Expect(key).To(Equal(expectedRoutingKey))
					Expect(modTag).To(Equal(changedDesiredLRP.ModificationTag))
				})

				It("emits whatever the table tells it to emit", func() {
					Eventually(emitter.EmitCallCount).Should(Equal(1))

					messagesToEmit := emitter.EmitArgsForCall(0)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
				})
			})
		})

		Context("when a delete event occurs", func() {
			var desiredLRP *models.DesiredLRP

			BeforeEach(func() {
				table.RemoveRoutesReturns(dummyMessagesToEmit)
				routes := cfroutes.CFRoutes{expectedCFRoute}
				desiredLRP = &models.DesiredLRP{
					// Domain:          "tests",
					ProcessGuid:     expectedProcessGuid,
					Ports:           []uint32{expectedContainerPort},
					Routes:          routes,
					LogGuid:         logGuid,
					ModificationTag: &models.ModificationTag{Epoch: "defg", Index: 2},
				}
			})

			JustBeforeEach(func() {
				events <- &models.DesiredLRPRemovedEvent{desiredLRP}
			})

			It("should remove the routes from the table", func() {
				Eventually(table.RemoveRoutesCallCount).Should(Equal(1))
				key, modTag := table.RemoveRoutesArgsForCall(0)
				Expect(key).To(Equal(expectedRoutingKey))
				Expect(modTag).To(Equal(desiredLRP.ModificationTag))
			})

			It("should emit whatever the table tells it to emit", func() {
				Eventually(emitter.EmitCallCount).Should(Equal(1))

				messagesToEmit := emitter.EmitArgsForCall(0)
				Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
			})

			Context("when there are multiple CF routes", func() {
				BeforeEach(func() {
					routes := cfroutes.CFRoutes{expectedCFRoute, expectedAdditionalCFRoute}
					desiredLRP.Routes = routes
				})

				It("should remove the routes from the table", func() {
					Eventually(table.RemoveRoutesCallCount).Should(Equal(2))

					key, modTag := table.RemoveRoutesArgsForCall(0)
					Expect(key).To(Equal(expectedRoutingKey))
					Expect(modTag).To(Equal(desiredLRP.ModificationTag))

					key, modTag = table.RemoveRoutesArgsForCall(1)
					Expect(key).To(Equal(expectedAdditionalRoutingKey))

					key, modTag = table.RemoveRoutesArgsForCall(0)
					Expect(key).To(Equal(expectedRoutingKey))
					Expect(modTag).To(Equal(desiredLRP.ModificationTag))
				})

				It("emits whatever the table tells it to emit", func() {
					Eventually(emitter.EmitCallCount).Should(Equal(2))

					messagesToEmit := emitter.EmitArgsForCall(0)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))

					messagesToEmit = emitter.EmitArgsForCall(1)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
				})
			})
		})
	})

	Describe("Actual LRP changes", func() {
		Context("when a create event occurs", func() {
			var (
				actualLRP            *models.ActualLRP
				actualLRPRoutingInfo *routing_table.ActualLRPRoutingInfo
			)

			Context("when the resulting LRP is in the RUNNING state", func() {
				BeforeEach(func() {
					actualLRP = &models.ActualLRP{
						ProcessGuid: expectedProcessGuid,
						Index:       expectedIndex,
						// Domain:"domain",
						InstanceGuid: expectedInstanceGuid,
						Address:      expectedHost,
						Ports:        []models.PortMapping{{expectedExternalPort, expectedContainerPort}, {expectedAdditionalExternalPort, expectedAdditionalContainerPort}},
						State:        models.ActualLRPStateRunning,
					}

					actualLRPRoutingInfo = &routing_table.ActualLRPRoutingInfo{
						ActualLRP:  actualLRP,
						Evacuating: false,
					}
				})

				JustBeforeEach(func() {
					table.AddEndpointReturns(dummyMessagesToEmit)
					events <- &models.ActualLRPCreatedEvent{actualLRP}
				})

				It("should log the net info", func() {
					Eventually(logger).Should(Say(
						fmt.Sprintf(
							`"net_info":\{"address":"%s","ports":\[\{"HostPort":%d,"ContainerPort":%d\},\{"HostPort":%d,"ContainerPort":%d\}\]\}`,
							expectedHost,
							expectedExternalPort,
							expectedContainerPort,
							expectedAdditionalExternalPort,
							expectedAdditionalContainerPort,
						),
					))
				})

				It("should add/update the endpoints on the table", func() {
					Eventually(table.AddEndpointCallCount).Should(Equal(2))

					keys := routing_table.RoutingKeysFromActual(actualLRP)
					endpoints, err := routing_table.EndpointsFromActual(actualLRPRoutingInfo)
					Expect(err).NotTo(HaveOccurred())

					key, endpoint := table.AddEndpointArgsForCall(0)
					Expect(keys).To(ContainElement(key))
					Expect(endpoint).To(Equal(endpoints[key.ContainerPort]))

					key, endpoint = table.AddEndpointArgsForCall(1)
					Expect(keys).To(ContainElement(key))
					Expect(endpoint).To(Equal(endpoints[key.ContainerPort]))
				})

				It("should emit whatever the table tells it to emit", func() {
					Eventually(emitter.EmitCallCount).Should(Equal(2))

					messagesToEmit := emitter.EmitArgsForCall(0)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
				})

				It("sends a 'routes registered' metric", func() {
					Eventually(func() uint64 {
						return fakeMetricSender.GetCounter("RoutesRegistered")
					}).Should(BeEquivalentTo(4))
				})

				It("sends a 'routes unregistered' metric", func() {
					Eventually(func() uint64 {
						return fakeMetricSender.GetCounter("RoutesUnregistered")
					}).Should(BeEquivalentTo(0))
				})
			})

			Context("when the resulting LRP is not in the RUNNING state", func() {
				JustBeforeEach(func() {
					actualLRP = &models.ActualLRP{
						ProcessGuid: expectedProcessGuid,
						Index:       expectedIndex,
						// Domain:"domain",
						InstanceGuid: expectedInstanceGuid,
						Address:      expectedHost,
						Ports:        []models.PortMapping{{expectedExternalPort, expectedContainerPort}, {expectedAdditionalExternalPort, expectedAdditionalContainerPort}},
						State:        models.ActualLRPStateUnclaimed,
					}

					events <- &models.ActualLRPCreatedEvent{actualLRP}
				})

				It("should NOT log the net info", func() {
					Consistently(logger).ShouldNot(Say(
						fmt.Sprintf(
							`"net_info":\{"address":"%s","ports":\[\{"HostPort":%d,"ContainerPort":%d\},\{"HostPort":%d,"ContainerPort":%d\}\]\}`,
							expectedHost,
							expectedExternalPort,
							expectedContainerPort,
							expectedExternalPort,
							expectedAdditionalContainerPort,
						),
					))
				})

				It("doesn't add/update the endpoint on the table", func() {
					Consistently(table.AddEndpointCallCount).Should(Equal(0))
				})

				It("doesn't emit", func() {
					Consistently(emitter.EmitCallCount).Should(Equal(0))
				})
			})
		})

		Context("when a change event occurs", func() {
			Context("when the resulting LRP is in the RUNNING state", func() {
				BeforeEach(func() {
					table.AddEndpointReturns(dummyMessagesToEmit)
				})

				JustBeforeEach(func() {
					beforeActualLRP := &models.ActualLRP{
						ProcessGuid: expectedProcessGuid,
						Index:       expectedIndex,
						// Domain:"domain",
						InstanceGuid: expectedInstanceGuid,
						State:        models.ActualLRPStateClaimed,
					}

					afterActualLRP := &models.ActualLRP{
						ProcessGuid: expectedProcessGuid,
						Index:       expectedIndex,
						// Domain:"domain",
						InstanceGuid: expectedInstanceGuid,
						Address:      expectedHost,
						Ports:        []models.PortMapping{{expectedExternalPort, expectedContainerPort}, {expectedAdditionalExternalPort, expectedAdditionalContainerPort}},
						State:        models.ActualLRPStateRunning,
					}

					events <- &models.ActualLRPChangedEvent{beforeActualLRP, afterActualLRP}
				})

				It("should log the new net info", func() {
					Eventually(logger).Should(Say(
						fmt.Sprintf(
							`"net_info":\{"address":"%s","ports":\[\{"HostPort":%d,"ContainerPort":%d\},\{"HostPort":%d,"ContainerPort":%d\}\]\}`,
							expectedHost,
							expectedExternalPort,
							expectedContainerPort,
							expectedAdditionalExternalPort,
							expectedAdditionalContainerPort,
						),
					))
				})

				It("should add/update the endpoint on the table", func() {
					Eventually(table.AddEndpointCallCount).Should(Equal(2))

					// Verify the arguments that were passed to AddEndpoint independent of which call was made first.
					type endpointArgs struct {
						key      routing_table.RoutingKey
						endpoint routing_table.Endpoint
					}
					args := make([]endpointArgs, 2)
					key, endpoint := table.AddEndpointArgsForCall(0)
					args[0] = endpointArgs{key, endpoint}
					key, endpoint = table.AddEndpointArgsForCall(1)
					args[1] = endpointArgs{key, endpoint}

					Expect(args).To(ConsistOf([]endpointArgs{
						endpointArgs{expectedRoutingKey, routing_table.Endpoint{
							InstanceGuid: expectedInstanceGuid,
							Index:        expectedIndex,
							Host:         expectedHost,
							//							Domain:        expectedDomain,
							Port:          expectedExternalPort,
							ContainerPort: expectedContainerPort,
						}},
						endpointArgs{expectedAdditionalRoutingKey, routing_table.Endpoint{
							InstanceGuid: expectedInstanceGuid,
							Index:        expectedIndex,
							Host:         expectedHost,
							//							Domain:        expectedDomain,
							Port:          expectedAdditionalExternalPort,
							ContainerPort: expectedAdditionalContainerPort,
						}},
					}))
				})

				It("should emit whatever the table tells it to emit", func() {
					Eventually(emitter.EmitCallCount).Should(Equal(2))

					messagesToEmit := emitter.EmitArgsForCall(0)
					Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
				})

				It("sends a 'routes registered' metric", func() {
					Eventually(func() uint64 {
						return fakeMetricSender.GetCounter("RoutesRegistered")
					}).Should(BeEquivalentTo(4))
				})

				It("sends a 'routes unregistered' metric", func() {
					Eventually(func() uint64 {
						return fakeMetricSender.GetCounter("RoutesUnregistered")
					}).Should(BeEquivalentTo(0))
				})
			})
		})

		Context("when the resulting LRP transitions away from the RUNNING state", func() {
			JustBeforeEach(func() {
				table.RemoveEndpointReturns(dummyMessagesToEmit)
				beforeActualLRP := &models.ActualLRP{
					ProcessGuid: expectedProcessGuid,
					Index:       expectedIndex,
					// Domain:"domain",
					InstanceGuid: expectedInstanceGuid,
					Address:      expectedHost,
					Ports:        []models.PortMapping{{expectedExternalPort, expectedContainerPort}, {expectedAdditionalExternalPort, expectedAdditionalContainerPort}},
					State:        models.ActualLRPStateRunning,
				}

				afterActualLRP := &models.ActualLRP{
					ProcessGuid: expectedProcessGuid,
					Index:       expectedIndex,

					State: models.ActualLRPStateUnclaimed,
				}

				events <- &models.ActualLRPChangedEvent{beforeActualLRP, afterActualLRP}
			})

			It("should log the previous net info", func() {
				Eventually(logger).Should(Say(
					fmt.Sprintf(
						`"net_info":\{"address":"%s","ports":\[\{"HostPort":%d,"ContainerPort":%d\},\{"HostPort":%d,"ContainerPort":%d\}\]\}`,
						expectedHost,
						expectedExternalPort,
						expectedContainerPort,
						expectedAdditionalExternalPort,
						expectedAdditionalContainerPort,
					),
				))
			})

			It("should remove the endpoint from the table", func() {
				Eventually(table.RemoveEndpointCallCount).Should(Equal(2))

				key, endpoint := table.RemoveEndpointArgsForCall(0)
				Expect(key).To(Equal(expectedRoutingKey))
				Expect(endpoint).To(Equal(routing_table.Endpoint{
					InstanceGuid: expectedInstanceGuid,
					Index:        expectedIndex,
					Host:         expectedHost,
					//	Domain:        expectedDomain,
					Port:          expectedExternalPort,
					ContainerPort: expectedContainerPort,
				}))

				key, endpoint = table.RemoveEndpointArgsForCall(1)
				Expect(key).To(Equal(expectedAdditionalRoutingKey))
				Expect(endpoint).To(Equal(routing_table.Endpoint{
					InstanceGuid: expectedInstanceGuid,
					Index:        expectedIndex,
					Host:         expectedHost,
					//Domain:        expectedDomain,
					Port:          expectedAdditionalExternalPort,
					ContainerPort: expectedAdditionalContainerPort,
				}))

			})

			It("should emit whatever the table tells it to emit", func() {
				Eventually(emitter.EmitCallCount).Should(Equal(2))

				messagesToEmit := emitter.EmitArgsForCall(0)
				Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
			})
		})

		Context("when the endpoint neither starts nor ends in the RUNNING state", func() {
			JustBeforeEach(func() {
				beforeActualLRP := &models.ActualLRP{
					ProcessGuid: expectedProcessGuid,
					Index:       expectedIndex,

					State: models.ActualLRPStateUnclaimed,
				}
				afterActualLRP := &models.ActualLRP{
					ProcessGuid:  expectedProcessGuid,
					Index:        expectedIndex,
					InstanceGuid: expectedInstanceGuid,

					State: models.ActualLRPStateUnclaimed,
				}

				events <- &models.ActualLRPChangedEvent{beforeActualLRP, afterActualLRP}
			})

			It("should NOT log the net info", func() {
				Consistently(logger).ShouldNot(Say(
					fmt.Sprintf(
						`"net_info":\{"address":"%s","ports":\[\{"HostPort":%d,"ContainerPort":%d\},\{"HostPort":%d,"ContainerPort":%d\}\]\}`,
						expectedHost,
						expectedExternalPort,
						expectedContainerPort,
						expectedExternalPort,
						expectedAdditionalContainerPort,
					),
				))
			})

			It("should not remove the endpoint", func() {
				Consistently(table.RemoveEndpointCallCount).Should(BeZero())
			})

			It("should not add or update the endpoint", func() {
				Consistently(table.AddEndpointCallCount).Should(BeZero())
			})

			It("should not emit anything", func() {
				Consistently(emitter.EmitCallCount).Should(Equal(0))
			})
		})
	})

	Context("when a delete event occurs", func() {
		Context("when the actual is in the RUNNING state", func() {
			BeforeEach(func() {
				table.RemoveEndpointReturns(dummyMessagesToEmit)
			})

			JustBeforeEach(func() {
				actualLRP := &models.ActualLRP{
					ProcessGuid: expectedProcessGuid,
					Index:       expectedIndex,
					// Domain:"domain",
					InstanceGuid: expectedInstanceGuid,
					Address:      expectedHost,
					Ports:        []models.PortMapping{{expectedExternalPort, expectedContainerPort}, {expectedAdditionalExternalPort, expectedAdditionalContainerPort}},
					State:        models.ActualLRPStateRunning,
				}

				events <- &models.ActualLRPRemovedEvent{actualLRP}
			})

			It("should log the previous net info", func() {
				Eventually(logger).Should(Say(
					fmt.Sprintf(
						`"net_info":\{"address":"%s","ports":\[\{"HostPort":%d,"ContainerPort":%d\},\{"HostPort":%d,"ContainerPort":%d\}\]\}`,
						expectedHost,
						expectedExternalPort,
						expectedContainerPort,
						expectedAdditionalExternalPort,
						expectedAdditionalContainerPort,
					),
				))
			})

			It("should remove the endpoint from the table", func() {
				Eventually(table.RemoveEndpointCallCount).Should(Equal(2))

				key, endpoint := table.RemoveEndpointArgsForCall(0)
				Expect(key).To(Equal(expectedRoutingKey))
				Expect(endpoint).To(Equal(routing_table.Endpoint{
					InstanceGuid: expectedInstanceGuid,
					Index:        expectedIndex,
					Host:         expectedHost,
					//	Domain:        expectedDomain,
					Port:          expectedExternalPort,
					ContainerPort: expectedContainerPort,
				}))

				key, endpoint = table.RemoveEndpointArgsForCall(1)
				Expect(key).To(Equal(expectedAdditionalRoutingKey))
				Expect(endpoint).To(Equal(routing_table.Endpoint{
					InstanceGuid: expectedInstanceGuid,
					Index:        expectedIndex,
					Host:         expectedHost,
					//	Domain:        expectedDomain,
					Port:          expectedAdditionalExternalPort,
					ContainerPort: expectedAdditionalContainerPort,
				}))

			})

			It("should emit whatever the table tells it to emit", func() {
				Eventually(emitter.EmitCallCount).Should(Equal(2))

				messagesToEmit := emitter.EmitArgsForCall(0)
				Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))

				messagesToEmit = emitter.EmitArgsForCall(1)
				Expect(messagesToEmit).To(Equal(dummyMessagesToEmit))
			})
		})

		Context("when the actual is not in the RUNNING state", func() {
			JustBeforeEach(func() {
				actualLRP := &models.ActualLRP{
					ProcessGuid: expectedProcessGuid,
					Index:       expectedIndex,
					// Domain:"domain",
					State: models.ActualLRPStateCrashed,
				}

				events <- &models.ActualLRPRemovedEvent{actualLRP}
			})

			It("should NOT log the net info", func() {
				Consistently(logger).ShouldNot(Say(
					fmt.Sprintf(
						`"net_info":\{"address":"%s","ports":\[\{"HostPort":%d,"ContainerPort":%d\},\{"HostPort":%d,"ContainerPort":%d\}\]\}`,
						expectedHost,
						expectedExternalPort,
						expectedContainerPort,
						expectedAdditionalExternalPort,
						expectedAdditionalContainerPort,
					),
				))
			})

			It("doesn't remove the endpoint from the table", func() {
				Consistently(table.RemoveEndpointCallCount).Should(Equal(0))
			})

			It("doesn't emit", func() {
				Consistently(emitter.EmitCallCount).Should(Equal(0))
			})
		})
	})

	Describe("Unrecognized events", func() {
		BeforeEach(func() {
			events <- &unrecognizedEvent{}
		})

		It("does not emit any more messages", func() {
			Consistently(emitter.EmitCallCount).Should(Equal(0))
		})
	})

	Describe("interrupting the process", func() {
		It("should be possible to SIGINT the route emitter", func() {
			process.Signal(os.Interrupt)
			Eventually(process.Wait()).Should(Receive())
		})
	})
})

type unrecognizedEvent struct{}

func (u *unrecognizedEvent) EventType() string { return "unrecognized-event" }
func (u *unrecognizedEvent) Key() string       { return "" }
