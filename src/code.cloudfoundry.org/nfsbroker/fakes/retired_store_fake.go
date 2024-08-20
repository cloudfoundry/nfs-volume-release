// ---------------------------------------------------------------------------------
// NOTE: the last line of this file had to be removed to avoid a circular dependency
// ---------------------------------------------------------------------------------

// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/service-broker-store/brokerstore"
	"github.com/pivotal-cf/brokerapi/v11/domain"
)

type FakeRetiredStore struct {
	CleanupStub        func() error
	cleanupMutex       sync.RWMutex
	cleanupArgsForCall []struct {
	}
	cleanupReturns struct {
		result1 error
	}
	cleanupReturnsOnCall map[int]struct {
		result1 error
	}
	CreateBindingDetailsStub        func(string, domain.BindDetails) error
	createBindingDetailsMutex       sync.RWMutex
	createBindingDetailsArgsForCall []struct {
		arg1 string
		arg2 domain.BindDetails
	}
	createBindingDetailsReturns struct {
		result1 error
	}
	createBindingDetailsReturnsOnCall map[int]struct {
		result1 error
	}
	CreateInstanceDetailsStub        func(string, brokerstore.ServiceInstance) error
	createInstanceDetailsMutex       sync.RWMutex
	createInstanceDetailsArgsForCall []struct {
		arg1 string
		arg2 brokerstore.ServiceInstance
	}
	createInstanceDetailsReturns struct {
		result1 error
	}
	createInstanceDetailsReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteBindingDetailsStub        func(string) error
	deleteBindingDetailsMutex       sync.RWMutex
	deleteBindingDetailsArgsForCall []struct {
		arg1 string
	}
	deleteBindingDetailsReturns struct {
		result1 error
	}
	deleteBindingDetailsReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteInstanceDetailsStub        func(string) error
	deleteInstanceDetailsMutex       sync.RWMutex
	deleteInstanceDetailsArgsForCall []struct {
		arg1 string
	}
	deleteInstanceDetailsReturns struct {
		result1 error
	}
	deleteInstanceDetailsReturnsOnCall map[int]struct {
		result1 error
	}
	IsBindingConflictStub        func(string, domain.BindDetails) bool
	isBindingConflictMutex       sync.RWMutex
	isBindingConflictArgsForCall []struct {
		arg1 string
		arg2 domain.BindDetails
	}
	isBindingConflictReturns struct {
		result1 bool
	}
	isBindingConflictReturnsOnCall map[int]struct {
		result1 bool
	}
	IsInstanceConflictStub        func(string, brokerstore.ServiceInstance) bool
	isInstanceConflictMutex       sync.RWMutex
	isInstanceConflictArgsForCall []struct {
		arg1 string
		arg2 brokerstore.ServiceInstance
	}
	isInstanceConflictReturns struct {
		result1 bool
	}
	isInstanceConflictReturnsOnCall map[int]struct {
		result1 bool
	}
	IsRetiredStub        func() (bool, error)
	isRetiredMutex       sync.RWMutex
	isRetiredArgsForCall []struct {
	}
	isRetiredReturns struct {
		result1 bool
		result2 error
	}
	isRetiredReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	RestoreStub        func(lager.Logger) error
	restoreMutex       sync.RWMutex
	restoreArgsForCall []struct {
		arg1 lager.Logger
	}
	restoreReturns struct {
		result1 error
	}
	restoreReturnsOnCall map[int]struct {
		result1 error
	}
	RetrieveAllBindingDetailsStub        func() (map[string]domain.BindDetails, error)
	retrieveAllBindingDetailsMutex       sync.RWMutex
	retrieveAllBindingDetailsArgsForCall []struct {
	}
	retrieveAllBindingDetailsReturns struct {
		result1 map[string]domain.BindDetails
		result2 error
	}
	retrieveAllBindingDetailsReturnsOnCall map[int]struct {
		result1 map[string]domain.BindDetails
		result2 error
	}
	RetrieveAllInstanceDetailsStub        func() (map[string]brokerstore.ServiceInstance, error)
	retrieveAllInstanceDetailsMutex       sync.RWMutex
	retrieveAllInstanceDetailsArgsForCall []struct {
	}
	retrieveAllInstanceDetailsReturns struct {
		result1 map[string]brokerstore.ServiceInstance
		result2 error
	}
	retrieveAllInstanceDetailsReturnsOnCall map[int]struct {
		result1 map[string]brokerstore.ServiceInstance
		result2 error
	}
	RetrieveBindingDetailsStub        func(string) (domain.BindDetails, error)
	retrieveBindingDetailsMutex       sync.RWMutex
	retrieveBindingDetailsArgsForCall []struct {
		arg1 string
	}
	retrieveBindingDetailsReturns struct {
		result1 domain.BindDetails
		result2 error
	}
	retrieveBindingDetailsReturnsOnCall map[int]struct {
		result1 domain.BindDetails
		result2 error
	}
	RetrieveInstanceDetailsStub        func(string) (brokerstore.ServiceInstance, error)
	retrieveInstanceDetailsMutex       sync.RWMutex
	retrieveInstanceDetailsArgsForCall []struct {
		arg1 string
	}
	retrieveInstanceDetailsReturns struct {
		result1 brokerstore.ServiceInstance
		result2 error
	}
	retrieveInstanceDetailsReturnsOnCall map[int]struct {
		result1 brokerstore.ServiceInstance
		result2 error
	}
	SaveStub        func(lager.Logger) error
	saveMutex       sync.RWMutex
	saveArgsForCall []struct {
		arg1 lager.Logger
	}
	saveReturns struct {
		result1 error
	}
	saveReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRetiredStore) Cleanup() error {
	fake.cleanupMutex.Lock()
	ret, specificReturn := fake.cleanupReturnsOnCall[len(fake.cleanupArgsForCall)]
	fake.cleanupArgsForCall = append(fake.cleanupArgsForCall, struct {
	}{})
	stub := fake.CleanupStub
	fakeReturns := fake.cleanupReturns
	fake.recordInvocation("Cleanup", []interface{}{})
	fake.cleanupMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRetiredStore) CleanupCallCount() int {
	fake.cleanupMutex.RLock()
	defer fake.cleanupMutex.RUnlock()
	return len(fake.cleanupArgsForCall)
}

func (fake *FakeRetiredStore) CleanupCalls(stub func() error) {
	fake.cleanupMutex.Lock()
	defer fake.cleanupMutex.Unlock()
	fake.CleanupStub = stub
}

func (fake *FakeRetiredStore) CleanupReturns(result1 error) {
	fake.cleanupMutex.Lock()
	defer fake.cleanupMutex.Unlock()
	fake.CleanupStub = nil
	fake.cleanupReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) CleanupReturnsOnCall(i int, result1 error) {
	fake.cleanupMutex.Lock()
	defer fake.cleanupMutex.Unlock()
	fake.CleanupStub = nil
	if fake.cleanupReturnsOnCall == nil {
		fake.cleanupReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.cleanupReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) CreateBindingDetails(arg1 string, arg2 domain.BindDetails) error {
	fake.createBindingDetailsMutex.Lock()
	ret, specificReturn := fake.createBindingDetailsReturnsOnCall[len(fake.createBindingDetailsArgsForCall)]
	fake.createBindingDetailsArgsForCall = append(fake.createBindingDetailsArgsForCall, struct {
		arg1 string
		arg2 domain.BindDetails
	}{arg1, arg2})
	stub := fake.CreateBindingDetailsStub
	fakeReturns := fake.createBindingDetailsReturns
	fake.recordInvocation("CreateBindingDetails", []interface{}{arg1, arg2})
	fake.createBindingDetailsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRetiredStore) CreateBindingDetailsCallCount() int {
	fake.createBindingDetailsMutex.RLock()
	defer fake.createBindingDetailsMutex.RUnlock()
	return len(fake.createBindingDetailsArgsForCall)
}

func (fake *FakeRetiredStore) CreateBindingDetailsCalls(stub func(string, domain.BindDetails) error) {
	fake.createBindingDetailsMutex.Lock()
	defer fake.createBindingDetailsMutex.Unlock()
	fake.CreateBindingDetailsStub = stub
}

func (fake *FakeRetiredStore) CreateBindingDetailsArgsForCall(i int) (string, domain.BindDetails) {
	fake.createBindingDetailsMutex.RLock()
	defer fake.createBindingDetailsMutex.RUnlock()
	argsForCall := fake.createBindingDetailsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRetiredStore) CreateBindingDetailsReturns(result1 error) {
	fake.createBindingDetailsMutex.Lock()
	defer fake.createBindingDetailsMutex.Unlock()
	fake.CreateBindingDetailsStub = nil
	fake.createBindingDetailsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) CreateBindingDetailsReturnsOnCall(i int, result1 error) {
	fake.createBindingDetailsMutex.Lock()
	defer fake.createBindingDetailsMutex.Unlock()
	fake.CreateBindingDetailsStub = nil
	if fake.createBindingDetailsReturnsOnCall == nil {
		fake.createBindingDetailsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createBindingDetailsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) CreateInstanceDetails(arg1 string, arg2 brokerstore.ServiceInstance) error {
	fake.createInstanceDetailsMutex.Lock()
	ret, specificReturn := fake.createInstanceDetailsReturnsOnCall[len(fake.createInstanceDetailsArgsForCall)]
	fake.createInstanceDetailsArgsForCall = append(fake.createInstanceDetailsArgsForCall, struct {
		arg1 string
		arg2 brokerstore.ServiceInstance
	}{arg1, arg2})
	stub := fake.CreateInstanceDetailsStub
	fakeReturns := fake.createInstanceDetailsReturns
	fake.recordInvocation("CreateInstanceDetails", []interface{}{arg1, arg2})
	fake.createInstanceDetailsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRetiredStore) CreateInstanceDetailsCallCount() int {
	fake.createInstanceDetailsMutex.RLock()
	defer fake.createInstanceDetailsMutex.RUnlock()
	return len(fake.createInstanceDetailsArgsForCall)
}

func (fake *FakeRetiredStore) CreateInstanceDetailsCalls(stub func(string, brokerstore.ServiceInstance) error) {
	fake.createInstanceDetailsMutex.Lock()
	defer fake.createInstanceDetailsMutex.Unlock()
	fake.CreateInstanceDetailsStub = stub
}

func (fake *FakeRetiredStore) CreateInstanceDetailsArgsForCall(i int) (string, brokerstore.ServiceInstance) {
	fake.createInstanceDetailsMutex.RLock()
	defer fake.createInstanceDetailsMutex.RUnlock()
	argsForCall := fake.createInstanceDetailsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRetiredStore) CreateInstanceDetailsReturns(result1 error) {
	fake.createInstanceDetailsMutex.Lock()
	defer fake.createInstanceDetailsMutex.Unlock()
	fake.CreateInstanceDetailsStub = nil
	fake.createInstanceDetailsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) CreateInstanceDetailsReturnsOnCall(i int, result1 error) {
	fake.createInstanceDetailsMutex.Lock()
	defer fake.createInstanceDetailsMutex.Unlock()
	fake.CreateInstanceDetailsStub = nil
	if fake.createInstanceDetailsReturnsOnCall == nil {
		fake.createInstanceDetailsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createInstanceDetailsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) DeleteBindingDetails(arg1 string) error {
	fake.deleteBindingDetailsMutex.Lock()
	ret, specificReturn := fake.deleteBindingDetailsReturnsOnCall[len(fake.deleteBindingDetailsArgsForCall)]
	fake.deleteBindingDetailsArgsForCall = append(fake.deleteBindingDetailsArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.DeleteBindingDetailsStub
	fakeReturns := fake.deleteBindingDetailsReturns
	fake.recordInvocation("DeleteBindingDetails", []interface{}{arg1})
	fake.deleteBindingDetailsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRetiredStore) DeleteBindingDetailsCallCount() int {
	fake.deleteBindingDetailsMutex.RLock()
	defer fake.deleteBindingDetailsMutex.RUnlock()
	return len(fake.deleteBindingDetailsArgsForCall)
}

func (fake *FakeRetiredStore) DeleteBindingDetailsCalls(stub func(string) error) {
	fake.deleteBindingDetailsMutex.Lock()
	defer fake.deleteBindingDetailsMutex.Unlock()
	fake.DeleteBindingDetailsStub = stub
}

func (fake *FakeRetiredStore) DeleteBindingDetailsArgsForCall(i int) string {
	fake.deleteBindingDetailsMutex.RLock()
	defer fake.deleteBindingDetailsMutex.RUnlock()
	argsForCall := fake.deleteBindingDetailsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRetiredStore) DeleteBindingDetailsReturns(result1 error) {
	fake.deleteBindingDetailsMutex.Lock()
	defer fake.deleteBindingDetailsMutex.Unlock()
	fake.DeleteBindingDetailsStub = nil
	fake.deleteBindingDetailsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) DeleteBindingDetailsReturnsOnCall(i int, result1 error) {
	fake.deleteBindingDetailsMutex.Lock()
	defer fake.deleteBindingDetailsMutex.Unlock()
	fake.DeleteBindingDetailsStub = nil
	if fake.deleteBindingDetailsReturnsOnCall == nil {
		fake.deleteBindingDetailsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteBindingDetailsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) DeleteInstanceDetails(arg1 string) error {
	fake.deleteInstanceDetailsMutex.Lock()
	ret, specificReturn := fake.deleteInstanceDetailsReturnsOnCall[len(fake.deleteInstanceDetailsArgsForCall)]
	fake.deleteInstanceDetailsArgsForCall = append(fake.deleteInstanceDetailsArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.DeleteInstanceDetailsStub
	fakeReturns := fake.deleteInstanceDetailsReturns
	fake.recordInvocation("DeleteInstanceDetails", []interface{}{arg1})
	fake.deleteInstanceDetailsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRetiredStore) DeleteInstanceDetailsCallCount() int {
	fake.deleteInstanceDetailsMutex.RLock()
	defer fake.deleteInstanceDetailsMutex.RUnlock()
	return len(fake.deleteInstanceDetailsArgsForCall)
}

func (fake *FakeRetiredStore) DeleteInstanceDetailsCalls(stub func(string) error) {
	fake.deleteInstanceDetailsMutex.Lock()
	defer fake.deleteInstanceDetailsMutex.Unlock()
	fake.DeleteInstanceDetailsStub = stub
}

func (fake *FakeRetiredStore) DeleteInstanceDetailsArgsForCall(i int) string {
	fake.deleteInstanceDetailsMutex.RLock()
	defer fake.deleteInstanceDetailsMutex.RUnlock()
	argsForCall := fake.deleteInstanceDetailsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRetiredStore) DeleteInstanceDetailsReturns(result1 error) {
	fake.deleteInstanceDetailsMutex.Lock()
	defer fake.deleteInstanceDetailsMutex.Unlock()
	fake.DeleteInstanceDetailsStub = nil
	fake.deleteInstanceDetailsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) DeleteInstanceDetailsReturnsOnCall(i int, result1 error) {
	fake.deleteInstanceDetailsMutex.Lock()
	defer fake.deleteInstanceDetailsMutex.Unlock()
	fake.DeleteInstanceDetailsStub = nil
	if fake.deleteInstanceDetailsReturnsOnCall == nil {
		fake.deleteInstanceDetailsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteInstanceDetailsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) IsBindingConflict(arg1 string, arg2 domain.BindDetails) bool {
	fake.isBindingConflictMutex.Lock()
	ret, specificReturn := fake.isBindingConflictReturnsOnCall[len(fake.isBindingConflictArgsForCall)]
	fake.isBindingConflictArgsForCall = append(fake.isBindingConflictArgsForCall, struct {
		arg1 string
		arg2 domain.BindDetails
	}{arg1, arg2})
	stub := fake.IsBindingConflictStub
	fakeReturns := fake.isBindingConflictReturns
	fake.recordInvocation("IsBindingConflict", []interface{}{arg1, arg2})
	fake.isBindingConflictMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRetiredStore) IsBindingConflictCallCount() int {
	fake.isBindingConflictMutex.RLock()
	defer fake.isBindingConflictMutex.RUnlock()
	return len(fake.isBindingConflictArgsForCall)
}

func (fake *FakeRetiredStore) IsBindingConflictCalls(stub func(string, domain.BindDetails) bool) {
	fake.isBindingConflictMutex.Lock()
	defer fake.isBindingConflictMutex.Unlock()
	fake.IsBindingConflictStub = stub
}

func (fake *FakeRetiredStore) IsBindingConflictArgsForCall(i int) (string, domain.BindDetails) {
	fake.isBindingConflictMutex.RLock()
	defer fake.isBindingConflictMutex.RUnlock()
	argsForCall := fake.isBindingConflictArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRetiredStore) IsBindingConflictReturns(result1 bool) {
	fake.isBindingConflictMutex.Lock()
	defer fake.isBindingConflictMutex.Unlock()
	fake.IsBindingConflictStub = nil
	fake.isBindingConflictReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeRetiredStore) IsBindingConflictReturnsOnCall(i int, result1 bool) {
	fake.isBindingConflictMutex.Lock()
	defer fake.isBindingConflictMutex.Unlock()
	fake.IsBindingConflictStub = nil
	if fake.isBindingConflictReturnsOnCall == nil {
		fake.isBindingConflictReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.isBindingConflictReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeRetiredStore) IsInstanceConflict(arg1 string, arg2 brokerstore.ServiceInstance) bool {
	fake.isInstanceConflictMutex.Lock()
	ret, specificReturn := fake.isInstanceConflictReturnsOnCall[len(fake.isInstanceConflictArgsForCall)]
	fake.isInstanceConflictArgsForCall = append(fake.isInstanceConflictArgsForCall, struct {
		arg1 string
		arg2 brokerstore.ServiceInstance
	}{arg1, arg2})
	stub := fake.IsInstanceConflictStub
	fakeReturns := fake.isInstanceConflictReturns
	fake.recordInvocation("IsInstanceConflict", []interface{}{arg1, arg2})
	fake.isInstanceConflictMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRetiredStore) IsInstanceConflictCallCount() int {
	fake.isInstanceConflictMutex.RLock()
	defer fake.isInstanceConflictMutex.RUnlock()
	return len(fake.isInstanceConflictArgsForCall)
}

func (fake *FakeRetiredStore) IsInstanceConflictCalls(stub func(string, brokerstore.ServiceInstance) bool) {
	fake.isInstanceConflictMutex.Lock()
	defer fake.isInstanceConflictMutex.Unlock()
	fake.IsInstanceConflictStub = stub
}

func (fake *FakeRetiredStore) IsInstanceConflictArgsForCall(i int) (string, brokerstore.ServiceInstance) {
	fake.isInstanceConflictMutex.RLock()
	defer fake.isInstanceConflictMutex.RUnlock()
	argsForCall := fake.isInstanceConflictArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRetiredStore) IsInstanceConflictReturns(result1 bool) {
	fake.isInstanceConflictMutex.Lock()
	defer fake.isInstanceConflictMutex.Unlock()
	fake.IsInstanceConflictStub = nil
	fake.isInstanceConflictReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeRetiredStore) IsInstanceConflictReturnsOnCall(i int, result1 bool) {
	fake.isInstanceConflictMutex.Lock()
	defer fake.isInstanceConflictMutex.Unlock()
	fake.IsInstanceConflictStub = nil
	if fake.isInstanceConflictReturnsOnCall == nil {
		fake.isInstanceConflictReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.isInstanceConflictReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeRetiredStore) IsRetired() (bool, error) {
	fake.isRetiredMutex.Lock()
	ret, specificReturn := fake.isRetiredReturnsOnCall[len(fake.isRetiredArgsForCall)]
	fake.isRetiredArgsForCall = append(fake.isRetiredArgsForCall, struct {
	}{})
	stub := fake.IsRetiredStub
	fakeReturns := fake.isRetiredReturns
	fake.recordInvocation("IsRetired", []interface{}{})
	fake.isRetiredMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRetiredStore) IsRetiredCallCount() int {
	fake.isRetiredMutex.RLock()
	defer fake.isRetiredMutex.RUnlock()
	return len(fake.isRetiredArgsForCall)
}

func (fake *FakeRetiredStore) IsRetiredCalls(stub func() (bool, error)) {
	fake.isRetiredMutex.Lock()
	defer fake.isRetiredMutex.Unlock()
	fake.IsRetiredStub = stub
}

func (fake *FakeRetiredStore) IsRetiredReturns(result1 bool, result2 error) {
	fake.isRetiredMutex.Lock()
	defer fake.isRetiredMutex.Unlock()
	fake.IsRetiredStub = nil
	fake.isRetiredReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) IsRetiredReturnsOnCall(i int, result1 bool, result2 error) {
	fake.isRetiredMutex.Lock()
	defer fake.isRetiredMutex.Unlock()
	fake.IsRetiredStub = nil
	if fake.isRetiredReturnsOnCall == nil {
		fake.isRetiredReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.isRetiredReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) Restore(arg1 lager.Logger) error {
	fake.restoreMutex.Lock()
	ret, specificReturn := fake.restoreReturnsOnCall[len(fake.restoreArgsForCall)]
	fake.restoreArgsForCall = append(fake.restoreArgsForCall, struct {
		arg1 lager.Logger
	}{arg1})
	stub := fake.RestoreStub
	fakeReturns := fake.restoreReturns
	fake.recordInvocation("Restore", []interface{}{arg1})
	fake.restoreMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRetiredStore) RestoreCallCount() int {
	fake.restoreMutex.RLock()
	defer fake.restoreMutex.RUnlock()
	return len(fake.restoreArgsForCall)
}

func (fake *FakeRetiredStore) RestoreCalls(stub func(lager.Logger) error) {
	fake.restoreMutex.Lock()
	defer fake.restoreMutex.Unlock()
	fake.RestoreStub = stub
}

func (fake *FakeRetiredStore) RestoreArgsForCall(i int) lager.Logger {
	fake.restoreMutex.RLock()
	defer fake.restoreMutex.RUnlock()
	argsForCall := fake.restoreArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRetiredStore) RestoreReturns(result1 error) {
	fake.restoreMutex.Lock()
	defer fake.restoreMutex.Unlock()
	fake.RestoreStub = nil
	fake.restoreReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) RestoreReturnsOnCall(i int, result1 error) {
	fake.restoreMutex.Lock()
	defer fake.restoreMutex.Unlock()
	fake.RestoreStub = nil
	if fake.restoreReturnsOnCall == nil {
		fake.restoreReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.restoreReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) RetrieveAllBindingDetails() (map[string]domain.BindDetails, error) {
	fake.retrieveAllBindingDetailsMutex.Lock()
	ret, specificReturn := fake.retrieveAllBindingDetailsReturnsOnCall[len(fake.retrieveAllBindingDetailsArgsForCall)]
	fake.retrieveAllBindingDetailsArgsForCall = append(fake.retrieveAllBindingDetailsArgsForCall, struct {
	}{})
	stub := fake.RetrieveAllBindingDetailsStub
	fakeReturns := fake.retrieveAllBindingDetailsReturns
	fake.recordInvocation("RetrieveAllBindingDetails", []interface{}{})
	fake.retrieveAllBindingDetailsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRetiredStore) RetrieveAllBindingDetailsCallCount() int {
	fake.retrieveAllBindingDetailsMutex.RLock()
	defer fake.retrieveAllBindingDetailsMutex.RUnlock()
	return len(fake.retrieveAllBindingDetailsArgsForCall)
}

func (fake *FakeRetiredStore) RetrieveAllBindingDetailsCalls(stub func() (map[string]domain.BindDetails, error)) {
	fake.retrieveAllBindingDetailsMutex.Lock()
	defer fake.retrieveAllBindingDetailsMutex.Unlock()
	fake.RetrieveAllBindingDetailsStub = stub
}

func (fake *FakeRetiredStore) RetrieveAllBindingDetailsReturns(result1 map[string]domain.BindDetails, result2 error) {
	fake.retrieveAllBindingDetailsMutex.Lock()
	defer fake.retrieveAllBindingDetailsMutex.Unlock()
	fake.RetrieveAllBindingDetailsStub = nil
	fake.retrieveAllBindingDetailsReturns = struct {
		result1 map[string]domain.BindDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) RetrieveAllBindingDetailsReturnsOnCall(i int, result1 map[string]domain.BindDetails, result2 error) {
	fake.retrieveAllBindingDetailsMutex.Lock()
	defer fake.retrieveAllBindingDetailsMutex.Unlock()
	fake.RetrieveAllBindingDetailsStub = nil
	if fake.retrieveAllBindingDetailsReturnsOnCall == nil {
		fake.retrieveAllBindingDetailsReturnsOnCall = make(map[int]struct {
			result1 map[string]domain.BindDetails
			result2 error
		})
	}
	fake.retrieveAllBindingDetailsReturnsOnCall[i] = struct {
		result1 map[string]domain.BindDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) RetrieveAllInstanceDetails() (map[string]brokerstore.ServiceInstance, error) {
	fake.retrieveAllInstanceDetailsMutex.Lock()
	ret, specificReturn := fake.retrieveAllInstanceDetailsReturnsOnCall[len(fake.retrieveAllInstanceDetailsArgsForCall)]
	fake.retrieveAllInstanceDetailsArgsForCall = append(fake.retrieveAllInstanceDetailsArgsForCall, struct {
	}{})
	stub := fake.RetrieveAllInstanceDetailsStub
	fakeReturns := fake.retrieveAllInstanceDetailsReturns
	fake.recordInvocation("RetrieveAllInstanceDetails", []interface{}{})
	fake.retrieveAllInstanceDetailsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRetiredStore) RetrieveAllInstanceDetailsCallCount() int {
	fake.retrieveAllInstanceDetailsMutex.RLock()
	defer fake.retrieveAllInstanceDetailsMutex.RUnlock()
	return len(fake.retrieveAllInstanceDetailsArgsForCall)
}

func (fake *FakeRetiredStore) RetrieveAllInstanceDetailsCalls(stub func() (map[string]brokerstore.ServiceInstance, error)) {
	fake.retrieveAllInstanceDetailsMutex.Lock()
	defer fake.retrieveAllInstanceDetailsMutex.Unlock()
	fake.RetrieveAllInstanceDetailsStub = stub
}

func (fake *FakeRetiredStore) RetrieveAllInstanceDetailsReturns(result1 map[string]brokerstore.ServiceInstance, result2 error) {
	fake.retrieveAllInstanceDetailsMutex.Lock()
	defer fake.retrieveAllInstanceDetailsMutex.Unlock()
	fake.RetrieveAllInstanceDetailsStub = nil
	fake.retrieveAllInstanceDetailsReturns = struct {
		result1 map[string]brokerstore.ServiceInstance
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) RetrieveAllInstanceDetailsReturnsOnCall(i int, result1 map[string]brokerstore.ServiceInstance, result2 error) {
	fake.retrieveAllInstanceDetailsMutex.Lock()
	defer fake.retrieveAllInstanceDetailsMutex.Unlock()
	fake.RetrieveAllInstanceDetailsStub = nil
	if fake.retrieveAllInstanceDetailsReturnsOnCall == nil {
		fake.retrieveAllInstanceDetailsReturnsOnCall = make(map[int]struct {
			result1 map[string]brokerstore.ServiceInstance
			result2 error
		})
	}
	fake.retrieveAllInstanceDetailsReturnsOnCall[i] = struct {
		result1 map[string]brokerstore.ServiceInstance
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) RetrieveBindingDetails(arg1 string) (domain.BindDetails, error) {
	fake.retrieveBindingDetailsMutex.Lock()
	ret, specificReturn := fake.retrieveBindingDetailsReturnsOnCall[len(fake.retrieveBindingDetailsArgsForCall)]
	fake.retrieveBindingDetailsArgsForCall = append(fake.retrieveBindingDetailsArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.RetrieveBindingDetailsStub
	fakeReturns := fake.retrieveBindingDetailsReturns
	fake.recordInvocation("RetrieveBindingDetails", []interface{}{arg1})
	fake.retrieveBindingDetailsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRetiredStore) RetrieveBindingDetailsCallCount() int {
	fake.retrieveBindingDetailsMutex.RLock()
	defer fake.retrieveBindingDetailsMutex.RUnlock()
	return len(fake.retrieveBindingDetailsArgsForCall)
}

func (fake *FakeRetiredStore) RetrieveBindingDetailsCalls(stub func(string) (domain.BindDetails, error)) {
	fake.retrieveBindingDetailsMutex.Lock()
	defer fake.retrieveBindingDetailsMutex.Unlock()
	fake.RetrieveBindingDetailsStub = stub
}

func (fake *FakeRetiredStore) RetrieveBindingDetailsArgsForCall(i int) string {
	fake.retrieveBindingDetailsMutex.RLock()
	defer fake.retrieveBindingDetailsMutex.RUnlock()
	argsForCall := fake.retrieveBindingDetailsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRetiredStore) RetrieveBindingDetailsReturns(result1 domain.BindDetails, result2 error) {
	fake.retrieveBindingDetailsMutex.Lock()
	defer fake.retrieveBindingDetailsMutex.Unlock()
	fake.RetrieveBindingDetailsStub = nil
	fake.retrieveBindingDetailsReturns = struct {
		result1 domain.BindDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) RetrieveBindingDetailsReturnsOnCall(i int, result1 domain.BindDetails, result2 error) {
	fake.retrieveBindingDetailsMutex.Lock()
	defer fake.retrieveBindingDetailsMutex.Unlock()
	fake.RetrieveBindingDetailsStub = nil
	if fake.retrieveBindingDetailsReturnsOnCall == nil {
		fake.retrieveBindingDetailsReturnsOnCall = make(map[int]struct {
			result1 domain.BindDetails
			result2 error
		})
	}
	fake.retrieveBindingDetailsReturnsOnCall[i] = struct {
		result1 domain.BindDetails
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) RetrieveInstanceDetails(arg1 string) (brokerstore.ServiceInstance, error) {
	fake.retrieveInstanceDetailsMutex.Lock()
	ret, specificReturn := fake.retrieveInstanceDetailsReturnsOnCall[len(fake.retrieveInstanceDetailsArgsForCall)]
	fake.retrieveInstanceDetailsArgsForCall = append(fake.retrieveInstanceDetailsArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.RetrieveInstanceDetailsStub
	fakeReturns := fake.retrieveInstanceDetailsReturns
	fake.recordInvocation("RetrieveInstanceDetails", []interface{}{arg1})
	fake.retrieveInstanceDetailsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRetiredStore) RetrieveInstanceDetailsCallCount() int {
	fake.retrieveInstanceDetailsMutex.RLock()
	defer fake.retrieveInstanceDetailsMutex.RUnlock()
	return len(fake.retrieveInstanceDetailsArgsForCall)
}

func (fake *FakeRetiredStore) RetrieveInstanceDetailsCalls(stub func(string) (brokerstore.ServiceInstance, error)) {
	fake.retrieveInstanceDetailsMutex.Lock()
	defer fake.retrieveInstanceDetailsMutex.Unlock()
	fake.RetrieveInstanceDetailsStub = stub
}

func (fake *FakeRetiredStore) RetrieveInstanceDetailsArgsForCall(i int) string {
	fake.retrieveInstanceDetailsMutex.RLock()
	defer fake.retrieveInstanceDetailsMutex.RUnlock()
	argsForCall := fake.retrieveInstanceDetailsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRetiredStore) RetrieveInstanceDetailsReturns(result1 brokerstore.ServiceInstance, result2 error) {
	fake.retrieveInstanceDetailsMutex.Lock()
	defer fake.retrieveInstanceDetailsMutex.Unlock()
	fake.RetrieveInstanceDetailsStub = nil
	fake.retrieveInstanceDetailsReturns = struct {
		result1 brokerstore.ServiceInstance
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) RetrieveInstanceDetailsReturnsOnCall(i int, result1 brokerstore.ServiceInstance, result2 error) {
	fake.retrieveInstanceDetailsMutex.Lock()
	defer fake.retrieveInstanceDetailsMutex.Unlock()
	fake.RetrieveInstanceDetailsStub = nil
	if fake.retrieveInstanceDetailsReturnsOnCall == nil {
		fake.retrieveInstanceDetailsReturnsOnCall = make(map[int]struct {
			result1 brokerstore.ServiceInstance
			result2 error
		})
	}
	fake.retrieveInstanceDetailsReturnsOnCall[i] = struct {
		result1 brokerstore.ServiceInstance
		result2 error
	}{result1, result2}
}

func (fake *FakeRetiredStore) Save(arg1 lager.Logger) error {
	fake.saveMutex.Lock()
	ret, specificReturn := fake.saveReturnsOnCall[len(fake.saveArgsForCall)]
	fake.saveArgsForCall = append(fake.saveArgsForCall, struct {
		arg1 lager.Logger
	}{arg1})
	stub := fake.SaveStub
	fakeReturns := fake.saveReturns
	fake.recordInvocation("Save", []interface{}{arg1})
	fake.saveMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRetiredStore) SaveCallCount() int {
	fake.saveMutex.RLock()
	defer fake.saveMutex.RUnlock()
	return len(fake.saveArgsForCall)
}

func (fake *FakeRetiredStore) SaveCalls(stub func(lager.Logger) error) {
	fake.saveMutex.Lock()
	defer fake.saveMutex.Unlock()
	fake.SaveStub = stub
}

func (fake *FakeRetiredStore) SaveArgsForCall(i int) lager.Logger {
	fake.saveMutex.RLock()
	defer fake.saveMutex.RUnlock()
	argsForCall := fake.saveArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRetiredStore) SaveReturns(result1 error) {
	fake.saveMutex.Lock()
	defer fake.saveMutex.Unlock()
	fake.SaveStub = nil
	fake.saveReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) SaveReturnsOnCall(i int, result1 error) {
	fake.saveMutex.Lock()
	defer fake.saveMutex.Unlock()
	fake.SaveStub = nil
	if fake.saveReturnsOnCall == nil {
		fake.saveReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.saveReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRetiredStore) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.cleanupMutex.RLock()
	defer fake.cleanupMutex.RUnlock()
	fake.createBindingDetailsMutex.RLock()
	defer fake.createBindingDetailsMutex.RUnlock()
	fake.createInstanceDetailsMutex.RLock()
	defer fake.createInstanceDetailsMutex.RUnlock()
	fake.deleteBindingDetailsMutex.RLock()
	defer fake.deleteBindingDetailsMutex.RUnlock()
	fake.deleteInstanceDetailsMutex.RLock()
	defer fake.deleteInstanceDetailsMutex.RUnlock()
	fake.isBindingConflictMutex.RLock()
	defer fake.isBindingConflictMutex.RUnlock()
	fake.isInstanceConflictMutex.RLock()
	defer fake.isInstanceConflictMutex.RUnlock()
	fake.isRetiredMutex.RLock()
	defer fake.isRetiredMutex.RUnlock()
	fake.restoreMutex.RLock()
	defer fake.restoreMutex.RUnlock()
	fake.retrieveAllBindingDetailsMutex.RLock()
	defer fake.retrieveAllBindingDetailsMutex.RUnlock()
	fake.retrieveAllInstanceDetailsMutex.RLock()
	defer fake.retrieveAllInstanceDetailsMutex.RUnlock()
	fake.retrieveBindingDetailsMutex.RLock()
	defer fake.retrieveBindingDetailsMutex.RUnlock()
	fake.retrieveInstanceDetailsMutex.RLock()
	defer fake.retrieveInstanceDetailsMutex.RUnlock()
	fake.saveMutex.RLock()
	defer fake.saveMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRetiredStore) recordInvocation(key string, args []interface{}) {
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

// NOTE: needed to comment this out to avoid circular dependency
// var _ main.RetiredStore = new(FakeRetiredStore)