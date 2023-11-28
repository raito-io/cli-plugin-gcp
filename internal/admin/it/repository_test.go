//go:build integration

package it

import (
	"context"
	"testing"

	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/admin"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/it"
)

func TestAdminRepository_GetUsers(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()
	repo, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	var users []*iam.UserEntity

	// When
	err = repo.GetUsers(ctx, func(ctx context.Context, entity *iam.UserEntity) error {
		users = append(users, entity)

		return nil
	})

	// Then
	require.NoError(t, err)

	expectedUsers := []*iam.UserEntity{
		{
			ExternalId: "user:b_stewart@raito.dev",
			Name:       "Benjamin Stewart",
			Email:      "b_stewart@raito.dev",
		},
		{
			ExternalId: "user:c_harris@raito.dev",
			Name:       "Carla Harris",
			Email:      "c_harris@raito.dev",
		},
		{
			ExternalId: "user:d_hayden@raito.dev",
			Name:       "Dustin Hayden",
			Email:      "d_hayden@raito.dev",
		},
		{
			ExternalId: "user:m_carissa@raito.dev",
			Name:       "Mary Carissa",
			Email:      "m_carissa@raito.dev",
		},
		{
			ExternalId: "user:n_nguyen@raito.dev",
			Name:       "Nick Nguyen",
			Email:      "n_nguyen@raito.dev",
		},
	}

	for _, expectedUser := range expectedUsers {
		assert.Contains(t, users, expectedUser)
	}
}

func TestAdminRepository_GetGroups(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()
	repo, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	var groups []*iam.GroupEntity

	// When
	err = repo.GetGroups(ctx, func(ctx context.Context, entity *iam.GroupEntity) error {
		groups = append(groups, entity)

		return nil
	})

	// Then
	require.NoError(t, err)

	expectedGroups := []*iam.GroupEntity{
		{
			ExternalId: "group:dev@raito.dev",
			Email:      "dev@raito.dev",
			Members: []string{
				"user:b_stewart@raito.dev",
				"user:c_harris@raito.dev",
				"group:sales@raito.dev",
			},
		},
		{
			ExternalId: "group:sales@raito.dev",
			Email:      "sales@raito.dev",
			Members: []string{
				"user:m_carissa@raito.dev",
			},
		},
	}

	assert.ElementsMatch(t, expectedGroups, groups)
}

func createRepository(ctx context.Context, t *testing.T) (*admin.AdminRepository, *config.ConfigMap, func(), error) {
	t.Helper()

	configMap := it.IntegrationTestConfigMap()
	repo, cleanup, err := InitializeAdminClient(ctx, configMap)

	return repo, configMap, cleanup, err
}
