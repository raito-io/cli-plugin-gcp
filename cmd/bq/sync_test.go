//go:build syncintegration

package main

import (
	"context"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/access_provider"
	"github.com/raito-io/cli/base/access_provider/sync_from_target"
	"github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/access_provider/types"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/tag"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	"github.com/raito-io/cli-plugin-gcp/internal/it"
)

func TestBigQuerySync(t *testing.T) {
	ctx := context.Background()
	cfg := it.IntegrationTestConfigMap()

	testMethod := func(tf func(ctx context.Context, cfg *config.ConfigMap, t *testing.T)) func(t *testing.T) {
		return func(t *testing.T) {
			tf(ctx, cfg, t)
		}
	}

	t.Run("DataSource Sync", testMethod(DataSourceSync))

	t.Run("Identity Store Sync", testMethod(IdentityStoreSync))

	t.Run("Access Provider Sync", testMethod(AccessSync))

	t.Run("Data Usage Sync", testMethod(DataUsageSync))
}

func DataSourceSync(ctx context.Context, config *config.ConfigMap, t *testing.T) {
	syncer, cleanup, err := InitializeDataSourceSyncer(ctx, config)
	require.NoError(t, err)

	defer cleanup()

	dsHandler := mocks.NewSimpleDataSourceObjectHandler(t, 1)

	// When
	err = syncer.SyncDataSource(ctx, dsHandler, &data_source.DataSourceSyncConfig{ConfigMap: config})

	// Then
	require.NoError(t, err)

	expectedDos := []data_source.DataObject{
		{
			ExternalId:  "raito-integration-test",
			Name:        "raito-integration-test",
			FullName:    "raito-integration-test",
			Type:        "datasource",
			Description: "",
		},
		{
			ExternalId:       "raito-integration-test.RAITO_TESTING",
			Name:             "RAITO_TESTING",
			FullName:         "raito-integration-test.RAITO_TESTING",
			Type:             "dataset",
			Description:      "",
			ParentExternalId: "raito-integration-test",
		},
		{
			ExternalId:       "raito-integration-test.RAITO_TESTING.HumanResources_Department",
			Name:             "HumanResources_Department",
			FullName:         "raito-integration-test.RAITO_TESTING.HumanResources_Department",
			Type:             "table",
			Description:      "Human resource department table",
			ParentExternalId: "raito-integration-test.RAITO_TESTING",
			Tags: []*tag.Tag{
				{Key: "label1", Value: "value1", Source: "gcp"},
			},
		},
		{
			ExternalId:       "raito-integration-test.RAITO_TESTING.HumanResources_Department.DepartmentID",
			Name:             "DepartmentID",
			FullName:         "raito-integration-test.RAITO_TESTING.HumanResources_Department.DepartmentID",
			Type:             "column",
			Description:      "",
			ParentExternalId: "raito-integration-test.RAITO_TESTING.HumanResources_Department",
			DataType:         ptr.String("INTEGER"),
		},
	}

	assert.GreaterOrEqual(t, len(dsHandler.DataObjects), 567)

	for _, do := range expectedDos {
		assert.Containsf(t, dsHandler.DataObjects, do, "Data object %+v not found", do)
	}

	assert.Equal(t, dsHandler.DataSourceFullName, "")
	assert.Equal(t, dsHandler.DataSourceName, "")
	assert.Equal(t, dsHandler.DataSourceDescription, "")
}

func IdentityStoreSync(ctx context.Context, config *config.ConfigMap, t *testing.T) {
	// Given
	syncer, cleanup, err := InitializeIdentityStoreSyncer(ctx, config)
	require.NoError(t, err)

	defer cleanup()

	isHandler := mocks.NewSimpleIdentityStoreIdentityHandler(t, 1)

	// When
	err = syncer.SyncIdentityStore(ctx, isHandler, config)

	// Then
	require.NoError(t, err)

	assert.Empty(t, isHandler.Users)
}

func AccessSync(ctx context.Context, config *config.ConfigMap, t *testing.T) {
	syncer, cleanup, err := InitializeDataAccessSyncer(ctx, config)
	require.NoError(t, err)

	defer cleanup()

	// We always set the same ap, and just check the errors
	t.Run("Sync to target", func(t *testing.T) {
		// Given
		feedbackHandler := mocks.NewSimpleAccessProviderFeedbackHandler(t)

		apImport := sync_to_target.AccessProviderImport{
			AccessProviders: []*sync_to_target.AccessProvider{
				{
					Name:   "Simple AP",
					Id:     "simple-ap-id",
					Action: types.Grant,
					Who: sync_to_target.WhoItem{
						Users: []string{"d_hayden@raito.dev"},
					},
					What: []sync_to_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "raito-integration-test.RAITO_TESTING.HumanResources_Department",
								Type:     "table",
							},
							Permissions: []string{
								roles.RolesBigQueryDataViewer.Name,
							},
						},
					},
				},
			},
		}

		// When
		err = syncer.SyncAccessProviderToTarget(ctx, &apImport, feedbackHandler, config)

		// Then
		require.NoError(t, err)
		assert.ElementsMatch(t, feedbackHandler.AccessProviderFeedback, []sync_to_target.AccessProviderSyncFeedback{
			{
				AccessProvider: "simple-ap-id",
				ActualName:     "simple-ap-id",
				Type:           ptr.String(access_provider.AclSet),
				Errors:         nil,
				Warnings:       nil,
			},
		})
	})

	t.Run("Sync from target", func(t *testing.T) {
		// Given
		apHandler := mocks.NewSimpleAccessProviderHandler(t, 35)

		// When
		err = syncer.SyncAccessProvidersFromTarget(ctx, apHandler, config)

		// Then
		require.NoError(t, err)

		expectedAPs := []sync_from_target.AccessProvider{
			{
				ExternalId: "raito-integration-test.RAITO_TESTING.Person_Address.person_address_group",
				Name:       "person_address_group",
				NamingHint: "person_address_group",
				Type:       nil,
				Action:     types.Filtered,
				Policy:     "StateProvinceID = 0",
				Who: &sync_from_target.WhoItem{
					Users:           nil,
					Groups:          []string{"dev@raito.dev"},
					AccessProviders: nil,
				},
				ActualName: "person_address_group",
				What: []sync_from_target.WhatItem{
					{
						DataObject: &data_source.DataObjectReference{
							FullName: "raito-integration-test.RAITO_TESTING.Person_Address",
							Type:     "table",
						},
					},
				},
			},
			{
				ExternalId: "table_raito-integration-test.RAITO_TESTING.HumanResources_Department_roles_bigquery.dataViewer",
				Name:       "Table RAITO_TESTING.HumanResources_Department - Bigquery Data Viewer",
				NamingHint: "table_raito-integration-test.RAITO_TESTING.HumanResources_Department_roles_bigquery.dataViewer",
				Type:       ptr.String(access_provider.AclSet),
				Action:     types.Grant,
				Who: &sync_from_target.WhoItem{
					Users:           []string{"m_carissa@raito.dev"},
					Groups:          []string{},
					AccessProviders: []string{},
				},
				WhoLocked:         ptr.Bool(false),
				WhatLocked:        ptr.Bool(false),
				NameLocked:        ptr.Bool(false),
				DeleteLocked:      ptr.Bool(false),
				NotInternalizable: false,
				ActualName:        "table_raito-integration-test.RAITO_TESTING.HumanResources_Department_roles_bigquery.dataViewer",
				What: []sync_from_target.WhatItem{
					{
						DataObject: &data_source.DataObjectReference{
							FullName: "raito-integration-test.RAITO_TESTING.HumanResources_Department",
							Type:     "table",
						},
						Permissions: []string{
							roles.RolesBigQueryDataViewer.Name,
						},
					},
				},
			},
		}

		assert.GreaterOrEqual(t, len(apHandler.AccessProviders), 10)

		for _, ap := range expectedAPs {
			assert.Containsf(t, apHandler.AccessProviders, ap, "Access provider %+v not found", ap)
		}
	})
}

func DataUsageSync(ctx context.Context, config *config.ConfigMap, t *testing.T) {
	syncer, cleanup, err := InitializeDataUsageSyncer(ctx, config)
	require.NoError(t, err)

	defer cleanup()

	dataUsageHandler := mocks.NewSimpleDataUsageStatementHandler(t)
	dataUsageHandler.EXPECT().GetImportFileSize().Return(uint64(1024 ^ 10))

	// When
	err = syncer.SyncDataUsage(ctx, dataUsageHandler, config)

	// Then
	require.NoError(t, err)

	assert.NotEmpty(t, dataUsageHandler.Statements)
}
