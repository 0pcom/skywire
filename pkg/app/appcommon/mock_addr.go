// Code generated by mockery v1.0.0. DO NOT EDIT.

package appcommon

import mock "github.com/stretchr/testify/mock"

// MockAddr is an autogenerated mock type for the Addr type
type MockAddr struct {
	mock.Mock
}

// Network provides a mock function with given fields:
func (_m *MockAddr) Network() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// String provides a mock function with given fields:
func (_m *MockAddr) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}