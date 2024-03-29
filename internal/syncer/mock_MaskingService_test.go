// Code generated by mockery v2.40.1. DO NOT EDIT.

package syncer

import (
	context "context"

	iam "github.com/raito-io/cli-plugin-gcp/internal/iam"
	mock "github.com/stretchr/testify/mock"

	set "github.com/raito-io/golang-set/set"

	sync_to_target "github.com/raito-io/cli/base/access_provider/sync_to_target"

	wrappers "github.com/raito-io/cli/base/wrappers"
)

// MockMaskingService is an autogenerated mock type for the MaskingService type
type MockMaskingService struct {
	mock.Mock
}

type MockMaskingService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockMaskingService) EXPECT() *MockMaskingService_Expecter {
	return &MockMaskingService_Expecter{mock: &_m.Mock}
}

// ExportMasks provides a mock function with given fields: ctx, accessProvider, accessProviderFeedbackHandler
func (_m *MockMaskingService) ExportMasks(ctx context.Context, accessProvider *sync_to_target.AccessProvider, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler) ([]string, error) {
	ret := _m.Called(ctx, accessProvider, accessProviderFeedbackHandler)

	if len(ret) == 0 {
		panic("no return value specified for ExportMasks")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *sync_to_target.AccessProvider, wrappers.AccessProviderFeedbackHandler) ([]string, error)); ok {
		return rf(ctx, accessProvider, accessProviderFeedbackHandler)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *sync_to_target.AccessProvider, wrappers.AccessProviderFeedbackHandler) []string); ok {
		r0 = rf(ctx, accessProvider, accessProviderFeedbackHandler)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *sync_to_target.AccessProvider, wrappers.AccessProviderFeedbackHandler) error); ok {
		r1 = rf(ctx, accessProvider, accessProviderFeedbackHandler)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMaskingService_ExportMasks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExportMasks'
type MockMaskingService_ExportMasks_Call struct {
	*mock.Call
}

// ExportMasks is a helper method to define mock.On call
//   - ctx context.Context
//   - accessProvider *sync_to_target.AccessProvider
//   - accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler
func (_e *MockMaskingService_Expecter) ExportMasks(ctx interface{}, accessProvider interface{}, accessProviderFeedbackHandler interface{}) *MockMaskingService_ExportMasks_Call {
	return &MockMaskingService_ExportMasks_Call{Call: _e.mock.On("ExportMasks", ctx, accessProvider, accessProviderFeedbackHandler)}
}

func (_c *MockMaskingService_ExportMasks_Call) Run(run func(ctx context.Context, accessProvider *sync_to_target.AccessProvider, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler)) *MockMaskingService_ExportMasks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*sync_to_target.AccessProvider), args[2].(wrappers.AccessProviderFeedbackHandler))
	})
	return _c
}

func (_c *MockMaskingService_ExportMasks_Call) Return(_a0 []string, _a1 error) *MockMaskingService_ExportMasks_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaskingService_ExportMasks_Call) RunAndReturn(run func(context.Context, *sync_to_target.AccessProvider, wrappers.AccessProviderFeedbackHandler) ([]string, error)) *MockMaskingService_ExportMasks_Call {
	_c.Call.Return(run)
	return _c
}

// ImportMasks provides a mock function with given fields: ctx, accessProviderHandler, locations, maskingTags, raitoMasks
func (_m *MockMaskingService) ImportMasks(ctx context.Context, accessProviderHandler wrappers.AccessProviderHandler, locations set.Set[string], maskingTags map[string][]string, raitoMasks set.Set[string]) error {
	ret := _m.Called(ctx, accessProviderHandler, locations, maskingTags, raitoMasks)

	if len(ret) == 0 {
		panic("no return value specified for ImportMasks")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, wrappers.AccessProviderHandler, set.Set[string], map[string][]string, set.Set[string]) error); ok {
		r0 = rf(ctx, accessProviderHandler, locations, maskingTags, raitoMasks)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockMaskingService_ImportMasks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ImportMasks'
type MockMaskingService_ImportMasks_Call struct {
	*mock.Call
}

// ImportMasks is a helper method to define mock.On call
//   - ctx context.Context
//   - accessProviderHandler wrappers.AccessProviderHandler
//   - locations set.Set[string]
//   - maskingTags map[string][]string
//   - raitoMasks set.Set[string]
func (_e *MockMaskingService_Expecter) ImportMasks(ctx interface{}, accessProviderHandler interface{}, locations interface{}, maskingTags interface{}, raitoMasks interface{}) *MockMaskingService_ImportMasks_Call {
	return &MockMaskingService_ImportMasks_Call{Call: _e.mock.On("ImportMasks", ctx, accessProviderHandler, locations, maskingTags, raitoMasks)}
}

func (_c *MockMaskingService_ImportMasks_Call) Run(run func(ctx context.Context, accessProviderHandler wrappers.AccessProviderHandler, locations set.Set[string], maskingTags map[string][]string, raitoMasks set.Set[string])) *MockMaskingService_ImportMasks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(wrappers.AccessProviderHandler), args[2].(set.Set[string]), args[3].(map[string][]string), args[4].(set.Set[string]))
	})
	return _c
}

func (_c *MockMaskingService_ImportMasks_Call) Return(_a0 error) *MockMaskingService_ImportMasks_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMaskingService_ImportMasks_Call) RunAndReturn(run func(context.Context, wrappers.AccessProviderHandler, set.Set[string], map[string][]string, set.Set[string]) error) *MockMaskingService_ImportMasks_Call {
	_c.Call.Return(run)
	return _c
}

// MaskedBinding provides a mock function with given fields: ctx, members
func (_m *MockMaskingService) MaskedBinding(ctx context.Context, members []string) ([]iam.IamBinding, error) {
	ret := _m.Called(ctx, members)

	if len(ret) == 0 {
		panic("no return value specified for MaskedBinding")
	}

	var r0 []iam.IamBinding
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]iam.IamBinding, error)); ok {
		return rf(ctx, members)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []iam.IamBinding); ok {
		r0 = rf(ctx, members)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iam.IamBinding)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, members)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockMaskingService_MaskedBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MaskedBinding'
type MockMaskingService_MaskedBinding_Call struct {
	*mock.Call
}

// MaskedBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - members []string
func (_e *MockMaskingService_Expecter) MaskedBinding(ctx interface{}, members interface{}) *MockMaskingService_MaskedBinding_Call {
	return &MockMaskingService_MaskedBinding_Call{Call: _e.mock.On("MaskedBinding", ctx, members)}
}

func (_c *MockMaskingService_MaskedBinding_Call) Run(run func(ctx context.Context, members []string)) *MockMaskingService_MaskedBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]string))
	})
	return _c
}

func (_c *MockMaskingService_MaskedBinding_Call) Return(_a0 []iam.IamBinding, _a1 error) *MockMaskingService_MaskedBinding_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockMaskingService_MaskedBinding_Call) RunAndReturn(run func(context.Context, []string) ([]iam.IamBinding, error)) *MockMaskingService_MaskedBinding_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockMaskingService creates a new instance of MockMaskingService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMaskingService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMaskingService {
	mock := &MockMaskingService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
