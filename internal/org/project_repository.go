package org

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"

	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

type projectClient interface {
	ListProjects(ctx context.Context, req *resourcemanagerpb.ListProjectsRequest, opts ...gax.CallOption) *resourcemanager.ProjectIterator
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}

type ProjectRepository struct {
	projectClient projectClient
}

func NewProjectRepository(projectClient projectClient) *ProjectRepository {
	return &ProjectRepository{
		projectClient: projectClient,
	}
}

func (r *ProjectRepository) GetProjects(ctx context.Context, parentName string, parent *GcpOrgEntity, fn func(ctx context.Context, project *GcpOrgEntity) error) error {
	projectIterator := r.projectClient.ListProjects(ctx, &resourcemanagerpb.ListProjectsRequest{
		Parent: parentName,
	})

	for {
		project, err := projectIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return fmt.Errorf("project iterator: %w", err)
		}

		res := GcpOrgEntity{
			EntryName: project.Name,
			Name:      project.DisplayName,
			Id:        project.ProjectId,
			Type:      "project",
			Parent:    parent,
		}

		err = fn(ctx, &res)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ProjectRepository) GetIamPolicies(ctx context.Context, projectId string) ([]types.IamBinding, error) {
	return parseBindings(ctx, r.projectClient, "project", projectId)
}
