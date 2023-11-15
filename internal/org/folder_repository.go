package org

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"

	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

type folderClient interface {
	ListFolders(ctx context.Context, req *resourcemanagerpb.ListFoldersRequest, opts ...gax.CallOption) *resourcemanager.FolderIterator
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
	SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}

type FolderRepository struct {
	folderClient folderClient
}

func NewFolderRepository(folderClient folderClient) *FolderRepository {
	return &FolderRepository{
		folderClient: folderClient,
	}
}

func (r *FolderRepository) GetFolders(ctx context.Context, parentName string, parent *GcpOrgEntity, fn func(ctx context.Context, folder *GcpOrgEntity) error) error {
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

		id := strings.Split(folder.Name, "/")[1]

		res := GcpOrgEntity{
			EntryName: folder.Name,
			Name:      folder.DisplayName,
			Id:        id,
			FullName:  id,
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

func (r *FolderRepository) GetIamPolicy(ctx context.Context, folderId string) ([]iam.IamBinding, error) {
	return getAndParseBindings(ctx, r.folderClient, "folder", folderId)
}

func (r *FolderRepository) AddBinding(ctx context.Context, binding *iam.IamBinding) error {
	return addBinding(ctx, r.folderClient, binding)
}

func (r *FolderRepository) RemoveBinding(ctx context.Context, binding *iam.IamBinding) error {
	return removeBinding(ctx, r.folderClient, binding)
}
