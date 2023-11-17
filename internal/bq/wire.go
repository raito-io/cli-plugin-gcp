//go:build wireinject
// +build wireinject

package bigquery

import (
	"github.com/google/wire"
)

var Wired = wire.NewSet(
	NewBiqQueryClient,
	NewPolicyTagClient,
	NewDataPolicyClient,

	NewRepository,
	NewDataCatalogRepository,
	NewDataObjectIterator,

	NewBqMaskingService,

	NewDataSourceMetaData,
	NewIdentityStoreMetadata,
)
