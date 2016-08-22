package routing_table_test

import (
	"github.com/cf-furnace/route-emitter/cfroutes"
	"github.com/cf-furnace/route-emitter/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRoutingTable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RoutingTable Suite")
}

func NewValidDesiredLRP(guid string) *models.DesiredLRP {
	modTag := models.ModificationTag{}
	return &models.DesiredLRP{
		ProcessGuid: guid,
		LogGuid:     "some-log-guid",
		Routes:      cfroutes.CFRoutes{{Hostnames: []string{"dora"}, Port: 8080}},

		ModificationTag: &modTag,
	}
}

func NewDesiredLRP(pguid, domain, logGuid string, routes cfroutes.CFRoutes) *models.DesiredLRP {
	modTag := models.ModificationTag{}

	return &models.DesiredLRP{
		ProcessGuid: pguid,
		LogGuid:     logGuid,
		Routes:      routes,

		ModificationTag: &modTag,
	}
}
