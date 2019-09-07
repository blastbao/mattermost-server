package sqlstore

import (
	"testing"

	"github.com/blastbao/mattermost-server/store/storetest"
)

func TestUserTermsOfServiceStore(t *testing.T) {
	StoreTest(t, storetest.TestUserTermsOfServiceStore)
}
