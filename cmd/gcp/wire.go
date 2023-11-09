//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/internal/gcp"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func InitializeDataSourceSyncer(ctx context.Context, configMap *config.ConfigMap) (wrappers.DataSourceSyncer, func(), error) {
	wire.Build(
		gcp.Wired,
		org.Wired,

		wire.Bind(new(wrappers.DataSourceSyncer), new(*gcp.DataSourceSyncer)),
		wire.Bind(new(gcp.DataSourceRepository), new(*org.GcpRepository)),
	)

	return nil, nil, nil
}
