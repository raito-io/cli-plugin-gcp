// Code generated by mockery v2.39.1. DO NOT EDIT.

package bigquery

import (
	context "context"

	gobigquery "cloud.google.com/go/bigquery"
	mock "github.com/stretchr/testify/mock"

	org "github.com/raito-io/cli-plugin-gcp/internal/org"
)

// mockDataCatalogBqRepository is an autogenerated mock type for the dataCatalogBqRepository type
type mockDataCatalogBqRepository struct {
	mock.Mock
}

type mockDataCatalogBqRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *mockDataCatalogBqRepository) EXPECT() *mockDataCatalogBqRepository_Expecter {
	return &mockDataCatalogBqRepository_Expecter{mock: &_m.Mock}
}

// ListDataSets provides a mock function with given fields: ctx, parent, fn
func (_m *mockDataCatalogBqRepository) ListDataSets(ctx context.Context, parent *org.GcpOrgEntity, fn func(context.Context, *org.GcpOrgEntity, *gobigquery.Dataset) error) error {
	ret := _m.Called(ctx, parent, fn)

	if len(ret) == 0 {
		panic("no return value specified for ListDataSets")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *org.GcpOrgEntity, func(context.Context, *org.GcpOrgEntity, *gobigquery.Dataset) error) error); ok {
		r0 = rf(ctx, parent, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDataCatalogBqRepository_ListDataSets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListDataSets'
type mockDataCatalogBqRepository_ListDataSets_Call struct {
	*mock.Call
}

// ListDataSets is a helper method to define mock.On call
//   - ctx context.Context
//   - parent *org.GcpOrgEntity
//   - fn func(context.Context , *org.GcpOrgEntity , *gobigquery.Dataset) error
func (_e *mockDataCatalogBqRepository_Expecter) ListDataSets(ctx interface{}, parent interface{}, fn interface{}) *mockDataCatalogBqRepository_ListDataSets_Call {
	return &mockDataCatalogBqRepository_ListDataSets_Call{Call: _e.mock.On("ListDataSets", ctx, parent, fn)}
}

func (_c *mockDataCatalogBqRepository_ListDataSets_Call) Run(run func(ctx context.Context, parent *org.GcpOrgEntity, fn func(context.Context, *org.GcpOrgEntity, *gobigquery.Dataset) error)) *mockDataCatalogBqRepository_ListDataSets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*org.GcpOrgEntity), args[2].(func(context.Context, *org.GcpOrgEntity, *gobigquery.Dataset) error))
	})
	return _c
}

func (_c *mockDataCatalogBqRepository_ListDataSets_Call) Return(_a0 error) *mockDataCatalogBqRepository_ListDataSets_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDataCatalogBqRepository_ListDataSets_Call) RunAndReturn(run func(context.Context, *org.GcpOrgEntity, func(context.Context, *org.GcpOrgEntity, *gobigquery.Dataset) error) error) *mockDataCatalogBqRepository_ListDataSets_Call {
	_c.Call.Return(run)
	return _c
}

// Project provides a mock function with given fields:
func (_m *mockDataCatalogBqRepository) Project() *org.GcpOrgEntity {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Project")
	}

	var r0 *org.GcpOrgEntity
	if rf, ok := ret.Get(0).(func() *org.GcpOrgEntity); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*org.GcpOrgEntity)
		}
	}

	return r0
}

// mockDataCatalogBqRepository_Project_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Project'
type mockDataCatalogBqRepository_Project_Call struct {
	*mock.Call
}

// Project is a helper method to define mock.On call
func (_e *mockDataCatalogBqRepository_Expecter) Project() *mockDataCatalogBqRepository_Project_Call {
	return &mockDataCatalogBqRepository_Project_Call{Call: _e.mock.On("Project")}
}

func (_c *mockDataCatalogBqRepository_Project_Call) Run(run func()) *mockDataCatalogBqRepository_Project_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockDataCatalogBqRepository_Project_Call) Return(_a0 *org.GcpOrgEntity) *mockDataCatalogBqRepository_Project_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDataCatalogBqRepository_Project_Call) RunAndReturn(run func() *org.GcpOrgEntity) *mockDataCatalogBqRepository_Project_Call {
	_c.Call.Return(run)
	return _c
}

// newMockDataCatalogBqRepository creates a new instance of mockDataCatalogBqRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockDataCatalogBqRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockDataCatalogBqRepository {
	mock := &mockDataCatalogBqRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
