// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import mock "github.com/stretchr/testify/mock"
import model "github.com/blastbao/mattermost-server/model"

// SystemStore is an autogenerated mock type for the SystemStore type
type SystemStore struct {
	mock.Mock
}

// Get provides a mock function with given fields:
func (_m *SystemStore) Get() (model.StringMap, *model.AppError) {
	ret := _m.Called()

	var r0 model.StringMap
	if rf, ok := ret.Get(0).(func() model.StringMap); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.StringMap)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func() *model.AppError); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetByName provides a mock function with given fields: name
func (_m *SystemStore) GetByName(name string) (*model.System, *model.AppError) {
	ret := _m.Called(name)

	var r0 *model.System
	if rf, ok := ret.Get(0).(func(string) *model.System); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.System)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(name)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// PermanentDeleteByName provides a mock function with given fields: name
func (_m *SystemStore) PermanentDeleteByName(name string) (*model.System, *model.AppError) {
	ret := _m.Called(name)

	var r0 *model.System
	if rf, ok := ret.Get(0).(func(string) *model.System); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.System)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(name)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// Save provides a mock function with given fields: system
func (_m *SystemStore) Save(system *model.System) *model.AppError {
	ret := _m.Called(system)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(*model.System) *model.AppError); ok {
		r0 = rf(system)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// SaveOrUpdate provides a mock function with given fields: system
func (_m *SystemStore) SaveOrUpdate(system *model.System) *model.AppError {
	ret := _m.Called(system)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(*model.System) *model.AppError); ok {
		r0 = rf(system)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// Update provides a mock function with given fields: system
func (_m *SystemStore) Update(system *model.System) *model.AppError {
	ret := _m.Called(system)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(*model.System) *model.AppError); ok {
		r0 = rf(system)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}
