package routing_table_test

import (
	"fmt"

	"code.cloudfoundry.org/lager/lagertest"

	"github.com/cf-furnace/route-emitter/models"
	"github.com/cf-furnace/route-emitter/routing_table"
	. "github.com/cf-furnace/route-emitter/routing_table/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("RoutingTable", func() {
	var (
		table          routing_table.RoutingTable
		messagesToEmit routing_table.MessagesToEmit
		logger         *lagertest.TestLogger
	)

	key := routing_table.RoutingKey{ProcessGuid: "some-process-guid", ContainerPort: 8080}

	hostname1 := "foo.example.com"
	hostname2 := "bar.example.com"
	hostname3 := "baz.example.com"

	domain := "domain"

	olderTag := &models.ModificationTag{Epoch: "abc", Index: 0}
	currentTag := &models.ModificationTag{Epoch: "abc", Index: 1}
	newerTag := &models.ModificationTag{Epoch: "def", Index: 0}

	endpoint1 := routing_table.Endpoint{InstanceGuid: "ig-1", Host: "1.1.1.1", Index: 0, Domain: domain, Port: 11, ContainerPort: 8080, Evacuating: false, ModificationTag: currentTag}
	endpoint2 := routing_table.Endpoint{InstanceGuid: "ig-2", Host: "2.2.2.2", Index: 1, Domain: domain, Port: 22, ContainerPort: 8080, Evacuating: false, ModificationTag: currentTag}
	endpoint3 := routing_table.Endpoint{InstanceGuid: "ig-3", Host: "3.3.3.3", Index: 2, Domain: domain, Port: 33, ContainerPort: 8080, Evacuating: false, ModificationTag: currentTag}
	collisionEndpoint := routing_table.Endpoint{
		InstanceGuid:    "ig-4",
		Host:            "1.1.1.1",
		Index:           3,
		Domain:          domain,
		Port:            11,
		ContainerPort:   8080,
		Evacuating:      false,
		ModificationTag: currentTag,
	}
	newInstanceEndpointAfterEvacuation := routing_table.Endpoint{InstanceGuid: "ig-5", Host: "5.5.5.5", Index: 0, Domain: domain, Port: 55, ContainerPort: 8080, Evacuating: false, ModificationTag: currentTag}

	evacuating1 := routing_table.Endpoint{InstanceGuid: "ig-1", Host: "1.1.1.1", Index: 0, Domain: domain, Port: 11, ContainerPort: 8080, Evacuating: true, ModificationTag: currentTag}

	logGuid := "some-log-guid"

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test-route-emitter")
		table = routing_table.NewTable(logger)
	})

	Describe("Evacuating endpoints", func() {
		BeforeEach(func() {
			messagesToEmit = table.SetRoutes(key, []routing_table.Route{routing_table.Route{Hostname: hostname1, LogGuid: logGuid}}, currentTag)
			Expect(messagesToEmit).To(BeZero())

			messagesToEmit = table.AddEndpoint(key, endpoint1)
			expected := routing_table.MessagesToEmit{
				RegistrationMessages: []routing_table.RegistryMessage{
					routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
				},
			}
			Expect(messagesToEmit).To(MatchMessagesToEmit(expected))

			messagesToEmit = table.AddEndpoint(key, evacuating1)
			Expect(messagesToEmit).To(BeZero())

			messagesToEmit = table.RemoveEndpoint(key, endpoint1)
			Expect(messagesToEmit).To(BeZero())
		})

		It("does not log an address collision", func() {
			Consistently(logger).ShouldNot(Say("collision-detected-with-endpoint"))
		})

		Context("when we have an evacuating endpoint and we add an instance for that added", func() {
			It("emits a registration for the instance and a unregister for the evacuating", func() {
				messagesToEmit = table.AddEndpoint(key, newInstanceEndpointAfterEvacuation)
				expected := routing_table.MessagesToEmit{
					RegistrationMessages: []routing_table.RegistryMessage{
						routing_table.RegistryMessageFor(newInstanceEndpointAfterEvacuation, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
					},
				}
				Expect(messagesToEmit).To(MatchMessagesToEmit(expected))

				messagesToEmit = table.RemoveEndpoint(key, evacuating1)
				expected = routing_table.MessagesToEmit{
					UnregistrationMessages: []routing_table.RegistryMessage{
						routing_table.RegistryMessageFor(evacuating1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
					},
				}
				Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
			})
		})
	})

	Describe("Processing deltas", func() {
		Context("when the table is empty", func() {
			Context("When setting routes", func() {
				It("emits nothing", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
						routing_table.Route{Hostname: hostname2, LogGuid: logGuid},
					}, currentTag)
					Expect(messagesToEmit).To(BeZero())
				})
			})

			Context("when removing routes", func() {
				It("emits nothing", func() {
					messagesToEmit = table.RemoveRoutes(key, currentTag)
					Expect(messagesToEmit).To(BeZero())
				})
			})

			Context("when adding/updating endpoints", func() {
				It("emits nothing", func() {
					messagesToEmit = table.AddEndpoint(key, endpoint1)
					Expect(messagesToEmit).To(BeZero())
				})
			})

			Context("when removing endpoints", func() {
				It("emits nothing", func() {
					messagesToEmit = table.RemoveEndpoint(key, endpoint1)
					Expect(messagesToEmit).To(BeZero())
				})
			})
		})

		Context("when there are both endpoints and routes in the table", func() {
			BeforeEach(func() {
				table.SetRoutes(key, []routing_table.Route{
					routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
					routing_table.Route{Hostname: hostname2, LogGuid: logGuid},
				}, currentTag)
				table.AddEndpoint(key, endpoint1)
				table.AddEndpoint(key, endpoint2)
			})

			Describe("SetRoutes", func() {
				It("emits nothing when the route's hostnames do not change", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
						routing_table.Route{Hostname: hostname2, LogGuid: logGuid},
					}, newerTag)
					Expect(messagesToEmit).To(BeZero())
				})

				It("emits registrations when route's hostnames do not change but the route service url does", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"},
						routing_table.Route{Hostname: hostname2, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"},
					}, newerTag)

					expected := routing_table.MessagesToEmit{
						RegistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname2, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})

				It("emits nothing when a hostname is added to a route with an older tag", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
						routing_table.Route{Hostname: hostname2, LogGuid: logGuid},
						routing_table.Route{Hostname: hostname3, LogGuid: logGuid},
					}, olderTag)
					Expect(messagesToEmit).To(BeZero())
				})

				It("emits registrations when a hostname is added to a route with a newer tag", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
						routing_table.Route{Hostname: hostname2, LogGuid: logGuid},
						routing_table.Route{Hostname: hostname3, LogGuid: logGuid},
					}, newerTag)

					expected := routing_table.MessagesToEmit{
						RegistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname3, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname3, LogGuid: logGuid}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})

				It("emits nothing when a hostname is removed from a route with an older tag", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
					}, olderTag)
					Expect(messagesToEmit).To(BeZero())
				})

				It("emits registrations and unregistrations when a hostname is removed from a route with a newer tag", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
					}, newerTag)

					expected := routing_table.MessagesToEmit{
						RegistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
						},
						UnregistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})

				It("emits nothing when hostnames are added and removed from a route with an older tag", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
						routing_table.Route{Hostname: hostname3, LogGuid: logGuid},
					}, olderTag)
					Expect(messagesToEmit).To(BeZero())
				})

				It("emits registrations and unregistrations when hostnames are added and removed from a route with a newer tag", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
						routing_table.Route{Hostname: hostname3, LogGuid: logGuid},
					}, newerTag)

					expected := routing_table.MessagesToEmit{
						RegistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname3, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname3, LogGuid: logGuid}),
						},
						UnregistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})
			})

			Context("RemoveRoutes", func() {
				It("emits unregistrations with a newer tag", func() {
					messagesToEmit = table.RemoveRoutes(key, newerTag)

					expected := routing_table.MessagesToEmit{
						UnregistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})

				It("updates routing table with a newer tag", func() {
					table.RemoveRoutes(key, newerTag)
					Expect(table.RouteCount()).To(Equal(0))
				})

				It("emits unregistrations with the same tag", func() {
					messagesToEmit = table.RemoveRoutes(key, currentTag)

					expected := routing_table.MessagesToEmit{
						UnregistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})

				It("updates routing table with a same tag", func() {
					table.RemoveRoutes(key, currentTag)
					Expect(table.RouteCount()).To(Equal(0))
				})

				It("emits nothing when the tag is older", func() {
					messagesToEmit = table.RemoveRoutes(key, olderTag)
					Expect(messagesToEmit).To(BeZero())
				})

				It("does NOT update routing table with an older tag", func() {
					beforeRouteCount := table.RouteCount()
					table.RemoveRoutes(key, olderTag)
					Expect(table.RouteCount()).To(Equal(beforeRouteCount))
				})
			})

			Context("AddEndpoint", func() {
				It("emits nothing when the tag is the same", func() {
					messagesToEmit = table.AddEndpoint(key, endpoint1)
					Expect(messagesToEmit).To(BeZero())
				})

				It("emits nothing when updating an endpoint with an older tag", func() {
					updatedEndpoint := endpoint1
					updatedEndpoint.ModificationTag = olderTag

					messagesToEmit = table.AddEndpoint(key, updatedEndpoint)
					Expect(messagesToEmit).To(BeZero())
				})

				It("emits nothing when updating an endpoint with a newer tag", func() {
					updatedEndpoint := endpoint1
					updatedEndpoint.ModificationTag = newerTag

					messagesToEmit = table.AddEndpoint(key, updatedEndpoint)
					Expect(messagesToEmit).To(BeZero())
				})

				It("emits registrations when adding an endpoint", func() {
					messagesToEmit = table.AddEndpoint(key, endpoint3)

					expected := routing_table.MessagesToEmit{
						RegistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint3, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint3, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})

				It("does not log a collision", func() {
					table.AddEndpoint(key, endpoint3)
					Consistently(logger).ShouldNot(Say("collision-detected-with-endpoint"))
				})

				Context("when adding an endpoint with IP and port that collide with existing endpoint", func() {
					It("logs the collision", func() {
						table.AddEndpoint(key, collisionEndpoint)
						Eventually(logger).Should(Say(
							fmt.Sprintf(
								`\{"Address":\{"Host":"%s","Port":%d\},"instance_guid_a":"%s","instance_guid_b":"%s"\}`,
								endpoint1.Host,
								endpoint1.Port,
								endpoint1.InstanceGuid,
								collisionEndpoint.InstanceGuid,
							),
						))
					})
				})

				Context("when an evacuating endpoint is added for an instance that already exists", func() {
					It("emits nothing", func() {
						messagesToEmit = table.AddEndpoint(key, evacuating1)
						Expect(messagesToEmit).To(BeZero())
					})
				})

				Context("when an instance endpoint is updated for an evacuating that already exists", func() {
					BeforeEach(func() {
						table.AddEndpoint(key, evacuating1)
					})

					It("emits nothing", func() {
						messagesToEmit = table.AddEndpoint(key, endpoint1)
						Expect(messagesToEmit).To(BeZero())
					})
				})
			})

			Context("RemoveEndpoint", func() {
				It("emits unregistrations with the same tag", func() {
					messagesToEmit = table.RemoveEndpoint(key, endpoint2)

					expected := routing_table.MessagesToEmit{
						UnregistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})

				It("emits unregistrations when the tag is newer", func() {
					newerEndpoint := endpoint2
					newerEndpoint.ModificationTag = newerTag
					messagesToEmit = table.RemoveEndpoint(key, newerEndpoint)

					expected := routing_table.MessagesToEmit{
						UnregistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})

				It("emits nothing when the tag is older", func() {
					olderEndpoint := endpoint2
					olderEndpoint.ModificationTag = olderTag
					messagesToEmit = table.RemoveEndpoint(key, olderEndpoint)
					Expect(messagesToEmit).To(BeZero())
				})

				Context("when an instance endpoint is removed for an instance that already exists", func() {
					BeforeEach(func() {
						table.AddEndpoint(key, evacuating1)
					})

					It("emits nothing", func() {
						messagesToEmit = table.RemoveEndpoint(key, endpoint1)
						Expect(messagesToEmit).To(BeZero())
					})
				})

				Context("when an evacuating endpoint is removed instance that already exists", func() {
					BeforeEach(func() {
						table.AddEndpoint(key, evacuating1)
					})

					It("emits nothing", func() {
						messagesToEmit = table.AddEndpoint(key, endpoint1)
						Expect(messagesToEmit).To(BeZero())
					})
				})

				Context("when a collision is avoided because the endpoint has already been removed", func() {
					It("does not log the collision", func() {
						table.RemoveEndpoint(key, endpoint1)
						table.AddEndpoint(key, collisionEndpoint)
						Consistently(logger).ShouldNot(Say("collision-detected-with-endpoint"))
					})
				})
			})
		})

		Context("when there are only routes in the table", func() {
			BeforeEach(func() {
				table.SetRoutes(key, []routing_table.Route{
					routing_table.Route{Hostname: hostname1, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"},
					routing_table.Route{Hostname: hostname2, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"},
				}, nil)
			})

			Context("When setting routes", func() {
				It("emits nothing", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
						routing_table.Route{Hostname: hostname3, LogGuid: logGuid},
					}, nil)
					Expect(messagesToEmit).To(BeZero())
				})
			})

			Context("when removing routes", func() {
				It("emits nothing", func() {
					messagesToEmit = table.RemoveRoutes(key, currentTag)
					Expect(messagesToEmit).To(BeZero())
				})
			})

			Context("when adding/updating endpoints", func() {
				It("emits registrations", func() {
					messagesToEmit = table.AddEndpoint(key, endpoint1)

					expected := routing_table.MessagesToEmit{
						RegistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname2, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})
			})
		})

		Context("when there are only endpoints in the table", func() {
			BeforeEach(func() {
				table.AddEndpoint(key, endpoint1)
				table.AddEndpoint(key, endpoint2)
			})

			Context("When setting routes", func() {
				It("emits registrations", func() {
					messagesToEmit = table.SetRoutes(key, []routing_table.Route{
						routing_table.Route{Hostname: hostname1, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"},
						routing_table.Route{Hostname: hostname2, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"},
					}, currentTag)

					expected := routing_table.MessagesToEmit{
						RegistrationMessages: []routing_table.RegistryMessage{
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
							routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname2, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
							routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid, RouteServiceUrl: "https://rs.example.com"}),
						},
					}
					Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
				})
			})

			Context("when removing routes", func() {
				It("emits nothing", func() {
					messagesToEmit = table.RemoveRoutes(key, currentTag)
					Expect(messagesToEmit).To(BeZero())
				})
			})

			Context("when adding/updating endpoints", func() {
				It("emits nothing", func() {
					messagesToEmit = table.AddEndpoint(key, endpoint2)
					Expect(messagesToEmit).To(BeZero())
				})
			})

			Context("when removing endpoints", func() {
				It("emits nothing", func() {
					messagesToEmit = table.RemoveEndpoint(key, endpoint1)
					Expect(messagesToEmit).To(BeZero())
				})
			})
		})
	})

	Describe("MessagesToEmit", func() {
		Context("when the table is empty", func() {
			It("should be empty", func() {
				messagesToEmit = table.MessagesToEmit()
				Expect(messagesToEmit).To(BeZero())
			})
		})

		Context("when the table has routes but no endpoints", func() {
			BeforeEach(func() {
				table.SetRoutes(key, []routing_table.Route{
					routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
					routing_table.Route{Hostname: hostname2, LogGuid: logGuid},
				}, nil)
			})

			It("should be empty", func() {
				messagesToEmit = table.MessagesToEmit()
				Expect(messagesToEmit).To(BeZero())
			})
		})

		Context("when the table has endpoints but no routes", func() {
			BeforeEach(func() {
				table.AddEndpoint(key, endpoint1)
				table.AddEndpoint(key, endpoint2)
			})

			It("should be empty", func() {
				messagesToEmit = table.MessagesToEmit()
				Expect(messagesToEmit).To(BeZero())
			})
		})

		Context("when the table has routes and endpoints", func() {
			BeforeEach(func() {
				table.SetRoutes(key, []routing_table.Route{
					routing_table.Route{Hostname: hostname1, LogGuid: logGuid},
					routing_table.Route{Hostname: hostname2, LogGuid: logGuid},
				}, nil)
				table.AddEndpoint(key, endpoint1)
				table.AddEndpoint(key, endpoint2)
			})

			It("emits the registrations", func() {
				messagesToEmit = table.MessagesToEmit()

				expected := routing_table.MessagesToEmit{
					RegistrationMessages: []routing_table.RegistryMessage{
						routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
						routing_table.RegistryMessageFor(endpoint1, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
						routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname1, LogGuid: logGuid}),
						routing_table.RegistryMessageFor(endpoint2, routing_table.Route{Hostname: hostname2, LogGuid: logGuid}),
					},
				}
				Expect(messagesToEmit).To(MatchMessagesToEmit(expected))
			})
		})
	})

	Describe("EndpointsForIndex", func() {
		It("returns endpoints for evacuation and non-evacuating instances", func() {
			table.SetRoutes(routing_table.RoutingKey{ProcessGuid: "fake-process-guid"}, []routing_table.Route{
				routing_table.Route{Hostname: "fake-route-url", LogGuid: logGuid},
			}, nil)
			table.AddEndpoint(key, endpoint1)
			table.AddEndpoint(key, endpoint2)
			table.AddEndpoint(key, evacuating1)

			Expect(table.EndpointsForIndex(key, 0)).To(ConsistOf([]routing_table.Endpoint{endpoint1, evacuating1}))
		})
	})

	Describe("RouteCount", func() {
		It("returns 0 on a new routing table", func() {
			Expect(table.RouteCount()).To(Equal(0))
		})

		It("returns 1 after adding a route to a single process", func() {
			table.SetRoutes(routing_table.RoutingKey{ProcessGuid: "fake-process-guid"}, []routing_table.Route{
				routing_table.Route{Hostname: "fake-route-url", LogGuid: logGuid},
			}, nil)

			Expect(table.RouteCount()).To(Equal(1))
		})

		It("returns 2 after associating 2 urls with a single process", func() {
			table.SetRoutes(routing_table.RoutingKey{ProcessGuid: "fake-process-guid"}, []routing_table.Route{
				routing_table.Route{Hostname: "fake-route-url-1", LogGuid: logGuid},
				routing_table.Route{Hostname: "fake-route-url-2", LogGuid: logGuid},
			}, nil)

			Expect(table.RouteCount()).To(Equal(2))
		})

		It("returns 4 after associating 2 urls with two processes", func() {
			table.SetRoutes(routing_table.RoutingKey{ProcessGuid: "fake-process-guid-a"}, []routing_table.Route{
				routing_table.Route{Hostname: "fake-route-url-a-1", LogGuid: logGuid},
				routing_table.Route{Hostname: "fake-route-url-a-2", LogGuid: logGuid},
			}, nil)
			table.SetRoutes(routing_table.RoutingKey{ProcessGuid: "fake-process-guid-b"}, []routing_table.Route{
				routing_table.Route{Hostname: "fake-route-url-b-1", LogGuid: logGuid},
				routing_table.Route{Hostname: "fake-route-url-b-2", LogGuid: logGuid},
			}, nil)

			Expect(table.RouteCount()).To(Equal(4))
		})
	})
})
