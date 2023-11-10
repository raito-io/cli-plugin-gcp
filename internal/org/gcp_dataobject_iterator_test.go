package org

import (
	"context"
	"errors"
	"testing"

	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

func TestGcpDataObjectIterator_DataObjects(t *testing.T) {
	type fields struct {
		organisationId string
		mockSetup      func(projectRepo *mockProjectRepo, folderRepo *mockFolderRepo)
	}
	tests := []struct {
		name            string
		fields          fields
		expectedObjects []*GcpOrgEntity
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "No dataobjects",
			fields: fields{
				organisationId: "orgId",
				mockSetup: func(projectRepo *mockProjectRepo, folderRepo *mockFolderRepo) {
					projectRepo.EXPECT().GetProjects(mock.Anything, "organizations/orgId", (*GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return nil
					})

					folderRepo.EXPECT().GetFolders(mock.Anything, "organizations/orgId", (*GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return nil
					})
				},
			},
			expectedObjects: []*GcpOrgEntity{},
			wantErr:         assert.NoError,
		},
		{
			name: "Only projects",
			fields: fields{
				organisationId: "orgId",
				mockSetup: func(projectRepo *mockProjectRepo, folderRepo *mockFolderRepo) {
					projectRepo.EXPECT().GetProjects(mock.Anything, "organizations/orgId", (*GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						err := f(ctx,
							&GcpOrgEntity{
								EntryName: "projects/projectId1",
								Id:        "projectId1",
								Name:      "projectName1",
								Type:      "project",
								Parent:    nil,
							},
						)
						if err != nil {
							return err
						}

						return f(ctx,
							&GcpOrgEntity{
								EntryName: "projects/projectId2",
								Id:        "projectId2",
								Name:      "projectName2",
								Type:      "project",
								Parent:    nil,
							},
						)
					})

					folderRepo.EXPECT().GetFolders(mock.Anything, "organizations/orgId", (*GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return nil
					})
				},
			},
			expectedObjects: []*GcpOrgEntity{
				{
					EntryName: "projects/projectId1",
					Id:        "projectId1",
					Name:      "projectName1",
					Type:      "project",
					Parent:    nil,
				},
				{
					EntryName: "projects/projectId2",
					Id:        "projectId2",
					Name:      "projectName2",
					Type:      "project",
					Parent:    nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Folder and projects",
			fields: fields{
				organisationId: "orgId",
				mockSetup: func(projectRepo *mockProjectRepo, folderRepo *mockFolderRepo) {
					projectRepo.EXPECT().GetProjects(mock.Anything, "organizations/orgId", (*GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return f(ctx,
							&GcpOrgEntity{
								EntryName: "projects/projectId1",
								Id:        "projectId1",
								Name:      "projectName1",
								Type:      "project",
								Parent:    nil,
							},
						)
					})

					folderRepo.EXPECT().GetFolders(mock.Anything, "organizations/orgId", (*GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						err := f(ctx,
							&GcpOrgEntity{
								EntryName: "folders/folder1",
								Id:        "folderId1",
								Name:      "folderName1",
								Type:      "folder",
								Parent:    nil,
							},
						)
						if err != nil {
							return err
						}

						return f(ctx,
							&GcpOrgEntity{
								EntryName: "folders/folder2",
								Id:        "folderId2",
								Name:      "folderName2",
								Type:      "folder",
								Parent:    nil,
							},
						)
					})

					projectRepo.EXPECT().GetProjects(mock.Anything, "folders/folder1", &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return f(ctx,
							&GcpOrgEntity{
								EntryName: "projects/projectId2",
								Id:        "projectId2",
								Name:      "projectName2",
								Type:      "project",
								Parent:    &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil},
							},
						)
					})

					folderRepo.EXPECT().GetFolders(mock.Anything, "folders/folder1", &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return f(ctx,
							&GcpOrgEntity{
								EntryName: "folders/folder3",
								Id:        "folderId3",
								Name:      "folderName3",
								Type:      "folder",
								Parent:    &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil},
							},
						)
					})

					projectRepo.EXPECT().GetProjects(mock.Anything, "folders/folder2", &GcpOrgEntity{EntryName: "folders/folder2", Id: "folderId2", Name: "folderName2", Type: "folder", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return nil
					})

					folderRepo.EXPECT().GetFolders(mock.Anything, "folders/folder2", &GcpOrgEntity{EntryName: "folders/folder2", Id: "folderId2", Name: "folderName2", Type: "folder", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return nil
					})

					projectRepo.EXPECT().GetProjects(mock.Anything, "folders/folder3", &GcpOrgEntity{EntryName: "folders/folder3", Id: "folderId3", Name: "folderName3", Type: "folder", Parent: &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil}}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return f(ctx, &GcpOrgEntity{
							EntryName: "projects/projectId3",
							Id:        "projectId3",
							Name:      "projectName3",
							Type:      "project",
							Parent:    &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil},
						})
					})

					folderRepo.EXPECT().GetFolders(mock.Anything, "folders/folder3", &GcpOrgEntity{EntryName: "folders/folder3", Id: "folderId3", Name: "folderName3", Type: "folder", Parent: &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil}}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return nil
					})
				},
			},
			expectedObjects: []*GcpOrgEntity{
				{
					EntryName: "projects/projectId1",
					Id:        "projectId1",
					Name:      "projectName1",
					Type:      "project",
					Parent:    nil,
				},
				{
					EntryName: "folders/folder1",
					Id:        "folderId1",
					Name:      "folderName1",
					Type:      "folder",
					Parent:    nil,
				},
				{
					EntryName: "folders/folder2",
					Id:        "folderId2",
					Name:      "folderName2",
					Type:      "folder",
					Parent:    nil,
				},
				{
					EntryName: "projects/projectId2",
					Id:        "projectId2",
					Name:      "projectName2",
					Type:      "project",
					Parent:    &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil},
				},
				{
					EntryName: "folders/folder3",
					Id:        "folderId3",
					Name:      "folderName3",
					Type:      "folder",
					Parent:    &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil},
				},
				{
					EntryName: "projects/projectId3",
					Id:        "projectId3",
					Name:      "projectName3",
					Type:      "project",
					Parent:    &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "processing errors",
			fields: fields{
				organisationId: "orgId",
				mockSetup: func(projectRepo *mockProjectRepo, folderRepo *mockFolderRepo) {
					projectRepo.EXPECT().GetProjects(mock.Anything, "organizations/orgId", (*GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return nil
					})

					folderRepo.EXPECT().GetFolders(mock.Anything, "organizations/orgId", (*GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return f(ctx,
							&GcpOrgEntity{
								EntryName: "folders/folder1",
								Id:        "folderId1",
								Name:      "folderName1",
								Type:      "folder",
								Parent:    nil,
							},
						)
					})

					projectRepo.EXPECT().GetProjects(mock.Anything, "folders/folder1", &GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", Type: "folder", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *GcpOrgEntity, f func(context.Context, *GcpOrgEntity) error) error {
						return errors.New("boom")
					})
				},
			},
			expectedObjects: []*GcpOrgEntity{
				{
					EntryName: "folders/folder1",
					Id:        "folderId1",
					Name:      "folderName1",
					Type:      "folder",
					Parent:    nil,
				},
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iterator, projectRepo, folderRepo := createGcpDataObjectIteratorTest(t, tt.fields.organisationId)
			tt.fields.mockSetup(projectRepo, folderRepo)

			var actualDataObjects []*GcpOrgEntity

			err := iterator.DataObjects(context.Background(), func(ctx context.Context, object *GcpOrgEntity) error {
				actualDataObjects = append(actualDataObjects, object)

				return nil
			})
			tt.wantErr(t, err)

			assert.ElementsMatch(t, tt.expectedObjects, actualDataObjects)
		})
	}
}

func createGcpDataObjectIteratorTest(t *testing.T, organisationId string) (*GcpDataObjectIterator, *mockProjectRepo, *mockFolderRepo) {
	t.Helper()

	projectRepo := newMockProjectRepo(t)
	folderRepo := newMockFolderRepo(t)

	r := NewGcpDataObjectIterator(projectRepo, folderRepo, &config.ConfigMap{Parameters: map[string]string{common.GcpOrgId: organisationId}})

	return r, projectRepo, folderRepo
}
