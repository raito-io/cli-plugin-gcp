package org

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

//go:generate go run github.com/vektra/mockery/v2 --name=projectRepo --with-expecter --inpackage
type projectRepo interface {
	GetProjects(ctx context.Context, parentName string, parent *GcpOrgEntity, fn func(ctx context.Context, project *GcpOrgEntity) error) error
	GetIamPolicies(ctx context.Context, projectId string) ([]types.IamBinding, error)
}

//go:generate go run github.com/vektra/mockery/v2 --name=folderRepo --with-expecter --inpackage
type folderRepo interface {
	GetFolders(ctx context.Context, parentName string, parent *GcpOrgEntity, fn func(ctx context.Context, folder *GcpOrgEntity) error) error
	GetIamPolicies(ctx context.Context, folderId string) ([]types.IamBinding, error)
}

//go:generate go run github.com/vektra/mockery/v2 --name=organizationRepo --with-expecter --inpackage
type organizationRepo interface {
	GetOrganization(ctx context.Context) (*GcpOrgEntity, error)
	GetIamPolicies(ctx context.Context) ([]types.IamBinding, error)
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

func (r *GcpDataObjectIterator) DataObjects(ctx context.Context, fn func(ctx context.Context, object *GcpOrgEntity) error) error {
	return r.sync(ctx, fn)
}

func (r *GcpDataObjectIterator) UserAndGroups(ctx context.Context, userFn func(ctx context.Context, userId string) error, groupFn func(ctx context.Context, groupId string) error) error {
	groupsAndUsers := set.NewSet[string]()

	return r.sync(ctx, func(ctx context.Context, dataObject *GcpOrgEntity) error {
		var bindings []types.IamBinding
		var err error

		switch dataObject.Type {
		case "organization":
			bindings, err = r.organizationRepo.GetIamPolicies(ctx)
		case "folder":
			bindings, err = r.folderRepo.GetIamPolicies(ctx, dataObject.Id)
		case "project":
			bindings, err = r.projectRepo.GetIamPolicies(ctx, dataObject.Id)
		default:
			return fmt.Errorf("unknown data object type: %s", dataObject.Type)
		}

		if err != nil {
			return fmt.Errorf("get iam policies of (%s, %s): %w", dataObject.Type, dataObject.Id, err)
		}

		for _, binding := range bindings {
			if groupsAndUsers.Contains(binding.Member) {
				continue
			}

			groupsAndUsers.Add(binding.Member)

			if strings.HasPrefix(binding.Member, "user:") || strings.HasPrefix(binding.Member, "serviceAccount:") {
				err = userFn(ctx, binding.Member)
				if err != nil {
					return err
				}
			} else if strings.HasPrefix(binding.Member, "group:") {
				err = groupFn(ctx, binding.Member)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unknown member type: %s", binding.Member)
			}
		}

		return nil
	})
}

func (r *GcpDataObjectIterator) sync(ctx context.Context, fn func(ctx context.Context, dataObject *GcpOrgEntity) error) error {
	organization, err := r.organizationRepo.GetOrganization(ctx)
	if err != nil {
		return fmt.Errorf("get organization: %w", err)
	}

	err = fn(ctx, organization)
	if err != nil {
		return err
	}

	return r.syncFolder(ctx, organization.EntryName, organization, fn)
}

func (r *GcpDataObjectIterator) syncFolder(ctx context.Context, parentId string, parent *GcpOrgEntity, fn func(ctx context.Context, dataObject *GcpOrgEntity) error) error {
	err := r.projectRepo.GetProjects(ctx, parentId, parent, fn)
	if err != nil {
		return fmt.Errorf("project syncs of %q: %w", parentId, err)
	}

	return r.folderRepo.GetFolders(ctx, parentId, parent, func(ctx context.Context, folder *GcpOrgEntity) error {
		err2 := fn(ctx, folder)
		if err2 != nil {
			return fmt.Errorf("folder syncs of %q: %w", parentId, err2)
		}

		return r.syncFolder(ctx, folder.EntryName, folder, fn)
	})
}
