package gcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/access_provider"
	ds "github.com/raito-io/cli/base/data_source"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
	"github.com/raito-io/cli-plugin-gcp/internal/org"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
)

//go:generate go run github.com/vektra/mockery/v2 --name=DataSourceRepository --with-expecter --inpackage
type DataSourceRepository interface {
	DataObjects(ctx context.Context, fn func(ctx context.Context, object *org.GcpOrgEntity) error) error
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

func (s *DataSourceSyncer) SyncDataSource(ctx context.Context, dataSourceHandler wrappers.DataSourceObjectHandler, configMap *config.ConfigMap) error {
	err := s.repoProvider.DataObjects(ctx, func(_ context.Context, object *org.GcpOrgEntity) error {
		return dataSourceHandler.AddDataObjects(handleGcpOrgEntities(object))
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
		FullName:         entity.Id,
		ExternalId:       entity.Id,
		ParentExternalId: parent,
	}
}

func (s *DataSourceSyncer) GetDataSourceMetaData(_ context.Context, _ *config.ConfigMap) (*ds.MetaData, error) {
	common.Logger.Info("DataSource meta data sync")

	return s.metadata, nil
}

func NewDataSourceMetaData() *ds.MetaData {
	managed_permissions := []*ds.DataObjectTypePermission{
		roles.RolesOwner.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesEditor.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesViewer.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryAdmin.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryConnectionAdmin.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryConnectionUser.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryEditor.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryDataOwner.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryDataViewer.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryFilteredDataViewer.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryJobUser.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryMetadataViewer.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryReadSessionUser.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryResourceAdmin.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryResourceEditor.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryResourceViewer.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryUser.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryMaskedReader.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryCatalogPolicyTagAdmin.ToDataObjectTypePermission(roles.ServiceGcp),
		roles.RolesBigQueryCatalogFineGrainedAccess.ToDataObjectTypePermission(roles.ServiceGcp),
	}

	org := strings.ToLower(types.Organization.String())
	project := strings.ToLower(types.Project.String())
	folder := strings.ToLower(types.Folder.String())

	return &ds.MetaData{
		Type:                  "gcp",
		SupportedFeatures:     []string{},
		SupportsApInheritance: false,
		DataObjectTypes: []*ds.DataObjectType{
			{
				Name:        ds.Datasource,
				Type:        ds.Datasource,
				Permissions: []*ds.DataObjectTypePermission{},
				Children:    []string{org},
			},
			{
				Name:        org,
				Type:        org,
				Permissions: managed_permissions,
				Children:    []string{folder, project},
			},
			{
				Name:        folder,
				Type:        folder,
				Permissions: managed_permissions,
				Children:    []string{folder, project},
			},
			{
				Name:        project,
				Type:        project,
				Permissions: managed_permissions,
				Children:    []string{},
			},
		},
		AccessProviderTypes: []*ds.AccessProviderType{
			{
				Type:          access_provider.AclSet,
				Label:         access_provider.AclSet,
				CanBeAssumed:  false,
				CanBeCreated:  true,
				IsNamedEntity: false,
			},
		},
	}
}
