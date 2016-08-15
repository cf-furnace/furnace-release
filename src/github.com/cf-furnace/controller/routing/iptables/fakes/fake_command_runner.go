// This file was generated by counterfeiter
package fakes

import (
	"os"
	"os/exec"
	"sync"
)

type FakeCommandRunner struct {
	RunStub        func(*exec.Cmd) error
	runMutex       sync.RWMutex
	runArgsForCall []struct {
		arg1 *exec.Cmd
	}
	runReturns struct {
		result1 error
	}
	StartStub        func(*exec.Cmd) error
	startMutex       sync.RWMutex
	startArgsForCall []struct {
		arg1 *exec.Cmd
	}
	startReturns struct {
		result1 error
	}
	BackgroundStub        func(*exec.Cmd) error
	backgroundMutex       sync.RWMutex
	backgroundArgsForCall []struct {
		arg1 *exec.Cmd
	}
	backgroundReturns struct {
		result1 error
	}
	WaitStub        func(*exec.Cmd) error
	waitMutex       sync.RWMutex
	waitArgsForCall []struct {
		arg1 *exec.Cmd
	}
	waitReturns struct {
		result1 error
	}
	KillStub        func(*exec.Cmd) error
	killMutex       sync.RWMutex
	killArgsForCall []struct {
		arg1 *exec.Cmd
	}
	killReturns struct {
		result1 error
	}
	SignalStub        func(*exec.Cmd, os.Signal) error
	signalMutex       sync.RWMutex
	signalArgsForCall []struct {
		arg1 *exec.Cmd
		arg2 os.Signal
	}
	signalReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCommandRunner) Run(arg1 *exec.Cmd) error {
	fake.runMutex.Lock()
	fake.runArgsForCall = append(fake.runArgsForCall, struct {
		arg1 *exec.Cmd
	}{arg1})
	fake.recordInvocation("Run", []interface{}{arg1})
	fake.runMutex.Unlock()
	if fake.RunStub != nil {
		return fake.RunStub(arg1)
	} else {
		return fake.runReturns.result1
	}
}

func (fake *FakeCommandRunner) RunCallCount() int {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return len(fake.runArgsForCall)
}

func (fake *FakeCommandRunner) RunArgsForCall(i int) *exec.Cmd {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return fake.runArgsForCall[i].arg1
}

func (fake *FakeCommandRunner) RunReturns(result1 error) {
	fake.RunStub = nil
	fake.runReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCommandRunner) Start(arg1 *exec.Cmd) error {
	fake.startMutex.Lock()
	fake.startArgsForCall = append(fake.startArgsForCall, struct {
		arg1 *exec.Cmd
	}{arg1})
	fake.recordInvocation("Start", []interface{}{arg1})
	fake.startMutex.Unlock()
	if fake.StartStub != nil {
		return fake.StartStub(arg1)
	} else {
		return fake.startReturns.result1
	}
}

func (fake *FakeCommandRunner) StartCallCount() int {
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	return len(fake.startArgsForCall)
}

func (fake *FakeCommandRunner) StartArgsForCall(i int) *exec.Cmd {
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	return fake.startArgsForCall[i].arg1
}

func (fake *FakeCommandRunner) StartReturns(result1 error) {
	fake.StartStub = nil
	fake.startReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCommandRunner) Background(arg1 *exec.Cmd) error {
	fake.backgroundMutex.Lock()
	fake.backgroundArgsForCall = append(fake.backgroundArgsForCall, struct {
		arg1 *exec.Cmd
	}{arg1})
	fake.recordInvocation("Background", []interface{}{arg1})
	fake.backgroundMutex.Unlock()
	if fake.BackgroundStub != nil {
		return fake.BackgroundStub(arg1)
	} else {
		return fake.backgroundReturns.result1
	}
}

func (fake *FakeCommandRunner) BackgroundCallCount() int {
	fake.backgroundMutex.RLock()
	defer fake.backgroundMutex.RUnlock()
	return len(fake.backgroundArgsForCall)
}

func (fake *FakeCommandRunner) BackgroundArgsForCall(i int) *exec.Cmd {
	fake.backgroundMutex.RLock()
	defer fake.backgroundMutex.RUnlock()
	return fake.backgroundArgsForCall[i].arg1
}

func (fake *FakeCommandRunner) BackgroundReturns(result1 error) {
	fake.BackgroundStub = nil
	fake.backgroundReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCommandRunner) Wait(arg1 *exec.Cmd) error {
	fake.waitMutex.Lock()
	fake.waitArgsForCall = append(fake.waitArgsForCall, struct {
		arg1 *exec.Cmd
	}{arg1})
	fake.recordInvocation("Wait", []interface{}{arg1})
	fake.waitMutex.Unlock()
	if fake.WaitStub != nil {
		return fake.WaitStub(arg1)
	} else {
		return fake.waitReturns.result1
	}
}

func (fake *FakeCommandRunner) WaitCallCount() int {
	fake.waitMutex.RLock()
	defer fake.waitMutex.RUnlock()
	return len(fake.waitArgsForCall)
}

func (fake *FakeCommandRunner) WaitArgsForCall(i int) *exec.Cmd {
	fake.waitMutex.RLock()
	defer fake.waitMutex.RUnlock()
	return fake.waitArgsForCall[i].arg1
}

func (fake *FakeCommandRunner) WaitReturns(result1 error) {
	fake.WaitStub = nil
	fake.waitReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCommandRunner) Kill(arg1 *exec.Cmd) error {
	fake.killMutex.Lock()
	fake.killArgsForCall = append(fake.killArgsForCall, struct {
		arg1 *exec.Cmd
	}{arg1})
	fake.recordInvocation("Kill", []interface{}{arg1})
	fake.killMutex.Unlock()
	if fake.KillStub != nil {
		return fake.KillStub(arg1)
	} else {
		return fake.killReturns.result1
	}
}

func (fake *FakeCommandRunner) KillCallCount() int {
	fake.killMutex.RLock()
	defer fake.killMutex.RUnlock()
	return len(fake.killArgsForCall)
}

func (fake *FakeCommandRunner) KillArgsForCall(i int) *exec.Cmd {
	fake.killMutex.RLock()
	defer fake.killMutex.RUnlock()
	return fake.killArgsForCall[i].arg1
}

func (fake *FakeCommandRunner) KillReturns(result1 error) {
	fake.KillStub = nil
	fake.killReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCommandRunner) Signal(arg1 *exec.Cmd, arg2 os.Signal) error {
	fake.signalMutex.Lock()
	fake.signalArgsForCall = append(fake.signalArgsForCall, struct {
		arg1 *exec.Cmd
		arg2 os.Signal
	}{arg1, arg2})
	fake.recordInvocation("Signal", []interface{}{arg1, arg2})
	fake.signalMutex.Unlock()
	if fake.SignalStub != nil {
		return fake.SignalStub(arg1, arg2)
	} else {
		return fake.signalReturns.result1
	}
}

func (fake *FakeCommandRunner) SignalCallCount() int {
	fake.signalMutex.RLock()
	defer fake.signalMutex.RUnlock()
	return len(fake.signalArgsForCall)
}

func (fake *FakeCommandRunner) SignalArgsForCall(i int) (*exec.Cmd, os.Signal) {
	fake.signalMutex.RLock()
	defer fake.signalMutex.RUnlock()
	return fake.signalArgsForCall[i].arg1, fake.signalArgsForCall[i].arg2
}

func (fake *FakeCommandRunner) SignalReturns(result1 error) {
	fake.SignalStub = nil
	fake.signalReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCommandRunner) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	fake.backgroundMutex.RLock()
	defer fake.backgroundMutex.RUnlock()
	fake.waitMutex.RLock()
	defer fake.waitMutex.RUnlock()
	fake.killMutex.RLock()
	defer fake.killMutex.RUnlock()
	fake.signalMutex.RLock()
	defer fake.signalMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeCommandRunner) recordInvocation(key string, args []interface{}) {
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
