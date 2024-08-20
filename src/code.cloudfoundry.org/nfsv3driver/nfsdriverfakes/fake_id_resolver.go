// Code generated by counterfeiter. DO NOT EDIT.
package nfsdriverfakes

import (
	"sync"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/nfsv3driver"
)

type FakeIdResolver struct {
	ResolveStub        func(dockerdriver.Env, string, string) (string, string, error)
	resolveMutex       sync.RWMutex
	resolveArgsForCall []struct {
		arg1 dockerdriver.Env
		arg2 string
		arg3 string
	}
	resolveReturns struct {
		result1 string
		result2 string
		result3 error
	}
	resolveReturnsOnCall map[int]struct {
		result1 string
		result2 string
		result3 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeIdResolver) Resolve(arg1 dockerdriver.Env, arg2 string, arg3 string) (string, string, error) {
	fake.resolveMutex.Lock()
	ret, specificReturn := fake.resolveReturnsOnCall[len(fake.resolveArgsForCall)]
	fake.resolveArgsForCall = append(fake.resolveArgsForCall, struct {
		arg1 dockerdriver.Env
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	stub := fake.ResolveStub
	fakeReturns := fake.resolveReturns
	fake.recordInvocation("Resolve", []interface{}{arg1, arg2, arg3})
	fake.resolveMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeIdResolver) ResolveCallCount() int {
	fake.resolveMutex.RLock()
	defer fake.resolveMutex.RUnlock()
	return len(fake.resolveArgsForCall)
}

func (fake *FakeIdResolver) ResolveCalls(stub func(dockerdriver.Env, string, string) (string, string, error)) {
	fake.resolveMutex.Lock()
	defer fake.resolveMutex.Unlock()
	fake.ResolveStub = stub
}

func (fake *FakeIdResolver) ResolveArgsForCall(i int) (dockerdriver.Env, string, string) {
	fake.resolveMutex.RLock()
	defer fake.resolveMutex.RUnlock()
	argsForCall := fake.resolveArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeIdResolver) ResolveReturns(result1 string, result2 string, result3 error) {
	fake.resolveMutex.Lock()
	defer fake.resolveMutex.Unlock()
	fake.ResolveStub = nil
	fake.resolveReturns = struct {
		result1 string
		result2 string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeIdResolver) ResolveReturnsOnCall(i int, result1 string, result2 string, result3 error) {
	fake.resolveMutex.Lock()
	defer fake.resolveMutex.Unlock()
	fake.ResolveStub = nil
	if fake.resolveReturnsOnCall == nil {
		fake.resolveReturnsOnCall = make(map[int]struct {
			result1 string
			result2 string
			result3 error
		})
	}
	fake.resolveReturnsOnCall[i] = struct {
		result1 string
		result2 string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeIdResolver) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.resolveMutex.RLock()
	defer fake.resolveMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeIdResolver) recordInvocation(key string, args []interface{}) {
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

var _ nfsv3driver.IdResolver = new(FakeIdResolver)