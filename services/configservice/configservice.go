// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package configservice

import (
	"crypto/ecdsa"

	"github.com/blastbao/mattermost-server/model"
)

// An interface representing something that contains a Config, such as the app.App struct
type ConfigService interface {
	Config() *model.Config
	AddConfigListener(func(old, current *model.Config)) string
	RemoveConfigListener(string)
	AsymmetricSigningKey() *ecdsa.PrivateKey
}
