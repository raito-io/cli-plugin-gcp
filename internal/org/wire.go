//go:build wireinject
// +build wireinject

package org

import (
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/google/wire"
)

var Wired = wire.NewSet(
	NewProjectsClient,
	NewFoldersClient,
	NewOrganizationsClient,

	NewFolderRepository,
	NewProjectRepository,
	NewOrganizationRepository,
	NewGcpDataObjectIterator,

	wire.Bind(new(projectClient), new(*resourcemanager.ProjectsClient)),
	wire.Bind(new(folderClient), new(*resourcemanager.FoldersClient)),
	wire.Bind(new(organizationClient), new(*resourcemanager.OrganizationsClient)),
	wire.Bind(new(projectRepo), new(*ProjectRepository)),
	wire.Bind(new(folderRepo), new(*FolderRepository)),
	wire.Bind(new(organizationRepo), new(*OrganizationRepository)),
)
