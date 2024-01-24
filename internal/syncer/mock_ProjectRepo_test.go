// Code generated by mockery v2.40.1. DO NOT EDIT.

package syncer

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockProjectRepo is an autogenerated mock type for the ProjectRepo type
type MockProjectRepo struct {
	mock.Mock
}

type MockProjectRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProjectRepo) EXPECT() *MockProjectRepo_Expecter {
	return &MockProjectRepo_Expecter{mock: &_m.Mock}
}

// GetProjectOwner provides a mock function with given fields: ctx, projectId
func (_m *MockProjectRepo) GetProjectOwner(ctx context.Context, projectId string) ([]string, []string, []string, error) {
	ret := _m.Called(ctx, projectId)

	if len(ret) == 0 {
		panic("no return value specified for GetProjectOwner")
	}

	var r0 []string
	var r1 []string
	var r2 []string
	var r3 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]string, []string, []string, error)); ok {
		return rf(ctx, projectId)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []string); ok {
		r0 = rf(ctx, projectId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) []string); ok {
		r1 = rf(ctx, projectId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]string)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, string) []string); ok {
		r2 = rf(ctx, projectId)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).([]string)
		}
	}

	if rf, ok := ret.Get(3).(func(context.Context, string) error); ok {
		r3 = rf(ctx, projectId)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// MockProjectRepo_GetProjectOwner_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProjectOwner'
type MockProjectRepo_GetProjectOwner_Call struct {
	*mock.Call
}

// GetProjectOwner is a helper method to define mock.On call
//   - ctx context.Context
//   - projectId string
func (_e *MockProjectRepo_Expecter) GetProjectOwner(ctx interface{}, projectId interface{}) *MockProjectRepo_GetProjectOwner_Call {
	return &MockProjectRepo_GetProjectOwner_Call{Call: _e.mock.On("GetProjectOwner", ctx, projectId)}
}

func (_c *MockProjectRepo_GetProjectOwner_Call) Run(run func(ctx context.Context, projectId string)) *MockProjectRepo_GetProjectOwner_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockProjectRepo_GetProjectOwner_Call) Return(owner []string, editor []string, viewer []string, err error) *MockProjectRepo_GetProjectOwner_Call {
	_c.Call.Return(owner, editor, viewer, err)
	return _c
}

func (_c *MockProjectRepo_GetProjectOwner_Call) RunAndReturn(run func(context.Context, string) ([]string, []string, []string, error)) *MockProjectRepo_GetProjectOwner_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProjectRepo creates a new instance of MockProjectRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProjectRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProjectRepo {
	mock := &MockProjectRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
