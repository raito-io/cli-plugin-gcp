//go:build wireinject
// +build wireinject

package it

import (
	"context"

	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"

	bigquery "github.com/raito-io/cli-plugin-gcp/internal/bq"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func InitializeBigqueryRepository(ctx context.Context, configMap *config.ConfigMap) (*TestRepositoryAndClient, func(), error) {
	wire.Build(
		bigquery.Wired,
		org.Wired,

		wire.Bind(new(bigquery.ProjectClient), new(*org.ProjectRepository)),

		wire.Struct(new(TestRepositoryAndClient), "Repository", "Client"),
	)

	return nil, nil, nil
}
