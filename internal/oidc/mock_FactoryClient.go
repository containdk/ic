// Code generated by mockery v2.53.2. DO NOT EDIT.

package oidc

import (
	context "context"
	slog "log/slog"

	mock "github.com/stretchr/testify/mock"
)

// MockFactoryClient is an autogenerated mock type for the FactoryClient type
type MockFactoryClient struct {
	mock.Mock
}

type MockFactoryClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockFactoryClient) EXPECT() *MockFactoryClient_Expecter {
	return &MockFactoryClient_Expecter{mock: &_m.Mock}
}

// New provides a mock function with given fields: ctx, p
func (_m *MockFactoryClient) New(ctx context.Context, p Provider) (Client, error) {
	ret := _m.Called(ctx, p)

	if len(ret) == 0 {
		panic("no return value specified for New")
	}

	var r0 Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, Provider) (Client, error)); ok {
		return rf(ctx, p)
	}
	if rf, ok := ret.Get(0).(func(context.Context, Provider) Client); ok {
		r0 = rf(ctx, p)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Client)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, Provider) error); ok {
		r1 = rf(ctx, p)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockFactoryClient_New_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'New'
type MockFactoryClient_New_Call struct {
	*mock.Call
}

// New is a helper method to define mock.On call
//   - ctx context.Context
//   - p Provider
func (_e *MockFactoryClient_Expecter) New(ctx interface{}, p interface{}) *MockFactoryClient_New_Call {
	return &MockFactoryClient_New_Call{Call: _e.mock.On("New", ctx, p)}
}

func (_c *MockFactoryClient_New_Call) Run(run func(ctx context.Context, p Provider)) *MockFactoryClient_New_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(Provider))
	})
	return _c
}

func (_c *MockFactoryClient_New_Call) Return(_a0 Client, _a1 error) *MockFactoryClient_New_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockFactoryClient_New_Call) RunAndReturn(run func(context.Context, Provider) (Client, error)) *MockFactoryClient_New_Call {
	_c.Call.Return(run)
	return _c
}

// SetLogger provides a mock function with given fields: _a0
func (_m *MockFactoryClient) SetLogger(_a0 *slog.Logger) {
	_m.Called(_a0)
}

// MockFactoryClient_SetLogger_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetLogger'
type MockFactoryClient_SetLogger_Call struct {
	*mock.Call
}

// SetLogger is a helper method to define mock.On call
//   - _a0 *slog.Logger
func (_e *MockFactoryClient_Expecter) SetLogger(_a0 interface{}) *MockFactoryClient_SetLogger_Call {
	return &MockFactoryClient_SetLogger_Call{Call: _e.mock.On("SetLogger", _a0)}
}

func (_c *MockFactoryClient_SetLogger_Call) Run(run func(_a0 *slog.Logger)) *MockFactoryClient_SetLogger_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*slog.Logger))
	})
	return _c
}

func (_c *MockFactoryClient_SetLogger_Call) Return() *MockFactoryClient_SetLogger_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockFactoryClient_SetLogger_Call) RunAndReturn(run func(*slog.Logger)) *MockFactoryClient_SetLogger_Call {
	_c.Run(run)
	return _c
}

// NewMockFactoryClient creates a new instance of MockFactoryClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockFactoryClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockFactoryClient {
	mock := &MockFactoryClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
