// Code generated by mockery v2.36.1. DO NOT EDIT.

package org

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/raito-io/cli-plugin-gcp/internal/iam"
)

// mockOrganizationRepo is an autogenerated mock type for the organizationRepo type
type mockOrganizationRepo struct {
	mock.Mock
}

type mockOrganizationRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *mockOrganizationRepo) EXPECT() *mockOrganizationRepo_Expecter {
	return &mockOrganizationRepo_Expecter{mock: &_m.Mock}
}

// AddBinding provides a mock function with given fields: ctx, binding
func (_m *mockOrganizationRepo) AddBinding(ctx context.Context, binding types.IamBinding) error {
	ret := _m.Called(ctx, binding)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.IamBinding) error); ok {
		r0 = rf(ctx, binding)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockOrganizationRepo_AddBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddBinding'
type mockOrganizationRepo_AddBinding_Call struct {
	*mock.Call
}

// AddBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - binding types.IamBinding
func (_e *mockOrganizationRepo_Expecter) AddBinding(ctx interface{}, binding interface{}) *mockOrganizationRepo_AddBinding_Call {
	return &mockOrganizationRepo_AddBinding_Call{Call: _e.mock.On("AddBinding", ctx, binding)}
}

func (_c *mockOrganizationRepo_AddBinding_Call) Run(run func(ctx context.Context, binding types.IamBinding)) *mockOrganizationRepo_AddBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(types.IamBinding))
	})
	return _c
}

func (_c *mockOrganizationRepo_AddBinding_Call) Return(_a0 error) *mockOrganizationRepo_AddBinding_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockOrganizationRepo_AddBinding_Call) RunAndReturn(run func(context.Context, types.IamBinding) error) *mockOrganizationRepo_AddBinding_Call {
	_c.Call.Return(run)
	return _c
}

// GetIamPolicy provides a mock function with given fields: ctx, projectId
func (_m *mockOrganizationRepo) GetIamPolicy(ctx context.Context, projectId string) ([]types.IamBinding, error) {
	ret := _m.Called(ctx, projectId)

	var r0 []types.IamBinding
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]types.IamBinding, error)); ok {
		return rf(ctx, projectId)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []types.IamBinding); ok {
		r0 = rf(ctx, projectId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.IamBinding)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, projectId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockOrganizationRepo_GetIamPolicy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIamPolicy'
type mockOrganizationRepo_GetIamPolicy_Call struct {
	*mock.Call
}

// GetIamPolicy is a helper method to define mock.On call
//   - ctx context.Context
//   - projectId string
func (_e *mockOrganizationRepo_Expecter) GetIamPolicy(ctx interface{}, projectId interface{}) *mockOrganizationRepo_GetIamPolicy_Call {
	return &mockOrganizationRepo_GetIamPolicy_Call{Call: _e.mock.On("GetIamPolicy", ctx, projectId)}
}

func (_c *mockOrganizationRepo_GetIamPolicy_Call) Run(run func(ctx context.Context, projectId string)) *mockOrganizationRepo_GetIamPolicy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *mockOrganizationRepo_GetIamPolicy_Call) Return(_a0 []types.IamBinding, _a1 error) *mockOrganizationRepo_GetIamPolicy_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockOrganizationRepo_GetIamPolicy_Call) RunAndReturn(run func(context.Context, string) ([]types.IamBinding, error)) *mockOrganizationRepo_GetIamPolicy_Call {
	_c.Call.Return(run)
	return _c
}

// GetOrganization provides a mock function with given fields: ctx
func (_m *mockOrganizationRepo) GetOrganization(ctx context.Context) (*GcpOrgEntity, error) {
	ret := _m.Called(ctx)

	var r0 *GcpOrgEntity
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*GcpOrgEntity, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *GcpOrgEntity); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*GcpOrgEntity)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockOrganizationRepo_GetOrganization_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOrganization'
type mockOrganizationRepo_GetOrganization_Call struct {
	*mock.Call
}

// GetOrganization is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockOrganizationRepo_Expecter) GetOrganization(ctx interface{}) *mockOrganizationRepo_GetOrganization_Call {
	return &mockOrganizationRepo_GetOrganization_Call{Call: _e.mock.On("GetOrganization", ctx)}
}

func (_c *mockOrganizationRepo_GetOrganization_Call) Run(run func(ctx context.Context)) *mockOrganizationRepo_GetOrganization_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockOrganizationRepo_GetOrganization_Call) Return(_a0 *GcpOrgEntity, _a1 error) *mockOrganizationRepo_GetOrganization_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockOrganizationRepo_GetOrganization_Call) RunAndReturn(run func(context.Context) (*GcpOrgEntity, error)) *mockOrganizationRepo_GetOrganization_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveBinding provides a mock function with given fields: ctx, binding
func (_m *mockOrganizationRepo) RemoveBinding(ctx context.Context, binding types.IamBinding) error {
	ret := _m.Called(ctx, binding)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.IamBinding) error); ok {
		r0 = rf(ctx, binding)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockOrganizationRepo_RemoveBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveBinding'
type mockOrganizationRepo_RemoveBinding_Call struct {
	*mock.Call
}

// RemoveBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - binding types.IamBinding
func (_e *mockOrganizationRepo_Expecter) RemoveBinding(ctx interface{}, binding interface{}) *mockOrganizationRepo_RemoveBinding_Call {
	return &mockOrganizationRepo_RemoveBinding_Call{Call: _e.mock.On("RemoveBinding", ctx, binding)}
}

func (_c *mockOrganizationRepo_RemoveBinding_Call) Run(run func(ctx context.Context, binding types.IamBinding)) *mockOrganizationRepo_RemoveBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(types.IamBinding))
	})
	return _c
}

func (_c *mockOrganizationRepo_RemoveBinding_Call) Return(_a0 error) *mockOrganizationRepo_RemoveBinding_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockOrganizationRepo_RemoveBinding_Call) RunAndReturn(run func(context.Context, types.IamBinding) error) *mockOrganizationRepo_RemoveBinding_Call {
	_c.Call.Return(run)
	return _c
}

// newMockOrganizationRepo creates a new instance of mockOrganizationRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockOrganizationRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockOrganizationRepo {
	mock := &mockOrganizationRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
