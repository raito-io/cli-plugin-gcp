// Code generated by mockery v2.37.1. DO NOT EDIT.

package org

import (
	context "context"

	iam "github.com/raito-io/cli-plugin-gcp/internal/iam"
	mock "github.com/stretchr/testify/mock"
)

// mockFolderRepo is an autogenerated mock type for the folderRepo type
type mockFolderRepo struct {
	mock.Mock
}

type mockFolderRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *mockFolderRepo) EXPECT() *mockFolderRepo_Expecter {
	return &mockFolderRepo_Expecter{mock: &_m.Mock}
}

// GetFolders provides a mock function with given fields: ctx, parentName, parent, fn
func (_m *mockFolderRepo) GetFolders(ctx context.Context, parentName string, parent *GcpOrgEntity, fn func(context.Context, *GcpOrgEntity) error) error {
	ret := _m.Called(ctx, parentName, parent, fn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *GcpOrgEntity, func(context.Context, *GcpOrgEntity) error) error); ok {
		r0 = rf(ctx, parentName, parent, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockFolderRepo_GetFolders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFolders'
type mockFolderRepo_GetFolders_Call struct {
	*mock.Call
}

// GetFolders is a helper method to define mock.On call
//   - ctx context.Context
//   - parentName string
//   - parent *GcpOrgEntity
//   - fn func(context.Context , *GcpOrgEntity) error
func (_e *mockFolderRepo_Expecter) GetFolders(ctx interface{}, parentName interface{}, parent interface{}, fn interface{}) *mockFolderRepo_GetFolders_Call {
	return &mockFolderRepo_GetFolders_Call{Call: _e.mock.On("GetFolders", ctx, parentName, parent, fn)}
}

func (_c *mockFolderRepo_GetFolders_Call) Run(run func(ctx context.Context, parentName string, parent *GcpOrgEntity, fn func(context.Context, *GcpOrgEntity) error)) *mockFolderRepo_GetFolders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*GcpOrgEntity), args[3].(func(context.Context, *GcpOrgEntity) error))
	})
	return _c
}

func (_c *mockFolderRepo_GetFolders_Call) Return(_a0 error) *mockFolderRepo_GetFolders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFolderRepo_GetFolders_Call) RunAndReturn(run func(context.Context, string, *GcpOrgEntity, func(context.Context, *GcpOrgEntity) error) error) *mockFolderRepo_GetFolders_Call {
	_c.Call.Return(run)
	return _c
}

// GetIamPolicy provides a mock function with given fields: ctx, projectId
func (_m *mockFolderRepo) GetIamPolicy(ctx context.Context, projectId string) ([]iam.IamBinding, error) {
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

// mockFolderRepo_GetIamPolicy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIamPolicy'
type mockFolderRepo_GetIamPolicy_Call struct {
	*mock.Call
}

// GetIamPolicy is a helper method to define mock.On call
//   - ctx context.Context
//   - projectId string
func (_e *mockFolderRepo_Expecter) GetIamPolicy(ctx interface{}, projectId interface{}) *mockFolderRepo_GetIamPolicy_Call {
	return &mockFolderRepo_GetIamPolicy_Call{Call: _e.mock.On("GetIamPolicy", ctx, projectId)}
}

func (_c *mockFolderRepo_GetIamPolicy_Call) Run(run func(ctx context.Context, projectId string)) *mockFolderRepo_GetIamPolicy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *mockFolderRepo_GetIamPolicy_Call) Return(_a0 []iam.IamBinding, _a1 error) *mockFolderRepo_GetIamPolicy_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockFolderRepo_GetIamPolicy_Call) RunAndReturn(run func(context.Context, string) ([]iam.IamBinding, error)) *mockFolderRepo_GetIamPolicy_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateBinding provides a mock function with given fields: ctx, dataObject, bindingsToAdd, bindingsToDelete
func (_m *mockFolderRepo) UpdateBinding(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding) error {
	ret := _m.Called(ctx, dataObject, bindingsToAdd, bindingsToDelete)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *iam.DataObjectReference, []iam.IamBinding, []iam.IamBinding) error); ok {
		r0 = rf(ctx, dataObject, bindingsToAdd, bindingsToDelete)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockFolderRepo_UpdateBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateBinding'
type mockFolderRepo_UpdateBinding_Call struct {
	*mock.Call
}

// UpdateBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - dataObject *iam.DataObjectReference
//   - bindingsToAdd []iam.IamBinding
//   - bindingsToDelete []iam.IamBinding
func (_e *mockFolderRepo_Expecter) UpdateBinding(ctx interface{}, dataObject interface{}, bindingsToAdd interface{}, bindingsToDelete interface{}) *mockFolderRepo_UpdateBinding_Call {
	return &mockFolderRepo_UpdateBinding_Call{Call: _e.mock.On("UpdateBinding", ctx, dataObject, bindingsToAdd, bindingsToDelete)}
}

func (_c *mockFolderRepo_UpdateBinding_Call) Run(run func(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding)) *mockFolderRepo_UpdateBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*iam.DataObjectReference), args[2].([]iam.IamBinding), args[3].([]iam.IamBinding))
	})
	return _c
}

func (_c *mockFolderRepo_UpdateBinding_Call) Return(_a0 error) *mockFolderRepo_UpdateBinding_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFolderRepo_UpdateBinding_Call) RunAndReturn(run func(context.Context, *iam.DataObjectReference, []iam.IamBinding, []iam.IamBinding) error) *mockFolderRepo_UpdateBinding_Call {
	_c.Call.Return(run)
	return _c
}

// newMockFolderRepo creates a new instance of mockFolderRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockFolderRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockFolderRepo {
	mock := &mockFolderRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}