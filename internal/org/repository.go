package org

import (
	"context"
	"errors"
	"fmt"
	"strings"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
)

type projectClient interface {
	ListProjects(ctx context.Context, req *resourcemanagerpb.ListProjectsRequest, opts ...gax.CallOption) *resourcemanager.ProjectIterator
}

type folderClient interface {
	ListFolders(ctx context.Context, req *resourcemanagerpb.ListFoldersRequest, opts ...gax.CallOption) *resourcemanager.FolderIterator
}

type GcpRepository struct {
	projectClient projectClient
	folderClient  folderClient
}

func NewGcpRepository(projectClient projectClient, folderClient folderClient) *GcpRepository {
	return &GcpRepository{
		projectClient: projectClient,
		folderClient:  folderClient,
	}
}

func (r *GcpRepository) GetProjects(ctx context.Context, parentName string, parent *GcpOrgEntity, fn func(ctx context.Context, project *GcpOrgEntity) error) error {
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

func (r *GcpRepository) GetFolders(ctx context.Context, parentName string, parent *GcpOrgEntity, fn func(ctx context.Context, folder *GcpOrgEntity) error) error {
	folderIterator := r.folderClient.ListFolders(ctx, &resourcemanagerpb.ListFoldersRequest{
		Parent: parentName,
	})

	for {
		folder, err := folderIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return fmt.Errorf("folder iterator: %w", err)
		}

		res := GcpOrgEntity{
			EntryName: folder.Name,
			Name:      folder.DisplayName,
			Id:        strings.Split(folder.Name, "/")[1],
			Type:      "folder",
			Parent:    parent,
		}

		err = fn(ctx, &res)
		if err != nil {
			return err
		}
	}

	return nil
}
