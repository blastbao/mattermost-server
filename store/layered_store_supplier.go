// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package store

import "github.com/blastbao/mattermost-server/model"
import "context"

type LayeredStoreSupplierResult struct {
	StoreResult
}

func NewSupplierResult() *LayeredStoreSupplierResult {
	return &LayeredStoreSupplierResult{}
}

type LayeredStoreSupplier interface {
	//
	// Control
	//
	SetChainNext(LayeredStoreSupplier)
	Next() LayeredStoreSupplier

	// Roles
	RoleSave(ctx context.Context, role *model.Role, hints ...LayeredStoreHint) (*model.Role, *model.AppError)
	RoleGet(ctx context.Context, roleId string, hints ...LayeredStoreHint) (*model.Role, *model.AppError)
	RoleGetAll(ctx context.Context, hints ...LayeredStoreHint) ([]*model.Role, *model.AppError)
	RoleGetByName(ctx context.Context, name string, hints ...LayeredStoreHint) (*model.Role, *model.AppError)
	RoleGetByNames(ctx context.Context, names []string, hints ...LayeredStoreHint) ([]*model.Role, *model.AppError)
	RoleDelete(ctx context.Context, roldId string, hints ...LayeredStoreHint) (*model.Role, *model.AppError)
	RolePermanentDeleteAll(ctx context.Context, hints ...LayeredStoreHint) *model.AppError

	// Schemes
	SchemeSave(ctx context.Context, scheme *model.Scheme, hints ...LayeredStoreHint) (*model.Scheme, *model.AppError)
	SchemeGet(ctx context.Context, schemeId string, hints ...LayeredStoreHint) (*model.Scheme, *model.AppError)
	SchemeGetByName(ctx context.Context, schemeName string, hints ...LayeredStoreHint) (*model.Scheme, *model.AppError)
	SchemeDelete(ctx context.Context, schemeId string, hints ...LayeredStoreHint) (*model.Scheme, *model.AppError)
	SchemeGetAllPage(ctx context.Context, scope string, offset int, limit int, hints ...LayeredStoreHint) ([]*model.Scheme, *model.AppError)
	SchemePermanentDeleteAll(ctx context.Context, hints ...LayeredStoreHint) *model.AppError
}
