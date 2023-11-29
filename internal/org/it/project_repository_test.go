//go:build integration

package it

import (
	"context"
	"testing"

	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/it"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func TestProjectRepository_GetProjects(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repo, _, cleanup, err := createProjectRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	type args struct {
		ctx        context.Context
		parentName string
		parent     *org.GcpOrgEntity
	}
	tests := []struct {
		name    string
		args    args
		want    []*org.GcpOrgEntity
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "No projects",
			args: args{
				ctx:        ctx,
				parentName: "folders/831872280962",
				parent: &org.GcpOrgEntity{
					EntryName: "folders/831872280962",
					Id:        "831872280962",
					Name:      "second_folder",
					FullName:  "831872280962",
					Type:      data_source.Folder,
					Parent: &org.GcpOrgEntity{
						EntryName: "folders/138023537297",
						Name:      "integration_tests",
						Id:        "folders/138023537297",
						FullName:  "138023537297",
						Type:      data_source.Folder,
					},
				},
			},
			want:    []*org.GcpOrgEntity{},
			wantErr: require.NoError,
		},
		{
			name: "Load projects in folder",
			args: args{
				ctx:        ctx,
				parentName: "folders/138023537297",
				parent: &org.GcpOrgEntity{
					EntryName: "folders/138023537297",
					Id:        "138023537297",
					Name:      "integration_tests",
					FullName:  "138023537297",
					Type:      data_source.Folder,
					Parent: &org.GcpOrgEntity{
						EntryName: "organizations/905493414429",
						Name:      "raito.dev",
						Id:        "organizations/905493414429",
						FullName:  "905493414429",
						Type:      data_source.Datasource,
					},
				},
			},
			want: []*org.GcpOrgEntity{
				{
					EntryName: "projects/204677507107",
					Id:        "raito-integration-test",
					Name:      "raito-integration-test",
					FullName:  "raito-integration-test",
					Type:      "project",
					Parent: &org.GcpOrgEntity{
						EntryName: "folders/138023537297",
						Id:        "138023537297",
						Name:      "integration_tests",
						FullName:  "138023537297",
						Type:      data_source.Folder,
						Parent: &org.GcpOrgEntity{
							EntryName: "organizations/905493414429",
							Name:      "raito.dev",
							Id:        "organizations/905493414429",
							FullName:  "905493414429",
							Type:      data_source.Datasource,
						},
					},
				},
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var projects []*org.GcpOrgEntity

			err := repo.GetProjects(tt.args.ctx, nil, tt.args.parentName, tt.args.parent, func(ctx context.Context, project *org.GcpOrgEntity) error {
				projects = append(projects, project)

				return nil
			})

			require.NoError(t, err)

			if err != nil {
				return
			}

			assert.ElementsMatch(t, tt.want, projects)
		})
	}
}

func TestProjectRepository_GetIamPolicy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repo, _, cleanup, err := createProjectRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	bindings, err := repo.GetIamPolicy(ctx, "204677507107")

	require.NoError(t, err)

	expectedBindings := []iam.IamBinding{
		{
			Member:       "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			Role:         "organizations/905493414429/roles/RaitoGcpRole",
			Resource:     "204677507107",
			ResourceType: "project",
		},
		{
			Member:       "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			Role:         "organizations/905493414429/roles/RaitoGcpRoleMasking",
			Resource:     "204677507107",
			ResourceType: "project",
		},
		{
			Member:       "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
			Role:         "roles/bigquery.admin",
			Resource:     "204677507107",
			ResourceType: "project",
		},
		{
			Member:       "user:ruben@raito.dev",
			Role:         "roles/owner",
			Resource:     "204677507107",
			ResourceType: "project",
		},
	}

	for _, binding := range expectedBindings {
		assert.Contains(t, bindings, binding)
	}
}

func TestProjectRepository_UpdateBinding(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repo, _, cleanup, err := createProjectRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	do := org.GcpOrgEntity{
		EntryName: "projects/204677507107",
		Id:        "raito-integration-test",
		Name:      "raito-integration-test",
		FullName:  "raito-integration-test",
		Type:      "project",
		Parent: &org.GcpOrgEntity{
			EntryName: "folders/138023537297",
			Id:        "138023537297",
			Name:      "integration_tests",
			FullName:  "138023537297",
			Type:      data_source.Folder,
			Parent: &org.GcpOrgEntity{
				EntryName: "organizations/905493414429",
				Name:      "raito.dev",
				Id:        "organizations/905493414429",
				FullName:  "905493414429",
				Type:      data_source.Datasource,
			},
		},
	}

	type args struct {
		ctx     context.Context
		binding []iam.IamBinding
	}
	tests := []struct {
		name    string
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "No bindings to update",
			args: args{
				ctx:     ctx,
				binding: nil,
			},
			wantErr: require.NoError,
		},
		{
			name: "Single binding",
			args: args{
				ctx: ctx,
				binding: []iam.IamBinding{
					{
						Role:         roles.RolesBigQueryMaskedReader.Name,
						ResourceType: "project",
						Resource:     "raito-integration-test",
						Member:       "group:sales@raito.dev",
					},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Multiple bindings",
			args: args{
				ctx: ctx,
				binding: []iam.IamBinding{
					{
						Role:         roles.RolesBigQueryJobUser.Name,
						ResourceType: "project",
						Resource:     "raito-integration-test",
						Member:       "group:dev@raito.dev",
					},
					{
						Role:         roles.RolesViewer.Name,
						ResourceType: "project",
						Resource:     "raito-integration-test",
						Member:       "user:c_harris@raito.dev",
					},
				},
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalBindings, err := repo.GetIamPolicy(ctx, do.FullName)
			require.NoError(t, err)

			do := iam.DataObjectReference{
				FullName:   do.FullName,
				ObjectType: do.Type,
			}

			t.Run("Add bindings", func(t *testing.T) {
				// When
				err = repo.UpdateBinding(ctx, &do, tt.args.binding, nil)

				// Then
				require.NoError(t, err)

				updatedBindings, err := repo.GetIamPolicy(ctx, do.FullName)
				require.NoError(t, err)

				assert.GreaterOrEqual(t, len(updatedBindings), len(originalBindings))

				for _, binding := range tt.args.binding {
					assert.Contains(t, updatedBindings, binding)
				}

				for _, binding := range originalBindings {
					assert.Contains(t, updatedBindings, binding)
				}

				originalBindings = updatedBindings
			})

			t.Run("Remove bindings", func(t *testing.T) {
				// When
				err = repo.UpdateBinding(ctx, &do, nil, tt.args.binding)

				// Then
				require.NoError(t, err)

				updatedBindings, err := repo.GetIamPolicy(ctx, do.FullName)
				require.NoError(t, err)

				assert.Equal(t, len(updatedBindings), len(originalBindings)-len(tt.args.binding))

				for _, binding := range tt.args.binding {
					assert.NotContains(t, updatedBindings, binding)
				}
			})
		})
	}

}

func createProjectRepository(ctx context.Context, t *testing.T) (*org.ProjectRepository, *config.ConfigMap, func(), error) {
	t.Helper()

	configMap := it.IntegrationTestConfigMap()

	repo, cleanup, err := InitializeProjectRepository(ctx, configMap)
	if err != nil {
		return nil, nil, nil, err
	}

	return repo, configMap, cleanup, nil
}
