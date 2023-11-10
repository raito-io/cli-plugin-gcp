// Code generated by mockery v2.36.1. DO NOT EDIT.

package iam

import (
	context "context"

	config "github.com/raito-io/cli/base/util/config"

	mock "github.com/stretchr/testify/mock"

	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

// MockIAMService is an autogenerated mock type for the IAMService type
type MockIAMService struct {
	mock.Mock
}

type MockIAMService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockIAMService) EXPECT() *MockIAMService_Expecter {
	return &MockIAMService_Expecter{mock: &_m.Mock}
}

// AccessProviderBindingHooks provides a mock function with given fields:
func (_m *MockIAMService) AccessProviderBindingHooks() []AccessProviderBindingHook {
	ret := _m.Called()

	var r0 []AccessProviderBindingHook
	if rf, ok := ret.Get(0).(func() []AccessProviderBindingHook); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]AccessProviderBindingHook)
		}
	}

	return r0
}

// MockIAMService_AccessProviderBindingHooks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AccessProviderBindingHooks'
type MockIAMService_AccessProviderBindingHooks_Call struct {
	*mock.Call
}

// AccessProviderBindingHooks is a helper method to define mock.On call
func (_e *MockIAMService_Expecter) AccessProviderBindingHooks() *MockIAMService_AccessProviderBindingHooks_Call {
	return &MockIAMService_AccessProviderBindingHooks_Call{Call: _e.mock.On("AccessProviderBindingHooks")}
}

func (_c *MockIAMService_AccessProviderBindingHooks_Call) Run(run func()) *MockIAMService_AccessProviderBindingHooks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockIAMService_AccessProviderBindingHooks_Call) Return(_a0 []AccessProviderBindingHook) *MockIAMService_AccessProviderBindingHooks_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIAMService_AccessProviderBindingHooks_Call) RunAndReturn(run func() []AccessProviderBindingHook) *MockIAMService_AccessProviderBindingHooks_Call {
	_c.Call.Return(run)
	return _c
}

// AddIamBinding provides a mock function with given fields: ctx, configMap, binding
func (_m *MockIAMService) AddIamBinding(ctx context.Context, configMap *config.ConfigMap, binding types.IamBinding) error {
	ret := _m.Called(ctx, configMap, binding)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap, types.IamBinding) error); ok {
		r0 = rf(ctx, configMap, binding)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockIAMService_AddIamBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddIamBinding'
type MockIAMService_AddIamBinding_Call struct {
	*mock.Call
}

// AddIamBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *config.ConfigMap
//   - binding IamBinding
func (_e *MockIAMService_Expecter) AddIamBinding(ctx interface{}, configMap interface{}, binding interface{}) *MockIAMService_AddIamBinding_Call {
	return &MockIAMService_AddIamBinding_Call{Call: _e.mock.On("AddIamBinding", ctx, configMap, binding)}
}

func (_c *MockIAMService_AddIamBinding_Call) Run(run func(ctx context.Context, configMap *config.ConfigMap, binding types.IamBinding)) *MockIAMService_AddIamBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*config.ConfigMap), args[2].(types.IamBinding))
	})
	return _c
}

func (_c *MockIAMService_AddIamBinding_Call) Return(_a0 error) *MockIAMService_AddIamBinding_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIAMService_AddIamBinding_Call) RunAndReturn(run func(context.Context, *config.ConfigMap, types.IamBinding) error) *MockIAMService_AddIamBinding_Call {
	_c.Call.Return(run)
	return _c
}

// GetGroups provides a mock function with given fields: ctx, configMap
func (_m *MockIAMService) GetGroups(ctx context.Context, configMap *config.ConfigMap) ([]types.GroupEntity, error) {
	ret := _m.Called(ctx, configMap)

	var r0 []types.GroupEntity
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap) ([]types.GroupEntity, error)); ok {
		return rf(ctx, configMap)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap) []types.GroupEntity); ok {
		r0 = rf(ctx, configMap)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.GroupEntity)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *config.ConfigMap) error); ok {
		r1 = rf(ctx, configMap)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockIAMService_GetGroups_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGroups'
type MockIAMService_GetGroups_Call struct {
	*mock.Call
}

// GetGroups is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *config.ConfigMap
func (_e *MockIAMService_Expecter) GetGroups(ctx interface{}, configMap interface{}) *MockIAMService_GetGroups_Call {
	return &MockIAMService_GetGroups_Call{Call: _e.mock.On("GetGroups", ctx, configMap)}
}

func (_c *MockIAMService_GetGroups_Call) Run(run func(ctx context.Context, configMap *config.ConfigMap)) *MockIAMService_GetGroups_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*config.ConfigMap))
	})
	return _c
}

func (_c *MockIAMService_GetGroups_Call) Return(_a0 []types.GroupEntity, _a1 error) *MockIAMService_GetGroups_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockIAMService_GetGroups_Call) RunAndReturn(run func(context.Context, *config.ConfigMap) ([]types.GroupEntity, error)) *MockIAMService_GetGroups_Call {
	_c.Call.Return(run)
	return _c
}

// GetIAMPolicyBindings provides a mock function with given fields: ctx, configMap
func (_m *MockIAMService) GetIAMPolicyBindings(ctx context.Context, configMap *config.ConfigMap) ([]types.IamBinding, error) {
	ret := _m.Called(ctx, configMap)

	var r0 []types.IamBinding
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap) ([]types.IamBinding, error)); ok {
		return rf(ctx, configMap)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap) []types.IamBinding); ok {
		r0 = rf(ctx, configMap)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.IamBinding)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *config.ConfigMap) error); ok {
		r1 = rf(ctx, configMap)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockIAMService_GetIAMPolicyBindings_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIAMPolicyBindings'
type MockIAMService_GetIAMPolicyBindings_Call struct {
	*mock.Call
}

// GetIAMPolicyBindings is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *config.ConfigMap
func (_e *MockIAMService_Expecter) GetIAMPolicyBindings(ctx interface{}, configMap interface{}) *MockIAMService_GetIAMPolicyBindings_Call {
	return &MockIAMService_GetIAMPolicyBindings_Call{Call: _e.mock.On("GetIAMPolicyBindings", ctx, configMap)}
}

func (_c *MockIAMService_GetIAMPolicyBindings_Call) Run(run func(ctx context.Context, configMap *config.ConfigMap)) *MockIAMService_GetIAMPolicyBindings_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*config.ConfigMap))
	})
	return _c
}

func (_c *MockIAMService_GetIAMPolicyBindings_Call) Return(_a0 []types.IamBinding, _a1 error) *MockIAMService_GetIAMPolicyBindings_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockIAMService_GetIAMPolicyBindings_Call) RunAndReturn(run func(context.Context, *config.ConfigMap) ([]types.IamBinding, error)) *MockIAMService_GetIAMPolicyBindings_Call {
	_c.Call.Return(run)
	return _c
}

// GetProjectOwners provides a mock function with given fields: ctx, configMap, projectId
func (_m *MockIAMService) GetProjectOwners(ctx context.Context, configMap *config.ConfigMap, projectId string) ([]string, []string, []string, error) {
	ret := _m.Called(ctx, configMap, projectId)

	var r0 []string
	var r1 []string
	var r2 []string
	var r3 error
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap, string) ([]string, []string, []string, error)); ok {
		return rf(ctx, configMap, projectId)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap, string) []string); ok {
		r0 = rf(ctx, configMap, projectId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *config.ConfigMap, string) []string); ok {
		r1 = rf(ctx, configMap, projectId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]string)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, *config.ConfigMap, string) []string); ok {
		r2 = rf(ctx, configMap, projectId)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).([]string)
		}
	}

	if rf, ok := ret.Get(3).(func(context.Context, *config.ConfigMap, string) error); ok {
		r3 = rf(ctx, configMap, projectId)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// MockIAMService_GetProjectOwners_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProjectOwners'
type MockIAMService_GetProjectOwners_Call struct {
	*mock.Call
}

// GetProjectOwners is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *config.ConfigMap
//   - projectId string
func (_e *MockIAMService_Expecter) GetProjectOwners(ctx interface{}, configMap interface{}, projectId interface{}) *MockIAMService_GetProjectOwners_Call {
	return &MockIAMService_GetProjectOwners_Call{Call: _e.mock.On("GetProjectOwners", ctx, configMap, projectId)}
}

func (_c *MockIAMService_GetProjectOwners_Call) Run(run func(ctx context.Context, configMap *config.ConfigMap, projectId string)) *MockIAMService_GetProjectOwners_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*config.ConfigMap), args[2].(string))
	})
	return _c
}

func (_c *MockIAMService_GetProjectOwners_Call) Return(owner []string, editor []string, viewer []string, err error) *MockIAMService_GetProjectOwners_Call {
	_c.Call.Return(owner, editor, viewer, err)
	return _c
}

func (_c *MockIAMService_GetProjectOwners_Call) RunAndReturn(run func(context.Context, *config.ConfigMap, string) ([]string, []string, []string, error)) *MockIAMService_GetProjectOwners_Call {
	_c.Call.Return(run)
	return _c
}

// GetServiceAccounts provides a mock function with given fields: ctx, configMap
func (_m *MockIAMService) GetServiceAccounts(ctx context.Context, configMap *config.ConfigMap) ([]types.UserEntity, error) {
	ret := _m.Called(ctx, configMap)

	var r0 []types.UserEntity
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap) ([]types.UserEntity, error)); ok {
		return rf(ctx, configMap)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap) []types.UserEntity); ok {
		r0 = rf(ctx, configMap)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.UserEntity)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *config.ConfigMap) error); ok {
		r1 = rf(ctx, configMap)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockIAMService_GetServiceAccounts_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetServiceAccounts'
type MockIAMService_GetServiceAccounts_Call struct {
	*mock.Call
}

// GetServiceAccounts is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *config.ConfigMap
func (_e *MockIAMService_Expecter) GetServiceAccounts(ctx interface{}, configMap interface{}) *MockIAMService_GetServiceAccounts_Call {
	return &MockIAMService_GetServiceAccounts_Call{Call: _e.mock.On("GetServiceAccounts", ctx, configMap)}
}

func (_c *MockIAMService_GetServiceAccounts_Call) Run(run func(ctx context.Context, configMap *config.ConfigMap)) *MockIAMService_GetServiceAccounts_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*config.ConfigMap))
	})
	return _c
}

func (_c *MockIAMService_GetServiceAccounts_Call) Return(_a0 []types.UserEntity, _a1 error) *MockIAMService_GetServiceAccounts_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockIAMService_GetServiceAccounts_Call) RunAndReturn(run func(context.Context, *config.ConfigMap) ([]types.UserEntity, error)) *MockIAMService_GetServiceAccounts_Call {
	_c.Call.Return(run)
	return _c
}

// GetUsers provides a mock function with given fields: ctx, configMap
func (_m *MockIAMService) GetUsers(ctx context.Context, configMap *config.ConfigMap) ([]types.UserEntity, error) {
	ret := _m.Called(ctx, configMap)

	var r0 []types.UserEntity
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap) ([]types.UserEntity, error)); ok {
		return rf(ctx, configMap)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap) []types.UserEntity); ok {
		r0 = rf(ctx, configMap)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.UserEntity)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *config.ConfigMap) error); ok {
		r1 = rf(ctx, configMap)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockIAMService_GetUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUsers'
type MockIAMService_GetUsers_Call struct {
	*mock.Call
}

// GetUsers is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *config.ConfigMap
func (_e *MockIAMService_Expecter) GetUsers(ctx interface{}, configMap interface{}) *MockIAMService_GetUsers_Call {
	return &MockIAMService_GetUsers_Call{Call: _e.mock.On("GetUsers", ctx, configMap)}
}

func (_c *MockIAMService_GetUsers_Call) Run(run func(ctx context.Context, configMap *config.ConfigMap)) *MockIAMService_GetUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*config.ConfigMap))
	})
	return _c
}

func (_c *MockIAMService_GetUsers_Call) Return(_a0 []types.UserEntity, _a1 error) *MockIAMService_GetUsers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockIAMService_GetUsers_Call) RunAndReturn(run func(context.Context, *config.ConfigMap) ([]types.UserEntity, error)) *MockIAMService_GetUsers_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveIamBinding provides a mock function with given fields: ctx, configMap, binding
func (_m *MockIAMService) RemoveIamBinding(ctx context.Context, configMap *config.ConfigMap, binding types.IamBinding) error {
	ret := _m.Called(ctx, configMap, binding)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *config.ConfigMap, types.IamBinding) error); ok {
		r0 = rf(ctx, configMap, binding)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockIAMService_RemoveIamBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveIamBinding'
type MockIAMService_RemoveIamBinding_Call struct {
	*mock.Call
}

// RemoveIamBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *config.ConfigMap
//   - binding IamBinding
func (_e *MockIAMService_Expecter) RemoveIamBinding(ctx interface{}, configMap interface{}, binding interface{}) *MockIAMService_RemoveIamBinding_Call {
	return &MockIAMService_RemoveIamBinding_Call{Call: _e.mock.On("RemoveIamBinding", ctx, configMap, binding)}
}

func (_c *MockIAMService_RemoveIamBinding_Call) Run(run func(ctx context.Context, configMap *config.ConfigMap, binding types.IamBinding)) *MockIAMService_RemoveIamBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*config.ConfigMap), args[2].(types.IamBinding))
	})
	return _c
}

func (_c *MockIAMService_RemoveIamBinding_Call) Return(_a0 error) *MockIAMService_RemoveIamBinding_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIAMService_RemoveIamBinding_Call) RunAndReturn(run func(context.Context, *config.ConfigMap, types.IamBinding) error) *MockIAMService_RemoveIamBinding_Call {
	_c.Call.Return(run)
	return _c
}

// WithBindingHook provides a mock function with given fields: hooks
func (_m *MockIAMService) WithBindingHook(hooks ...AccessProviderBindingHook) IAMService {
	_va := make([]interface{}, len(hooks))
	for _i := range hooks {
		_va[_i] = hooks[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 IAMService
	if rf, ok := ret.Get(0).(func(...AccessProviderBindingHook) IAMService); ok {
		r0 = rf(hooks...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(IAMService)
		}
	}

	return r0
}

// MockIAMService_WithBindingHook_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithBindingHook'
type MockIAMService_WithBindingHook_Call struct {
	*mock.Call
}

// WithBindingHook is a helper method to define mock.On call
//   - hooks ...AccessProviderBindingHook
func (_e *MockIAMService_Expecter) WithBindingHook(hooks ...interface{}) *MockIAMService_WithBindingHook_Call {
	return &MockIAMService_WithBindingHook_Call{Call: _e.mock.On("WithBindingHook",
		append([]interface{}{}, hooks...)...)}
}

func (_c *MockIAMService_WithBindingHook_Call) Run(run func(hooks ...AccessProviderBindingHook)) *MockIAMService_WithBindingHook_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]AccessProviderBindingHook, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(AccessProviderBindingHook)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *MockIAMService_WithBindingHook_Call) Return(_a0 IAMService) *MockIAMService_WithBindingHook_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIAMService_WithBindingHook_Call) RunAndReturn(run func(...AccessProviderBindingHook) IAMService) *MockIAMService_WithBindingHook_Call {
	_c.Call.Return(run)
	return _c
}

// WithServiceIamRepo provides a mock function with given fields: resourceTypes, localRepo, ids
func (_m *MockIAMService) WithServiceIamRepo(resourceTypes []string, localRepo IAMRepository, ids func(context.Context, *config.ConfigMap) ([]string, error)) IAMService {
	ret := _m.Called(resourceTypes, localRepo, ids)

	var r0 IAMService
	if rf, ok := ret.Get(0).(func([]string, IAMRepository, func(context.Context, *config.ConfigMap) ([]string, error)) IAMService); ok {
		r0 = rf(resourceTypes, localRepo, ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(IAMService)
		}
	}

	return r0
}

// MockIAMService_WithServiceIamRepo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithServiceIamRepo'
type MockIAMService_WithServiceIamRepo_Call struct {
	*mock.Call
}

// WithServiceIamRepo is a helper method to define mock.On call
//   - resourceTypes []string
//   - localRepo IAMRepository
//   - ids func(context.Context , *config.ConfigMap)([]string , error)
func (_e *MockIAMService_Expecter) WithServiceIamRepo(resourceTypes interface{}, localRepo interface{}, ids interface{}) *MockIAMService_WithServiceIamRepo_Call {
	return &MockIAMService_WithServiceIamRepo_Call{Call: _e.mock.On("WithServiceIamRepo", resourceTypes, localRepo, ids)}
}

func (_c *MockIAMService_WithServiceIamRepo_Call) Run(run func(resourceTypes []string, localRepo IAMRepository, ids func(context.Context, *config.ConfigMap) ([]string, error))) *MockIAMService_WithServiceIamRepo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string), args[1].(IAMRepository), args[2].(func(context.Context, *config.ConfigMap) ([]string, error)))
	})
	return _c
}

func (_c *MockIAMService_WithServiceIamRepo_Call) Return(_a0 IAMService) *MockIAMService_WithServiceIamRepo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIAMService_WithServiceIamRepo_Call) RunAndReturn(run func([]string, IAMRepository, func(context.Context, *config.ConfigMap) ([]string, error)) IAMService) *MockIAMService_WithServiceIamRepo_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockIAMService creates a new instance of MockIAMService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockIAMService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIAMService {
	mock := &MockIAMService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
