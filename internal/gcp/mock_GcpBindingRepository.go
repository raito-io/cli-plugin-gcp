// Code generated by mockery v2.36.1. DO NOT EDIT.

package gcp

import (
	context "context"

	org "github.com/raito-io/cli-plugin-gcp/internal/org"
	mock "github.com/stretchr/testify/mock"

	types "github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

// MockGcpBindingRepository is an autogenerated mock type for the GcpBindingRepository type
type MockGcpBindingRepository struct {
	mock.Mock
}

type MockGcpBindingRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockGcpBindingRepository) EXPECT() *MockGcpBindingRepository_Expecter {
	return &MockGcpBindingRepository_Expecter{mock: &_m.Mock}
}

// AddBinding provides a mock function with given fields: ctx, binding
func (_m *MockGcpBindingRepository) AddBinding(ctx context.Context, binding types.IamBinding) error {
	ret := _m.Called(ctx, binding)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.IamBinding) error); ok {
		r0 = rf(ctx, binding)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockGcpBindingRepository_AddBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddBinding'
type MockGcpBindingRepository_AddBinding_Call struct {
	*mock.Call
}

// AddBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - binding types.IamBinding
func (_e *MockGcpBindingRepository_Expecter) AddBinding(ctx interface{}, binding interface{}) *MockGcpBindingRepository_AddBinding_Call {
	return &MockGcpBindingRepository_AddBinding_Call{Call: _e.mock.On("AddBinding", ctx, binding)}
}

func (_c *MockGcpBindingRepository_AddBinding_Call) Run(run func(ctx context.Context, binding types.IamBinding)) *MockGcpBindingRepository_AddBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(types.IamBinding))
	})
	return _c
}

func (_c *MockGcpBindingRepository_AddBinding_Call) Return(_a0 error) *MockGcpBindingRepository_AddBinding_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockGcpBindingRepository_AddBinding_Call) RunAndReturn(run func(context.Context, types.IamBinding) error) *MockGcpBindingRepository_AddBinding_Call {
	_c.Call.Return(run)
	return _c
}

// Bindings provides a mock function with given fields: ctx, fn
func (_m *MockGcpBindingRepository) Bindings(ctx context.Context, fn func(context.Context, *org.GcpOrgEntity, []types.IamBinding) error) error {
	ret := _m.Called(ctx, fn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, func(context.Context, *org.GcpOrgEntity, []types.IamBinding) error) error); ok {
		r0 = rf(ctx, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockGcpBindingRepository_Bindings_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bindings'
type MockGcpBindingRepository_Bindings_Call struct {
	*mock.Call
}

// Bindings is a helper method to define mock.On call
//   - ctx context.Context
//   - fn func(context.Context , *org.GcpOrgEntity , []types.IamBinding) error
func (_e *MockGcpBindingRepository_Expecter) Bindings(ctx interface{}, fn interface{}) *MockGcpBindingRepository_Bindings_Call {
	return &MockGcpBindingRepository_Bindings_Call{Call: _e.mock.On("Bindings", ctx, fn)}
}

func (_c *MockGcpBindingRepository_Bindings_Call) Run(run func(ctx context.Context, fn func(context.Context, *org.GcpOrgEntity, []types.IamBinding) error)) *MockGcpBindingRepository_Bindings_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(func(context.Context, *org.GcpOrgEntity, []types.IamBinding) error))
	})
	return _c
}

func (_c *MockGcpBindingRepository_Bindings_Call) Return(_a0 error) *MockGcpBindingRepository_Bindings_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockGcpBindingRepository_Bindings_Call) RunAndReturn(run func(context.Context, func(context.Context, *org.GcpOrgEntity, []types.IamBinding) error) error) *MockGcpBindingRepository_Bindings_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveBinding provides a mock function with given fields: ctx, binding
func (_m *MockGcpBindingRepository) RemoveBinding(ctx context.Context, binding types.IamBinding) error {
	ret := _m.Called(ctx, binding)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.IamBinding) error); ok {
		r0 = rf(ctx, binding)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockGcpBindingRepository_RemoveBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveBinding'
type MockGcpBindingRepository_RemoveBinding_Call struct {
	*mock.Call
}

// RemoveBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - binding types.IamBinding
func (_e *MockGcpBindingRepository_Expecter) RemoveBinding(ctx interface{}, binding interface{}) *MockGcpBindingRepository_RemoveBinding_Call {
	return &MockGcpBindingRepository_RemoveBinding_Call{Call: _e.mock.On("RemoveBinding", ctx, binding)}
}

func (_c *MockGcpBindingRepository_RemoveBinding_Call) Run(run func(ctx context.Context, binding types.IamBinding)) *MockGcpBindingRepository_RemoveBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(types.IamBinding))
	})
	return _c
}

func (_c *MockGcpBindingRepository_RemoveBinding_Call) Return(_a0 error) *MockGcpBindingRepository_RemoveBinding_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockGcpBindingRepository_RemoveBinding_Call) RunAndReturn(run func(context.Context, types.IamBinding) error) *MockGcpBindingRepository_RemoveBinding_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockGcpBindingRepository creates a new instance of MockGcpBindingRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockGcpBindingRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGcpBindingRepository {
	mock := &MockGcpBindingRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
