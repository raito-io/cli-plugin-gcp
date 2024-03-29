// Code generated by mockery v2.40.1. DO NOT EDIT.

package bigquery

import (
	context "context"

	iam "github.com/raito-io/cli-plugin-gcp/internal/iam"
	mock "github.com/stretchr/testify/mock"
)

// MockProjectClient is an autogenerated mock type for the ProjectClient type
type MockProjectClient struct {
	mock.Mock
}

type MockProjectClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProjectClient) EXPECT() *MockProjectClient_Expecter {
	return &MockProjectClient_Expecter{mock: &_m.Mock}
}

// GetIamPolicy provides a mock function with given fields: ctx, projectId
func (_m *MockProjectClient) GetIamPolicy(ctx context.Context, projectId string) ([]iam.IamBinding, error) {
	ret := _m.Called(ctx, projectId)

	if len(ret) == 0 {
		panic("no return value specified for GetIamPolicy")
	}

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

// MockProjectClient_GetIamPolicy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIamPolicy'
type MockProjectClient_GetIamPolicy_Call struct {
	*mock.Call
}

// GetIamPolicy is a helper method to define mock.On call
//   - ctx context.Context
//   - projectId string
func (_e *MockProjectClient_Expecter) GetIamPolicy(ctx interface{}, projectId interface{}) *MockProjectClient_GetIamPolicy_Call {
	return &MockProjectClient_GetIamPolicy_Call{Call: _e.mock.On("GetIamPolicy", ctx, projectId)}
}

func (_c *MockProjectClient_GetIamPolicy_Call) Run(run func(ctx context.Context, projectId string)) *MockProjectClient_GetIamPolicy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockProjectClient_GetIamPolicy_Call) Return(_a0 []iam.IamBinding, _a1 error) *MockProjectClient_GetIamPolicy_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProjectClient_GetIamPolicy_Call) RunAndReturn(run func(context.Context, string) ([]iam.IamBinding, error)) *MockProjectClient_GetIamPolicy_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateBinding provides a mock function with given fields: ctx, dataObject, bindingsToAdd, bindingsToDelete
func (_m *MockProjectClient) UpdateBinding(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding) error {
	ret := _m.Called(ctx, dataObject, bindingsToAdd, bindingsToDelete)

	if len(ret) == 0 {
		panic("no return value specified for UpdateBinding")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *iam.DataObjectReference, []iam.IamBinding, []iam.IamBinding) error); ok {
		r0 = rf(ctx, dataObject, bindingsToAdd, bindingsToDelete)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockProjectClient_UpdateBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateBinding'
type MockProjectClient_UpdateBinding_Call struct {
	*mock.Call
}

// UpdateBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - dataObject *iam.DataObjectReference
//   - bindingsToAdd []iam.IamBinding
//   - bindingsToDelete []iam.IamBinding
func (_e *MockProjectClient_Expecter) UpdateBinding(ctx interface{}, dataObject interface{}, bindingsToAdd interface{}, bindingsToDelete interface{}) *MockProjectClient_UpdateBinding_Call {
	return &MockProjectClient_UpdateBinding_Call{Call: _e.mock.On("UpdateBinding", ctx, dataObject, bindingsToAdd, bindingsToDelete)}
}

func (_c *MockProjectClient_UpdateBinding_Call) Run(run func(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding)) *MockProjectClient_UpdateBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*iam.DataObjectReference), args[2].([]iam.IamBinding), args[3].([]iam.IamBinding))
	})
	return _c
}

func (_c *MockProjectClient_UpdateBinding_Call) Return(_a0 error) *MockProjectClient_UpdateBinding_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProjectClient_UpdateBinding_Call) RunAndReturn(run func(context.Context, *iam.DataObjectReference, []iam.IamBinding, []iam.IamBinding) error) *MockProjectClient_UpdateBinding_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProjectClient creates a new instance of MockProjectClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProjectClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProjectClient {
	mock := &MockProjectClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
