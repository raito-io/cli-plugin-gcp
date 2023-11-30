//go:build wireinject
// +build wireinject

package admin

import (
	"context"

	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"
)

var Wired = wire.NewSet(
	NewAdminRepository,

	NewGcpAdminService,
)

// TESTING

func InitializeAdminClient(ctx context.Context, configMap *config.ConfigMap) (*AdminRepository, func(), error) {
	wire.Build(
		Wired,
	)

	return nil, nil, nil
}
