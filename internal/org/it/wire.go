//go:build wireinject
// +build wireinject

package it

import (
	"context"

	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func InitializeFolderRepository(ctx context.Context, configMap *config.ConfigMap) (*org.FolderRepository, func(), error) {
	wire.Build(
		org.Wired,
	)

	return nil, nil, nil
}

func InitializeOrganizationRepository(ctx context.Context, configMap *config.ConfigMap) (*org.OrganizationRepository, func(), error) {
	wire.Build(
		org.Wired,
	)

	return nil, nil, nil
}

func InitializeProjectRepository(ctx context.Context, configMap *config.ConfigMap) (*org.ProjectRepository, func(), error) {
	wire.Build(
		org.Wired,
	)

	return nil, nil, nil
}