package bigquery

import (
	"context"

	"cloud.google.com/go/bigquery/datapolicies/apiv1/datapoliciespb"
	"github.com/raito-io/cli/base/access_provider"
	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
)

func NewDataSourceMetaData(_ context.Context, configParams *config.ConfigMap) (*ds.MetaData, error) {
	supportedFeatures := []string{ds.RowFiltering}

	catalogEnabled := configParams.GetBoolWithDefault(common.BqCatalogEnabled, false)

	if catalogEnabled {
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
		AccessProviderTypes: []*ds.AccessProviderType{
			{
				Type:          access_provider.AclSet,
				Label:         "IAM Policy",
				CanBeAssumed:  false,
				CanBeCreated:  true,
				IsNamedEntity: false,
			},
		},
	}

	if catalogEnabled {
		metaData.MaskingMetadata = &ds.MaskingMetadata{
			MaskTypes: []*ds.MaskingType{
				{
					DisplayName: "NULL",
					ExternalId:  datapoliciespb.DataMaskingPolicy_PredefinedExpression_name[int32(datapoliciespb.DataMaskingPolicy_ALWAYS_NULL)],
					Description: "Returns NULL instead of the column value. Use this when you want to hide both the value and the data type of the column. When this data masking rule is applied to a column, it makes it less useful in query JOIN operations for users with Masked Reader access. This is because a NULL value isn't sufficiently unique to be useful when joining tables.",
				},
				{
					DisplayName: "Hash (SHA-256)",
					ExternalId:  datapoliciespb.DataMaskingPolicy_PredefinedExpression_name[int32(datapoliciespb.DataMaskingPolicy_SHA256)],
					Description: "Returns the SHA-256 hash of the column's value. You can only use this rule with columns that use the STRING data type.",
					DataTypes: []string{
						"STRING",
					},
				},
				{
					DisplayName: "Last four characters",
					ExternalId:  datapoliciespb.DataMaskingPolicy_PredefinedExpression_name[int32(datapoliciespb.DataMaskingPolicy_LAST_FOUR_CHARACTERS)],
					Description: "Returns the last 4 characters of the column's value, replacing the rest of the string with XXXXX. If the column's value is equal to or less than 4 characters in length, then it returns the column's value after it has been run through the SHA-256 hash function. You can only use this rule with columns that use the STRING data type.",
					DataTypes: []string{
						"STRING",
					},
				},
				{
					DisplayName: "First four characters",
					ExternalId:  datapoliciespb.DataMaskingPolicy_PredefinedExpression_name[int32(datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS)],
					Description: "Returns the first 4 characters of the column's value, replacing the rest of the string with XXXXX. If the column's value is equal to or less than 4 characters in length, then it returns the column's value after it has been run through the SHA-256 hash function. You can only use this rule with columns that use the STRING data type.",
					DataTypes: []string{
						"STRING",
					},
				},
				{
					DisplayName: "Email mask",
					ExternalId:  datapoliciespb.DataMaskingPolicy_PredefinedExpression_name[int32(datapoliciespb.DataMaskingPolicy_EMAIL_MASK)],
					Description: "Returns the column's value after replacing the username of a valid email with XXXXX. If the column's value is not a valid email address, then it returns the column's value after it has been run through the SHA-256 hash function. You can only use this rule with columns that use the STRING data type.",
					DataTypes: []string{
						"STRING",
					},
				},
				{
					DisplayName: "Default masking value",
					ExternalId:  datapoliciespb.DataMaskingPolicy_PredefinedExpression_name[int32(datapoliciespb.DataMaskingPolicy_DEFAULT_MASKING_VALUE)],
					Description: "Returns a default masking value for the column based on the column's data type. Use this when you want to hide the value of the column but reveal the data type. When this data masking rule is applied to a column, it makes it less useful in query JOIN operations for users with Masked Reader access. This is because a default value isn't sufficiently unique to be useful when joining tables.",
				},
				{
					DisplayName: "Date year mask",
					ExternalId:  datapoliciespb.DataMaskingPolicy_PredefinedExpression_name[int32(datapoliciespb.DataMaskingPolicy_DATE_YEAR_MASK)],
					Description: "Returns the column's value after truncating the value to its year, setting all non-year parts of the value to the beginning of the year. You can only use this rule with columns that use the DATE, DATETIME, and TIMESTAMP data types.",
					DataTypes: []string{
						"DATE",
						"DATETIME",
						"TIMESTAMP",
					},
				},
			},
			DefaultMaskExternalName: datapoliciespb.DataMaskingPolicy_PredefinedExpression_name[int32(datapoliciespb.DataMaskingPolicy_ALWAYS_NULL)],
			MaskOverridePermissions: []string{
				roles.RolesBigQueryCatalogFineGrainedAccess.Name,
			},
		}
	}

	return metaData, nil
}
