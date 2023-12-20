//go:build wireinject
// +build wireinject

package gcp

import (
	"github.com/google/wire"
)

var Wired = wire.NewSet(
	NewDataSourceMetaData,
	NewIdentityStoreMetadata,
	NewNoMasking,
	NewNoFiltering,
)
