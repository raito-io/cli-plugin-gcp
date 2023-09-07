package gcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/access_provider"
	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli-plugin-gcp/gcp/common"
	"github.com/raito-io/cli-plugin-gcp/gcp/iam"
	"github.com/raito-io/cli-plugin-gcp/gcp/org"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
)

//go:generate go run github.com/vektra/mockery/v2 --name=dataSourceRepository --with-expecter --inpackage
type dataSourceRepository interface {
	GetProjects(ctx context.Context, configMap *config.ConfigMap) ([]org.GcpOrgEntity, error)
	GetFolders(ctx context.Context, configMap *config.ConfigMap) ([]org.GcpOrgEntity, error)
}

type DataSourceSyncer struct {
	repoProvider func() dataSourceRepository
}

func NewDataSourceSyncer() *DataSourceSyncer {
	return &DataSourceSyncer{repoProvider: newDsRepoProvider}
}

func newDsRepoProvider() dataSourceRepository {
	return org.NewGCPRepository()
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

	folders, err := s.repoProvider().GetFolders(ctx, configMap)

	if err != nil {
		return err
	}

	err = dataSourceHandler.AddDataObjects(handleGcpOrgEntities(folders, configMap)...)

	if err != nil {
		return err
	}

	projects, err := s.repoProvider().GetProjects(ctx, configMap)

	if err != nil {
		return err
	}

	err = dataSourceHandler.AddDataObjects(handleGcpOrgEntities(projects, configMap)...)

	if err != nil {
		return err
	}

	return nil
}

var externalIds = set.NewSet[string]()

func handleGcpOrgEntities(entities []org.GcpOrgEntity, configMap *config.ConfigMap) []*ds.DataObject {
	dos := make([]*ds.DataObject, len(entities))

	for i, p := range entities {
		externalIds.Add(p.Id)

		parent := GetOrgDataObjectName(configMap)

		if p.Parent != nil && !strings.EqualFold(p.Parent.Type, iam.Organization.String()) {
			if _, f := externalIds[p.Parent.Id]; f {
				parent = p.Parent.Id
			}
		}

		dos[i] = &ds.DataObject{
			Name:             p.Name,
			Type:             p.Type,
			FullName:         p.Id,
			ExternalId:       p.Id,
			ParentExternalId: parent,
		}
	}

	return dos
}

func (s *DataSourceSyncer) GetDataSourceMetaData(ctx context.Context) (*ds.MetaData, error) {
	return GetDataSourceMetaData(ctx)
}

func GetDataSourceMetaData(ctx context.Context) (*ds.MetaData, error) {
	common.Logger.Debug("Returning meta data for the GCP data source")

	managed_permissions := []*ds.DataObjectTypePermission{
		{
			Permission:             "roles/owner",
			Description:            "Full access to most Google Cloud resources. See the list of included permissions.",
			UsageGlobalPermissions: []string{ds.Read, ds.Write, ds.Admin},
			GlobalPermissions:      []string{ds.Admin},
		},
		{
			Permission:             "roles/editor",
			Description:            "View, create, update, and delete most Google Cloud resources. See the list of included permissions.",
			UsageGlobalPermissions: []string{ds.Read, ds.Write},
			GlobalPermissions:      []string{ds.Write},
		},
		{
			Permission:             "roles/viewer",
			Description:            "View most Google Cloud resources. See the list of included permissions.",
			UsageGlobalPermissions: []string{ds.Read},
			GlobalPermissions:      []string{ds.Read},
		},
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
