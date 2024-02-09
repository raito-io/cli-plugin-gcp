// Code generated by mockery v2.40.1. DO NOT EDIT.

package org

import (
	context "context"

	iam "github.com/raito-io/cli-plugin-gcp/internal/iam"
	mock "github.com/stretchr/testify/mock"
)

// mockProjectRepository is an autogenerated mock type for the projectRepository type
type mockProjectRepository struct {
	mock.Mock
}

type mockProjectRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *mockProjectRepository) EXPECT() *mockProjectRepository_Expecter {
	return &mockProjectRepository_Expecter{mock: &_m.Mock}
}

// GetUsers provides a mock function with given fields: ctx, projectName, fn
func (_m *mockProjectRepository) GetUsers(ctx context.Context, projectName string, fn func(context.Context, *iam.UserEntity) error) error {
	ret := _m.Called(ctx, projectName, fn)

	if len(ret) == 0 {
		panic("no return value specified for GetUsers")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, func(context.Context, *iam.UserEntity) error) error); ok {
		r0 = rf(ctx, projectName, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockProjectRepository_GetUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUsers'
type mockProjectRepository_GetUsers_Call struct {
	*mock.Call
}

// GetUsers is a helper method to define mock.On call
//   - ctx context.Context
//   - projectName string
//   - fn func(context.Context , *iam.UserEntity) error
func (_e *mockProjectRepository_Expecter) GetUsers(ctx interface{}, projectName interface{}, fn interface{}) *mockProjectRepository_GetUsers_Call {
	return &mockProjectRepository_GetUsers_Call{Call: _e.mock.On("GetUsers", ctx, projectName, fn)}
}

func (_c *mockProjectRepository_GetUsers_Call) Run(run func(ctx context.Context, projectName string, fn func(context.Context, *iam.UserEntity) error)) *mockProjectRepository_GetUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(func(context.Context, *iam.UserEntity) error))
	})
	return _c
}

func (_c *mockProjectRepository_GetUsers_Call) Return(_a0 error) *mockProjectRepository_GetUsers_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockProjectRepository_GetUsers_Call) RunAndReturn(run func(context.Context, string, func(context.Context, *iam.UserEntity) error) error) *mockProjectRepository_GetUsers_Call {
	_c.Call.Return(run)
	return _c
}

// newMockProjectRepository creates a new instance of mockProjectRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockProjectRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockProjectRepository {
	mock := &mockProjectRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
