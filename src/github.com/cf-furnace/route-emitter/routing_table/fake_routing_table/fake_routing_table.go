// This file was generated by counterfeiter
package fake_routing_table

import (
	"sync"

	"github.com/cf-furnace/route-emitter/models"
	"github.com/cf-furnace/route-emitter/routing_table"
)

type FakeRoutingTable struct {
	RouteCountStub        func() int
	routeCountMutex       sync.RWMutex
	routeCountArgsForCall []struct{}
	routeCountReturns     struct {
		result1 int
	}
	SetRoutesStub        func(key routing_table.RoutingKey, routes []routing_table.Route, modTag *models.ModificationTag) routing_table.MessagesToEmit
	setRoutesMutex       sync.RWMutex
	setRoutesArgsForCall []struct {
		key    routing_table.RoutingKey
		routes []routing_table.Route
		modTag *models.ModificationTag
	}
	setRoutesReturns struct {
		result1 routing_table.MessagesToEmit
	}
	RemoveRoutesStub        func(key routing_table.RoutingKey, modTag *models.ModificationTag) routing_table.MessagesToEmit
	removeRoutesMutex       sync.RWMutex
	removeRoutesArgsForCall []struct {
		key    routing_table.RoutingKey
		modTag *models.ModificationTag
	}
	removeRoutesReturns struct {
		result1 routing_table.MessagesToEmit
	}
	AddEndpointStub        func(key routing_table.RoutingKey, endpoint routing_table.Endpoint) routing_table.MessagesToEmit
	addEndpointMutex       sync.RWMutex
	addEndpointArgsForCall []struct {
		key      routing_table.RoutingKey
		endpoint routing_table.Endpoint
	}
	addEndpointReturns struct {
		result1 routing_table.MessagesToEmit
	}
	RemoveEndpointStub        func(key routing_table.RoutingKey, endpoint routing_table.Endpoint) routing_table.MessagesToEmit
	removeEndpointMutex       sync.RWMutex
	removeEndpointArgsForCall []struct {
		key      routing_table.RoutingKey
		endpoint routing_table.Endpoint
	}
	removeEndpointReturns struct {
		result1 routing_table.MessagesToEmit
	}
	EndpointsForIndexStub        func(key routing_table.RoutingKey, index int32) []routing_table.Endpoint
	endpointsForIndexMutex       sync.RWMutex
	endpointsForIndexArgsForCall []struct {
		key   routing_table.RoutingKey
		index int32
	}
	endpointsForIndexReturns struct {
		result1 []routing_table.Endpoint
	}
	MessagesToEmitStub        func() routing_table.MessagesToEmit
	messagesToEmitMutex       sync.RWMutex
	messagesToEmitArgsForCall []struct{}
	messagesToEmitReturns     struct {
		result1 routing_table.MessagesToEmit
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRoutingTable) RouteCount() int {
	fake.routeCountMutex.Lock()
	fake.routeCountArgsForCall = append(fake.routeCountArgsForCall, struct{}{})
	fake.recordInvocation("RouteCount", []interface{}{})
	fake.routeCountMutex.Unlock()
	if fake.RouteCountStub != nil {
		return fake.RouteCountStub()
	} else {
		return fake.routeCountReturns.result1
	}
}

func (fake *FakeRoutingTable) RouteCountCallCount() int {
	fake.routeCountMutex.RLock()
	defer fake.routeCountMutex.RUnlock()
	return len(fake.routeCountArgsForCall)
}

func (fake *FakeRoutingTable) RouteCountReturns(result1 int) {
	fake.RouteCountStub = nil
	fake.routeCountReturns = struct {
		result1 int
	}{result1}
}

func (fake *FakeRoutingTable) SetRoutes(key routing_table.RoutingKey, routes []routing_table.Route, modTag *models.ModificationTag) routing_table.MessagesToEmit {
	var routesCopy []routing_table.Route
	if routes != nil {
		routesCopy = make([]routing_table.Route, len(routes))
		copy(routesCopy, routes)
	}
	fake.setRoutesMutex.Lock()
	fake.setRoutesArgsForCall = append(fake.setRoutesArgsForCall, struct {
		key    routing_table.RoutingKey
		routes []routing_table.Route
		modTag *models.ModificationTag
	}{key, routesCopy, modTag})
	fake.recordInvocation("SetRoutes", []interface{}{key, routesCopy, modTag})
	fake.setRoutesMutex.Unlock()
	if fake.SetRoutesStub != nil {
		return fake.SetRoutesStub(key, routes, modTag)
	} else {
		return fake.setRoutesReturns.result1
	}
}

func (fake *FakeRoutingTable) SetRoutesCallCount() int {
	fake.setRoutesMutex.RLock()
	defer fake.setRoutesMutex.RUnlock()
	return len(fake.setRoutesArgsForCall)
}

func (fake *FakeRoutingTable) SetRoutesArgsForCall(i int) (routing_table.RoutingKey, []routing_table.Route, *models.ModificationTag) {
	fake.setRoutesMutex.RLock()
	defer fake.setRoutesMutex.RUnlock()
	return fake.setRoutesArgsForCall[i].key, fake.setRoutesArgsForCall[i].routes, fake.setRoutesArgsForCall[i].modTag
}

func (fake *FakeRoutingTable) SetRoutesReturns(result1 routing_table.MessagesToEmit) {
	fake.SetRoutesStub = nil
	fake.setRoutesReturns = struct {
		result1 routing_table.MessagesToEmit
	}{result1}
}

func (fake *FakeRoutingTable) RemoveRoutes(key routing_table.RoutingKey, modTag *models.ModificationTag) routing_table.MessagesToEmit {
	fake.removeRoutesMutex.Lock()
	fake.removeRoutesArgsForCall = append(fake.removeRoutesArgsForCall, struct {
		key    routing_table.RoutingKey
		modTag *models.ModificationTag
	}{key, modTag})
	fake.recordInvocation("RemoveRoutes", []interface{}{key, modTag})
	fake.removeRoutesMutex.Unlock()
	if fake.RemoveRoutesStub != nil {
		return fake.RemoveRoutesStub(key, modTag)
	} else {
		return fake.removeRoutesReturns.result1
	}
}

func (fake *FakeRoutingTable) RemoveRoutesCallCount() int {
	fake.removeRoutesMutex.RLock()
	defer fake.removeRoutesMutex.RUnlock()
	return len(fake.removeRoutesArgsForCall)
}

func (fake *FakeRoutingTable) RemoveRoutesArgsForCall(i int) (routing_table.RoutingKey, *models.ModificationTag) {
	fake.removeRoutesMutex.RLock()
	defer fake.removeRoutesMutex.RUnlock()
	return fake.removeRoutesArgsForCall[i].key, fake.removeRoutesArgsForCall[i].modTag
}

func (fake *FakeRoutingTable) RemoveRoutesReturns(result1 routing_table.MessagesToEmit) {
	fake.RemoveRoutesStub = nil
	fake.removeRoutesReturns = struct {
		result1 routing_table.MessagesToEmit
	}{result1}
}

func (fake *FakeRoutingTable) AddEndpoint(key routing_table.RoutingKey, endpoint routing_table.Endpoint) routing_table.MessagesToEmit {
	fake.addEndpointMutex.Lock()
	fake.addEndpointArgsForCall = append(fake.addEndpointArgsForCall, struct {
		key      routing_table.RoutingKey
		endpoint routing_table.Endpoint
	}{key, endpoint})
	fake.recordInvocation("AddEndpoint", []interface{}{key, endpoint})
	fake.addEndpointMutex.Unlock()
	if fake.AddEndpointStub != nil {
		return fake.AddEndpointStub(key, endpoint)
	} else {
		return fake.addEndpointReturns.result1
	}
}

func (fake *FakeRoutingTable) AddEndpointCallCount() int {
	fake.addEndpointMutex.RLock()
	defer fake.addEndpointMutex.RUnlock()
	return len(fake.addEndpointArgsForCall)
}

func (fake *FakeRoutingTable) AddEndpointArgsForCall(i int) (routing_table.RoutingKey, routing_table.Endpoint) {
	fake.addEndpointMutex.RLock()
	defer fake.addEndpointMutex.RUnlock()
	return fake.addEndpointArgsForCall[i].key, fake.addEndpointArgsForCall[i].endpoint
}

func (fake *FakeRoutingTable) AddEndpointReturns(result1 routing_table.MessagesToEmit) {
	fake.AddEndpointStub = nil
	fake.addEndpointReturns = struct {
		result1 routing_table.MessagesToEmit
	}{result1}
}

func (fake *FakeRoutingTable) RemoveEndpoint(key routing_table.RoutingKey, endpoint routing_table.Endpoint) routing_table.MessagesToEmit {
	fake.removeEndpointMutex.Lock()
	fake.removeEndpointArgsForCall = append(fake.removeEndpointArgsForCall, struct {
		key      routing_table.RoutingKey
		endpoint routing_table.Endpoint
	}{key, endpoint})
	fake.recordInvocation("RemoveEndpoint", []interface{}{key, endpoint})
	fake.removeEndpointMutex.Unlock()
	if fake.RemoveEndpointStub != nil {
		return fake.RemoveEndpointStub(key, endpoint)
	} else {
		return fake.removeEndpointReturns.result1
	}
}

func (fake *FakeRoutingTable) RemoveEndpointCallCount() int {
	fake.removeEndpointMutex.RLock()
	defer fake.removeEndpointMutex.RUnlock()
	return len(fake.removeEndpointArgsForCall)
}

func (fake *FakeRoutingTable) RemoveEndpointArgsForCall(i int) (routing_table.RoutingKey, routing_table.Endpoint) {
	fake.removeEndpointMutex.RLock()
	defer fake.removeEndpointMutex.RUnlock()
	return fake.removeEndpointArgsForCall[i].key, fake.removeEndpointArgsForCall[i].endpoint
}

func (fake *FakeRoutingTable) RemoveEndpointReturns(result1 routing_table.MessagesToEmit) {
	fake.RemoveEndpointStub = nil
	fake.removeEndpointReturns = struct {
		result1 routing_table.MessagesToEmit
	}{result1}
}

func (fake *FakeRoutingTable) EndpointsForIndex(key routing_table.RoutingKey, index int32) []routing_table.Endpoint {
	fake.endpointsForIndexMutex.Lock()
	fake.endpointsForIndexArgsForCall = append(fake.endpointsForIndexArgsForCall, struct {
		key   routing_table.RoutingKey
		index int32
	}{key, index})
	fake.recordInvocation("EndpointsForIndex", []interface{}{key, index})
	fake.endpointsForIndexMutex.Unlock()
	if fake.EndpointsForIndexStub != nil {
		return fake.EndpointsForIndexStub(key, index)
	} else {
		return fake.endpointsForIndexReturns.result1
	}
}

func (fake *FakeRoutingTable) EndpointsForIndexCallCount() int {
	fake.endpointsForIndexMutex.RLock()
	defer fake.endpointsForIndexMutex.RUnlock()
	return len(fake.endpointsForIndexArgsForCall)
}

func (fake *FakeRoutingTable) EndpointsForIndexArgsForCall(i int) (routing_table.RoutingKey, int32) {
	fake.endpointsForIndexMutex.RLock()
	defer fake.endpointsForIndexMutex.RUnlock()
	return fake.endpointsForIndexArgsForCall[i].key, fake.endpointsForIndexArgsForCall[i].index
}

func (fake *FakeRoutingTable) EndpointsForIndexReturns(result1 []routing_table.Endpoint) {
	fake.EndpointsForIndexStub = nil
	fake.endpointsForIndexReturns = struct {
		result1 []routing_table.Endpoint
	}{result1}
}

func (fake *FakeRoutingTable) MessagesToEmit() routing_table.MessagesToEmit {
	fake.messagesToEmitMutex.Lock()
	fake.messagesToEmitArgsForCall = append(fake.messagesToEmitArgsForCall, struct{}{})
	fake.recordInvocation("MessagesToEmit", []interface{}{})
	fake.messagesToEmitMutex.Unlock()
	if fake.MessagesToEmitStub != nil {
		return fake.MessagesToEmitStub()
	} else {
		return fake.messagesToEmitReturns.result1
	}
}

func (fake *FakeRoutingTable) MessagesToEmitCallCount() int {
	fake.messagesToEmitMutex.RLock()
	defer fake.messagesToEmitMutex.RUnlock()
	return len(fake.messagesToEmitArgsForCall)
}

func (fake *FakeRoutingTable) MessagesToEmitReturns(result1 routing_table.MessagesToEmit) {
	fake.MessagesToEmitStub = nil
	fake.messagesToEmitReturns = struct {
		result1 routing_table.MessagesToEmit
	}{result1}
}

func (fake *FakeRoutingTable) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.routeCountMutex.RLock()
	defer fake.routeCountMutex.RUnlock()
	fake.setRoutesMutex.RLock()
	defer fake.setRoutesMutex.RUnlock()
	fake.removeRoutesMutex.RLock()
	defer fake.removeRoutesMutex.RUnlock()
	fake.addEndpointMutex.RLock()
	defer fake.addEndpointMutex.RUnlock()
	fake.removeEndpointMutex.RLock()
	defer fake.removeEndpointMutex.RUnlock()
	fake.endpointsForIndexMutex.RLock()
	defer fake.endpointsForIndexMutex.RUnlock()
	fake.messagesToEmitMutex.RLock()
	defer fake.messagesToEmitMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeRoutingTable) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ routing_table.RoutingTable = new(FakeRoutingTable)
