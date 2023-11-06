package bigquery

import (
	"context"
	"fmt"
	"sync"

	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/gcp/common"
)

var metaData *ds.MetaData
var mu sync.Mutex

//go:generate go run github.com/vektra/mockery/v2 --name=dataSourceRepository --with-expecter --inpackage
type dataSourceRepository interface {
	GetDataSets(ctx context.Context, configMap *config.ConfigMap) ([]BQEntity, error)
	GetTables(ctx context.Context, configMap *config.ConfigMap, parent BQEntity) ([]BQEntity, error)
}

type DataSourceSyncer struct {
	repoProvider func() dataSourceRepository
}

func NewDataSourceSyncer() *DataSourceSyncer {
	return &DataSourceSyncer{repoProvider: newDsRepoProvider}
}

func newDsRepoProvider() dataSourceRepository {
	return &BigQueryRepository{}
}

func (s *DataSourceSyncer) SyncDataSource(ctx context.Context, dataSourceHandler wrappers.DataSourceObjectHandler, configMap *config.ConfigMap) error {
	// add gcp project as DataObject of type DataSource
	err := s.addGcpProject(dataSourceHandler, configMap)

	if err != nil {
		return err
	}

	// handle datasets
	datasets, err := s.repoProvider().GetDataSets(ctx, configMap)

	if err != nil {
		return err
	}

	err = s.addBqEntities(datasets, dataSourceHandler, configMap)

	if err != nil {
		return err
	}

	// handle tables
	for _, d := range datasets {
		tables, err := s.repoProvider().GetTables(ctx, configMap, d)

		if err != nil {
			return err
		}

		err = s.addBqEntities(tables, dataSourceHandler, configMap)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DataSourceSyncer) GetDataSourceMetaData(ctx context.Context, configParams *config.ConfigMap) (*ds.MetaData, error) {
	return GetDataSourceMetaData(ctx, configParams)
}

func GetDataSourceMetaData(_ context.Context, configParams *config.ConfigMap) (*ds.MetaData, error) {
	mu.Lock()
	defer mu.Unlock()

	var supportedFeatures []string

	if configParams.GetBoolWithDefault(BqCatalogEnabled, false) {
		supportedFeatures = append(supportedFeatures, ds.ColumnMasking) // TODO include row filtering
	}

	if metaData == nil {
		metaData = &ds.MetaData{
			Type:                  "bigquery",
			SupportedFeatures:     supportedFeatures,
			SupportsApInheritance: false,
			DataObjectTypes: []*ds.DataObjectType{
				{
					Name: ds.Datasource,
					Type: ds.Datasource,
					Permissions: []*ds.DataObjectTypePermission{
						&rolesOwner,
						&rolesEditor,
						&rolesViewer,
						&rolesBigQueryAdmin,
						&rolesBigQueryConnectionAdmin,
						&rolesBigQueryConnectionUser,
						&rolesBigQueryEditor,
						&rolesBigQueryDataOwner,
						&rolesBigQueryDataViewer,
						&rolesBigQueryFilteredDataViewer,
						&rolesBigQueryJobUser,
						&rolesBigQueryMetadataViewer,
						&rolesBigQueryReadSessionUser,
						&rolesBigQueryResourceAdmin,
						&rolesBigQueryResourceEditor,
						&rolesBigQueryResourceViewer,
						&rolesBigQueryUser,
						&rolesBigQueryMaskedReader,
						&rolesBigQueryCatalogPolicyTagAdmin,
						&rolesBigQueryCatalogFineGrainedAccess,
					},
					Children: []string{ds.Dataset},
				},
				{
					Name: ds.Dataset,
					Type: ds.Dataset,
					Permissions: []*ds.DataObjectTypePermission{
						&rolesOwner,
						&rolesEditor,
						&rolesViewer,
						&rolesBigQueryAdmin,
						&rolesBigQueryEditor,
						&rolesBigQueryDataOwner,
						&rolesBigQueryDataViewer,
						&rolesBigQueryFilteredDataViewer,
						&rolesBigQueryMetadataViewer,
						&rolesBigQueryUser,
						&rolesBigQueryCatalogFineGrainedAccess,
					},
					Children: []string{ds.Table, ds.View},
				},
				{
					Name: ds.Table,
					Type: ds.Table,
					Permissions: []*ds.DataObjectTypePermission{
						&rolesOwner,
						&rolesEditor,
						&rolesViewer,
						&rolesBigQueryAdmin,
						&rolesBigQueryEditor,
						&rolesBigQueryDataOwner,
						&rolesBigQueryDataViewer,
						&rolesBigQueryFilteredDataViewer,
						&rolesBigQueryMetadataViewer,
						&rolesBigQueryCatalogFineGrainedAccess,
					},
					Actions: []*ds.DataObjectTypeAction{
						{
							Action:        "SELECT",
							GlobalActions: []string{ds.Read},
						},
						{
							Action:        "INSERT",
							GlobalActions: []string{ds.Write},
						},
						{
							Action:        "UPDATE",
							GlobalActions: []string{ds.Write},
						},
						{
							Action:        "DELETE",
							GlobalActions: []string{ds.Write},
						},
						{
							Action:        "TRUNCATE",
							GlobalActions: []string{ds.Write},
						},
					},
					Children: []string{ds.Column},
				},
				{
					Name: ds.View,
					Type: ds.View,
					Permissions: []*ds.DataObjectTypePermission{
						&rolesOwner,
						&rolesEditor,
						&rolesViewer,
						&rolesBigQueryAdmin,
						&rolesBigQueryEditor,
						&rolesBigQueryDataOwner,
						&rolesBigQueryDataViewer,
						&rolesBigQueryFilteredDataViewer,
						&rolesBigQueryMetadataViewer,
						&rolesBigQueryCatalogFineGrainedAccess,
					},
					Actions: []*ds.DataObjectTypeAction{
						{
							Action:        "SELECT",
							GlobalActions: []string{ds.Read},
						},
						{
							Action:        "INSERT",
							GlobalActions: []string{ds.Write},
						},
						{
							Action:        "UPDATE",
							GlobalActions: []string{ds.Write},
						},
						{
							Action:        "DELETE",
							GlobalActions: []string{ds.Write},
						},
						{
							Action:        "TRUNCATE",
							GlobalActions: []string{ds.Write},
						},
					},
					Children: []string{ds.Column},
				},
				{
					Name:        ds.Column,
					Type:        ds.Column,
					Permissions: []*ds.DataObjectTypePermission{},
					Children:    []string{},
				},
			},
			UsageMetaInfo: &ds.UsageMetaInput{
				DefaultLevel: ds.Table,
				Levels: []*ds.UsageMetaInputDetail{
					{
						Name:            ds.Table,
						DataObjectTypes: []string{ds.Table, ds.View},
					},
					{
						Name:            ds.Dataset,
						DataObjectTypes: []string{ds.Dataset},
					},
				},
			},
		}
	}

	return metaData, nil
}

func (s *DataSourceSyncer) addGcpProject(dataSourceHandler wrappers.DataSourceObjectHandler, configMap *config.ConfigMap) error {
	gcpProject := configMap.GetString(common.GcpProjectId)

	return dataSourceHandler.AddDataObjects(&ds.DataObject{
		ExternalId:       gcpProject,
		Name:             gcpProject,
		FullName:         gcpProject,
		Type:             ds.Datasource,
		Description:      fmt.Sprintf("BigQuery DataSource for GCP project %s", gcpProject),
		ParentExternalId: "",
	})
}

func (s *DataSourceSyncer) addBqEntities(entities []BQEntity, dataSourceHandler wrappers.DataSourceObjectHandler, configMap *config.ConfigMap) error {
	gcpProject := configMap.GetString(common.GcpProjectId)

	for _, d := range entities {
		err := dataSourceHandler.AddDataObjects(&ds.DataObject{
			ExternalId:       d.FullName,
			Name:             d.Name,
			FullName:         d.FullName,
			Type:             d.Type,
			Description:      fmt.Sprintf("BigQuery project %s %s", gcpProject, d.Type),
			ParentExternalId: d.ParentId,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
