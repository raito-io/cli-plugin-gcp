package org

import (
	"context"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/raito-io/cli/base/util/config"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

func NewProjectsClient(ctx context.Context, configMap *config.ConfigMap) (*resourcemanager.ProjectsClient, func(), error) {
	c, err := resourcemanager.NewProjectsClient(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)))
	if err != nil {
		return nil, nil, fmt.Errorf("new projects client: %w", err)
	}

	return c, func() { c.Close() }, nil
}

func NewFoldersClient(ctx context.Context, configMap *config.ConfigMap) (*resourcemanager.FoldersClient, func(), error) {
	c, err := resourcemanager.NewFoldersClient(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)))
	if err != nil {
		return nil, nil, fmt.Errorf("new folders client: %w", err)
	}

	return c, func() { c.Close() }, nil
}

func NewOrganizationsClient(ctx context.Context, configMap *config.ConfigMap) (*resourcemanager.OrganizationsClient, func(), error) {
	c, err := resourcemanager.NewOrganizationsClient(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)))
	if err != nil {
		return nil, nil, fmt.Errorf("new organizations client: %w", err)
	}

	return c, func() { c.Close() }, nil
}

func NewIamClient(ctx context.Context, configMap *config.ConfigMap) (*iam.ProjectsServiceAccountsService, error) {
	c, err := iam.NewService(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)))
	if err != nil {
		return nil, fmt.Errorf("new iam client: %w", err)
	}

	return c.Projects.ServiceAccounts, nil
}
