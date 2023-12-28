//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/internal/admin"
	bigquery "github.com/raito-io/cli-plugin-gcp/internal/bq"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
	"github.com/raito-io/cli-plugin-gcp/internal/syncer"
)

var optionSet = wire.NewSet(
	wire.Value(&bigquery.RepositoryOptions{EnableCache: true}),
)

func InitializeDataSourceSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.DataSourceSyncer, func(), error) {
	wire.Build(
		optionSet,
		bigquery.Wired,
		org.Wired,
		syncer.Wired,

		wire.Bind(new(wrappers.DataSourceSyncer), new(*syncer.DataSourceSyncer)),
		wire.Bind(new(syncer.DataSourceRepository), new(*bigquery.DataObjectIterator)),
		wire.Bind(new(bigquery.ProjectClient), new(*org.ProjectRepository)),
	)

	return nil, nil, nil
}

func InitializeIdentityStoreSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.IdentityStoreSyncer, func(), error) {
	wire.Build(
		optionSet,
		bigquery.Wired,
		org.Wired,
		admin.Wired,
		syncer.Wired,

		wire.Bind(new(wrappers.IdentityStoreSyncer), new(*syncer.IdentityStoreSyncer)),
		wire.Bind(new(syncer.AdminRepository), new(*admin.AdminRepository)),
		wire.Bind(new(syncer.DataObjectRepository), new(*bigquery.DataObjectIterator)),
		wire.Bind(new(bigquery.ProjectClient), new(*org.ProjectRepository)),
	)

	return nil, nil, nil
}

func InitializeDataAccessSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.AccessProviderSyncer, func(), error) {
	wire.Build(
		optionSet,
		bigquery.Wired,
		syncer.Wired,
		org.Wired,

		wire.Bind(new(wrappers.AccessProviderSyncer), new(*syncer.AccessSyncer)),
		wire.Bind(new(syncer.ProjectRepo), new(*org.ProjectRepository)),
		wire.Bind(new(syncer.BindingRepository), new(*bigquery.DataObjectIterator)),
		wire.Bind(new(syncer.MaskingService), new(*bigquery.BqMaskingService)),
		wire.Bind(new(bigquery.ProjectClient), new(*org.ProjectRepository)),
		wire.Bind(new(syncer.FilteringService), new(*bigquery.BqFilteringService)),
	)

	return nil, nil, nil
}

func InitializeDataUsageSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.DataUsageSyncer, func(), error) {
	wire.Build(
		optionSet,
		bigquery.Wired,
		syncer.Wired,
		org.Wired,

		wire.Bind(new(wrappers.DataUsageSyncer), new(*syncer.DataUsageSyncer)),
		wire.Bind(new(syncer.DataUsageRepository), new(*bigquery.Repository)),
		wire.Bind(new(bigquery.ProjectClient), new(*org.ProjectRepository)),
	)

	return nil, nil, nil
}
