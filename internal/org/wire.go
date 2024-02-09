//go:build wireinject
// +build wireinject

package org

import (
	"context"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/google/wire"
	"github.com/raito-io/cli/base/util/config"
	iam2 "google.golang.org/api/iam/v1"
)

var Wired = wire.NewSet(
	NewProjectsClient,
	NewFoldersClient,
	NewOrganizationsClient,
	NewIamClient,

	NewFolderRepository,
	NewProjectRepository,
	NewOrganizationRepository,
	NewGcpDataObjectIterator,
	NewOrgIdentityStoreSyncer,

	wire.Bind(new(projectClient), new(*resourcemanager.ProjectsClient)),
	wire.Bind(new(folderClient), new(*resourcemanager.FoldersClient)),
	wire.Bind(new(organizationClient), new(*resourcemanager.OrganizationsClient)),
	wire.Bind(new(projectRepo), new(*ProjectRepository)),
	wire.Bind(new(folderRepo), new(*FolderRepository)),
	wire.Bind(new(organizationRepo), new(*OrganizationRepository)),
	wire.Bind(new(gcpDataIterator), new(*GcpDataObjectIterator)),
	wire.Bind(new(projectRepository), new(*ProjectRepository)),
	wire.Bind(new(serviceAccountClient), new(*iam2.ProjectsServiceAccountsService)),
)

// TESTING

func InitializeFolderRepository(ctx context.Context, configMap *config.ConfigMap) (*FolderRepository, func(), error) {
	wire.Build(
		Wired,
	)

	return nil, nil, nil
}

func InitializeOrganizationRepository(ctx context.Context, configMap *config.ConfigMap) (*OrganizationRepository, func(), error) {
	wire.Build(
		Wired,
	)

	return nil, nil, nil
}

func InitializeProjectRepository(ctx context.Context, configMap *config.ConfigMap) (*ProjectRepository, func(), error) {
	wire.Build(
		Wired,
	)

	return nil, nil, nil
}
