package bigquery

import (
	is "github.com/raito-io/cli/base/identity_store"
)

func NewIdentityStoreMetadata() *is.MetaData {
	return &is.MetaData{
		Type:        "bigquery",
		CanBeLinked: false,
		CanBeMaster: false,
	}
}
