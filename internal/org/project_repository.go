package org

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/googleapis/gax-go/v2"
	ds "github.com/raito-io/cli/base/data_source"
	"google.golang.org/api/iterator"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

const ownerRole = "roles/owner"
const editorRole = "roles/editor"
const viewerRole = "roles/viewer"

type projectClient interface {
	ListProjects(ctx context.Context, req *resourcemanagerpb.ListProjectsRequest, opts ...gax.CallOption) *resourcemanager.ProjectIterator
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
	SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}
type ProjectRepository struct {
	projectClient projectClient
}

func NewProjectRepository(projectClient projectClient) *ProjectRepository {
	return &ProjectRepository{
		projectClient: projectClient,
	}
}

func (r *ProjectRepository) GetProjects(ctx context.Context, _ *ds.DataSourceSyncConfig, parentName string, parent *GcpOrgEntity, fn func(ctx context.Context, project *GcpOrgEntity) error) error {
	projectIterator := r.projectClient.ListProjects(ctx, &resourcemanagerpb.ListProjectsRequest{
		Parent: parentName,
	})

	for {
		project, err := projectIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching project in %q: %s", parentName, err.Error()))

			continue
		} else if err != nil {
			return fmt.Errorf("project iterator: %w", err)
		}

		if errors.Is(err, iterator.Done) {
			break
		} else if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching project in %q: %s", parentName, err.Error()))

			continue
		} else if err != nil {
			return fmt.Errorf("project tags iterator: %w", err)
		}

		res := GcpOrgEntity{
			EntryName: project.Name,
			Name:      project.DisplayName,
			Id:        project.ProjectId,
			FullName:  project.ProjectId,
			Type:      "project",
			Parent:    parent,
			Tags:      project.Labels,
		}

		err = fn(ctx, &res)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ProjectRepository) GetProjectOwner(ctx context.Context, projectId string) (owner []string, editor []string, viewer []string, err error) {
	bindings, err := r.GetIamPolicy(ctx, projectId)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get iam bindings for project %s: %w", projectId, err)
	}

	for _, binding := range bindings {
		if binding.Role == ownerRole {
			owner = append(owner, binding.Member)
		} else if binding.Role == editorRole {
			editor = append(editor, binding.Member)
		} else if binding.Role == viewerRole {
			viewer = append(viewer, binding.Member)
		}
	}

	return owner, editor, viewer, nil
}

func (r *ProjectRepository) GetIamPolicy(ctx context.Context, projectId string) ([]iam.IamBinding, error) {
	return getAndParseBindings(ctx, r.projectClient, "project", projectId)
}

func (r *ProjectRepository) UpdateBinding(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding) error {
	dataObject.ObjectType = "project"

	return updateBindings(ctx, r.projectClient, dataObject, bindingsToAdd, bindingsToDelete)
}
