package gcp

import (
	"strings"

	"github.com/raito-io/cli/base/access_provider"
	ds "github.com/raito-io/cli/base/data_source"

	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

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
