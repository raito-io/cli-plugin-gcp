//go:build wireinject
// +build wireinject

package bigquery

import (
	"context"

	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/org"
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

	wire.Bind(new(maskingDataCatalogRepository), new(*DataCatalogRepository)),
	wire.Bind(new(dataCatalogBqRepository), new(*Repository)),
)

// TESTING

func InitializeBigqueryRepository(ctx context.Context, configMap *config.ConfigMap) (*TestRepositoryAndClient, func(), error) {
	wire.Build(
		Wired,
		org.Wired,

		wire.Bind(new(ProjectClient), new(*org.ProjectRepository)),

		wire.Struct(new(TestRepositoryAndClient), "Repository", "Client"),

		wire.Value(&RepositoryOptions{EnableCache: false}),
	)

	return nil, nil, nil
}

func InitializeDataObjectIterator(ctx context.Context, configMap *config.ConfigMap) (*DataObjectIterator, func(), error) {
	wire.Build(
		Wired,
		org.Wired,

		wire.Bind(new(ProjectClient), new(*org.ProjectRepository)),

		wire.Value(&RepositoryOptions{EnableCache: false}),
	)

	return nil, nil, nil
}

func InitializeDataCatalogRepository(ctx context.Context, configMap *config.ConfigMap) (*DataCatalogRepository, func(), error) {
	wire.Build(
		Wired,
		org.Wired,

		wire.Bind(new(ProjectClient), new(*org.ProjectRepository)),

		wire.Value(&RepositoryOptions{EnableCache: false}),
	)

	return nil, nil, nil
}
