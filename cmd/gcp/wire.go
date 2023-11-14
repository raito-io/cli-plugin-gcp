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
)

func InitializeDataSourceSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.DataSourceSyncer, func(), error) {
	wire.Build(
		gcp.Wired,
		org.Wired,

		wire.Bind(new(wrappers.DataSourceSyncer), new(*gcp.DataSourceSyncer)),
		wire.Bind(new(gcp.DataSourceRepository), new(*org.GcpDataObjectIterator)),
	)

	return nil, nil, nil
}

func InitializeIdentityStoreSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.IdentityStoreSyncer, func(), error) {
	wire.Build(
		gcp.Wired,
		admin.Wired,
		org.Wired,

		wire.Bind(new(wrappers.IdentityStoreSyncer), new(*gcp.IdentityStoreSyncer)),
		wire.Bind(new(gcp.AdminRepository), new(*admin.AdminRepository)),
		wire.Bind(new(gcp.DataObjectRepository), new(*org.GcpDataObjectIterator)),
	)

	return nil, nil, nil
}

func InitializeDataAccessSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.AccessProviderSyncer, func(), error) {
	wire.Build(
		gcp.Wired,
		org.Wired,

		wire.Bind(new(wrappers.AccessProviderSyncer), new(*gcp.AccessSyncer)),
		wire.Bind(new(gcp.ProjectRepo), new(*org.ProjectRepository)),
		wire.Bind(new(gcp.GcpBindingRepository), new(*org.GcpDataObjectIterator)),
	)

	return nil, nil, nil
}
