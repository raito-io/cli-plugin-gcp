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

func InitializeBqRepository(ctx context.Context, configMap *config.ConfigMap) (*bigquery.Repository, func(), error) {
	wire.Build(
		bigquery.Wired,
		org.Wired,

		wire.Bind(new(bigquery.ProjectClient), new(*org.ProjectRepository)),
	)

	return nil, nil, nil
}
