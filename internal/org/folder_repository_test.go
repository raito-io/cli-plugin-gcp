//go:build integration

package org

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
)

func TestFolderRepository_GetFolders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repo, _, cleanup, err := createFolderRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	type args struct {
		ctx        context.Context
		parentName string
		parent     *GcpOrgEntity
	}
	tests := []struct {
		name    string
		args    args
		want    []*GcpOrgEntity
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "folders in organisation",
			args: args{
				ctx:        ctx,
				parentName: "organizations/905493414429",
				parent: &GcpOrgEntity{
					EntryName: "organizations/905493414429",
					Name:      "raito.dev",
					Id:        "organizations/905493414429",
					FullName:  "905493414429",
					Type:      data_source.Datasource,
				},
			},
			want: []*GcpOrgEntity{
				{
					EntryName: "folders/894564211610",
					Id:        "894564211610",
					Name:      "e2e_tests",
					FullName:  "894564211610",
					Type:      data_source.Folder,
					Parent: &GcpOrgEntity{
						EntryName: "organizations/905493414429",
						Name:      "raito.dev",
						Id:        "organizations/905493414429",
						FullName:  "905493414429",
						Type:      data_source.Datasource,
					},
				},
				{
					EntryName: "folders/138023537297",
					Id:        "138023537297",
					Name:      "integration_tests",
					FullName:  "138023537297",
					Type:      data_source.Folder,
					Parent: &GcpOrgEntity{
						EntryName: "organizations/905493414429",
						Name:      "raito.dev",
						Id:        "organizations/905493414429",
						FullName:  "905493414429",
						Type:      data_source.Datasource,
					},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "folders in folder",
			args: args{
				ctx:        ctx,
				parentName: "folders/138023537297",
				parent: &GcpOrgEntity{
					EntryName: "folders/138023537297",
					Name:      "integration_tests",
					Id:        "folders/138023537297",
					FullName:  "138023537297",
					Type:      data_source.Folder,
				},
			},
			want: []*GcpOrgEntity{
				{
					EntryName: "folders/831872280962",
					Id:        "831872280962",
					Name:      "second_folder",
					FullName:  "831872280962",
					Type:      data_source.Folder,
					Parent: &GcpOrgEntity{
						EntryName: "folders/138023537297",
						Name:      "integration_tests",
						Id:        "folders/138023537297",
						FullName:  "138023537297",
						Type:      data_source.Folder,
					},
				},
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var folders []*GcpOrgEntity

			err := repo.GetFolders(tt.args.ctx, tt.args.parentName, tt.args.parent, func(ctx context.Context, folder *GcpOrgEntity) error {
				folders = append(folders, folder)

				return nil
			})

			require.NoError(t, err)

			if err != nil {
				return
			}

			assert.ElementsMatch(t, folders, tt.want)
		})
	}
}

func TestFolderRepository_GetIamPolicy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repo, _, cleanup, err := createFolderRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	type args struct {
		ctx      context.Context
		folderId string
	}
	tests := []struct {
		name    string
		args    args
		want    []iam.IamBinding
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "folder has no bindings",
			args: args{
				ctx:      ctx,
				folderId: "831872280962",
			},
			want: []iam.IamBinding{
				{
					Role:         "roles/resourcemanager.folderAdmin",
					ResourceType: "folder",
					Resource:     "831872280962",
					Member:       "user:ruben@raito.dev",
				},
				{
					Role:         "roles/resourcemanager.folderEditor",
					ResourceType: "folder",
					Resource:     "831872280962",
					Member:       "user:ruben@raito.dev",
				},
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetIamPolicy(tt.args.ctx, tt.args.folderId)
			require.NoError(t, err)

			if err != nil {
				return
			}

			for _, binding := range tt.want {
				assert.Contains(t, result, binding)
			}
		})
	}
}

func TestFolderRepository_UpdateBinding(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dataObject := GcpOrgEntity{
		EntryName: "folders/831872280962",
		Id:        "831872280962",
		Name:      "second_folder",
		FullName:  "831872280962",
		Type:      data_source.Folder,
		Parent: &GcpOrgEntity{
			EntryName: "folders/138023537297",
			Name:      "integration_tests",
			Id:        "folders/138023537297",
			FullName:  "138023537297",
			Type:      data_source.Folder,
		},
	}

	repo, _, cleanup, err := createFolderRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	type args struct {
		ctx        context.Context
		dataObject *GcpOrgEntity
		binding    []iam.IamBinding
	}
	tests := []struct {
		name    string
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "No bindings",
			args: args{
				ctx:        ctx,
				dataObject: &dataObject,
			},
			wantErr: require.NoError,
		},
		{
			name: "Single binding",
			args: args{
				ctx:        ctx,
				dataObject: &dataObject,
				binding: []iam.IamBinding{
					{
						Role:         roles.RolesBigQueryMaskedReader.Name,
						ResourceType: "folder",
						Resource:     "831872280962",
						Member:       "group:sales@raito.dev",
					},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Multiple bindings",
			args: args{
				ctx:        ctx,
				dataObject: &dataObject,
				binding: []iam.IamBinding{
					{
						Role:         roles.RolesBigQueryJobUser.Name,
						ResourceType: "folder",
						Resource:     "831872280962",
						Member:       "group:dev@raito.dev",
					},
					{
						Role:         roles.RolesViewer.Name,
						ResourceType: "folder",
						Resource:     "831872280962",
						Member:       "user:c_harris@raito.dev",
					},
				},
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalBindings, err := repo.GetIamPolicy(ctx, tt.args.dataObject.FullName)
			require.NoError(t, err)

			do := iam.DataObjectReference{
				FullName:   tt.args.dataObject.FullName,
				ObjectType: tt.args.dataObject.Type,
			}

			t.Run("Add bindings", func(t *testing.T) {
				// When
				err = repo.UpdateBinding(ctx, &do, tt.args.binding, nil)

				// Then
				require.NoError(t, err)

				updatedBindings, err := repo.GetIamPolicy(ctx, tt.args.dataObject.FullName)
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

				updatedBindings, err := repo.GetIamPolicy(ctx, tt.args.dataObject.FullName)
				require.NoError(t, err)

				assert.Equal(t, len(updatedBindings), len(originalBindings)-len(tt.args.binding))

				for _, binding := range tt.args.binding {
					assert.NotContains(t, updatedBindings, binding)
				}
			})
		})
	}
}

func createFolderRepository(ctx context.Context, t *testing.T) (*FolderRepository, *config.ConfigMap, func(), error) {
	t.Helper()

	configMap := it.IntegrationTestConfigMap()

	repo, cleanup, err := InitializeFolderRepository(ctx, configMap)
	if err != nil {
		return nil, nil, nil, err
	}

	return repo, configMap, cleanup, nil
}
