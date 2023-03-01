package gcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/raito-io/cli-plugin-gcp/gcp/org"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDataSourceSyncer_GetMetaData(t *testing.T) {
	//Given
	syncer := DataSourceSyncer{repoProvider: func() dataSourceRepository {
		return nil
	}}

	//When
	result, err := syncer.GetDataSourceMetaData(context.Background())

	//Then
	assert.NoError(t, err)
	assert.Equal(t, "gcp", result.Type)
	assert.NotEmpty(t, result.DataObjectTypes)
}

func TestDataSourceSyncer_SyncDataSource(t *testing.T) {

	repoMock := newMockDataSourceRepository(t)
	dataSourceObjectHandlerMock := mocks.NewSimpleDataSourceObjectHandler(t, 1)

	repoMock.EXPECT().GetFolders(mock.Anything, mock.Anything).Return([]org.GcpOrgEntity{
		{
			Id:     "f1",
			Name:   "f1",
			Type:   "folder",
			Parent: nil,
		},
	}, nil).Once()
	repoMock.EXPECT().GetProjects(mock.Anything, mock.Anything).Return([]org.GcpOrgEntity{
		{
			Id:   "p1",
			Name: "p1",
			Type: "project",
			Parent: &org.GcpOrgEntity{
				Id: "f1",
			},
		},
	}, nil).Once()

	//Given
	configParams := config.ConfigMap{
		Parameters: map[string]string{"key": "value"},
	}

	syncer := DataSourceSyncer{repoProvider: func() dataSourceRepository {
		return repoMock
	}}

	//When
	err := syncer.SyncDataSource(context.Background(), dataSourceObjectHandlerMock, &configParams)

	//Then
	assert.NoError(t, err)
	assert.Len(t, dataSourceObjectHandlerMock.DataObjects, 2)

	for _, do := range dataSourceObjectHandlerMock.DataObjects {
		if do.ExternalId == "p1" {
			assert.Equal(t, "f1", do.ParentExternalId)
		}
	}
}

func TestDataSourceSyncer_ProjectError(t *testing.T) {

	repoMock := newMockDataSourceRepository(t)
	dataSourceObjectHandlerMock := mocks.NewSimpleDataSourceObjectHandler(t, 1)

	repoMock.EXPECT().GetFolders(mock.Anything, mock.Anything).Return([]org.GcpOrgEntity{
		{
			Id:     "f1",
			Name:   "f1",
			Type:   "folder",
			Parent: nil,
		},
	}, nil).Once()
	repoMock.EXPECT().GetProjects(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error!")).Once()

	//Given
	configParams := config.ConfigMap{
		Parameters: map[string]string{"key": "value"},
	}

	syncer := DataSourceSyncer{repoProvider: func() dataSourceRepository {
		return repoMock
	}}

	//When
	err := syncer.SyncDataSource(context.Background(), dataSourceObjectHandlerMock, &configParams)

	//Then
	assert.Error(t, err)
}

func TestDataSourceSyncer_FolderError(t *testing.T) {

	repoMock := newMockDataSourceRepository(t)
	dataSourceObjectHandlerMock := mocks.NewSimpleDataSourceObjectHandler(t, 1)

	repoMock.EXPECT().GetFolders(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error!")).Once()

	//Given
	configParams := config.ConfigMap{
		Parameters: map[string]string{"key": "value"},
	}

	syncer := DataSourceSyncer{repoProvider: func() dataSourceRepository {
		return repoMock
	}}

	//When
	err := syncer.SyncDataSource(context.Background(), dataSourceObjectHandlerMock, &configParams)

	//Then
	assert.Error(t, err)
}
