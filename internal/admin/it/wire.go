//go:build wireinject
// +build wireinject

package it

import (
	"context"

	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/admin"
)

func InitializeAdminClient(ctx context.Context, configMap *config.ConfigMap) (*admin.AdminRepository, func(), error) {
	wire.Build(
		admin.Wired,
	)

	return nil, nil, nil
}
