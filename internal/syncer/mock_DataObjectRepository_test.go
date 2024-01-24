// Code generated by mockery v2.40.1. DO NOT EDIT.

package syncer

import (
	context "context"

	iam "github.com/raito-io/cli-plugin-gcp/internal/iam"
	data_source "github.com/raito-io/cli/base/data_source"

	mock "github.com/stretchr/testify/mock"

	org "github.com/raito-io/cli-plugin-gcp/internal/org"
)

// MockDataObjectRepository is an autogenerated mock type for the DataObjectRepository type
type MockDataObjectRepository struct {
	mock.Mock
}

type MockDataObjectRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDataObjectRepository) EXPECT() *MockDataObjectRepository_Expecter {
	return &MockDataObjectRepository_Expecter{mock: &_m.Mock}
}

// Bindings provides a mock function with given fields: ctx, config, fn
func (_m *MockDataObjectRepository) Bindings(ctx context.Context, config *data_source.DataSourceSyncConfig, fn func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
	ret := _m.Called(ctx, config, fn)

	if len(ret) == 0 {
		panic("no return value specified for Bindings")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *data_source.DataSourceSyncConfig, func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error); ok {
		r0 = rf(ctx, config, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataObjectRepository_Bindings_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bindings'
type MockDataObjectRepository_Bindings_Call struct {
	*mock.Call
}

// Bindings is a helper method to define mock.On call
//   - ctx context.Context
//   - config *data_source.DataSourceSyncConfig
//   - fn func(context.Context , *org.GcpOrgEntity , []iam.IamBinding) error
func (_e *MockDataObjectRepository_Expecter) Bindings(ctx interface{}, config interface{}, fn interface{}) *MockDataObjectRepository_Bindings_Call {
	return &MockDataObjectRepository_Bindings_Call{Call: _e.mock.On("Bindings", ctx, config, fn)}
}

func (_c *MockDataObjectRepository_Bindings_Call) Run(run func(ctx context.Context, config *data_source.DataSourceSyncConfig, fn func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error)) *MockDataObjectRepository_Bindings_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*data_source.DataSourceSyncConfig), args[2].(func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error))
	})
	return _c
}

func (_c *MockDataObjectRepository_Bindings_Call) Return(_a0 error) *MockDataObjectRepository_Bindings_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDataObjectRepository_Bindings_Call) RunAndReturn(run func(context.Context, *data_source.DataSourceSyncConfig, func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error) *MockDataObjectRepository_Bindings_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDataObjectRepository creates a new instance of MockDataObjectRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDataObjectRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDataObjectRepository {
	mock := &MockDataObjectRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
