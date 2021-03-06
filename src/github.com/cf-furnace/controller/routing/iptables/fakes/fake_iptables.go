// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/cf-furnace/controller/routing/iptables"
)

type FakeIPTables struct {
	CreateChainStub        func(table iptables.Table, name string) error
	createChainMutex       sync.RWMutex
	createChainArgsForCall []struct {
		table iptables.Table
		name  string
	}
	createChainReturns struct {
		result1 error
	}
	DeleteChainStub        func(table iptables.Table, name string) error
	deleteChainMutex       sync.RWMutex
	deleteChainArgsForCall []struct {
		table iptables.Table
		name  string
	}
	deleteChainReturns struct {
		result1 error
	}
	DeleteChainReferencesStub        func(table iptables.Table, chain, referenced string) error
	deleteChainReferencesMutex       sync.RWMutex
	deleteChainReferencesArgsForCall []struct {
		table      iptables.Table
		chain      string
		referenced string
	}
	deleteChainReferencesReturns struct {
		result1 error
	}
	FlushChainStub        func(table iptables.Table, name string) error
	flushChainMutex       sync.RWMutex
	flushChainArgsForCall []struct {
		table iptables.Table
		name  string
	}
	flushChainReturns struct {
		result1 error
	}
	ListChainsStub        func(table iptables.Table) ([]string, error)
	listChainsMutex       sync.RWMutex
	listChainsArgsForCall []struct {
		table iptables.Table
	}
	listChainsReturns struct {
		result1 []string
		result2 error
	}
	PrependRuleStub        func(table iptables.Table, chain string, rule iptables.Rule) error
	prependRuleMutex       sync.RWMutex
	prependRuleArgsForCall []struct {
		table iptables.Table
		chain string
		rule  iptables.Rule
	}
	prependRuleReturns struct {
		result1 error
	}
	AppendRuleStub        func(table iptables.Table, chain string, rule iptables.Rule) error
	appendRuleMutex       sync.RWMutex
	appendRuleArgsForCall []struct {
		table iptables.Table
		chain string
		rule  iptables.Rule
	}
	appendRuleReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeIPTables) CreateChain(table iptables.Table, name string) error {
	fake.createChainMutex.Lock()
	fake.createChainArgsForCall = append(fake.createChainArgsForCall, struct {
		table iptables.Table
		name  string
	}{table, name})
	fake.recordInvocation("CreateChain", []interface{}{table, name})
	fake.createChainMutex.Unlock()
	if fake.CreateChainStub != nil {
		return fake.CreateChainStub(table, name)
	} else {
		return fake.createChainReturns.result1
	}
}

func (fake *FakeIPTables) CreateChainCallCount() int {
	fake.createChainMutex.RLock()
	defer fake.createChainMutex.RUnlock()
	return len(fake.createChainArgsForCall)
}

func (fake *FakeIPTables) CreateChainArgsForCall(i int) (iptables.Table, string) {
	fake.createChainMutex.RLock()
	defer fake.createChainMutex.RUnlock()
	return fake.createChainArgsForCall[i].table, fake.createChainArgsForCall[i].name
}

func (fake *FakeIPTables) CreateChainReturns(result1 error) {
	fake.CreateChainStub = nil
	fake.createChainReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeIPTables) DeleteChain(table iptables.Table, name string) error {
	fake.deleteChainMutex.Lock()
	fake.deleteChainArgsForCall = append(fake.deleteChainArgsForCall, struct {
		table iptables.Table
		name  string
	}{table, name})
	fake.recordInvocation("DeleteChain", []interface{}{table, name})
	fake.deleteChainMutex.Unlock()
	if fake.DeleteChainStub != nil {
		return fake.DeleteChainStub(table, name)
	} else {
		return fake.deleteChainReturns.result1
	}
}

func (fake *FakeIPTables) DeleteChainCallCount() int {
	fake.deleteChainMutex.RLock()
	defer fake.deleteChainMutex.RUnlock()
	return len(fake.deleteChainArgsForCall)
}

func (fake *FakeIPTables) DeleteChainArgsForCall(i int) (iptables.Table, string) {
	fake.deleteChainMutex.RLock()
	defer fake.deleteChainMutex.RUnlock()
	return fake.deleteChainArgsForCall[i].table, fake.deleteChainArgsForCall[i].name
}

func (fake *FakeIPTables) DeleteChainReturns(result1 error) {
	fake.DeleteChainStub = nil
	fake.deleteChainReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeIPTables) DeleteChainReferences(table iptables.Table, chain string, referenced string) error {
	fake.deleteChainReferencesMutex.Lock()
	fake.deleteChainReferencesArgsForCall = append(fake.deleteChainReferencesArgsForCall, struct {
		table      iptables.Table
		chain      string
		referenced string
	}{table, chain, referenced})
	fake.recordInvocation("DeleteChainReferences", []interface{}{table, chain, referenced})
	fake.deleteChainReferencesMutex.Unlock()
	if fake.DeleteChainReferencesStub != nil {
		return fake.DeleteChainReferencesStub(table, chain, referenced)
	} else {
		return fake.deleteChainReferencesReturns.result1
	}
}

func (fake *FakeIPTables) DeleteChainReferencesCallCount() int {
	fake.deleteChainReferencesMutex.RLock()
	defer fake.deleteChainReferencesMutex.RUnlock()
	return len(fake.deleteChainReferencesArgsForCall)
}

func (fake *FakeIPTables) DeleteChainReferencesArgsForCall(i int) (iptables.Table, string, string) {
	fake.deleteChainReferencesMutex.RLock()
	defer fake.deleteChainReferencesMutex.RUnlock()
	return fake.deleteChainReferencesArgsForCall[i].table, fake.deleteChainReferencesArgsForCall[i].chain, fake.deleteChainReferencesArgsForCall[i].referenced
}

func (fake *FakeIPTables) DeleteChainReferencesReturns(result1 error) {
	fake.DeleteChainReferencesStub = nil
	fake.deleteChainReferencesReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeIPTables) FlushChain(table iptables.Table, name string) error {
	fake.flushChainMutex.Lock()
	fake.flushChainArgsForCall = append(fake.flushChainArgsForCall, struct {
		table iptables.Table
		name  string
	}{table, name})
	fake.recordInvocation("FlushChain", []interface{}{table, name})
	fake.flushChainMutex.Unlock()
	if fake.FlushChainStub != nil {
		return fake.FlushChainStub(table, name)
	} else {
		return fake.flushChainReturns.result1
	}
}

func (fake *FakeIPTables) FlushChainCallCount() int {
	fake.flushChainMutex.RLock()
	defer fake.flushChainMutex.RUnlock()
	return len(fake.flushChainArgsForCall)
}

func (fake *FakeIPTables) FlushChainArgsForCall(i int) (iptables.Table, string) {
	fake.flushChainMutex.RLock()
	defer fake.flushChainMutex.RUnlock()
	return fake.flushChainArgsForCall[i].table, fake.flushChainArgsForCall[i].name
}

func (fake *FakeIPTables) FlushChainReturns(result1 error) {
	fake.FlushChainStub = nil
	fake.flushChainReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeIPTables) ListChains(table iptables.Table) ([]string, error) {
	fake.listChainsMutex.Lock()
	fake.listChainsArgsForCall = append(fake.listChainsArgsForCall, struct {
		table iptables.Table
	}{table})
	fake.recordInvocation("ListChains", []interface{}{table})
	fake.listChainsMutex.Unlock()
	if fake.ListChainsStub != nil {
		return fake.ListChainsStub(table)
	} else {
		return fake.listChainsReturns.result1, fake.listChainsReturns.result2
	}
}

func (fake *FakeIPTables) ListChainsCallCount() int {
	fake.listChainsMutex.RLock()
	defer fake.listChainsMutex.RUnlock()
	return len(fake.listChainsArgsForCall)
}

func (fake *FakeIPTables) ListChainsArgsForCall(i int) iptables.Table {
	fake.listChainsMutex.RLock()
	defer fake.listChainsMutex.RUnlock()
	return fake.listChainsArgsForCall[i].table
}

func (fake *FakeIPTables) ListChainsReturns(result1 []string, result2 error) {
	fake.ListChainsStub = nil
	fake.listChainsReturns = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeIPTables) PrependRule(table iptables.Table, chain string, rule iptables.Rule) error {
	fake.prependRuleMutex.Lock()
	fake.prependRuleArgsForCall = append(fake.prependRuleArgsForCall, struct {
		table iptables.Table
		chain string
		rule  iptables.Rule
	}{table, chain, rule})
	fake.recordInvocation("PrependRule", []interface{}{table, chain, rule})
	fake.prependRuleMutex.Unlock()
	if fake.PrependRuleStub != nil {
		return fake.PrependRuleStub(table, chain, rule)
	} else {
		return fake.prependRuleReturns.result1
	}
}

func (fake *FakeIPTables) PrependRuleCallCount() int {
	fake.prependRuleMutex.RLock()
	defer fake.prependRuleMutex.RUnlock()
	return len(fake.prependRuleArgsForCall)
}

func (fake *FakeIPTables) PrependRuleArgsForCall(i int) (iptables.Table, string, iptables.Rule) {
	fake.prependRuleMutex.RLock()
	defer fake.prependRuleMutex.RUnlock()
	return fake.prependRuleArgsForCall[i].table, fake.prependRuleArgsForCall[i].chain, fake.prependRuleArgsForCall[i].rule
}

func (fake *FakeIPTables) PrependRuleReturns(result1 error) {
	fake.PrependRuleStub = nil
	fake.prependRuleReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeIPTables) AppendRule(table iptables.Table, chain string, rule iptables.Rule) error {
	fake.appendRuleMutex.Lock()
	fake.appendRuleArgsForCall = append(fake.appendRuleArgsForCall, struct {
		table iptables.Table
		chain string
		rule  iptables.Rule
	}{table, chain, rule})
	fake.recordInvocation("AppendRule", []interface{}{table, chain, rule})
	fake.appendRuleMutex.Unlock()
	if fake.AppendRuleStub != nil {
		return fake.AppendRuleStub(table, chain, rule)
	} else {
		return fake.appendRuleReturns.result1
	}
}

func (fake *FakeIPTables) AppendRuleCallCount() int {
	fake.appendRuleMutex.RLock()
	defer fake.appendRuleMutex.RUnlock()
	return len(fake.appendRuleArgsForCall)
}

func (fake *FakeIPTables) AppendRuleArgsForCall(i int) (iptables.Table, string, iptables.Rule) {
	fake.appendRuleMutex.RLock()
	defer fake.appendRuleMutex.RUnlock()
	return fake.appendRuleArgsForCall[i].table, fake.appendRuleArgsForCall[i].chain, fake.appendRuleArgsForCall[i].rule
}

func (fake *FakeIPTables) AppendRuleReturns(result1 error) {
	fake.AppendRuleStub = nil
	fake.appendRuleReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeIPTables) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createChainMutex.RLock()
	defer fake.createChainMutex.RUnlock()
	fake.deleteChainMutex.RLock()
	defer fake.deleteChainMutex.RUnlock()
	fake.deleteChainReferencesMutex.RLock()
	defer fake.deleteChainReferencesMutex.RUnlock()
	fake.flushChainMutex.RLock()
	defer fake.flushChainMutex.RUnlock()
	fake.listChainsMutex.RLock()
	defer fake.listChainsMutex.RUnlock()
	fake.prependRuleMutex.RLock()
	defer fake.prependRuleMutex.RUnlock()
	fake.appendRuleMutex.RLock()
	defer fake.appendRuleMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeIPTables) recordInvocation(key string, args []interface{}) {
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

var _ iptables.IPTables = new(FakeIPTables)
