package routing_table

import (
	"errors"

	"github.com/cf-furnace/route-emitter/models"
)

type RoutesByRoutingKey map[RoutingKey][]Route
type EndpointsByRoutingKey map[RoutingKey][]Endpoint

func RoutesByRoutingKeyFromDesireds(schedulingInfos []*models.DesiredLRP) RoutesByRoutingKey {
	routesByRoutingKey := RoutesByRoutingKey{}
	for _, desired := range schedulingInfos {
		for _, cfRoute := range desired.Routes {
			key := RoutingKey{ProcessGuid: desired.ProcessGuid, ContainerPort: cfRoute.Port}
			var routeEntries []Route
			for _, hostname := range cfRoute.Hostnames {
				routeEntries = append(routeEntries, Route{
					Hostname:        hostname,
					LogGuid:         desired.LogGuid,
					RouteServiceUrl: cfRoute.RouteServiceUrl,
				})
			}
			routesByRoutingKey[key] = append(routesByRoutingKey[key], routeEntries...)
		}
	}

	return routesByRoutingKey
}

func EndpointsByRoutingKeyFromActuals(actuals []*ActualLRPRoutingInfo, schedInfos map[string]*models.DesiredLRP) EndpointsByRoutingKey {
	endpointsByRoutingKey := EndpointsByRoutingKey{}
	for _, actual := range actuals {
		endpoints, err := EndpointsFromActual(actual)
		if err != nil {
			continue
		}

		for containerPort, endpoint := range endpoints {
			key := RoutingKey{ProcessGuid: actual.ActualLRP.ProcessGuid, ContainerPort: containerPort}
			endpointsByRoutingKey[key] = append(endpointsByRoutingKey[key], endpoint)
		}
	}

	return endpointsByRoutingKey
}

func EndpointsFromActual(actualLRPInfo *ActualLRPRoutingInfo) (map[uint32]Endpoint, error) {
	endpoints := map[uint32]Endpoint{}
	actual := actualLRPInfo.ActualLRP

	if len(actual.Ports) == 0 {
		return endpoints, errors.New("missing ports")
	}

	for _, portMapping := range actual.Ports {
		endpoint := Endpoint{
			InstanceGuid: actual.InstanceGuid,
			Index:        actual.Index,
			Host:         actual.Address,
			// Domain:        actual.Domain,
			Port:          portMapping.HostPort,
			ContainerPort: portMapping.ContainerPort,
			Evacuating:    actualLRPInfo.Evacuating,
		}
		endpoints[portMapping.ContainerPort] = endpoint
	}

	return endpoints, nil
}

func RoutingKeysFromActual(actual *models.ActualLRP) []RoutingKey {
	keys := []RoutingKey{}
	for _, portMapping := range actual.Ports {
		keys = append(keys, RoutingKey{ProcessGuid: actual.ProcessGuid, ContainerPort: uint32(portMapping.ContainerPort)})
	}

	return keys
}

func RoutingKeysFromDesired(desiredLRP *models.DesiredLRP) []RoutingKey {
	keys := []RoutingKey{}

	for _, cfRoute := range desiredLRP.Routes {
		keys = append(keys, RoutingKey{ProcessGuid: desiredLRP.ProcessGuid, ContainerPort: cfRoute.Port})
	}
	return keys
}
