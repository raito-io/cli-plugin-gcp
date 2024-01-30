//go:build integration

package bigquery

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"cloud.google.com/go/bigquery/datapolicies/apiv1/datapoliciespb"
	"github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/it"
)

func TestDataCatalogRepository_CrudPolicyTags(t *testing.T) {
	ctx := context.Background()

	repo, _, cleanup, err := createDataCatalogRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	var maskingInformation *BQMaskingInformation

	t.Run("CreatePolicyTag", func(t *testing.T) {
		maskingInformation, err = repo.CreatePolicyTagWithDataPolicy(ctx, "europe-west1", datapoliciespb.DataMaskingPolicy_ALWAYS_NULL, &sync_to_target.AccessProvider{
			Name:       "test_policy_tag",
			NamingHint: "test_policy_tag",
		})

		require.NoError(t, err)

		assert.True(t, strings.HasPrefix(maskingInformation.PolicyTag.Name, "test_policy_tag_"))
		assert.Equal(t, maskingInformation.DataPolicy.PolicyType, datapoliciespb.DataMaskingPolicy_ALWAYS_NULL)
	})

	t.Run("GetMaskingInformation", func(t *testing.T) {
		getMaskingInformation, err := repo.GetMaskingInformationForDataPolicy(ctx, maskingInformation.DataPolicy.FullName)

		require.NoError(t, err)

		assert.Equal(t, getMaskingInformation.DataPolicy.FullName, maskingInformation.DataPolicy.FullName)
		assert.Equal(t, getMaskingInformation.DataPolicy.PolicyType, maskingInformation.DataPolicy.PolicyType)
		assert.Equal(t, getMaskingInformation.PolicyTag.Name, maskingInformation.PolicyTag.Name)

		getMaskingInformationFullNameParts := strings.Split(getMaskingInformation.DataPolicy.FullName, "/")
		maskingInformationFullNameParts := strings.Split(maskingInformation.DataPolicy.FullName, "/")

		getMaskingInformationFullNameParts[1] = ""
		maskingInformationFullNameParts[1] = ""

		assert.Equal(t, getMaskingInformationFullNameParts, maskingInformationFullNameParts)
	})

	t.Run("DeletePolicyTag", func(t *testing.T) {
		err = repo.DeletePolicyAndTag(ctx, maskingInformation.DataPolicy.FullName)

		require.NoError(t, err)

		maskingInformation, err = repo.GetMaskingInformationForDataPolicy(ctx, maskingInformation.DataPolicy.FullName)
		require.Error(t, err)
	})
}

func TestDataCatalogRepository_UpdateWhatOfDataPolicy(t *testing.T) {
	ctx := context.Background()

	repo, _, cleanup, err := createDataCatalogRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	maskingInformation, err := repo.CreatePolicyTagWithDataPolicy(ctx, "eu", datapoliciespb.DataMaskingPolicy_ALWAYS_NULL, &sync_to_target.AccessProvider{
		Name: "update_what_test",
	})

	require.NoError(t, err)

	defer func(maskingInformation *BQMaskingInformation) {
		err = repo.DeletePolicyAndTag(ctx, maskingInformation.DataPolicy.FullName)
	}(maskingInformation)

	columnName := "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_deaths"

	t.Run("Add policy tag to column", func(t *testing.T) {
		err = repo.UpdateWhatOfDataPolicy(ctx, maskingInformation, []string{columnName}, nil)
		require.NoError(t, err)

		metadata, err := repo.bigQueryClient.Dataset("public_dataset").Table("covid_19_geographic_distribution_worldwide").Metadata(ctx)
		require.NoError(t, err)

		for _, column := range metadata.Schema {
			if column.Name == columnName {
				require.NotNil(t, column.PolicyTags)
				assert.Contains(t, column.PolicyTags.Names, maskingInformation.PolicyTag.FullName)
			}
		}
	})

	t.Run("List data policies", func(t *testing.T) {
		policies, err := repo.ListDataPolicies(ctx)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(policies), 1)

		for k, policy := range policies {
			fmt.Printf("%s: %+v\n", k, policy)
		}

		fmt.Printf("%+v\n", maskingInformation)

		maskingPolicy, found := policies[maskingInformation.PolicyTag.FullName]
		require.True(t, found)

		assert.Equal(t, maskingPolicy.DataPolicy.FullName, maskingInformation.DataPolicy.FullName)
		assert.Equal(t, maskingPolicy.DataPolicy.PolicyType, maskingInformation.DataPolicy.PolicyType)
		assert.Equal(t, maskingPolicy.PolicyTag.Name, maskingInformation.PolicyTag.Name)
	})

	t.Run("Delete policy tag from column", func(t *testing.T) {
		err = repo.UpdateWhatOfDataPolicy(ctx, maskingInformation, nil, []string{columnName})
		require.NoError(t, err)

		metadata, err := repo.bigQueryClient.Dataset("public_dataset").Table("covid_19_geographic_distribution_worldwide").Metadata(ctx)
		require.NoError(t, err)

		for _, column := range metadata.Schema {
			if column.Name == columnName {
				require.NotNil(t, column.PolicyTags)
				assert.NotContains(t, column.PolicyTags.Names, maskingInformation.PolicyTag.FullName)
			}
		}
	})
}

func TestDataCatalogRepository_GetLocationsForDataObjects(t *testing.T) {
	ctx := context.Background()

	repo, _, cleanup, err := createDataCatalogRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	ap := &sync_to_target.AccessProvider{
		What: []sync_to_target.WhatItem{
			{
				DataObject: &data_source.DataObjectReference{
					FullName: "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.deaths",
				},
			},
		},
		DeleteWhat: []sync_to_target.WhatItem{
			{
				DataObject: &data_source.DataObjectReference{
					FullName: "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.confirmed_cases",
				},
			},
		},
	}

	// When
	resultDos, resultDeletedDos, err := repo.GetLocationsForDataObjects(ctx, ap)

	// Then
	require.NoError(t, err)
	assert.Equal(t, resultDos, map[string]string{"raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.deaths": "eu"})
	assert.Equal(t, resultDeletedDos, map[string]string{"raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.confirmed_cases": "eu"})
}

func TestDataCatalogRepository_UpdateAccess(t *testing.T) {
	ctx := context.Background()

	repo, _, cleanup, err := createDataCatalogRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	maskingInformation, err := repo.CreatePolicyTagWithDataPolicy(ctx, "eu", datapoliciespb.DataMaskingPolicy_ALWAYS_NULL, &sync_to_target.AccessProvider{
		Name: "update_access_test",
	})

	require.NoError(t, err)

	defer func() {
		err = repo.DeletePolicyAndTag(ctx, maskingInformation.DataPolicy.FullName)
	}()

	columnName := "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_confirmed_cases"

	err = repo.UpdateWhatOfDataPolicy(ctx, maskingInformation, []string{columnName}, nil)
	require.NoError(t, err)

	whoItem := sync_to_target.WhoItem{
		Users:  []string{"d_hayden@raito.dev"},
		Groups: []string{"sales@raito.dev"},
	}

	t.Run("Add Access to mask", func(t *testing.T) {
		err := repo.UpdateAccess(ctx, maskingInformation, &whoItem, nil)
		require.NoError(t, err)

		members, err := repo.GetFineGrainedReaderMembers(ctx, maskingInformation.PolicyTag.FullName)
		require.NoError(t, err)

		assert.ElementsMatch(t, members, []string{"user:d_hayden@raito.dev", "group:sales@raito.dev"})
	})

	t.Run("Delete Access from mask", func(t *testing.T) {
		err := repo.UpdateAccess(ctx, maskingInformation, nil, &whoItem)
		require.NoError(t, err)

		members, err := repo.GetFineGrainedReaderMembers(ctx, maskingInformation.PolicyTag.FullName)
		require.NoError(t, err)

		assert.Empty(t, members)
	})

}

func createDataCatalogRepository(ctx context.Context, t *testing.T) (*DataCatalogRepository, *config.ConfigMap, func(), error) {
	t.Helper()

	configMap := it.IntegrationTestConfigMap()

	repo, cleanup, err := InitializeDataCatalogRepository(ctx, configMap)
	if err != nil {
		return nil, nil, nil, err
	}

	return repo, configMap, cleanup, nil
}
