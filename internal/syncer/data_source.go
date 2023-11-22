package syncer

import (
	"context"
	"fmt"

	ds "github.com/raito-io/cli/base/data_source"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/org"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
)

//go:generate go run github.com/vektra/mockery/v2 --name=DataSourceRepository --with-expecter --inpackage
type DataSourceRepository interface {
	DataObjects(ctx context.Context, config *ds.DataSourceSyncConfig, fn func(ctx context.Context, object *org.GcpOrgEntity) error) error
}

type DataSourceSyncer struct {
	repoProvider DataSourceRepository
	metadata     *ds.MetaData
}

func NewDataSourceSyncer(repository DataSourceRepository, metadata *ds.MetaData) *DataSourceSyncer {
	return &DataSourceSyncer{repoProvider: repository, metadata: metadata}
}

func GetOrgDataObjectName(configmap *config.ConfigMap) string {
	return fmt.Sprintf("gcp-org-%s", configmap.GetString(common.GcpOrgId))
}

func (s *DataSourceSyncer) SyncDataSource(ctx context.Context, dataSourceHandler wrappers.DataSourceObjectHandler, config *ds.DataSourceSyncConfig) error {
	err := s.repoProvider.DataObjects(ctx, config, func(_ context.Context, object *org.GcpOrgEntity) error {
		err := dataSourceHandler.AddDataObjects(handleGcpOrgEntities(object))
		if err != nil {
			return fmt.Errorf("add data object to handler: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("data object iterator: %w", err)
	}

	return nil
}

func handleGcpOrgEntities(entity *org.GcpOrgEntity) *ds.DataObject {
	var parent string
	if entity.Parent != nil {
		parent = entity.Parent.Id
	}

	return &ds.DataObject{
		Name:             entity.Name,
		Type:             entity.Type,
		FullName:         entity.FullName,
		ExternalId:       entity.Id,
		Description:      entity.Description,
		ParentExternalId: parent,
		DataType:         entity.DataType,
	}
}

func (s *DataSourceSyncer) GetDataSourceMetaData(_ context.Context, _ *config.ConfigMap) (*ds.MetaData, error) {
	common.Logger.Info("DataSource meta data sync")

	return s.metadata, nil
}
