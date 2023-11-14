package gcp

import is "github.com/raito-io/cli/base/identity_store"

func NewIdentityStoreMetadata() *is.MetaData {
	return &is.MetaData{
		Type:        "gcp",
		CanBeLinked: true,
		CanBeMaster: true,
	}
}
