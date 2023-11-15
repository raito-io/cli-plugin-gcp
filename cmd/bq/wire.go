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

func InitializeDataSourceSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.DataSourceSyncer, func(), error) {
	wire.Build(
		bigquery.Wired,
		syncer.Wired,

		wire.Bind(new(wrappers.DataSourceSyncer), new(*syncer.DataSourceSyncer)),
		wire.Bind(new(syncer.DataSourceRepository), new(*bigquery.DataObjectIterator)),
	)

	return nil, nil, nil
}

func InitializeIdentityStoreSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.IdentityStoreSyncer, func(), error) {
	wire.Build(
		bigquery.Wired,
		admin.Wired,
		syncer.Wired,

		wire.Bind(new(wrappers.IdentityStoreSyncer), new(*syncer.IdentityStoreSyncer)),
		wire.Bind(new(syncer.AdminRepository), new(*admin.AdminRepository)),
		wire.Bind(new(syncer.DataObjectRepository), new(*bigquery.DataObjectIterator)),
	)

	return nil, nil, nil
}

func InitializeDataAccessSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.AccessProviderSyncer, func(), error) {
	wire.Build(
		bigquery.Wired,
		syncer.Wired,
		org.Wired,

		wire.Bind(new(wrappers.AccessProviderSyncer), new(*syncer.AccessSyncer)),
		wire.Bind(new(syncer.ProjectRepo), new(*org.ProjectRepository)),
		wire.Bind(new(syncer.BindingRepository), new(*bigquery.DataObjectIterator)),
		wire.Bind(new(syncer.MaskingService), new(*bigquery.BqMaskingService)),
	)

	return nil, nil, nil
}

func InitializeDataUsageSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.DataUsageSyncer, func(), error) {
	wire.Build(
		bigquery.Wired,
		syncer.Wired,

		wire.Bind(new(wrappers.DataUsageSyncer), new(*syncer.DataUsageSyncer)),
		wire.Bind(new(syncer.DataUsageRepository), new(*bigquery.Repository)),
	)

	return nil, nil, nil
}
