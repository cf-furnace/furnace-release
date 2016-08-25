package cfroutes

import "encoding/json"

const CF_ROUTER = "cf-router"

type Routes map[string]*json.RawMessage

type CFRoutes []CFRoute

type CFRoute struct {
	Hostnames       []string `json:"hostnames"`
	Port            uint32   `json:"port"`
	RouteServiceUrl string   `json:"route_service_url,omitempty"`
}

func (c CFRoutes) RoutingInfo() Routes {
	data, _ := json.Marshal(c)
	routingInfo := json.RawMessage(data)
	return Routes{
		CF_ROUTER: &routingInfo,
	}
}

func CFRoutesFromRoutingInfo(routingInfo Routes) (CFRoutes, error) {
	if routingInfo == nil {
		return nil, nil
	}

	routes := routingInfo
	data, found := routes[CF_ROUTER]
	if !found {
		return nil, nil
	}

	if data == nil {
		return nil, nil
	}

	cfRoutes := CFRoutes{}
	err := json.Unmarshal(*data, &cfRoutes)

	return cfRoutes, err
}
