package org

import (
	"context"
	"fmt"

	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

type AdminRepository interface {
	GetUsers(ctx context.Context, fn func(ctx context.Context, entity *iam.UserEntity) error) error
	GetGroups(ctx context.Context, fn func(ctx context.Context, entity *iam.GroupEntity) error) error
}

type gcpDataIterator interface {
	DataObjects(ctx context.Context, config *ds.DataSourceSyncConfig, fn func(ctx context.Context, object *GcpOrgEntity) error) error
}

type projectRepository interface {
	GetUsers(ctx context.Context, projectName string, fn func(ctx context.Context, entity *iam.UserEntity) error) error
}

type OrgIdenityStoreSyncer struct {
	configMap       *config.ConfigMap
	adminRepo       AdminRepository
	projectRepo     projectRepository
	gcpDataIterator gcpDataIterator
}

func NewOrgIdentityStoreSyncer(configMap *config.ConfigMap, adminRepo AdminRepository, projectRepo projectRepository, gcpDataIterator gcpDataIterator) *OrgIdenityStoreSyncer {
	return &OrgIdenityStoreSyncer{
		configMap:       configMap,
		adminRepo:       adminRepo,
		projectRepo:     projectRepo,
		gcpDataIterator: gcpDataIterator,
	}
}

func (r *OrgIdenityStoreSyncer) GetUsers(ctx context.Context, fn func(ctx context.Context, entity *iam.UserEntity) error) error {
	err := r.adminRepo.GetUsers(ctx, fn)
	if err != nil {
		return fmt.Errorf("get users in google admin: %w", err)
	}

	err = r.gcpDataIterator.DataObjects(ctx, &ds.DataSourceSyncConfig{ConfigMap: r.configMap}, func(ctx context.Context, object *GcpOrgEntity) error {
		if object.Type == "project" {
			getUserErr := r.projectRepo.GetUsers(ctx, object.Name, fn)
			if getUserErr != nil {
				return fmt.Errorf("get users in project %s: %w", object.Name, getUserErr)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("get service accounts: %w", err)
	}

	return nil
}

func (r *OrgIdenityStoreSyncer) GetGroups(ctx context.Context, fn func(ctx context.Context, entity *iam.GroupEntity) error) error {
	return r.adminRepo.GetGroups(ctx, fn)
}
