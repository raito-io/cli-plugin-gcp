package gcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/access_provider"
	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/org"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
)

//go:generate go run github.com/vektra/mockery/v2 --name=DataSourceRepository --with-expecter --inpackage
type DataSourceRepository interface {
	GetProjects(ctx context.Context, parentName string, parent *org.GcpOrgEntity, fn func(ctx context.Context, project *org.GcpOrgEntity) error) error
	GetFolders(ctx context.Context, parentName string, parent *org.GcpOrgEntity, fn func(ctx context.Context, folder *org.GcpOrgEntity) error) error
}

type DataSourceSyncer struct {
	repoProvider DataSourceRepository
}

func NewDataSourceSyncer(repository DataSourceRepository) *DataSourceSyncer {
	return &DataSourceSyncer{repoProvider: repository}
}

func GetOrgDataObjectName(configmap *config.ConfigMap) string {
	return fmt.Sprintf("gcp-org-%s", configmap.GetString(common.GcpOrgId))
}

func (s *DataSourceSyncer) SyncDataSource(ctx context.Context, dataSourceHandler wrappers.DataSourceObjectHandler, configMap *config.ConfigMap) error {
	err := dataSourceHandler.AddDataObjects(&ds.DataObject{
		Name:       GetOrgDataObjectName(configMap),
		Type:       strings.ToLower(iam.Organization.String()),
		FullName:   GetOrgDataObjectName(configMap),
		ExternalId: GetOrgDataObjectName(configMap),
	})

	if err != nil {
		return err
	}

	//sync projects that are childs of the organization
	err = s.syncProjects(ctx, dataSourceHandler, configMap, nil)
	if err != nil {
		return fmt.Errorf("project syncs of organisation: %w", err)
	}

	// sync folders (and child projects)
	err = s.syncFolders(ctx, dataSourceHandler, configMap, nil)
	if err != nil {
		return fmt.Errorf("folder syncs of organisation: %w", err)
	}

	return nil
}

func (s *DataSourceSyncer) syncFolders(ctx context.Context, dataSourceHandler wrappers.DataSourceObjectHandler, configMap *config.ConfigMap, parent *org.GcpOrgEntity) error {
	syncFunc := func(ctx context.Context, folder *org.GcpOrgEntity) error {
		err := dataSourceHandler.AddDataObjects(handleGcpOrgEntities(folder, configMap))
		if err != nil {
			return fmt.Errorf("add data object %q to data object file: %w", folder.Id, err)
		}

		// Search for projects in folder
		err = s.syncProjects(ctx, dataSourceHandler, configMap, folder)
		if err != nil {
			return fmt.Errorf("sync projects of folder %q: %w", folder.Id, err)
		}

		// Search for sub folders
		err = s.syncFolders(ctx, dataSourceHandler, configMap, folder)
		if err != nil {
			return fmt.Errorf("sync folders of folder %q: %w", folder.Id, err)
		}

		return nil
	}

	parentName := fmt.Sprintf("organizations/%s", configMap.GetString(common.GcpOrgId))

	if parent != nil {
		parentName = parent.EntryName
	}

	return s.repoProvider.GetFolders(ctx, parentName, parent, syncFunc)
}

func (s *DataSourceSyncer) syncProjects(ctx context.Context, dataSourceHandler wrappers.DataSourceObjectHandler, configMap *config.ConfigMap, parent *org.GcpOrgEntity) error {
	syncFunc := func(ctx context.Context, project *org.GcpOrgEntity) error {
		err := dataSourceHandler.AddDataObjects(handleGcpOrgEntities(project, configMap))
		if err != nil {
			return fmt.Errorf("add data object %q to data object file: %w", project.Id, err)
		}

		return nil
	}

	parentName := fmt.Sprintf("organizations/%s", configMap.GetString(common.GcpOrgId))

	if parent != nil {
		parentName = parent.EntryName
	}

	return s.repoProvider.GetProjects(ctx, parentName, parent, syncFunc)
}

var externalIds = set.NewSet[string]()

func handleGcpOrgEntities(entity *org.GcpOrgEntity, configMap *config.ConfigMap) *ds.DataObject {
	externalIds.Add(entity.Id)

	parent := GetOrgDataObjectName(configMap)

	if entity.Parent != nil && !strings.EqualFold(entity.Parent.Type, iam.Organization.String()) {
		if _, f := externalIds[entity.Parent.Id]; f {
			parent = entity.Parent.Id
		}
	}

	return &ds.DataObject{
		Name:             entity.Name,
		Type:             entity.Type,
		FullName:         entity.Id,
		ExternalId:       entity.Id,
		ParentExternalId: parent,
	}
}

func (s *DataSourceSyncer) GetDataSourceMetaData(ctx context.Context, configParams *config.ConfigMap) (*ds.MetaData, error) {
	common.Logger.Info("DataSource meta data sync")
	return GetDataSourceMetaData(ctx, configParams)
}

func GetDataSourceMetaData(_ context.Context, _ *config.ConfigMap) (*ds.MetaData, error) {
	common.Logger.Debug("Returning meta data for the GCP data source")

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

	org := strings.ToLower(iam.Organization.String())
	project := strings.ToLower(iam.Project.String())
	folder := strings.ToLower(iam.Folder.String())

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
	}, nil
}
