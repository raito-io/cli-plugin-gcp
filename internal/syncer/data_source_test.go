package syncer

import (
	"context"
	"errors"
	"testing"

	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/gcp"
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

func TestDataSourceSyncer_SyncDataSource(t *testing.T) {
	type fields struct {
		mocksSetup func(repository *MockDataSourceRepository)
	}
	type args struct {
		ctx    context.Context
		config *data_source.DataSourceSyncConfig
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		expectedDataObjects []data_source.DataObject
		wantErr             assert.ErrorAssertionFunc
	}{
		{
			name: "Successfully synced data source",
			fields: fields{
				mocksSetup: func(repository *MockDataSourceRepository) {
					repository.EXPECT().DataObjects(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, config *data_source.DataSourceSyncConfig, f func(context.Context, *org.GcpOrgEntity) error) error {
						err := f(ctx, &org.GcpOrgEntity{
							EntryName: "projects/projectId1",
							Id:        "projectId1",
							Name:      "projectName1",
							FullName:  "gcp.projectId1",
							Type:      "project",
							Parent:    nil,
						})
						if err != nil {
							return err
						}

						err = f(ctx, &org.GcpOrgEntity{
							EntryName: "folders/folder1",
							Id:        "folderId1",
							Name:      "folderName1",
							FullName:  "gcp.folderId1",
							Type:      "folder",
							Parent:    nil,
						})
						if err != nil {
							return err
						}

						return f(ctx, &org.GcpOrgEntity{
							EntryName: "projects/projectId3",
							Id:        "projectId3",
							Name:      "projectName3",
							FullName:  "gcp.folderId1.projectId3",
							Type:      "project",
							Parent:    &org.GcpOrgEntity{EntryName: "folders/folder1", Id: "folderId1", Name: "folderName1", FullName: "gcp.folderId1", Type: "folder", Parent: nil},
						})
					})
				},
			},
			args: args{
				ctx:    context.Background(),
				config: &data_source.DataSourceSyncConfig{ConfigMap: &config.ConfigMap{Parameters: map[string]string{common.GcpOrgId: "orgId"}}},
			},
			expectedDataObjects: []data_source.DataObject{
				{
					ExternalId:       "projectId1",
					Name:             "projectName1",
					FullName:         "gcp.projectId1",
					Type:             "project",
					ParentExternalId: "",
				},
				{
					ExternalId:       "folderId1",
					Name:             "folderName1",
					FullName:         "gcp.folderId1",
					Type:             "folder",
					ParentExternalId: "",
				},
				{
					ExternalId:       "projectId3",
					Name:             "projectName3",
					FullName:         "gcp.folderId1.projectId3",
					Type:             "project",
					ParentExternalId: "folderId1",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "processing error",
			fields: fields{
				mocksSetup: func(repository *MockDataSourceRepository) {
					repository.EXPECT().DataObjects(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, config *data_source.DataSourceSyncConfig, f func(context.Context, *org.GcpOrgEntity) error) error {
						err := f(ctx, &org.GcpOrgEntity{
							EntryName: "projects/projectId1",
							Id:        "projectId1",
							Name:      "projectName1",
							FullName:  "gcp.projectId1",
							Type:      "project",
							Parent:    nil,
						})
						if err != nil {
							return err
						}

						err = f(ctx, &org.GcpOrgEntity{
							EntryName: "folders/folder1",
							Id:        "folderId1",
							Name:      "folderName1",
							FullName:  "gcp.folderId1",
							Type:      "folder",
							Parent:    nil,
						})
						if err != nil {
							return err
						}

						return errors.New("boom")
					})
				},
			},
			args: args{
				ctx:    context.Background(),
				config: &data_source.DataSourceSyncConfig{ConfigMap: &config.ConfigMap{Parameters: map[string]string{common.GcpOrgId: "orgId"}}},
			},
			expectedDataObjects: []data_source.DataObject{
				{
					ExternalId:       "projectId1",
					Name:             "projectName1",
					FullName:         "gcp.projectId1",
					Type:             "project",
					ParentExternalId: "",
				},
				{
					ExternalId:       "folderId1",
					Name:             "folderName1",
					FullName:         "gcp.folderId1",
					Type:             "folder",
					ParentExternalId: "",
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, repo := createTestDataSourceSyncer(t)
			tt.fields.mocksSetup(repo)

			dataSourceObjectHandler := mocks.NewSimpleDataSourceObjectHandler(t, 1)
			err := s.SyncDataSource(tt.args.ctx, dataSourceObjectHandler, tt.args.config)

			tt.wantErr(t, err)

			assert.ElementsMatch(t, tt.expectedDataObjects, dataSourceObjectHandler.DataObjects)
		})
	}
}

func createTestDataSourceSyncer(t *testing.T) (*DataSourceSyncer, *MockDataSourceRepository) {
	t.Helper()

	repoMock := NewMockDataSourceRepository(t)

	return NewDataSourceSyncer(repoMock, gcp.NewDataSourceMetaData()), repoMock
}
