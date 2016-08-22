package routing_table_test

import (
	"github.com/cf-furnace/route-emitter/cfroutes"
	"github.com/cf-furnace/route-emitter/models"
	"github.com/cf-furnace/route-emitter/routing_table"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ByRoutingKey", func() {
	Describe("RoutesByRoutingKeyFromSchedulingInfos", func() {
		It("should build a map of routes", func() {
			abcRoutes := cfroutes.CFRoutes{
				{Hostnames: []string{"foo.com", "bar.com"}, Port: 8080, RouteServiceUrl: "https://something.creative"},
				{Hostnames: []string{"foo.example.com"}, Port: 9090},
			}
			defRoutes := cfroutes.CFRoutes{
				{Hostnames: []string{"baz.com"}, Port: 8080},
			}

			routes := routing_table.RoutesByRoutingKeyFromDesireds([]*models.DesiredLRP{
				NewDesiredLRP("abc", "tests", "abc-guid", abcRoutes),
				NewDesiredLRP("def", "tests", "def-guid", defRoutes),
			})

			Expect(routes).To(HaveLen(3))
			Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][0].Hostname).To(Equal("foo.com"))
			Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][0].LogGuid).To(Equal("abc-guid"))
			Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][0].RouteServiceUrl).To(Equal("https://something.creative"))

			Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][1].Hostname).To(Equal("bar.com"))
			Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][1].LogGuid).To(Equal("abc-guid"))
			Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][1].RouteServiceUrl).To(Equal("https://something.creative"))

			Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 9090}][0].Hostname).To(Equal("foo.example.com"))
			Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 9090}][0].LogGuid).To(Equal("abc-guid"))

			Expect(routes[routing_table.RoutingKey{ProcessGuid: "def", ContainerPort: 8080}][0].Hostname).To(Equal("baz.com"))
			Expect(routes[routing_table.RoutingKey{ProcessGuid: "def", ContainerPort: 8080}][0].LogGuid).To(Equal("def-guid"))
		})

		Context("when multiple hosts have the same key, but one hostname is bound to a route service and the other is not", func() {
			It("should build a map of routes", func() {
				abcRoutes := cfroutes.CFRoutes{
					{Hostnames: []string{"foo.com"}, Port: 8080, RouteServiceUrl: "https://something.creative"},
					{Hostnames: []string{"bar.com"}, Port: 8080},
				}
				defRoutes := cfroutes.CFRoutes{
					{Hostnames: []string{"baz.com"}, Port: 8080},
				}

				routes := routing_table.RoutesByRoutingKeyFromDesireds([]*models.DesiredLRP{
					NewDesiredLRP("abc", "tests", "abc-guid", abcRoutes),
					NewDesiredLRP("def", "tests", "def-guid", defRoutes),
				})

				Expect(routes).To(HaveLen(2))
				Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][0].Hostname).To(Equal("foo.com"))
				Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][0].LogGuid).To(Equal("abc-guid"))
				Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][0].RouteServiceUrl).To(Equal("https://something.creative"))

				Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][1].Hostname).To(Equal("bar.com"))
				Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][1].LogGuid).To(Equal("abc-guid"))
				Expect(routes[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 8080}][1].RouteServiceUrl).To(Equal(""))

				Expect(routes[routing_table.RoutingKey{ProcessGuid: "def", ContainerPort: 8080}][0].Hostname).To(Equal("baz.com"))
				Expect(routes[routing_table.RoutingKey{ProcessGuid: "def", ContainerPort: 8080}][0].LogGuid).To(Equal("def-guid"))
			})
		})
		Context("when the routing info is nil", func() {
			It("should not be included in the results", func() {
				routes := routing_table.RoutesByRoutingKeyFromDesireds([]*models.DesiredLRP{
					NewDesiredLRP("abc", "tests", "abc-guid", nil),
				})
				Expect(routes).To(HaveLen(0))
			})
		})
	})

	Describe("EndpointsByRoutingKeyFromActuals", func() {
		Context("when some actuals don't have port mappings", func() {
			var endpoints routing_table.EndpointsByRoutingKey

			BeforeEach(func() {
				schedInfo1 := NewValidDesiredLRP("abc")
				schedInfo1.Instances = 2
				schedInfo2 := NewValidDesiredLRP("def")
				schedInfo2.Instances = 2

				endpoints = routing_table.EndpointsByRoutingKeyFromActuals([]*routing_table.ActualLRPRoutingInfo{
					{
						ActualLRP: &models.ActualLRP{
							ProcessGuid: schedInfo1.ProcessGuid,
							Index:       0,
							Address:     "1.1.1.1",
							Ports:       []models.PortMapping{models.PortMapping{11, 44}, models.PortMapping{66, 99}},
						},
					},
					{
						ActualLRP: &models.ActualLRP{
							ProcessGuid: schedInfo1.ProcessGuid,
							Index:       1,
							Address:     "2.2.2.2",
							Ports:       []models.PortMapping{models.PortMapping{22, 44}, models.PortMapping{88, 99}},
						},
					},
					{
						ActualLRP: &models.ActualLRP{
							ProcessGuid: schedInfo2.ProcessGuid,
							Index:       0,
							Address:     "3.3.3.3",
							Ports:       []models.PortMapping{models.PortMapping{33, 55}},
						},
					},
					{
						ActualLRP: &models.ActualLRP{
							ProcessGuid: schedInfo2.ProcessGuid,
							Index:       1,
							Address:     "4.4.4.4",
							Ports:       nil,
						},
					},
				}, map[string]*models.DesiredLRP{
					schedInfo1.ProcessGuid: schedInfo1,
					schedInfo2.ProcessGuid: schedInfo2,
				},
				)
			})

			It("should build a map of endpoints, ignoring those without ports", func() {
				Expect(endpoints).To(HaveLen(3))

				Expect(endpoints[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 44}]).To(ConsistOf([]routing_table.Endpoint{
					routing_table.Endpoint{Host: "1.1.1.1", Index: 0, Domain: "", Port: 11, ContainerPort: 44},
					routing_table.Endpoint{Host: "2.2.2.2", Index: 1, Domain: "", Port: 22, ContainerPort: 44}}))

				Expect(endpoints[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 99}]).To(ConsistOf([]routing_table.Endpoint{
					routing_table.Endpoint{Host: "1.1.1.1", Index: 0, Domain: "", Port: 66, ContainerPort: 99},
					routing_table.Endpoint{Host: "2.2.2.2", Index: 1, Domain: "", Port: 88, ContainerPort: 99}}))

				Expect(endpoints[routing_table.RoutingKey{ProcessGuid: "def", ContainerPort: 55}]).To(ConsistOf([]routing_table.Endpoint{
					routing_table.Endpoint{Host: "3.3.3.3", Index: 0, Domain: "", Port: 33, ContainerPort: 55}}))
			})
		})

		Context("when not all running actuals are desired", func() {
			var endpoints routing_table.EndpointsByRoutingKey

			BeforeEach(func() {
				schedInfo1 := NewValidDesiredLRP("abc")
				schedInfo1.Instances = 1
				schedInfo2 := NewValidDesiredLRP("def")
				schedInfo2.Instances = 1

				endpoints = routing_table.EndpointsByRoutingKeyFromActuals([]*routing_table.ActualLRPRoutingInfo{
					{
						ActualLRP: &models.ActualLRP{
							ProcessGuid: "abc",
							Index:       0,
							Address:     "1.1.1.1",
							Ports:       []models.PortMapping{{11, 44}, {66, 99}},
						},
					},
					{
						ActualLRP: &models.ActualLRP{
							ProcessGuid: "abc",
							Index:       1,
							Address:     "1.1.1.1",
							Ports:       []models.PortMapping{{22, 55}, {88, 99}},
						},
					},
					{
						ActualLRP: &models.ActualLRP{
							ProcessGuid: "def",
							Index:       0,
							Address:     "3.3.3.3",
							Ports:       []models.PortMapping{{33, 55}},
						},
					},
				}, map[string]*models.DesiredLRP{
					"abc": schedInfo1,
					"def": schedInfo2,
				},
				)
			})

			It("should build a map of endpoints, regardless whether actuals are desired", func() {
				Expect(endpoints).To(HaveLen(4))

				Expect(endpoints[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 44}]).To(ConsistOf([]routing_table.Endpoint{
					routing_table.Endpoint{Host: "1.1.1.1", Domain: "", Port: 11, ContainerPort: 44}}))
				Expect(endpoints[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 99}]).To(ConsistOf([]routing_table.Endpoint{
					routing_table.Endpoint{Host: "1.1.1.1", Domain: "", Port: 66, ContainerPort: 99},
					routing_table.Endpoint{Host: "1.1.1.1", Index: 1, Domain: "", Port: 88, ContainerPort: 99}}))
				Expect(endpoints[routing_table.RoutingKey{ProcessGuid: "abc", ContainerPort: 55}]).To(ConsistOf([]routing_table.Endpoint{
					routing_table.Endpoint{Host: "1.1.1.1", Index: 1, Domain: "", Port: 22, ContainerPort: 55}}))
				Expect(endpoints[routing_table.RoutingKey{ProcessGuid: "def", ContainerPort: 55}]).To(ConsistOf([]routing_table.Endpoint{
					routing_table.Endpoint{Host: "3.3.3.3", Domain: "", Port: 33, ContainerPort: 55}}))
			})
		})

	})

	Describe("EndpointsFromActual", func() {
		It("builds a map of container port to endpoint", func() {
			endpoints, err := routing_table.EndpointsFromActual(&routing_table.ActualLRPRoutingInfo{
				ActualLRP: &models.ActualLRP{
					ProcessGuid:  "process-guid",
					Index:        0,
					InstanceGuid: "instance-guid",
					Address:      "1.1.1.1",
					Ports:        []models.PortMapping{{11, 44}, {66, 99}},
				},
				Evacuating: true,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(endpoints).To(ConsistOf([]routing_table.Endpoint{
				routing_table.Endpoint{Host: "1.1.1.1", Domain: "", Port: 11, InstanceGuid: "instance-guid", ContainerPort: 44, Evacuating: true, Index: 0},
				routing_table.Endpoint{Host: "1.1.1.1", Domain: "", Port: 66, InstanceGuid: "instance-guid", ContainerPort: 99, Evacuating: true, Index: 0},
			}))
		})
	})

	Describe("RoutingKeysFromActual", func() {
		It("creates a list of keys for an actual LRP", func() {
			keys := routing_table.RoutingKeysFromActual(&models.ActualLRP{
				ProcessGuid:  "process-guid",
				Index:        0,
				InstanceGuid: "instance-guid",
				Address:      "1.1.1.1",
				Ports:        []models.PortMapping{{11, 44}, {66, 99}},
			})
			Expect(keys).To(HaveLen(2))
			Expect(keys).To(ContainElement(routing_table.RoutingKey{ProcessGuid: "process-guid", ContainerPort: 44}))
			Expect(keys).To(ContainElement(routing_table.RoutingKey{ProcessGuid: "process-guid", ContainerPort: 99}))
		})

		Context("when the actual lrp has no port mappings", func() {
			It("returns no keys", func() {
				keys := routing_table.RoutingKeysFromActual(&models.ActualLRP{
					ProcessGuid:  "process-guid",
					Index:        0,
					InstanceGuid: "instance-guid",
					Address:      "1.1.1.1",
					Ports:        nil,
				})

				Expect(keys).To(HaveLen(0))
			})
		})
	})

	Describe("RoutingKeysFromDesired", func() {
		It("creates a list of keys for an actual LRP", func() {
			routes := cfroutes.CFRoutes{
				{Hostnames: []string{"foo.com", "bar.com"}, Port: 8080},
				{Hostnames: []string{"foo.example.com"}, Port: 9090},
			}

			schedulingInfo := &models.DesiredLRP{
				ProcessGuid: "process-guid",
				LogGuid:     "log-guid",
				Routes:      routes,
			}

			keys := routing_table.RoutingKeysFromDesired(schedulingInfo)

			Expect(keys).To(HaveLen(2))
			Expect(keys).To(ContainElement(routing_table.RoutingKey{ProcessGuid: "process-guid", ContainerPort: 8080}))
			Expect(keys).To(ContainElement(routing_table.RoutingKey{ProcessGuid: "process-guid", ContainerPort: 9090}))
		})

		Context("when the desired LRP does not define any container ports", func() {
			It("still uses the routes property", func() {
				schedulingInfo := &models.DesiredLRP{
					ProcessGuid: "process-guid",
					LogGuid:     "log-guid",
					Routes:      cfroutes.CFRoutes{{Hostnames: []string{"foo.com", "bar.com"}, Port: 8080}},
				}

				keys := routing_table.RoutingKeysFromDesired(schedulingInfo)
				Expect(keys).To(HaveLen(1))
				Expect(keys).To(ContainElement(routing_table.RoutingKey{ProcessGuid: "process-guid", ContainerPort: 8080}))
			})
		})
	})
})
