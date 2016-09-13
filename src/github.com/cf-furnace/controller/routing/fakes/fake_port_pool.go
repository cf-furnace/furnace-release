// This file was generated by counterfeiter
package fakes

import "sync"

type FakePortPool struct {
	AcquireStub        func() (uint32, error)
	acquireMutex       sync.RWMutex
	acquireArgsForCall []struct{}
	acquireReturns     struct {
		result1 uint32
		result2 error
	}
	ReleaseStub        func(port uint32)
	releaseMutex       sync.RWMutex
	releaseArgsForCall []struct {
		port uint32
	}
	RemoveStub        func(port uint32) error
	removeMutex       sync.RWMutex
	removeArgsForCall []struct {
		port uint32
	}
	removeReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakePortPool) Acquire() (uint32, error) {
	fake.acquireMutex.Lock()
	fake.acquireArgsForCall = append(fake.acquireArgsForCall, struct{}{})
	fake.recordInvocation("Acquire", []interface{}{})
	fake.acquireMutex.Unlock()
	if fake.AcquireStub != nil {
		return fake.AcquireStub()
	} else {
		return fake.acquireReturns.result1, fake.acquireReturns.result2
	}
}

func (fake *FakePortPool) AcquireCallCount() int {
	fake.acquireMutex.RLock()
	defer fake.acquireMutex.RUnlock()
	return len(fake.acquireArgsForCall)
}

func (fake *FakePortPool) AcquireReturns(result1 uint32, result2 error) {
	fake.AcquireStub = nil
	fake.acquireReturns = struct {
		result1 uint32
		result2 error
	}{result1, result2}
}

func (fake *FakePortPool) Release(port uint32) {
	fake.releaseMutex.Lock()
	fake.releaseArgsForCall = append(fake.releaseArgsForCall, struct {
		port uint32
	}{port})
	fake.recordInvocation("Release", []interface{}{port})
	fake.releaseMutex.Unlock()
	if fake.ReleaseStub != nil {
		fake.ReleaseStub(port)
	}
}

func (fake *FakePortPool) ReleaseCallCount() int {
	fake.releaseMutex.RLock()
	defer fake.releaseMutex.RUnlock()
	return len(fake.releaseArgsForCall)
}

func (fake *FakePortPool) ReleaseArgsForCall(i int) uint32 {
	fake.releaseMutex.RLock()
	defer fake.releaseMutex.RUnlock()
	return fake.releaseArgsForCall[i].port
}

func (fake *FakePortPool) Remove(port uint32) error {
	fake.removeMutex.Lock()
	fake.removeArgsForCall = append(fake.removeArgsForCall, struct {
		port uint32
	}{port})
	fake.recordInvocation("Remove", []interface{}{port})
	fake.removeMutex.Unlock()
	if fake.RemoveStub != nil {
		return fake.RemoveStub(port)
	} else {
		return fake.removeReturns.result1
	}
}

func (fake *FakePortPool) RemoveCallCount() int {
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	return len(fake.removeArgsForCall)
}

func (fake *FakePortPool) RemoveArgsForCall(i int) uint32 {
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	return fake.removeArgsForCall[i].port
}

func (fake *FakePortPool) RemoveReturns(result1 error) {
	fake.RemoveStub = nil
	fake.removeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakePortPool) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.acquireMutex.RLock()
	defer fake.acquireMutex.RUnlock()
	fake.releaseMutex.RLock()
	defer fake.releaseMutex.RUnlock()
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	return fake.invocations
}

func (fake *FakePortPool) recordInvocation(key string, args []interface{}) {
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