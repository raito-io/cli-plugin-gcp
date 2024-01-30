//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/internal/admin"
	"github.com/raito-io/cli-plugin-gcp/internal/gcp"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
	"github.com/raito-io/cli-plugin-gcp/internal/syncer"
)

func InitializeDataSourceSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.DataSourceSyncer, func(), error) {
	wire.Build(
		gcp.Wired,
		syncer.Wired,
		org.Wired,

		wire.Bind(new(wrappers.DataSourceSyncer), new(*syncer.DataSourceSyncer)),
		wire.Bind(new(syncer.DataSourceRepository), new(*org.GcpDataObjectIterator)),
	)

	return nil, nil, nil
}

func InitializeIdentityStoreSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.IdentityStoreSyncer, func(), error) {
	wire.Build(
		gcp.Wired,
		syncer.Wired,
		admin.Wired,

		wire.Bind(new(wrappers.IdentityStoreSyncer), new(*syncer.IdentityStoreSyncer)),
		wire.Bind(new(syncer.AdminRepository), new(*admin.AdminRepository)),
	)

	return nil, nil, nil
}

func InitializeDataAccessSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.AccessProviderSyncer, func(), error) {
	wire.Build(
		gcp.Wired,
		org.Wired,
		syncer.Wired,

		wire.Bind(new(wrappers.AccessProviderSyncer), new(*syncer.AccessSyncer)),
		wire.Bind(new(syncer.ProjectRepo), new(*org.ProjectRepository)),
		wire.Bind(new(syncer.BindingRepository), new(*org.GcpDataObjectIterator)),
		wire.Bind(new(syncer.MaskingService), new(*gcp.NoMasking)),
		wire.Bind(new(syncer.FilteringService), new(*gcp.NoFiltering)),
	)

	return nil, nil, nil
}
