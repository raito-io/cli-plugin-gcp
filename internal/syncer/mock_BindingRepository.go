// Code generated by mockery v2.36.1. DO NOT EDIT.

package syncer

import (
	context "context"

	iam "github.com/raito-io/cli-plugin-gcp/internal/iam"
	mock "github.com/stretchr/testify/mock"

	org "github.com/raito-io/cli-plugin-gcp/internal/org"
)

// MockBindingRepository is an autogenerated mock type for the BindingRepository type
type MockBindingRepository struct {
	mock.Mock
}

type MockBindingRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockBindingRepository) EXPECT() *MockBindingRepository_Expecter {
	return &MockBindingRepository_Expecter{mock: &_m.Mock}
}

// Bindings provides a mock function with given fields: ctx, fn
func (_m *MockBindingRepository) Bindings(ctx context.Context, fn func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
	ret := _m.Called(ctx, fn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error); ok {
		r0 = rf(ctx, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockBindingRepository_Bindings_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bindings'
type MockBindingRepository_Bindings_Call struct {
	*mock.Call
}

// Bindings is a helper method to define mock.On call
//   - ctx context.Context
//   - fn func(context.Context , *org.GcpOrgEntity , []iam.IamBinding) error
func (_e *MockBindingRepository_Expecter) Bindings(ctx interface{}, fn interface{}) *MockBindingRepository_Bindings_Call {
	return &MockBindingRepository_Bindings_Call{Call: _e.mock.On("Bindings", ctx, fn)}
}

func (_c *MockBindingRepository_Bindings_Call) Run(run func(ctx context.Context, fn func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error)) *MockBindingRepository_Bindings_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error))
	})
	return _c
}

func (_c *MockBindingRepository_Bindings_Call) Return(_a0 error) *MockBindingRepository_Bindings_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockBindingRepository_Bindings_Call) RunAndReturn(run func(context.Context, func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error) *MockBindingRepository_Bindings_Call {
	_c.Call.Return(run)
	return _c
}

// DataSourceType provides a mock function with given fields:
func (_m *MockBindingRepository) DataSourceType() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockBindingRepository_DataSourceType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DataSourceType'
type MockBindingRepository_DataSourceType_Call struct {
	*mock.Call
}

// DataSourceType is a helper method to define mock.On call
func (_e *MockBindingRepository_Expecter) DataSourceType() *MockBindingRepository_DataSourceType_Call {
	return &MockBindingRepository_DataSourceType_Call{Call: _e.mock.On("DataSourceType")}
}

func (_c *MockBindingRepository_DataSourceType_Call) Run(run func()) *MockBindingRepository_DataSourceType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockBindingRepository_DataSourceType_Call) Return(_a0 string) *MockBindingRepository_DataSourceType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockBindingRepository_DataSourceType_Call) RunAndReturn(run func() string) *MockBindingRepository_DataSourceType_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateBindings provides a mock function with given fields: ctx, dataObject, addBindings, removeBindings
func (_m *MockBindingRepository) UpdateBindings(ctx context.Context, dataObject *iam.DataObjectReference, addBindings []iam.IamBinding, removeBindings []iam.IamBinding) error {
	ret := _m.Called(ctx, dataObject, addBindings, removeBindings)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *iam.DataObjectReference, []iam.IamBinding, []iam.IamBinding) error); ok {
		r0 = rf(ctx, dataObject, addBindings, removeBindings)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockBindingRepository_UpdateBindings_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateBindings'
type MockBindingRepository_UpdateBindings_Call struct {
	*mock.Call
}

// UpdateBindings is a helper method to define mock.On call
//   - ctx context.Context
//   - dataObject *iam.DataObjectReference
//   - addBindings []iam.IamBinding
//   - removeBindings []iam.IamBinding
func (_e *MockBindingRepository_Expecter) UpdateBindings(ctx interface{}, dataObject interface{}, addBindings interface{}, removeBindings interface{}) *MockBindingRepository_UpdateBindings_Call {
	return &MockBindingRepository_UpdateBindings_Call{Call: _e.mock.On("UpdateBindings", ctx, dataObject, addBindings, removeBindings)}
}

func (_c *MockBindingRepository_UpdateBindings_Call) Run(run func(ctx context.Context, dataObject *iam.DataObjectReference, addBindings []iam.IamBinding, removeBindings []iam.IamBinding)) *MockBindingRepository_UpdateBindings_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*iam.DataObjectReference), args[2].([]iam.IamBinding), args[3].([]iam.IamBinding))
	})
	return _c
}

func (_c *MockBindingRepository_UpdateBindings_Call) Return(_a0 error) *MockBindingRepository_UpdateBindings_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockBindingRepository_UpdateBindings_Call) RunAndReturn(run func(context.Context, *iam.DataObjectReference, []iam.IamBinding, []iam.IamBinding) error) *MockBindingRepository_UpdateBindings_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockBindingRepository creates a new instance of MockBindingRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockBindingRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBindingRepository {
	mock := &MockBindingRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
