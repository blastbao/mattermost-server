// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import context "context"
import mock "github.com/stretchr/testify/mock"
import model "github.com/blastbao/mattermost-server/model"
import store "github.com/blastbao/mattermost-server/store"

// LayeredStoreSupplier is an autogenerated mock type for the LayeredStoreSupplier type
type LayeredStoreSupplier struct {
	mock.Mock
}

// Next provides a mock function with given fields:
func (_m *LayeredStoreSupplier) Next() store.LayeredStoreSupplier {
	ret := _m.Called()

	var r0 store.LayeredStoreSupplier
	if rf, ok := ret.Get(0).(func() store.LayeredStoreSupplier); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(store.LayeredStoreSupplier)
		}
	}

	return r0
}

// RoleDelete provides a mock function with given fields: ctx, roldId, hints
func (_m *LayeredStoreSupplier) RoleDelete(ctx context.Context, roldId string, hints ...store.LayeredStoreHint) (*model.Role, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, roldId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.Role
	if rf, ok := ret.Get(0).(func(context.Context, string, ...store.LayeredStoreHint) *model.Role); ok {
		r0 = rf(ctx, roldId, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Role)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, string, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, roldId, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// RoleGet provides a mock function with given fields: ctx, roleId, hints
func (_m *LayeredStoreSupplier) RoleGet(ctx context.Context, roleId string, hints ...store.LayeredStoreHint) (*model.Role, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, roleId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.Role
	if rf, ok := ret.Get(0).(func(context.Context, string, ...store.LayeredStoreHint) *model.Role); ok {
		r0 = rf(ctx, roleId, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Role)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, string, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, roleId, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// RoleGetAll provides a mock function with given fields: ctx, hints
func (_m *LayeredStoreSupplier) RoleGetAll(ctx context.Context, hints ...store.LayeredStoreHint) ([]*model.Role, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []*model.Role
	if rf, ok := ret.Get(0).(func(context.Context, ...store.LayeredStoreHint) []*model.Role); ok {
		r0 = rf(ctx, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Role)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// RoleGetByName provides a mock function with given fields: ctx, name, hints
func (_m *LayeredStoreSupplier) RoleGetByName(ctx context.Context, name string, hints ...store.LayeredStoreHint) (*model.Role, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.Role
	if rf, ok := ret.Get(0).(func(context.Context, string, ...store.LayeredStoreHint) *model.Role); ok {
		r0 = rf(ctx, name, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Role)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, string, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, name, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// RoleGetByNames provides a mock function with given fields: ctx, names, hints
func (_m *LayeredStoreSupplier) RoleGetByNames(ctx context.Context, names []string, hints ...store.LayeredStoreHint) ([]*model.Role, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, names)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []*model.Role
	if rf, ok := ret.Get(0).(func(context.Context, []string, ...store.LayeredStoreHint) []*model.Role); ok {
		r0 = rf(ctx, names, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Role)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, []string, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, names, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// RolePermanentDeleteAll provides a mock function with given fields: ctx, hints
func (_m *LayeredStoreSupplier) RolePermanentDeleteAll(ctx context.Context, hints ...store.LayeredStoreHint) *model.AppError {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(context.Context, ...store.LayeredStoreHint) *model.AppError); ok {
		r0 = rf(ctx, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// RoleSave provides a mock function with given fields: ctx, role, hints
func (_m *LayeredStoreSupplier) RoleSave(ctx context.Context, role *model.Role, hints ...store.LayeredStoreHint) (*model.Role, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, role)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.Role
	if rf, ok := ret.Get(0).(func(context.Context, *model.Role, ...store.LayeredStoreHint) *model.Role); ok {
		r0 = rf(ctx, role, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Role)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, *model.Role, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, role, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SchemeDelete provides a mock function with given fields: ctx, schemeId, hints
func (_m *LayeredStoreSupplier) SchemeDelete(ctx context.Context, schemeId string, hints ...store.LayeredStoreHint) (*model.Scheme, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, schemeId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.Scheme
	if rf, ok := ret.Get(0).(func(context.Context, string, ...store.LayeredStoreHint) *model.Scheme); ok {
		r0 = rf(ctx, schemeId, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Scheme)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, string, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, schemeId, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SchemeGet provides a mock function with given fields: ctx, schemeId, hints
func (_m *LayeredStoreSupplier) SchemeGet(ctx context.Context, schemeId string, hints ...store.LayeredStoreHint) (*model.Scheme, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, schemeId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.Scheme
	if rf, ok := ret.Get(0).(func(context.Context, string, ...store.LayeredStoreHint) *model.Scheme); ok {
		r0 = rf(ctx, schemeId, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Scheme)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, string, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, schemeId, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SchemeGetAllPage provides a mock function with given fields: ctx, scope, offset, limit, hints
func (_m *LayeredStoreSupplier) SchemeGetAllPage(ctx context.Context, scope string, offset int, limit int, hints ...store.LayeredStoreHint) ([]*model.Scheme, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, scope, offset, limit)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []*model.Scheme
	if rf, ok := ret.Get(0).(func(context.Context, string, int, int, ...store.LayeredStoreHint) []*model.Scheme); ok {
		r0 = rf(ctx, scope, offset, limit, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Scheme)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, string, int, int, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, scope, offset, limit, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SchemeGetByName provides a mock function with given fields: ctx, schemeName, hints
func (_m *LayeredStoreSupplier) SchemeGetByName(ctx context.Context, schemeName string, hints ...store.LayeredStoreHint) (*model.Scheme, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, schemeName)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.Scheme
	if rf, ok := ret.Get(0).(func(context.Context, string, ...store.LayeredStoreHint) *model.Scheme); ok {
		r0 = rf(ctx, schemeName, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Scheme)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, string, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, schemeName, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SchemePermanentDeleteAll provides a mock function with given fields: ctx, hints
func (_m *LayeredStoreSupplier) SchemePermanentDeleteAll(ctx context.Context, hints ...store.LayeredStoreHint) *model.AppError {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(context.Context, ...store.LayeredStoreHint) *model.AppError); ok {
		r0 = rf(ctx, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// SchemeSave provides a mock function with given fields: ctx, scheme, hints
func (_m *LayeredStoreSupplier) SchemeSave(ctx context.Context, scheme *model.Scheme, hints ...store.LayeredStoreHint) (*model.Scheme, *model.AppError) {
	_va := make([]interface{}, len(hints))
	for _i := range hints {
		_va[_i] = hints[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, scheme)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.Scheme
	if rf, ok := ret.Get(0).(func(context.Context, *model.Scheme, ...store.LayeredStoreHint) *model.Scheme); ok {
		r0 = rf(ctx, scheme, hints...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Scheme)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, *model.Scheme, ...store.LayeredStoreHint) *model.AppError); ok {
		r1 = rf(ctx, scheme, hints...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SetChainNext provides a mock function with given fields: _a0
func (_m *LayeredStoreSupplier) SetChainNext(_a0 store.LayeredStoreSupplier) {
	_m.Called(_a0)
}
