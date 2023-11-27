// Code generated by mockery v2.37.1. DO NOT EDIT.

package org

import (
	context "context"

	iam "github.com/raito-io/cli-plugin-gcp/internal/iam"
	data_source "github.com/raito-io/cli/base/data_source"

	mock "github.com/stretchr/testify/mock"
)

// mockProjectRepo is an autogenerated mock type for the projectRepo type
type mockProjectRepo struct {
	mock.Mock
}

type mockProjectRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *mockProjectRepo) EXPECT() *mockProjectRepo_Expecter {
	return &mockProjectRepo_Expecter{mock: &_m.Mock}
}

// GetIamPolicy provides a mock function with given fields: ctx, projectId
func (_m *mockProjectRepo) GetIamPolicy(ctx context.Context, projectId string) ([]iam.IamBinding, error) {
	ret := _m.Called(ctx, projectId)

	var r0 []iam.IamBinding
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]iam.IamBinding, error)); ok {
		return rf(ctx, projectId)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []iam.IamBinding); ok {
		r0 = rf(ctx, projectId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iam.IamBinding)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, projectId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockProjectRepo_GetIamPolicy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIamPolicy'
type mockProjectRepo_GetIamPolicy_Call struct {
	*mock.Call
}

// GetIamPolicy is a helper method to define mock.On call
//   - ctx context.Context
//   - projectId string
func (_e *mockProjectRepo_Expecter) GetIamPolicy(ctx interface{}, projectId interface{}) *mockProjectRepo_GetIamPolicy_Call {
	return &mockProjectRepo_GetIamPolicy_Call{Call: _e.mock.On("GetIamPolicy", ctx, projectId)}
}

func (_c *mockProjectRepo_GetIamPolicy_Call) Run(run func(ctx context.Context, projectId string)) *mockProjectRepo_GetIamPolicy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *mockProjectRepo_GetIamPolicy_Call) Return(_a0 []iam.IamBinding, _a1 error) *mockProjectRepo_GetIamPolicy_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockProjectRepo_GetIamPolicy_Call) RunAndReturn(run func(context.Context, string) ([]iam.IamBinding, error)) *mockProjectRepo_GetIamPolicy_Call {
	_c.Call.Return(run)
	return _c
}

// GetProjects provides a mock function with given fields: ctx, config, parentName, parent, fn
func (_m *mockProjectRepo) GetProjects(ctx context.Context, config *data_source.DataSourceSyncConfig, parentName string, parent *GcpOrgEntity, fn func(context.Context, *GcpOrgEntity) error) error {
	ret := _m.Called(ctx, config, parentName, parent, fn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *data_source.DataSourceSyncConfig, string, *GcpOrgEntity, func(context.Context, *GcpOrgEntity) error) error); ok {
		r0 = rf(ctx, config, parentName, parent, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockProjectRepo_GetProjects_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProjects'
type mockProjectRepo_GetProjects_Call struct {
	*mock.Call
}

// GetProjects is a helper method to define mock.On call
//   - ctx context.Context
//   - config *data_source.DataSourceSyncConfig
//   - parentName string
//   - parent *GcpOrgEntity
//   - fn func(context.Context , *GcpOrgEntity) error
func (_e *mockProjectRepo_Expecter) GetProjects(ctx interface{}, config interface{}, parentName interface{}, parent interface{}, fn interface{}) *mockProjectRepo_GetProjects_Call {
	return &mockProjectRepo_GetProjects_Call{Call: _e.mock.On("GetProjects", ctx, config, parentName, parent, fn)}
}

func (_c *mockProjectRepo_GetProjects_Call) Run(run func(ctx context.Context, config *data_source.DataSourceSyncConfig, parentName string, parent *GcpOrgEntity, fn func(context.Context, *GcpOrgEntity) error)) *mockProjectRepo_GetProjects_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*data_source.DataSourceSyncConfig), args[2].(string), args[3].(*GcpOrgEntity), args[4].(func(context.Context, *GcpOrgEntity) error))
	})
	return _c
}

func (_c *mockProjectRepo_GetProjects_Call) Return(_a0 error) *mockProjectRepo_GetProjects_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockProjectRepo_GetProjects_Call) RunAndReturn(run func(context.Context, *data_source.DataSourceSyncConfig, string, *GcpOrgEntity, func(context.Context, *GcpOrgEntity) error) error) *mockProjectRepo_GetProjects_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateBinding provides a mock function with given fields: ctx, dataObject, bindingsToAdd, bindingsToDelete
func (_m *mockProjectRepo) UpdateBinding(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding) error {
	ret := _m.Called(ctx, dataObject, bindingsToAdd, bindingsToDelete)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *iam.DataObjectReference, []iam.IamBinding, []iam.IamBinding) error); ok {
		r0 = rf(ctx, dataObject, bindingsToAdd, bindingsToDelete)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockProjectRepo_UpdateBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateBinding'
type mockProjectRepo_UpdateBinding_Call struct {
	*mock.Call
}

// UpdateBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - dataObject *iam.DataObjectReference
//   - bindingsToAdd []iam.IamBinding
//   - bindingsToDelete []iam.IamBinding
func (_e *mockProjectRepo_Expecter) UpdateBinding(ctx interface{}, dataObject interface{}, bindingsToAdd interface{}, bindingsToDelete interface{}) *mockProjectRepo_UpdateBinding_Call {
	return &mockProjectRepo_UpdateBinding_Call{Call: _e.mock.On("UpdateBinding", ctx, dataObject, bindingsToAdd, bindingsToDelete)}
}

func (_c *mockProjectRepo_UpdateBinding_Call) Run(run func(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding)) *mockProjectRepo_UpdateBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*iam.DataObjectReference), args[2].([]iam.IamBinding), args[3].([]iam.IamBinding))
	})
	return _c
}

func (_c *mockProjectRepo_UpdateBinding_Call) Return(_a0 error) *mockProjectRepo_UpdateBinding_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockProjectRepo_UpdateBinding_Call) RunAndReturn(run func(context.Context, *iam.DataObjectReference, []iam.IamBinding, []iam.IamBinding) error) *mockProjectRepo_UpdateBinding_Call {
	_c.Call.Return(run)
	return _c
}

// newMockProjectRepo creates a new instance of mockProjectRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockProjectRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockProjectRepo {
	mock := &mockProjectRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}