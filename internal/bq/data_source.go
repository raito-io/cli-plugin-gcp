package bigquery

import (
	"context"

	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
)

func NewDataSourceMetaData(_ context.Context, configParams *config.ConfigMap) (*ds.MetaData, error) {
	var supportedFeatures []string

	if configParams.GetBoolWithDefault(common.BqCatalogEnabled, false) {
		supportedFeatures = append(supportedFeatures, ds.ColumnMasking) // TODO include row filtering
	}

	metaData := &ds.MetaData{
		Type:                  "bigquery",
		SupportedFeatures:     supportedFeatures,
		SupportsApInheritance: false,
		DataObjectTypes: []*ds.DataObjectType{
			{
				Name: ds.Datasource,
				Type: ds.Datasource,
				Permissions: []*ds.DataObjectTypePermission{
					roles.RolesOwner.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesEditor.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryAdmin.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryConnectionAdmin.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryConnectionUser.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryEditor.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryDataOwner.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryDataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryFilteredDataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryJobUser.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryMetadataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryReadSessionUser.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryResourceAdmin.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryResourceEditor.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryResourceViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryUser.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryMaskedReader.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryCatalogPolicyTagAdmin.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryCatalogFineGrainedAccess.ToDataObjectTypePermission(roles.ServiceBigQuery),
				},
				Children: []string{ds.Dataset},
			},
			{
				Name: ds.Dataset,
				Type: ds.Dataset,
				Permissions: []*ds.DataObjectTypePermission{
					roles.RolesOwner.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesEditor.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryAdmin.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryEditor.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryDataOwner.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryDataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryFilteredDataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryMetadataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryUser.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryCatalogFineGrainedAccess.ToDataObjectTypePermission(roles.ServiceBigQuery),
				},
				Children: []string{ds.Table, ds.View},
			},
			{
				Name: ds.Table,
				Type: ds.Table,
				Permissions: []*ds.DataObjectTypePermission{
					roles.RolesOwner.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesEditor.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryAdmin.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryEditor.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryDataOwner.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryDataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryFilteredDataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryMetadataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryCatalogFineGrainedAccess.ToDataObjectTypePermission(roles.ServiceBigQuery),
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
					roles.RolesOwner.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesEditor.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryAdmin.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryEditor.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryDataOwner.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryDataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryFilteredDataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryMetadataViewer.ToDataObjectTypePermission(roles.ServiceBigQuery),
					roles.RolesBigQueryCatalogFineGrainedAccess.ToDataObjectTypePermission(roles.ServiceBigQuery),
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
						Action:        "MERGE",
						GlobalActions: []string{ds.Write},
					},
					{
						Action:        "TRUNCATE_TABLE",
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

	return metaData, nil
}
