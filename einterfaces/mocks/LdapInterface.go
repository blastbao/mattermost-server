// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make einterfaces-mocks`.

package mocks

import mock "github.com/stretchr/testify/mock"
import model "github.com/blastbao/mattermost-server/model"

// LdapInterface is an autogenerated mock type for the LdapInterface type
type LdapInterface struct {
	mock.Mock
}

// CheckPassword provides a mock function with given fields: id, password
func (_m *LdapInterface) CheckPassword(id string, password string) *model.AppError {
	ret := _m.Called(id, password)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string, string) *model.AppError); ok {
		r0 = rf(id, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// CheckPasswordAuthData provides a mock function with given fields: authData, password
func (_m *LdapInterface) CheckPasswordAuthData(authData string, password string) *model.AppError {
	ret := _m.Called(authData, password)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string, string) *model.AppError); ok {
		r0 = rf(authData, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// DoLogin provides a mock function with given fields: id, password
func (_m *LdapInterface) DoLogin(id string, password string) (*model.User, *model.AppError) {
	ret := _m.Called(id, password)

	var r0 *model.User
	if rf, ok := ret.Get(0).(func(string, string) *model.User); ok {
		r0 = rf(id, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, string) *model.AppError); ok {
		r1 = rf(id, password)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// FirstLoginSync provides a mock function with given fields: userID, userAuthService, userAuthData, email
func (_m *LdapInterface) FirstLoginSync(userID string, userAuthService string, userAuthData string, email string) *model.AppError {
	ret := _m.Called(userID, userAuthService, userAuthData, email)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string, string, string, string) *model.AppError); ok {
		r0 = rf(userID, userAuthService, userAuthData, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// GetAllGroupsPage provides a mock function with given fields: page, perPage, opts
func (_m *LdapInterface) GetAllGroupsPage(page int, perPage int, opts model.LdapGroupSearchOpts) ([]*model.Group, int, *model.AppError) {
	ret := _m.Called(page, perPage, opts)

	var r0 []*model.Group
	if rf, ok := ret.Get(0).(func(int, int, model.LdapGroupSearchOpts) []*model.Group); ok {
		r0 = rf(page, perPage, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Group)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func(int, int, model.LdapGroupSearchOpts) int); ok {
		r1 = rf(page, perPage, opts)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 *model.AppError
	if rf, ok := ret.Get(2).(func(int, int, model.LdapGroupSearchOpts) *model.AppError); ok {
		r2 = rf(page, perPage, opts)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*model.AppError)
		}
	}

	return r0, r1, r2
}

// GetAllLdapUsers provides a mock function with given fields:
func (_m *LdapInterface) GetAllLdapUsers() ([]*model.User, *model.AppError) {
	ret := _m.Called()

	var r0 []*model.User
	if rf, ok := ret.Get(0).(func() []*model.User); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.User)
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

// GetGroup provides a mock function with given fields: groupUID
func (_m *LdapInterface) GetGroup(groupUID string) (*model.Group, *model.AppError) {
	ret := _m.Called(groupUID)

	var r0 *model.Group
	if rf, ok := ret.Get(0).(func(string) *model.Group); ok {
		r0 = rf(groupUID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Group)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(groupUID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetUser provides a mock function with given fields: id
func (_m *LdapInterface) GetUser(id string) (*model.User, *model.AppError) {
	ret := _m.Called(id)

	var r0 *model.User
	if rf, ok := ret.Get(0).(func(string) *model.User); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(id)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetUserAttributes provides a mock function with given fields: id, attributes
func (_m *LdapInterface) GetUserAttributes(id string, attributes []string) (map[string]string, *model.AppError) {
	ret := _m.Called(id, attributes)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(string, []string) map[string]string); ok {
		r0 = rf(id, attributes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, []string) *model.AppError); ok {
		r1 = rf(id, attributes)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// MigrateIDAttribute provides a mock function with given fields: toAttribute
func (_m *LdapInterface) MigrateIDAttribute(toAttribute string) error {
	ret := _m.Called(toAttribute)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(toAttribute)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RunTest provides a mock function with given fields:
func (_m *LdapInterface) RunTest() *model.AppError {
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

// StartSynchronizeJob provides a mock function with given fields: waitForJobToFinish
func (_m *LdapInterface) StartSynchronizeJob(waitForJobToFinish bool) (*model.Job, *model.AppError) {
	ret := _m.Called(waitForJobToFinish)

	var r0 *model.Job
	if rf, ok := ret.Get(0).(func(bool) *model.Job); ok {
		r0 = rf(waitForJobToFinish)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(bool) *model.AppError); ok {
		r1 = rf(waitForJobToFinish)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SwitchToLdap provides a mock function with given fields: userId, ldapId, ldapPassword
func (_m *LdapInterface) SwitchToLdap(userId string, ldapId string, ldapPassword string) *model.AppError {
	ret := _m.Called(userId, ldapId, ldapPassword)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string, string, string) *model.AppError); ok {
		r0 = rf(userId, ldapId, ldapPassword)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}
