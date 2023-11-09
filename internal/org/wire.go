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

	NewGcpRepository,

	wire.Bind(new(projectClient), new(*resourcemanager.ProjectsClient)),
	wire.Bind(new(folderClient), new(*resourcemanager.FoldersClient)),
)
