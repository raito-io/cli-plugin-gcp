// Code generated by mockery v2.40.1. DO NOT EDIT.

package syncer

import (
	context "context"

	data_source "github.com/raito-io/cli/base/data_source"
	mock "github.com/stretchr/testify/mock"

	set "github.com/raito-io/golang-set/set"

	sync_to_target "github.com/raito-io/cli/base/access_provider/sync_to_target"

	wrappers "github.com/raito-io/cli/base/wrappers"
)

// MockFilteringService is an autogenerated mock type for the FilteringService type
type MockFilteringService struct {
	mock.Mock
}

type MockFilteringService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockFilteringService) EXPECT() *MockFilteringService_Expecter {
	return &MockFilteringService_Expecter{mock: &_m.Mock}
}

// ExportFilter provides a mock function with given fields: ctx, accessProvider, accessProviderFeedbackHandler
func (_m *MockFilteringService) ExportFilter(ctx context.Context, accessProvider *sync_to_target.AccessProvider, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler) (*string, error) {
	ret := _m.Called(ctx, accessProvider, accessProviderFeedbackHandler)

	if len(ret) == 0 {
		panic("no return value specified for ExportFilter")
	}

	var r0 *string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *sync_to_target.AccessProvider, wrappers.AccessProviderFeedbackHandler) (*string, error)); ok {
		return rf(ctx, accessProvider, accessProviderFeedbackHandler)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *sync_to_target.AccessProvider, wrappers.AccessProviderFeedbackHandler) *string); ok {
		r0 = rf(ctx, accessProvider, accessProviderFeedbackHandler)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *sync_to_target.AccessProvider, wrappers.AccessProviderFeedbackHandler) error); ok {
		r1 = rf(ctx, accessProvider, accessProviderFeedbackHandler)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockFilteringService_ExportFilter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExportFilter'
type MockFilteringService_ExportFilter_Call struct {
	*mock.Call
}

// ExportFilter is a helper method to define mock.On call
//   - ctx context.Context
//   - accessProvider *sync_to_target.AccessProvider
//   - accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler
func (_e *MockFilteringService_Expecter) ExportFilter(ctx interface{}, accessProvider interface{}, accessProviderFeedbackHandler interface{}) *MockFilteringService_ExportFilter_Call {
	return &MockFilteringService_ExportFilter_Call{Call: _e.mock.On("ExportFilter", ctx, accessProvider, accessProviderFeedbackHandler)}
}

func (_c *MockFilteringService_ExportFilter_Call) Run(run func(ctx context.Context, accessProvider *sync_to_target.AccessProvider, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler)) *MockFilteringService_ExportFilter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*sync_to_target.AccessProvider), args[2].(wrappers.AccessProviderFeedbackHandler))
	})
	return _c
}

func (_c *MockFilteringService_ExportFilter_Call) Return(_a0 *string, _a1 error) *MockFilteringService_ExportFilter_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockFilteringService_ExportFilter_Call) RunAndReturn(run func(context.Context, *sync_to_target.AccessProvider, wrappers.AccessProviderFeedbackHandler) (*string, error)) *MockFilteringService_ExportFilter_Call {
	_c.Call.Return(run)
	return _c
}

// ImportFilters provides a mock function with given fields: ctx, config, accessProviderHandler, raitoFilters
func (_m *MockFilteringService) ImportFilters(ctx context.Context, config *data_source.DataSourceSyncConfig, accessProviderHandler wrappers.AccessProviderHandler, raitoFilters set.Set[string]) error {
	ret := _m.Called(ctx, config, accessProviderHandler, raitoFilters)

	if len(ret) == 0 {
		panic("no return value specified for ImportFilters")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *data_source.DataSourceSyncConfig, wrappers.AccessProviderHandler, set.Set[string]) error); ok {
		r0 = rf(ctx, config, accessProviderHandler, raitoFilters)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockFilteringService_ImportFilters_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ImportFilters'
type MockFilteringService_ImportFilters_Call struct {
	*mock.Call
}

// ImportFilters is a helper method to define mock.On call
//   - ctx context.Context
//   - config *data_source.DataSourceSyncConfig
//   - accessProviderHandler wrappers.AccessProviderHandler
//   - raitoFilters set.Set[string]
func (_e *MockFilteringService_Expecter) ImportFilters(ctx interface{}, config interface{}, accessProviderHandler interface{}, raitoFilters interface{}) *MockFilteringService_ImportFilters_Call {
	return &MockFilteringService_ImportFilters_Call{Call: _e.mock.On("ImportFilters", ctx, config, accessProviderHandler, raitoFilters)}
}

func (_c *MockFilteringService_ImportFilters_Call) Run(run func(ctx context.Context, config *data_source.DataSourceSyncConfig, accessProviderHandler wrappers.AccessProviderHandler, raitoFilters set.Set[string])) *MockFilteringService_ImportFilters_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*data_source.DataSourceSyncConfig), args[2].(wrappers.AccessProviderHandler), args[3].(set.Set[string]))
	})
	return _c
}

func (_c *MockFilteringService_ImportFilters_Call) Return(_a0 error) *MockFilteringService_ImportFilters_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFilteringService_ImportFilters_Call) RunAndReturn(run func(context.Context, *data_source.DataSourceSyncConfig, wrappers.AccessProviderHandler, set.Set[string]) error) *MockFilteringService_ImportFilters_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockFilteringService creates a new instance of MockFilteringService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockFilteringService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockFilteringService {
	mock := &MockFilteringService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
