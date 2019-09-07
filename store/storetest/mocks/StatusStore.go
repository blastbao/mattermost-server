// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import mock "github.com/stretchr/testify/mock"
import model "github.com/blastbao/mattermost-server/model"

// StatusStore is an autogenerated mock type for the StatusStore type
type StatusStore struct {
	mock.Mock
}

// Get provides a mock function with given fields: userId
func (_m *StatusStore) Get(userId string) (*model.Status, *model.AppError) {
	ret := _m.Called(userId)

	var r0 *model.Status
	if rf, ok := ret.Get(0).(func(string) *model.Status); ok {
		r0 = rf(userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Status)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(userId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetByIds provides a mock function with given fields: userIds
func (_m *StatusStore) GetByIds(userIds []string) ([]*model.Status, *model.AppError) {
	ret := _m.Called(userIds)

	var r0 []*model.Status
	if rf, ok := ret.Get(0).(func([]string) []*model.Status); ok {
		r0 = rf(userIds)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Status)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func([]string) *model.AppError); ok {
		r1 = rf(userIds)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetTotalActiveUsersCount provides a mock function with given fields:
func (_m *StatusStore) GetTotalActiveUsersCount() (int64, *model.AppError) {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
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

// ResetAll provides a mock function with given fields:
func (_m *StatusStore) ResetAll() *model.AppError {
	ret := _m.Called()

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func() *model.AppError); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// SaveOrUpdate provides a mock function with given fields: status
func (_m *StatusStore) SaveOrUpdate(status *model.Status) *model.AppError {
	ret := _m.Called(status)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(*model.Status) *model.AppError); ok {
		r0 = rf(status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// UpdateLastActivityAt provides a mock function with given fields: userId, lastActivityAt
func (_m *StatusStore) UpdateLastActivityAt(userId string, lastActivityAt int64) *model.AppError {
	ret := _m.Called(userId, lastActivityAt)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string, int64) *model.AppError); ok {
		r0 = rf(userId, lastActivityAt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}
