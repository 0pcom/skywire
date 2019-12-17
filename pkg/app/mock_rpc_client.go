// Code generated by mockery v1.0.0. DO NOT EDIT.

package app

import (
	appnet "github.com/SkycoinProject/skywire-mainnet/pkg/app/appnet"
	mock "github.com/stretchr/testify/mock"

	routing "github.com/SkycoinProject/skywire-mainnet/pkg/routing"

	time "time"
)

// MockRPCClient is an autogenerated mock type for the RPCClient type
type MockRPCClient struct {
	mock.Mock
}

// Accept provides a mock function with given fields: lisID
func (_m *MockRPCClient) Accept(lisID uint16) (uint16, appnet.Addr, error) {
	ret := _m.Called(lisID)

	var r0 uint16
	if rf, ok := ret.Get(0).(func(uint16) uint16); ok {
		r0 = rf(lisID)
	} else {
		r0 = ret.Get(0).(uint16)
	}

	var r1 appnet.Addr
	if rf, ok := ret.Get(1).(func(uint16) appnet.Addr); ok {
		r1 = rf(lisID)
	} else {
		r1 = ret.Get(1).(appnet.Addr)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(uint16) error); ok {
		r2 = rf(lisID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CloseConn provides a mock function with given fields: id
func (_m *MockRPCClient) CloseConn(id uint16) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint16) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CloseListener provides a mock function with given fields: id
func (_m *MockRPCClient) CloseListener(id uint16) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint16) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Dial provides a mock function with given fields: remote
func (_m *MockRPCClient) Dial(remote appnet.Addr) (uint16, routing.Port, error) {
	ret := _m.Called(remote)

	var r0 uint16
	if rf, ok := ret.Get(0).(func(appnet.Addr) uint16); ok {
		r0 = rf(remote)
	} else {
		r0 = ret.Get(0).(uint16)
	}

	var r1 routing.Port
	if rf, ok := ret.Get(1).(func(appnet.Addr) routing.Port); ok {
		r1 = rf(remote)
	} else {
		r1 = ret.Get(1).(routing.Port)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(appnet.Addr) error); ok {
		r2 = rf(remote)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Listen provides a mock function with given fields: local
func (_m *MockRPCClient) Listen(local appnet.Addr) (uint16, error) {
	ret := _m.Called(local)

	var r0 uint16
	if rf, ok := ret.Get(0).(func(appnet.Addr) uint16); ok {
		r0 = rf(local)
	} else {
		r0 = ret.Get(0).(uint16)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(appnet.Addr) error); ok {
		r1 = rf(local)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Read provides a mock function with given fields: connID, b
func (_m *MockRPCClient) Read(connID uint16, b []byte) (int, error) {
	ret := _m.Called(connID, b)

	var r0 int
	if rf, ok := ret.Get(0).(func(uint16, []byte) int); ok {
		r0 = rf(connID, b)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint16, []byte) error); ok {
		r1 = rf(connID, b)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetDeadline provides a mock function with given fields: id, t
func (_m *MockRPCClient) SetDeadline(connID uint16, d time.Time) error {
	ret := _m.Called(connID, d)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint16, time.Time) error); ok {
		r0 = rf(connID, d)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetReadDeadline provides a mock function with given fields: id, t
func (_m *MockRPCClient) SetReadDeadline(connID uint16, d time.Time) error {
	ret := _m.Called(connID, d)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint16, time.Time) error); ok {
		r0 = rf(connID, d)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetWriteDeadline provides a mock function with given fields: id, t
func (_m *MockRPCClient) SetWriteDeadline(connID uint16, d time.Time) error {
	ret := _m.Called(connID, d)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint16, time.Time) error); ok {
		r0 = rf(connID, d)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Write provides a mock function with given fields: connID, b
func (_m *MockRPCClient) Write(connID uint16, b []byte) (int, error) {
	ret := _m.Called(connID, b)

	var r0 int
	if rf, ok := ret.Get(0).(func(uint16, []byte) int); ok {
		r0 = rf(connID, b)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint16, []byte) error); ok {
		r1 = rf(connID, b)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
