//go:build integration

package org

import (
	"context"
	"testing"

	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/it"
)

func TestOrganizationRepository_GetOrganization(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repo, _, cleanup, err := createOrganizationRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	organization, err := repo.GetOrganization(ctx)

	require.NoError(t, err)
	assert.Equal(t, organization, &GcpOrgEntity{
		EntryName: "organizations/905493414429",
		Id:        "gcp-org-905493414429",
		Name:      "raito.dev",
		FullName:  "gcp-org-905493414429",
		Type:      "organization",
	})
}

func TestOrganizationRepository_GetIamPolicy(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	repo, _, cleanup, err := createOrganizationRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	// When
	bindings, err := repo.GetIamPolicy(ctx, "")

	// Then
	require.NoError(t, err)

	expectedBindings := []iam.IamBinding{
		{
			Member:       "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			Role:         "organizations/905493414429/roles/RaitoGcpRole",
			Resource:     "gcp-org-905493414429",
			ResourceType: "organization",
		},
		{
			Member:       "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			Role:         "roles/bigquery.admin",
			Resource:     "gcp-org-905493414429",
			ResourceType: "organization",
		},
		{
			Member:       "domain:raito.dev",
			Role:         "roles/billing.creator",
			Resource:     "gcp-org-905493414429",
			ResourceType: "organization",
		},
	}

	for _, binding := range expectedBindings {
		assert.Contains(t, bindings, binding)
	}
}

func TestOrganizationRepository_UpdateBinding(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dataObject := GcpOrgEntity{
		EntryName: "organizations/905493414429",
		Id:        "gcp-org-905493414429",
		Name:      "raito.dev",
		FullName:  "gcp-org-905493414429",
		Type:      "organization",
	}

	repo, _, cleanup, err := createOrganizationRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	type args struct {
		ctx      context.Context
		bindings []iam.IamBinding
	}
	tests := []struct {
		name    string
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "No bindings",
			args: args{
				ctx:      ctx,
				bindings: []iam.IamBinding{},
			},
			wantErr: require.NoError,
		},
		{
			name: "Single binding",
			args: args{
				ctx: ctx,
				bindings: []iam.IamBinding{
					{
						Role:         roles.RolesBigQueryDataViewer.Name,
						ResourceType: "organization",
						Resource:     "gcp-org-905493414429",
						Member:       "group:sales@raito.dev",
					},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Multiple binding",
			args: args{
				ctx: ctx,
				bindings: []iam.IamBinding{
					{
						Role:         roles.RolesBigQueryJobUser.Name,
						ResourceType: "organization",
						Resource:     "gcp-org-905493414429",
						Member:       "group:dev@raito.dev",
					},
					{
						Role:         roles.RolesEditor.Name,
						ResourceType: "organization",
						Resource:     "gcp-org-905493414429",
						Member:       "user:c_harris@raito.dev",
					},
				},
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalBindings, err := repo.GetIamPolicy(ctx, dataObject.FullName)
			require.NoError(t, err)

			do := iam.DataObjectReference{
				FullName:   dataObject.FullName,
				ObjectType: dataObject.Type,
			}

			t.Run("Add bindings", func(t *testing.T) {
				// When
				err = repo.UpdateBinding(ctx, &do, tt.args.bindings, nil)

				// Then
				require.NoError(t, err)

				updatedBindings, err := repo.GetIamPolicy(ctx, dataObject.FullName)
				require.NoError(t, err)

				assert.GreaterOrEqual(t, len(updatedBindings), len(originalBindings))

				for _, binding := range tt.args.bindings {
					assert.Contains(t, updatedBindings, binding)
				}

				for _, binding := range originalBindings {
					assert.Contains(t, updatedBindings, binding)
				}

				originalBindings = updatedBindings
			})

			t.Run("Remove bindings", func(t *testing.T) {
				// When
				err = repo.UpdateBinding(ctx, &do, nil, tt.args.bindings)

				// Then
				require.NoError(t, err)

				updatedBindings, err := repo.GetIamPolicy(ctx, dataObject.FullName)
				require.NoError(t, err)

				assert.Equal(t, len(updatedBindings), len(originalBindings)-len(tt.args.bindings))

				for _, binding := range tt.args.bindings {
					assert.NotContains(t, updatedBindings, binding)
				}
			})
		})
	}
}

func createOrganizationRepository(ctx context.Context, t *testing.T) (*OrganizationRepository, *config.ConfigMap, func(), error) {
	t.Helper()

	configMap := it.IntegrationTestConfigMap()

	repo, cleanup, err := InitializeOrganizationRepository(ctx, configMap)
	if err != nil {
		return nil, nil, nil, err
	}

	return repo, configMap, cleanup, nil
}
