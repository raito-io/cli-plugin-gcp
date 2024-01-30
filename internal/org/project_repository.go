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
	iam2 "google.golang.org/api/iam/v1"
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

type serviceAccountClient interface {
	List(name string) *iam2.ProjectsServiceAccountsListCall
}

type ProjectRepository struct {
	projectClient        projectClient
	serviceAccountClient serviceAccountClient
}

func NewProjectRepository(projectClient projectClient, serviceAccountClient serviceAccountClient) *ProjectRepository {
	return &ProjectRepository{
		projectClient:        projectClient,
		serviceAccountClient: serviceAccountClient,
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

		res := GcpOrgEntity{
			EntryName: project.Name,
			Name:      project.DisplayName,
			Id:        project.ProjectId,
			FullName:  project.ProjectId,
			Type:      TypeProject,
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
	return getAndParseBindings(ctx, r.projectClient, TypeProject, projectId)
}

func (r *ProjectRepository) UpdateBinding(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding) error {
	dataObject.ObjectType = TypeProject

	return updateBindings(ctx, r.projectClient, dataObject, bindingsToAdd, bindingsToDelete)
}

func (r *ProjectRepository) GetUsers(ctx context.Context, projectEntryName string, fn func(ctx context.Context, entity *iam.UserEntity) error) error {
	nextPageToken := ""

	for {
		saCall := r.serviceAccountClient.List(projectEntryName).PageSize(64)

		if nextPageToken != "" {
			saCall = saCall.PageToken(nextPageToken)
		}

		serviceAccounts, err := saCall.Do()
		if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching users: %s", err.Error()))

			return nil
		} else if common.IsGoogle403Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 403 error while loading service accounts. Make sure Identity and Access Management (IAM) API is enabled and user has iam.serviceAccounts.list permission: %s", err.Error()))

			return nil
		} else if err != nil {
			return fmt.Errorf("listing service account: %s", err.Error())
		}

		for _, sa := range serviceAccounts.Accounts {
			err = fn(ctx, &iam.UserEntity{
				ExternalId: fmt.Sprintf("serviceAccount:%s", sa.Email),
				Name:       sa.DisplayName,
				Email:      sa.Email,
			})
			if err != nil {
				return err
			}
		}

		if serviceAccounts.NextPageToken != "" {
			nextPageToken = serviceAccounts.NextPageToken
		} else {
			break
		}
	}

	return nil
}
