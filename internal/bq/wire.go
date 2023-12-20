//go:build wireinject
// +build wireinject

package bigquery

import (
	"context"

	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"
	bigquery2 "google.golang.org/api/bigquery/v2"

	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

var Wired = wire.NewSet(
	NewBiqQueryClient,
	NewPolicyTagClient,
	NewDataPolicyClient,
	NewServiceClient,
	NewRowAccessClient,

	NewRepository,
	NewDataCatalogRepository,
	NewDataObjectIterator,
	NewBqFilteringService,

	NewBqMaskingService,

	NewDataSourceMetaData,
	NewIdentityStoreMetadata,

	wire.Bind(new(maskingDataCatalogRepository), new(*DataCatalogRepository)),
	wire.Bind(new(dataCatalogBqRepository), new(*Repository)),
	wire.Bind(new(filteringRepository), new(*Repository)),
	wire.Bind(new(filteringDataObjectIterator), new(*DataObjectIterator)),
	wire.Bind(new(BigQueryRowAccessPoliciesService), new(*bigquery2.RowAccessPoliciesService)),
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
