// Code generated by mockery v2.40.1. DO NOT EDIT.

package bigquery

import (
	context "context"

	datapoliciespb "cloud.google.com/go/bigquery/datapolicies/apiv1/datapoliciespb"
	mock "github.com/stretchr/testify/mock"

	sync_to_target "github.com/raito-io/cli/base/access_provider/sync_to_target"
)

// mockMaskingDataCatalogRepository is an autogenerated mock type for the maskingDataCatalogRepository type
type mockMaskingDataCatalogRepository struct {
	mock.Mock
}

type mockMaskingDataCatalogRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *mockMaskingDataCatalogRepository) EXPECT() *mockMaskingDataCatalogRepository_Expecter {
	return &mockMaskingDataCatalogRepository_Expecter{mock: &_m.Mock}
}

// CreatePolicyTagWithDataPolicy provides a mock function with given fields: ctx, location, maskingType, ap
func (_m *mockMaskingDataCatalogRepository) CreatePolicyTagWithDataPolicy(ctx context.Context, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *sync_to_target.AccessProvider) (*BQMaskingInformation, error) {
	ret := _m.Called(ctx, location, maskingType, ap)

	if len(ret) == 0 {
		panic("no return value specified for CreatePolicyTagWithDataPolicy")
	}

	var r0 *BQMaskingInformation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, datapoliciespb.DataMaskingPolicy_PredefinedExpression, *sync_to_target.AccessProvider) (*BQMaskingInformation, error)); ok {
		return rf(ctx, location, maskingType, ap)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, datapoliciespb.DataMaskingPolicy_PredefinedExpression, *sync_to_target.AccessProvider) *BQMaskingInformation); ok {
		r0 = rf(ctx, location, maskingType, ap)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*BQMaskingInformation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, datapoliciespb.DataMaskingPolicy_PredefinedExpression, *sync_to_target.AccessProvider) error); ok {
		r1 = rf(ctx, location, maskingType, ap)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreatePolicyTagWithDataPolicy'
type mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call struct {
	*mock.Call
}

// CreatePolicyTagWithDataPolicy is a helper method to define mock.On call
//   - ctx context.Context
//   - location string
//   - maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression
//   - ap *sync_to_target.AccessProvider
func (_e *mockMaskingDataCatalogRepository_Expecter) CreatePolicyTagWithDataPolicy(ctx interface{}, location interface{}, maskingType interface{}, ap interface{}) *mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call {
	return &mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call{Call: _e.mock.On("CreatePolicyTagWithDataPolicy", ctx, location, maskingType, ap)}
}

func (_c *mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call) Run(run func(ctx context.Context, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *sync_to_target.AccessProvider)) *mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(datapoliciespb.DataMaskingPolicy_PredefinedExpression), args[3].(*sync_to_target.AccessProvider))
	})
	return _c
}

func (_c *mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call) Return(_a0 *BQMaskingInformation, err error) *mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call {
	_c.Call.Return(_a0, err)
	return _c
}

func (_c *mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call) RunAndReturn(run func(context.Context, string, datapoliciespb.DataMaskingPolicy_PredefinedExpression, *sync_to_target.AccessProvider) (*BQMaskingInformation, error)) *mockMaskingDataCatalogRepository_CreatePolicyTagWithDataPolicy_Call {
	_c.Call.Return(run)
	return _c
}

// DeletePolicyAndTag provides a mock function with given fields: ctx, policyTagId
func (_m *mockMaskingDataCatalogRepository) DeletePolicyAndTag(ctx context.Context, policyTagId string) error {
	ret := _m.Called(ctx, policyTagId)

	if len(ret) == 0 {
		panic("no return value specified for DeletePolicyAndTag")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, policyTagId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeletePolicyAndTag'
type mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call struct {
	*mock.Call
}

// DeletePolicyAndTag is a helper method to define mock.On call
//   - ctx context.Context
//   - policyTagId string
func (_e *mockMaskingDataCatalogRepository_Expecter) DeletePolicyAndTag(ctx interface{}, policyTagId interface{}) *mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call {
	return &mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call{Call: _e.mock.On("DeletePolicyAndTag", ctx, policyTagId)}
}

func (_c *mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call) Run(run func(ctx context.Context, policyTagId string)) *mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call) Return(_a0 error) *mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call) RunAndReturn(run func(context.Context, string) error) *mockMaskingDataCatalogRepository_DeletePolicyAndTag_Call {
	_c.Call.Return(run)
	return _c
}

// GetFineGrainedReaderMembers provides a mock function with given fields: ctx, tagId
func (_m *mockMaskingDataCatalogRepository) GetFineGrainedReaderMembers(ctx context.Context, tagId string) ([]string, error) {
	ret := _m.Called(ctx, tagId)

	if len(ret) == 0 {
		panic("no return value specified for GetFineGrainedReaderMembers")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]string, error)); ok {
		return rf(ctx, tagId)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []string); ok {
		r0 = rf(ctx, tagId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tagId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFineGrainedReaderMembers'
type mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call struct {
	*mock.Call
}

// GetFineGrainedReaderMembers is a helper method to define mock.On call
//   - ctx context.Context
//   - tagId string
func (_e *mockMaskingDataCatalogRepository_Expecter) GetFineGrainedReaderMembers(ctx interface{}, tagId interface{}) *mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call {
	return &mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call{Call: _e.mock.On("GetFineGrainedReaderMembers", ctx, tagId)}
}

func (_c *mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call) Run(run func(ctx context.Context, tagId string)) *mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call) Return(_a0 []string, _a1 error) *mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call) RunAndReturn(run func(context.Context, string) ([]string, error)) *mockMaskingDataCatalogRepository_GetFineGrainedReaderMembers_Call {
	_c.Call.Return(run)
	return _c
}

// GetLocationsForDataObjects provides a mock function with given fields: ctx, ap
func (_m *mockMaskingDataCatalogRepository) GetLocationsForDataObjects(ctx context.Context, ap *sync_to_target.AccessProvider) (map[string]string, map[string]string, error) {
	ret := _m.Called(ctx, ap)

	if len(ret) == 0 {
		panic("no return value specified for GetLocationsForDataObjects")
	}

	var r0 map[string]string
	var r1 map[string]string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, *sync_to_target.AccessProvider) (map[string]string, map[string]string, error)); ok {
		return rf(ctx, ap)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *sync_to_target.AccessProvider) map[string]string); ok {
		r0 = rf(ctx, ap)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *sync_to_target.AccessProvider) map[string]string); ok {
		r1 = rf(ctx, ap)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(map[string]string)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, *sync_to_target.AccessProvider) error); ok {
		r2 = rf(ctx, ap)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLocationsForDataObjects'
type mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call struct {
	*mock.Call
}

// GetLocationsForDataObjects is a helper method to define mock.On call
//   - ctx context.Context
//   - ap *sync_to_target.AccessProvider
func (_e *mockMaskingDataCatalogRepository_Expecter) GetLocationsForDataObjects(ctx interface{}, ap interface{}) *mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call {
	return &mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call{Call: _e.mock.On("GetLocationsForDataObjects", ctx, ap)}
}

func (_c *mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call) Run(run func(ctx context.Context, ap *sync_to_target.AccessProvider)) *mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*sync_to_target.AccessProvider))
	})
	return _c
}

func (_c *mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call) Return(_a0 map[string]string, _a1 map[string]string, _a2 error) *mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call) RunAndReturn(run func(context.Context, *sync_to_target.AccessProvider) (map[string]string, map[string]string, error)) *mockMaskingDataCatalogRepository_GetLocationsForDataObjects_Call {
	_c.Call.Return(run)
	return _c
}

// ListDataPolicies provides a mock function with given fields: ctx
func (_m *mockMaskingDataCatalogRepository) ListDataPolicies(ctx context.Context) (map[string]BQMaskingInformation, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ListDataPolicies")
	}

	var r0 map[string]BQMaskingInformation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (map[string]BQMaskingInformation, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) map[string]BQMaskingInformation); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]BQMaskingInformation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMaskingDataCatalogRepository_ListDataPolicies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListDataPolicies'
type mockMaskingDataCatalogRepository_ListDataPolicies_Call struct {
	*mock.Call
}

// ListDataPolicies is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockMaskingDataCatalogRepository_Expecter) ListDataPolicies(ctx interface{}) *mockMaskingDataCatalogRepository_ListDataPolicies_Call {
	return &mockMaskingDataCatalogRepository_ListDataPolicies_Call{Call: _e.mock.On("ListDataPolicies", ctx)}
}

func (_c *mockMaskingDataCatalogRepository_ListDataPolicies_Call) Run(run func(ctx context.Context)) *mockMaskingDataCatalogRepository_ListDataPolicies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockMaskingDataCatalogRepository_ListDataPolicies_Call) Return(_a0 map[string]BQMaskingInformation, _a1 error) *mockMaskingDataCatalogRepository_ListDataPolicies_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMaskingDataCatalogRepository_ListDataPolicies_Call) RunAndReturn(run func(context.Context) (map[string]BQMaskingInformation, error)) *mockMaskingDataCatalogRepository_ListDataPolicies_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateAccess provides a mock function with given fields: ctx, maskingInformation, who, deletedWho
func (_m *mockMaskingDataCatalogRepository) UpdateAccess(ctx context.Context, maskingInformation *BQMaskingInformation, who *sync_to_target.WhoItem, deletedWho *sync_to_target.WhoItem) error {
	ret := _m.Called(ctx, maskingInformation, who, deletedWho)

	if len(ret) == 0 {
		panic("no return value specified for UpdateAccess")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *BQMaskingInformation, *sync_to_target.WhoItem, *sync_to_target.WhoItem) error); ok {
		r0 = rf(ctx, maskingInformation, who, deletedWho)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockMaskingDataCatalogRepository_UpdateAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateAccess'
type mockMaskingDataCatalogRepository_UpdateAccess_Call struct {
	*mock.Call
}

// UpdateAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - maskingInformation *BQMaskingInformation
//   - who *sync_to_target.WhoItem
//   - deletedWho *sync_to_target.WhoItem
func (_e *mockMaskingDataCatalogRepository_Expecter) UpdateAccess(ctx interface{}, maskingInformation interface{}, who interface{}, deletedWho interface{}) *mockMaskingDataCatalogRepository_UpdateAccess_Call {
	return &mockMaskingDataCatalogRepository_UpdateAccess_Call{Call: _e.mock.On("UpdateAccess", ctx, maskingInformation, who, deletedWho)}
}

func (_c *mockMaskingDataCatalogRepository_UpdateAccess_Call) Run(run func(ctx context.Context, maskingInformation *BQMaskingInformation, who *sync_to_target.WhoItem, deletedWho *sync_to_target.WhoItem)) *mockMaskingDataCatalogRepository_UpdateAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*BQMaskingInformation), args[2].(*sync_to_target.WhoItem), args[3].(*sync_to_target.WhoItem))
	})
	return _c
}

func (_c *mockMaskingDataCatalogRepository_UpdateAccess_Call) Return(_a0 error) *mockMaskingDataCatalogRepository_UpdateAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockMaskingDataCatalogRepository_UpdateAccess_Call) RunAndReturn(run func(context.Context, *BQMaskingInformation, *sync_to_target.WhoItem, *sync_to_target.WhoItem) error) *mockMaskingDataCatalogRepository_UpdateAccess_Call {
	_c.Call.Return(run)
	return _c
}

// UpdatePolicyTag provides a mock function with given fields: ctx, location, maskingType, ap, dataPolicyId
func (_m *mockMaskingDataCatalogRepository) UpdatePolicyTag(ctx context.Context, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *sync_to_target.AccessProvider, dataPolicyId string) (*BQMaskingInformation, error) {
	ret := _m.Called(ctx, location, maskingType, ap, dataPolicyId)

	if len(ret) == 0 {
		panic("no return value specified for UpdatePolicyTag")
	}

	var r0 *BQMaskingInformation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, datapoliciespb.DataMaskingPolicy_PredefinedExpression, *sync_to_target.AccessProvider, string) (*BQMaskingInformation, error)); ok {
		return rf(ctx, location, maskingType, ap, dataPolicyId)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, datapoliciespb.DataMaskingPolicy_PredefinedExpression, *sync_to_target.AccessProvider, string) *BQMaskingInformation); ok {
		r0 = rf(ctx, location, maskingType, ap, dataPolicyId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*BQMaskingInformation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, datapoliciespb.DataMaskingPolicy_PredefinedExpression, *sync_to_target.AccessProvider, string) error); ok {
		r1 = rf(ctx, location, maskingType, ap, dataPolicyId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMaskingDataCatalogRepository_UpdatePolicyTag_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdatePolicyTag'
type mockMaskingDataCatalogRepository_UpdatePolicyTag_Call struct {
	*mock.Call
}

// UpdatePolicyTag is a helper method to define mock.On call
//   - ctx context.Context
//   - location string
//   - maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression
//   - ap *sync_to_target.AccessProvider
//   - dataPolicyId string
func (_e *mockMaskingDataCatalogRepository_Expecter) UpdatePolicyTag(ctx interface{}, location interface{}, maskingType interface{}, ap interface{}, dataPolicyId interface{}) *mockMaskingDataCatalogRepository_UpdatePolicyTag_Call {
	return &mockMaskingDataCatalogRepository_UpdatePolicyTag_Call{Call: _e.mock.On("UpdatePolicyTag", ctx, location, maskingType, ap, dataPolicyId)}
}

func (_c *mockMaskingDataCatalogRepository_UpdatePolicyTag_Call) Run(run func(ctx context.Context, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *sync_to_target.AccessProvider, dataPolicyId string)) *mockMaskingDataCatalogRepository_UpdatePolicyTag_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(datapoliciespb.DataMaskingPolicy_PredefinedExpression), args[3].(*sync_to_target.AccessProvider), args[4].(string))
	})
	return _c
}

func (_c *mockMaskingDataCatalogRepository_UpdatePolicyTag_Call) Return(_a0 *BQMaskingInformation, _a1 error) *mockMaskingDataCatalogRepository_UpdatePolicyTag_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMaskingDataCatalogRepository_UpdatePolicyTag_Call) RunAndReturn(run func(context.Context, string, datapoliciespb.DataMaskingPolicy_PredefinedExpression, *sync_to_target.AccessProvider, string) (*BQMaskingInformation, error)) *mockMaskingDataCatalogRepository_UpdatePolicyTag_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateWhatOfDataPolicy provides a mock function with given fields: ctx, policy, dataObjects, deletedDataObjects
func (_m *mockMaskingDataCatalogRepository) UpdateWhatOfDataPolicy(ctx context.Context, policy *BQMaskingInformation, dataObjects []string, deletedDataObjects []string) error {
	ret := _m.Called(ctx, policy, dataObjects, deletedDataObjects)

	if len(ret) == 0 {
		panic("no return value specified for UpdateWhatOfDataPolicy")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *BQMaskingInformation, []string, []string) error); ok {
		r0 = rf(ctx, policy, dataObjects, deletedDataObjects)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateWhatOfDataPolicy'
type mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call struct {
	*mock.Call
}

// UpdateWhatOfDataPolicy is a helper method to define mock.On call
//   - ctx context.Context
//   - policy *BQMaskingInformation
//   - dataObjects []string
//   - deletedDataObjects []string
func (_e *mockMaskingDataCatalogRepository_Expecter) UpdateWhatOfDataPolicy(ctx interface{}, policy interface{}, dataObjects interface{}, deletedDataObjects interface{}) *mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call {
	return &mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call{Call: _e.mock.On("UpdateWhatOfDataPolicy", ctx, policy, dataObjects, deletedDataObjects)}
}

func (_c *mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call) Run(run func(ctx context.Context, policy *BQMaskingInformation, dataObjects []string, deletedDataObjects []string)) *mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*BQMaskingInformation), args[2].([]string), args[3].([]string))
	})
	return _c
}

func (_c *mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call) Return(_a0 error) *mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call) RunAndReturn(run func(context.Context, *BQMaskingInformation, []string, []string) error) *mockMaskingDataCatalogRepository_UpdateWhatOfDataPolicy_Call {
	_c.Call.Return(run)
	return _c
}

// newMockMaskingDataCatalogRepository creates a new instance of mockMaskingDataCatalogRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockMaskingDataCatalogRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockMaskingDataCatalogRepository {
	mock := &mockMaskingDataCatalogRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
