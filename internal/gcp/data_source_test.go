package gcp

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func TestDataSourceSyncer_GetMetaData(t *testing.T) {
	//Given
	syncer, _ := createTestDataSourceSyncer(t)

	//When
	result, err := syncer.GetDataSourceMetaData(context.Background(), nil)

	//Then
	assert.NoError(t, err)
	assert.Equal(t, "gcp", result.Type)
	assert.NotEmpty(t, result.DataObjectTypes)
}

func createTestDataSourceSyncer(t *testing.T) (*DataSourceSyncer, *MockDataSourceRepository) {
	t.Helper()

	repoMock := NewMockDataSourceRepository(t)

	return NewDataSourceSyncer(repoMock), repoMock
}

func TestDataSourceSyncer_SyncDataSource(t *testing.T) {
	orgId := "orgId"

	configMap := config.ConfigMap{
		Parameters: map[string]string{
			common.GcpOrgId: orgId,
		},
	}

	type args struct {
		configMap *config.ConfigMap
	}
	tests := []struct {
		name                  string
		mock                  func(repo *MockDataSourceRepository)
		args                  args
		wantErr               assert.ErrorAssertionFunc
		wantDataSourceHandler func(t *testing.T, handler *mocks.SimpleDataSourceObjectHandler)
	}{
		{
			name: "No folders",
			mock: func(repo *MockDataSourceRepository) {
				repo.EXPECT().GetFolders(mock.Anything, "organizations/"+orgId, (*org.GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return nil
				})

				repo.EXPECT().GetProjects(mock.Anything, "organizations/"+orgId, (*org.GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return f(ctx, &org.GcpOrgEntity{
						Id:        "projectId",
						Name:      "projectName",
						Type:      iam.Project.String(),
						EntryName: "projectId",
						Parent:    nil,
					})
				})
			},
			args: args{
				configMap: &configMap,
			},
			wantErr: assert.NoError,
			wantDataSourceHandler: func(t *testing.T, handler *mocks.SimpleDataSourceObjectHandler) {
				assert.Equal(t, []data_source.DataObject{
					{
						ExternalId:       "gcp-org-orgId",
						Name:             "gcp-org-orgId",
						FullName:         "gcp-org-orgId",
						Type:             "organization",
						Description:      "",
						ParentExternalId: "",
					}, {
						ExternalId:       "projectId",
						Name:             "projectName",
						FullName:         "projectId",
						Type:             "Project",
						Description:      "",
						ParentExternalId: "gcp-org-orgId",
					},
				}, handler.DataObjects)
			},
		},
		{
			name: "Folders and projects",
			mock: func(repo *MockDataSourceRepository) {
				repo.EXPECT().GetFolders(mock.Anything, "organizations/"+orgId, (*org.GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					err := f(ctx, &org.GcpOrgEntity{
						Id:        "folderId1",
						Name:      "folder1",
						Type:      iam.Folder.String(),
						EntryName: "folderId1",
						Parent:    nil,
					})

					if err != nil {
						return err
					}

					return f(ctx, &org.GcpOrgEntity{
						Id:        "folderId2",
						Name:      "folder2",
						Type:      iam.Folder.String(),
						EntryName: "folderId2",
						Parent:    nil,
					})
				})

				repo.EXPECT().GetFolders(mock.Anything, "folderId1", &org.GcpOrgEntity{Id: "folderId1", Name: "folder1", Type: iam.Folder.String(), EntryName: "folderId1", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return f(ctx, &org.GcpOrgEntity{
						Id:        "folderId3",
						Name:      "folder3",
						Type:      iam.Folder.String(),
						EntryName: "folderId3",
						Parent:    &org.GcpOrgEntity{Id: "folderId1", Name: "folder1", Type: iam.Folder.String(), EntryName: "folderId1", Parent: nil},
					})
				})

				repo.EXPECT().GetFolders(mock.Anything, "folderId2", &org.GcpOrgEntity{Id: "folderId2", Name: "folder2", Type: iam.Folder.String(), EntryName: "folderId2", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return nil
				})

				repo.EXPECT().GetFolders(mock.Anything, "folderId3", &org.GcpOrgEntity{Id: "folderId3", Name: "folder3", Type: iam.Folder.String(), EntryName: "folderId3", Parent: &org.GcpOrgEntity{Id: "folderId1", Name: "folder1", Type: iam.Folder.String(), EntryName: "folderId1", Parent: nil}}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return nil
				})

				repo.EXPECT().GetProjects(mock.Anything, "organizations/"+orgId, (*org.GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return f(ctx, &org.GcpOrgEntity{
						Id:        "projectId1",
						Name:      "projectName1",
						Type:      iam.Project.String(),
						EntryName: "projectId1",
						Parent:    nil,
					})
				})

				repo.EXPECT().GetProjects(mock.Anything, "folderId1", &org.GcpOrgEntity{Id: "folderId1", Name: "folder1", Type: iam.Folder.String(), EntryName: "folderId1", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return f(ctx, &org.GcpOrgEntity{
						Id:        "projectId2",
						Name:      "projectName2",
						Type:      iam.Project.String(),
						EntryName: "projectId2",
						Parent:    nil,
					})
				})

				repo.EXPECT().GetProjects(mock.Anything, "folderId2", &org.GcpOrgEntity{Id: "folderId2", Name: "folder2", Type: iam.Folder.String(), EntryName: "folderId2", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return nil
				})

				repo.EXPECT().GetProjects(mock.Anything, "folderId3", &org.GcpOrgEntity{Id: "folderId3", Name: "folder3", Type: iam.Folder.String(), EntryName: "folderId3", Parent: &org.GcpOrgEntity{Id: "folderId1", Name: "folder1", Type: iam.Folder.String(), EntryName: "folderId1", Parent: nil}}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return f(ctx, &org.GcpOrgEntity{
						Id:        "projectId3",
						Name:      "projectName3",
						Type:      iam.Project.String(),
						EntryName: "projectId3",
						Parent:    &org.GcpOrgEntity{Id: "folderId1", Name: "folder1", Type: iam.Folder.String(), EntryName: "folderId1", Parent: nil},
					})
				})
			},
			args: args{
				configMap: &configMap,
			},
			wantErr: assert.NoError,
			wantDataSourceHandler: func(t *testing.T, handler *mocks.SimpleDataSourceObjectHandler) {
				assert.ElementsMatch(t, []data_source.DataObject{
					{
						ExternalId:       "gcp-org-orgId",
						Name:             "gcp-org-orgId",
						FullName:         "gcp-org-orgId",
						Type:             "organization",
						Description:      "",
						ParentExternalId: "",
					}, {
						ExternalId:       "folderId1",
						Name:             "folder1",
						FullName:         "folderId1",
						Type:             "Folder",
						Description:      "",
						ParentExternalId: "gcp-org-orgId",
					},
					{
						ExternalId:       "folderId2",
						Name:             "folder2",
						FullName:         "folderId2",
						Type:             "Folder",
						Description:      "",
						ParentExternalId: "gcp-org-orgId",
					},
					{
						ExternalId:       "folderId3",
						Name:             "folder3",
						FullName:         "folderId3",
						Type:             "Folder",
						Description:      "",
						ParentExternalId: "folderId1",
					},
					{
						ExternalId:       "projectId1",
						Name:             "projectName1",
						FullName:         "projectId1",
						Type:             "Project",
						Description:      "",
						ParentExternalId: "gcp-org-orgId",
					},
					{
						ExternalId:       "projectId2",
						Name:             "projectName2",
						FullName:         "projectId2",
						Type:             "Project",
						Description:      "",
						ParentExternalId: "gcp-org-orgId",
					},
					{
						ExternalId:       "projectId3",
						Name:             "projectName3",
						FullName:         "projectId3",
						Type:             "Project",
						Description:      "",
						ParentExternalId: "folderId1",
					},
				}, handler.DataObjects)
			},
		},
		{
			name: "Error during processing",
			mock: func(repo *MockDataSourceRepository) {
				repo.EXPECT().GetProjects(mock.Anything, "organizations/"+orgId, (*org.GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return errors.New("boom")
				})
			},
			args: args{
				configMap: &configMap,
			},
			wantErr:               assert.Error,
			wantDataSourceHandler: func(t *testing.T, handler *mocks.SimpleDataSourceObjectHandler) {},
		},

		{
			name: "Error during processing recursively",
			mock: func(repo *MockDataSourceRepository) {
				repo.EXPECT().GetFolders(mock.Anything, "organizations/"+orgId, (*org.GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return f(ctx, &org.GcpOrgEntity{
						Id:        "folderId1",
						Name:      "folder1",
						Type:      iam.Folder.String(),
						EntryName: "folderId1",
						Parent:    nil,
					})
				})

				repo.EXPECT().GetFolders(mock.Anything, "folderId1", &org.GcpOrgEntity{Id: "folderId1", Name: "folder1", Type: iam.Folder.String(), EntryName: "folderId1", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return errors.New("boom")
				})

				repo.EXPECT().GetProjects(mock.Anything, "organizations/"+orgId, (*org.GcpOrgEntity)(nil), mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return nil
				})

				repo.EXPECT().GetProjects(mock.Anything, "folderId1", &org.GcpOrgEntity{Id: "folderId1", Name: "folder1", Type: iam.Folder.String(), EntryName: "folderId1", Parent: nil}, mock.Anything).RunAndReturn(func(ctx context.Context, s string, entity *org.GcpOrgEntity, f func(context.Context, *org.GcpOrgEntity) error) error {
					return nil
				})
			},
			args: args{
				configMap: &configMap,
			},
			wantErr:               assert.Error,
			wantDataSourceHandler: func(t *testing.T, handler *mocks.SimpleDataSourceObjectHandler) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer, repo := createTestDataSourceSyncer(t)
			tt.mock(repo)

			dataSourceObjectHandlerMock := mocks.NewSimpleDataSourceObjectHandler(t, 1)

			err := syncer.SyncDataSource(context.Background(), dataSourceObjectHandlerMock, tt.args.configMap)

			tt.wantErr(t, err, fmt.Sprintf("SyncDataSource(ctx, dsHandler, %v)", tt.args.configMap))
			tt.wantDataSourceHandler(t, dataSourceObjectHandlerMock)
		})
	}
}
