// This file was generated by counterfeiter
package fake_nats_emitter

import (
	"sync"

	"github.com/cf-furnace/route-emitter/nats_emitter"
	"github.com/cf-furnace/route-emitter/routing_table"
)

type FakeNATSEmitter struct {
	EmitStub        func(messagesToEmit routing_table.MessagesToEmit) error
	emitMutex       sync.RWMutex
	emitArgsForCall []struct {
		messagesToEmit routing_table.MessagesToEmit
	}
	emitReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeNATSEmitter) Emit(messagesToEmit routing_table.MessagesToEmit) error {
	fake.emitMutex.Lock()
	fake.emitArgsForCall = append(fake.emitArgsForCall, struct {
		messagesToEmit routing_table.MessagesToEmit
	}{messagesToEmit})
	fake.recordInvocation("Emit", []interface{}{messagesToEmit})
	fake.emitMutex.Unlock()
	if fake.EmitStub != nil {
		return fake.EmitStub(messagesToEmit)
	} else {
		return fake.emitReturns.result1
	}
}

func (fake *FakeNATSEmitter) EmitCallCount() int {
	fake.emitMutex.RLock()
	defer fake.emitMutex.RUnlock()
	return len(fake.emitArgsForCall)
}

func (fake *FakeNATSEmitter) EmitArgsForCall(i int) routing_table.MessagesToEmit {
	fake.emitMutex.RLock()
	defer fake.emitMutex.RUnlock()
	return fake.emitArgsForCall[i].messagesToEmit
}

func (fake *FakeNATSEmitter) EmitReturns(result1 error) {
	fake.EmitStub = nil
	fake.emitReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeNATSEmitter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.emitMutex.RLock()
	defer fake.emitMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeNATSEmitter) recordInvocation(key string, args []interface{}) {
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

var _ nats_emitter.NATSEmitter = new(FakeNATSEmitter)