// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import mock "github.com/stretchr/testify/mock"
import model "github.com/blastbao/mattermost-server/model"

// UserAccessTokenStore is an autogenerated mock type for the UserAccessTokenStore type
type UserAccessTokenStore struct {
	mock.Mock
}

// Delete provides a mock function with given fields: tokenId
func (_m *UserAccessTokenStore) Delete(tokenId string) *model.AppError {
	ret := _m.Called(tokenId)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string) *model.AppError); ok {
		r0 = rf(tokenId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// DeleteAllForUser provides a mock function with given fields: userId
func (_m *UserAccessTokenStore) DeleteAllForUser(userId string) *model.AppError {
	ret := _m.Called(userId)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string) *model.AppError); ok {
		r0 = rf(userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// Get provides a mock function with given fields: tokenId
func (_m *UserAccessTokenStore) Get(tokenId string) (*model.UserAccessToken, *model.AppError) {
	ret := _m.Called(tokenId)

	var r0 *model.UserAccessToken
	if rf, ok := ret.Get(0).(func(string) *model.UserAccessToken); ok {
		r0 = rf(tokenId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UserAccessToken)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(tokenId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetAll provides a mock function with given fields: offset, limit
func (_m *UserAccessTokenStore) GetAll(offset int, limit int) ([]*model.UserAccessToken, *model.AppError) {
	ret := _m.Called(offset, limit)

	var r0 []*model.UserAccessToken
	if rf, ok := ret.Get(0).(func(int, int) []*model.UserAccessToken); ok {
		r0 = rf(offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.UserAccessToken)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(int, int) *model.AppError); ok {
		r1 = rf(offset, limit)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetByToken provides a mock function with given fields: tokenString
func (_m *UserAccessTokenStore) GetByToken(tokenString string) (*model.UserAccessToken, *model.AppError) {
	ret := _m.Called(tokenString)

	var r0 *model.UserAccessToken
	if rf, ok := ret.Get(0).(func(string) *model.UserAccessToken); ok {
		r0 = rf(tokenString)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UserAccessToken)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(tokenString)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetByUser provides a mock function with given fields: userId, page, perPage
func (_m *UserAccessTokenStore) GetByUser(userId string, page int, perPage int) ([]*model.UserAccessToken, *model.AppError) {
	ret := _m.Called(userId, page, perPage)

	var r0 []*model.UserAccessToken
	if rf, ok := ret.Get(0).(func(string, int, int) []*model.UserAccessToken); ok {
		r0 = rf(userId, page, perPage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.UserAccessToken)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, int, int) *model.AppError); ok {
		r1 = rf(userId, page, perPage)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// Save provides a mock function with given fields: token
func (_m *UserAccessTokenStore) Save(token *model.UserAccessToken) (*model.UserAccessToken, *model.AppError) {
	ret := _m.Called(token)

	var r0 *model.UserAccessToken
	if rf, ok := ret.Get(0).(func(*model.UserAccessToken) *model.UserAccessToken); ok {
		r0 = rf(token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UserAccessToken)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*model.UserAccessToken) *model.AppError); ok {
		r1 = rf(token)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// Search provides a mock function with given fields: term
func (_m *UserAccessTokenStore) Search(term string) ([]*model.UserAccessToken, *model.AppError) {
	ret := _m.Called(term)

	var r0 []*model.UserAccessToken
	if rf, ok := ret.Get(0).(func(string) []*model.UserAccessToken); ok {
		r0 = rf(term)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.UserAccessToken)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(term)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// UpdateTokenDisable provides a mock function with given fields: tokenId
func (_m *UserAccessTokenStore) UpdateTokenDisable(tokenId string) *model.AppError {
	ret := _m.Called(tokenId)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string) *model.AppError); ok {
		r0 = rf(tokenId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// UpdateTokenEnable provides a mock function with given fields: tokenId
func (_m *UserAccessTokenStore) UpdateTokenEnable(tokenId string) *model.AppError {
	ret := _m.Called(tokenId)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string) *model.AppError); ok {
		r0 = rf(tokenId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}
