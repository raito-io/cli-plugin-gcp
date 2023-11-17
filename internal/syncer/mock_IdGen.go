// Code generated by mockery v2.36.1. DO NOT EDIT.

package syncer

import mock "github.com/stretchr/testify/mock"

// MockIdGen is an autogenerated mock type for the IdGen type
type MockIdGen struct {
	mock.Mock
}

type MockIdGen_Expecter struct {
	mock *mock.Mock
}

func (_m *MockIdGen) EXPECT() *MockIdGen_Expecter {
	return &MockIdGen_Expecter{mock: &_m.Mock}
}

// New provides a mock function with given fields:
func (_m *MockIdGen) New() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockIdGen_New_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'New'
type MockIdGen_New_Call struct {
	*mock.Call
}

// New is a helper method to define mock.On call
func (_e *MockIdGen_Expecter) New() *MockIdGen_New_Call {
	return &MockIdGen_New_Call{Call: _e.mock.On("New")}
}

func (_c *MockIdGen_New_Call) Run(run func()) *MockIdGen_New_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockIdGen_New_Call) Return(_a0 string) *MockIdGen_New_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIdGen_New_Call) RunAndReturn(run func() string) *MockIdGen_New_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockIdGen creates a new instance of MockIdGen. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockIdGen(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIdGen {
	mock := &MockIdGen{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}