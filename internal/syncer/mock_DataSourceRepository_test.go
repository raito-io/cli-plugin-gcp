// Code generated by mockery v2.39.1. DO NOT EDIT.

package syncer

import (
	context "context"

	data_source "github.com/raito-io/cli/base/data_source"
	mock "github.com/stretchr/testify/mock"

	org "github.com/raito-io/cli-plugin-gcp/internal/org"
)

// MockDataSourceRepository is an autogenerated mock type for the DataSourceRepository type
type MockDataSourceRepository struct {
	mock.Mock
}

type MockDataSourceRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDataSourceRepository) EXPECT() *MockDataSourceRepository_Expecter {
	return &MockDataSourceRepository_Expecter{mock: &_m.Mock}
}

// DataObjects provides a mock function with given fields: ctx, config, fn
func (_m *MockDataSourceRepository) DataObjects(ctx context.Context, config *data_source.DataSourceSyncConfig, fn func(context.Context, *org.GcpOrgEntity) error) error {
	ret := _m.Called(ctx, config, fn)

	if len(ret) == 0 {
		panic("no return value specified for DataObjects")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *data_source.DataSourceSyncConfig, func(context.Context, *org.GcpOrgEntity) error) error); ok {
		r0 = rf(ctx, config, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataSourceRepository_DataObjects_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DataObjects'
type MockDataSourceRepository_DataObjects_Call struct {
	*mock.Call
}

// DataObjects is a helper method to define mock.On call
//   - ctx context.Context
//   - config *data_source.DataSourceSyncConfig
//   - fn func(context.Context , *org.GcpOrgEntity) error
func (_e *MockDataSourceRepository_Expecter) DataObjects(ctx interface{}, config interface{}, fn interface{}) *MockDataSourceRepository_DataObjects_Call {
	return &MockDataSourceRepository_DataObjects_Call{Call: _e.mock.On("DataObjects", ctx, config, fn)}
}

func (_c *MockDataSourceRepository_DataObjects_Call) Run(run func(ctx context.Context, config *data_source.DataSourceSyncConfig, fn func(context.Context, *org.GcpOrgEntity) error)) *MockDataSourceRepository_DataObjects_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*data_source.DataSourceSyncConfig), args[2].(func(context.Context, *org.GcpOrgEntity) error))
	})
	return _c
}

func (_c *MockDataSourceRepository_DataObjects_Call) Return(_a0 error) *MockDataSourceRepository_DataObjects_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDataSourceRepository_DataObjects_Call) RunAndReturn(run func(context.Context, *data_source.DataSourceSyncConfig, func(context.Context, *org.GcpOrgEntity) error) error) *MockDataSourceRepository_DataObjects_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDataSourceRepository creates a new instance of MockDataSourceRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDataSourceRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDataSourceRepository {
	mock := &MockDataSourceRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
