package org

import (
	"context"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/raito-io/cli/base/util/config"
	"google.golang.org/api/option"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

func NewProjectsClient(ctx context.Context, configMap *config.ConfigMap) (*resourcemanager.ProjectsClient, func(), error) {
	c, err := resourcemanager.NewProjectsClient(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)))
	if err != nil {
		return nil, nil, err
	}

	return c, func() { c.Close() }, nil
}

func NewFoldersClient(ctx context.Context, configMap *config.ConfigMap) (*resourcemanager.FoldersClient, func(), error) {
	c, err := resourcemanager.NewFoldersClient(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)))
	if err != nil {
		return nil, nil, err
	}

	return c, func() { c.Close() }, nil
}

func NewOrganizationsClient(ctx context.Context, configMap *config.ConfigMap) (*resourcemanager.OrganizationsClient, func(), error) {
	c, err := resourcemanager.NewOrganizationsClient(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)))
	if err != nil {
		return nil, nil, err
	}

	return c, func() { c.Close() }, nil
}
