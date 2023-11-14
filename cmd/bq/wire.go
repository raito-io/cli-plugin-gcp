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

//
//func InitializeDataAccessSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.AccessProviderSyncer, func(), error) {
//	wire.Build(
//		gcp.Wired,
//		org.Wired,
//
//		wire.Bind(new(wrappers.AccessProviderSyncer), new(*gcp.AccessSyncer)),
//		wire.Bind(new(gcp.ProjectRepo), new(*org.ProjectRepository)),
//		wire.Bind(new(gcp.GcpBindingRepository), new(*org.GcpDataObjectIterator)),
//	)
//
//	return nil, nil, nil
//}
