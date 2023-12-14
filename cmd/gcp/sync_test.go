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
	"github.com/raito-io/cli/base/tag"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	"github.com/raito-io/cli-plugin-gcp/internal/it"
)

func TestGcpSync(t *testing.T) {
	ctx := context.Background()
	cfg := it.IntegrationTestConfigMap()

	cfg.Parameters[common.GsuiteIdentityStoreSync] = "true"

	testMethod := func(tf func(ctx context.Context, cfg *config.ConfigMap, t *testing.T)) func(t *testing.T) {
		return func(t *testing.T) {
			tf(ctx, cfg, t)
		}
	}

	t.Run("DataSource Sync", testMethod(DataSourceSync))
	t.Run("IdentityStore Sync", testMethod(IdentityStoreSync))
	t.Run("Access sync", testMethod(AccessSync))
}

func DataSourceSync(ctx context.Context, config *config.ConfigMap, t *testing.T) {
	// Given
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
			ExternalId: "gcp-org-905493414429",
			Name:       "raito.dev",
			FullName:   "gcp-org-905493414429",
			Type:       "organization",
		},
		{
			ExternalId:       "894564211610",
			Name:             "e2e_tests",
			FullName:         "894564211610",
			Type:             "folder",
			ParentExternalId: "gcp-org-905493414429",
		},
		{
			ExternalId:       "138023537297",
			Name:             "integration_tests",
			FullName:         "138023537297",
			Type:             "folder",
			ParentExternalId: "gcp-org-905493414429",
		},
		{
			ExternalId:       "raito-integration-test",
			Name:             "raito-integration-test",
			FullName:         "raito-integration-test",
			Type:             "project",
			ParentExternalId: "138023537297",
			Tags: []*tag.Tag{
				{Key: "test-type", Value: "integration", Source: "gcp-plugin"},
			},
		},
		{
			ExternalId:       "831872280962",
			Name:             "second_folder",
			FullName:         "831872280962",
			Type:             "folder",
			ParentExternalId: "138023537297",
		},
	}

	for _, do := range expectedDos {
		assert.Contains(t, dsHandler.DataObjects, do)
	}

	assert.Equal(t, dsHandler.DataSourceDescription, "")
	assert.Equal(t, dsHandler.DataSourceFullName, "")
	assert.Equal(t, dsHandler.DataSourceName, "")
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
			ExternalId:       "user:b_stewart@raito.dev",
			Name:             "Benjamin Stewart",
			UserName:         "b_stewart@raito.dev",
			Email:            "b_stewart@raito.dev",
			GroupExternalIds: []string{"group:dev@raito.dev"},
		},
		{
			ExternalId:       "user:c_harris@raito.dev",
			Name:             "Carla Harris",
			UserName:         "c_harris@raito.dev",
			Email:            "c_harris@raito.dev",
			GroupExternalIds: []string{"group:dev@raito.dev"},
		},
		{
			ExternalId:       "user:d_hayden@raito.dev",
			Name:             "Dustin Hayden",
			UserName:         "d_hayden@raito.dev",
			Email:            "d_hayden@raito.dev",
			GroupExternalIds: nil,
		},
		{
			ExternalId:       "user:m_carissa@raito.dev",
			Name:             "Mary Carissa",
			UserName:         "m_carissa@raito.dev",
			Email:            "m_carissa@raito.dev",
			GroupExternalIds: []string{"group:sales@raito.dev"},
		},
		{
			ExternalId:       "user:n_nguyen@raito.dev",
			Name:             "Nick Nguyen",
			UserName:         "n_nguyen@raito.dev",
			Email:            "n_nguyen@raito.dev",
			GroupExternalIds: nil,
		},
		{
			ExternalId: "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			Name:       "service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			UserName:   "service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			Email:      "service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
		},
	}

	for _, u := range expectedUsers {
		assert.Contains(t, isHandler.Users, u)
	}

	expectedGroups := []identity_store.Group{
		{
			ExternalId:  "group:dev@raito.dev",
			Name:        "dev@raito.dev",
			DisplayName: "dev@raito.dev",
		},
		{
			ExternalId:             "group:sales@raito.dev",
			Name:                   "sales@raito.dev",
			DisplayName:            "sales@raito.dev",
			ParentGroupExternalIds: []string{"group:dev@raito.dev"},
		},
	}

	for _, g := range expectedGroups {
		assert.Contains(t, isHandler.Groups, g)
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
								FullName: "831872280962",
								Type:     "folder",
							},
							Permissions: []string{
								roles.RolesViewer.Name,
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

		expectedAps := []sync_from_target.AccessProvider{
			{
				ExternalId: "organization_gcp-org-905493414429_roles_bigquery.admin",
				Name:       "organization_gcp-org-905493414429_roles_bigquery.admin",
				NamingHint: "organization_gcp-org-905493414429_roles_bigquery.admin",
				Type:       ptr.String(access_provider.AclSet),
				Action:     sync_from_target.Grant,
				Who: &sync_from_target.WhoItem{
					Users:           []string{"service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com", "thomas@raito.dev"},
					Groups:          []string{},
					AccessProviders: []string{},
				},
				WhoLocked:    ptr.Bool(false),
				WhatLocked:   ptr.Bool(false),
				NameLocked:   ptr.Bool(false),
				DeleteLocked: ptr.Bool(false),
				ActualName:   "organization_gcp-org-905493414429_roles_bigquery.admin",
				What: []sync_from_target.WhatItem{
					{
						DataObject: &data_source.DataObjectReference{
							FullName: "gcp-org-905493414429",
							Type:     "organization",
						},
						Permissions: []string{
							roles.RolesBigQueryAdmin.Name,
						},
					},
				},
			},
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
							Type:     "project",
						},
						Permissions: []string{
							roles.RolesBigQueryAdmin.Name,
						},
					},
				},
			},
		}
		
		for _, ap := range expectedAps {
			assert.Contains(t, apHandler.AccessProviders, ap)
		}

	})
}
