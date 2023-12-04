//go:build syncintegration

package main

import (
	"context"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/access_provider"
	"github.com/raito-io/cli/base/access_provider/sync_from_target"
	"github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/identity_store"
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
			ExternalId:       "raito-integration-test.public_dataset",
			Name:             "public_dataset",
			FullName:         "raito-integration-test.public_dataset",
			Type:             "dataset",
			Description:      "",
			ParentExternalId: "raito-integration-test",
		},
		{
			ExternalId:       "raito-integration-test.public_dataset.covid19_open_data",
			Name:             "covid19_open_data",
			FullName:         "raito-integration-test.public_dataset.covid19_open_data",
			Type:             "table",
			Description:      "",
			ParentExternalId: "raito-integration-test.public_dataset",
		},
		{
			ExternalId:       "raito-integration-test.public_dataset.covid19_open_data.location_key",
			Name:             "location_key",
			FullName:         "raito-integration-test.public_dataset.covid19_open_data.location_key",
			Type:             "column",
			Description:      "",
			ParentExternalId: "raito-integration-test.public_dataset.covid19_open_data",
			DataType:         ptr.String("STRING"),
		},
	}

	assert.GreaterOrEqual(t, len(dsHandler.DataObjects), 732)

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

	expectedUsers := []identity_store.User{
		{
			ExternalId: "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			Name:       "service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			UserName:   "service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			Email:      "service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
		},
		{
			ExternalId: "user:m_carissa@raito.dev",
			Name:       "m_carissa@raito.dev",
			UserName:   "m_carissa@raito.dev",
			Email:      "m_carissa@raito.dev",
		}, {
			ExternalId: "user:d_hayden@raito.dev",
			Name:       "d_hayden@raito.dev",
			UserName:   "d_hayden@raito.dev",
			Email:      "d_hayden@raito.dev",
		},
	}

	for _, user := range expectedUsers {
		assert.Containsf(t, isHandler.Users, user, "User %+v not found", user)
	}
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
					Action: sync_to_target.Grant,
					Who: sync_to_target.WhoItem{
						Users: []string{"d_hayden@raito.dev"},
					},
					What: []sync_to_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "raito-integration-test.public_dataset.covid19_open_data",
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
		apHandler := mocks.NewSimpleAccessProviderHandler(t, 10)

		// When
		err = syncer.SyncAccessProvidersFromTarget(ctx, apHandler, config)

		// Then
		require.NoError(t, err)

		expectedAPs := []sync_from_target.AccessProvider{
			{
				ExternalId: "project_raito-integration-test_roles_bigquery.admin",
				Name:       "project_raito-integration-test_roles_bigquery.admin",
				NamingHint: "project_raito-integration-test_roles_bigquery.admin",
				Type:       ptr.String(access_provider.AclSet),
				Action:     sync_from_target.Grant,
				Who: &sync_from_target.WhoItem{
					Users:           []string{"service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com"},
					Groups:          []string{},
					AccessProviders: []string{},
				},
				WhoLocked:    ptr.Bool(false),
				WhatLocked:   ptr.Bool(false),
				NameLocked:   ptr.Bool(false),
				DeleteLocked: ptr.Bool(false),
				ActualName:   "project_raito-integration-test_roles_bigquery.admin",
				What: []sync_from_target.WhatItem{
					{
						DataObject: &data_source.DataObjectReference{
							FullName: "raito-integration-test",
							Type:     "datasource",
						},
						Permissions: []string{
							roles.RolesBigQueryAdmin.Name,
						},
					},
				},
			},
			{
				ExternalId: "dataset_raito-integration-test.private_dataset_roles_bigquery.dataEditor",
				Name:       "dataset_raito-integration-test.private_dataset_roles_bigquery.dataEditor",
				NamingHint: "dataset_raito-integration-test.private_dataset_roles_bigquery.dataEditor",
				Type:       ptr.String(access_provider.AclSet),
				Action:     sync_from_target.Grant,
				Who: &sync_from_target.WhoItem{
					Users:           []string{"m_carissa@raito.dev"},
					Groups:          []string{},
					AccessProviders: []string{},
				},
				WhoLocked:    ptr.Bool(false),
				WhatLocked:   ptr.Bool(false),
				NameLocked:   ptr.Bool(false),
				DeleteLocked: ptr.Bool(false),
				ActualName:   "dataset_raito-integration-test.private_dataset_roles_bigquery.dataEditor",
				What: []sync_from_target.WhatItem{
					{
						DataObject: &data_source.DataObjectReference{
							FullName: "raito-integration-test.private_dataset",
							Type:     "dataset",
						},
						Permissions: []string{
							roles.RolesBigQueryEditor.Name,
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
