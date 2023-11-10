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

	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

type folderClient interface {
	ListFolders(ctx context.Context, req *resourcemanagerpb.ListFoldersRequest, opts ...gax.CallOption) *resourcemanager.FolderIterator
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
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

func (r *FolderRepository) GetIamPolicies(ctx context.Context, folderId string) ([]types.IamBinding, error) {
	return parseBindings(ctx, r.folderClient, "folder", folderId)
}
