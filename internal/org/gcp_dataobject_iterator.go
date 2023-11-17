package org

import (
	"context"
	"errors"
	"fmt"

	ds "github.com/raito-io/cli/base/data_source"

	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

type iamRepo interface {
	GetIamPolicy(ctx context.Context, projectId string) ([]iam.IamBinding, error)
	UpdateBinding(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding) error
}

//go:generate go run github.com/vektra/mockery/v2 --name=projectRepo --with-expecter --inpackage
type projectRepo interface {
	iamRepo
	GetProjects(ctx context.Context, config *ds.DataSourceSyncConfig, parentName string, parent *GcpOrgEntity, fn func(ctx context.Context, project *GcpOrgEntity) error) error
}

//go:generate go run github.com/vektra/mockery/v2 --name=folderRepo --with-expecter --inpackage
type folderRepo interface {
	iamRepo
	GetFolders(ctx context.Context, parentName string, parent *GcpOrgEntity, fn func(ctx context.Context, folder *GcpOrgEntity) error) error
}

//go:generate go run github.com/vektra/mockery/v2 --name=organizationRepo --with-expecter --inpackage
type organizationRepo interface {
	iamRepo
	GetOrganization(ctx context.Context) (*GcpOrgEntity, error)
}

type GcpDataObjectIterator struct {
	projectRepo      projectRepo
	folderRepo       folderRepo
	organizationRepo organizationRepo

	organisationId string
}

func NewGcpDataObjectIterator(projectRepo projectRepo, folderRepo folderRepo, organzationRepo organizationRepo, configMap *config.ConfigMap) *GcpDataObjectIterator {
	return &GcpDataObjectIterator{
		projectRepo:      projectRepo,
		folderRepo:       folderRepo,
		organizationRepo: organzationRepo,

		organisationId: configMap.GetString(common.GcpOrgId),
	}
}

func (r *GcpDataObjectIterator) DataObjects(ctx context.Context, config *ds.DataSourceSyncConfig, fn func(ctx context.Context, object *GcpOrgEntity) error) error {
	return r.sync(ctx, config, fn)
}

func (r *GcpDataObjectIterator) Bindings(ctx context.Context, config *ds.DataSourceSyncConfig, fn func(ctx context.Context, dataObject *GcpOrgEntity, bindings []iam.IamBinding) error) error {
	return r.sync(ctx, config, func(ctx context.Context, dataObject *GcpOrgEntity) error {
		common.Logger.Debug(fmt.Sprintf("Fetch bindings for %s", dataObject.Id))

		var bindings []iam.IamBinding
		var err error

		repo := r.getIamRepository(dataObject.Type)
		if repo == nil {
			return fmt.Errorf("unknown data object type: %s", dataObject.Type)
		}

		bindings, err = repo.GetIamPolicy(ctx, dataObject.Id)
		if err != nil {
			return fmt.Errorf("get iam policies of (%s, %s): %w", dataObject.Type, dataObject.Id, err)
		}

		return fn(ctx, dataObject, bindings)
	})
}

func (r *GcpDataObjectIterator) UpdateBindings(ctx context.Context, dataObject *iam.DataObjectReference, addBindings []iam.IamBinding, removeBindings []iam.IamBinding) error {
	repo := r.getIamRepository(dataObject.ObjectType)
	if repo == nil {
		return fmt.Errorf("unknown data object type: %s", dataObject.ObjectType)
	}

	err := repo.UpdateBinding(ctx, dataObject, addBindings, removeBindings)
	if err != nil {
		return fmt.Errorf("update bindings of %s: %w", dataObject.FullName, err)
	}

	return nil
}

func (r *GcpDataObjectIterator) DataSourceType() string {
	return "organization"
}

func (r *GcpDataObjectIterator) sync(ctx context.Context, config *ds.DataSourceSyncConfig, fn func(ctx context.Context, dataObject *GcpOrgEntity) error) error {
	organization, err := r.organizationRepo.GetOrganization(ctx)
	if err != nil {
		return fmt.Errorf("get organization: %w", err)
	}

	if organization == nil {
		return errors.New("organization not found")
	}

	if common.ShouldHandle(organization.FullName, config) {
		err = fn(ctx, organization)
		if err != nil {
			return err
		}
	}

	if !common.ShouldGoInto(organization.FullName, config) {
		return nil
	}

	return r.syncFolder(ctx, config, organization.EntryName, organization, fn)
}

func (r *GcpDataObjectIterator) syncFolder(ctx context.Context, config *ds.DataSourceSyncConfig, parentId string, parent *GcpOrgEntity, fn func(ctx context.Context, dataObject *GcpOrgEntity) error) error {
	err := r.projectRepo.GetProjects(ctx, config, parentId, parent, fn)
	if err != nil {
		return fmt.Errorf("project syncs of %q: %w", parentId, err)
	}

	err = r.folderRepo.GetFolders(ctx, parentId, parent, func(ctx context.Context, folder *GcpOrgEntity) error {
		if common.ShouldHandle(folder.FullName, config) {
			err2 := fn(ctx, folder)
			if err2 != nil {
				return fmt.Errorf("folder syncs of %q: %w", parentId, err2)
			}
		}

		if !common.ShouldGoInto(folder.FullName, config) {
			return nil
		}

		return r.syncFolder(ctx, config, folder.EntryName, folder, fn)
	})
	if err != nil {
		return fmt.Errorf("folder syncs of %q: %w", parentId, err)
	}

	return nil
}

func (r *GcpDataObjectIterator) getIamRepository(resourceType string) iamRepo {
	switch resourceType {
	case "project":
		return r.projectRepo
	case "folder":
		return r.folderRepo
	case "organization":
		return r.organizationRepo
	default:
		return nil
	}
}
