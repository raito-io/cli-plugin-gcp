package syncer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/access_provider"
	"github.com/raito-io/cli/base/access_provider/sync_from_target"
	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/raito-io/golang-set/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bigquery "github.com/raito-io/cli-plugin-gcp/internal/bq"
	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	"github.com/raito-io/cli-plugin-gcp/internal/gcp"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func TestAccessSyncer_GenerateAccessProviderDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		input    iam.IamBinding
		expected string
	}{
		{
			name: "Dataset dataEditor",
			input: iam.IamBinding{
				Role:         "roles/bigquery.dataEditor",
				ResourceType: "table",
				Resource:     "bq-demodata.MASTER_DATA.dbo_ErrorLog",
			},
			expected: "Table MASTER_DATA.dbo_ErrorLog - Bigquery Data Editor",
		},
		{
			name: "Dataset dataViewer",
			input: iam.IamBinding{
				Role:         "roles/bigquery.dataViewer",
				ResourceType: "dataset",
				Resource:     "bq-demodata.DEMO_VIEWS",
			},
			expected: "Dataset DEMO_VIEWS - Bigquery Data Viewer",
		},
		{
			name: "Project owner",
			input: iam.IamBinding{
				Role:         "roles/owner",
				ResourceType: "project",
				Resource:     "bq-demodata",
			},
			expected: "Project bq-demodata - Owner",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateAccessProviderDisplayName(tt.input.ResourceType, tt.input); got != tt.expected {
				t.Errorf("generateAccessProviderDisplayName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAccessSyncer_SyncAccessProvidersFromTarget(t *testing.T) {
	bqMetadata, err := bigquery.NewDataSourceMetaData(context.Background(), &config.ConfigMap{Parameters: map[string]string{
		common.BqCatalogEnabled: "true",
	}})

	require.NoError(t, err)

	type fields struct {
		mockSetup            func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService)
		metadata             *data_source.MetaData
		raitoManagedBindings []iam.IamBinding
		raitoMasks           set.Set[string]
	}
	type args struct {
		ctx       context.Context
		configMap *config.ConfigMap
	}
	tests := []struct {
		name                    string
		fields                  fields
		args                    args
		expectedAccessProviders []sync_from_target.AccessProvider
		wantErr                 assert.ErrorAssertionFunc
	}{
		{
			name: "No access providers",
			fields: fields{
				mockSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					gcpRepo.EXPECT().Bindings(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				},
				metadata: gcp.NewDataSourceMetaData(),
			},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{},
			},
			expectedAccessProviders: []sync_from_target.AccessProvider{},
			wantErr:                 assert.NoError,
		},
		{
			name: "Single access provider",
			fields: fields{
				mockSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					gcpRepo.EXPECT().Bindings(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, config *data_source.DataSourceSyncConfig, f func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
						return f(ctx, &org.GcpOrgEntity{
							EntryName: "projects/project1",
							Id:        "project1",
							Type:      "project",
							Name:      "project1",
						}, []iam.IamBinding{
							{
								Member:       "user:ruben@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
							{
								Member:       "user:dieter@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
							{
								Member:       "group:group1@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
						},
						)
					})
				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: []iam.IamBinding{},
			},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			expectedAccessProviders: []sync_from_target.AccessProvider{
				{
					ExternalId: "project_project1_roles_owner",
					Name:       "Project project1 - Owner",
					NamingHint: "project_project1_roles_owner",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"ruben@raito.io", "dieter@raito.io"},
						Groups:          []string{"group1@raito.io"},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "project_project1_roles_owner",
					What: []sync_from_target.WhatItem{
						{
							Permissions: []string{"roles/owner"},
							DataObject: &data_source.DataObjectReference{
								FullName: "project1",
								Type:     "project",
							},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Multiple access provider",
			fields: fields{
				mockSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					gcpRepo.EXPECT().Bindings(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, config *data_source.DataSourceSyncConfig, f func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
						err := f(ctx, &org.GcpOrgEntity{
							EntryName: "projects/project1",
							Id:        "project1",
							Type:      "project",
							Name:      "project1",
						}, []iam.IamBinding{
							{
								Member:       "user:ruben@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
							{
								Member:       "user:dieter@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/viewer",
							},
						},
						)
						if err != nil {
							return err
						}

						return f(ctx, &org.GcpOrgEntity{
							EntryName: "folders/folder1",
							Id:        "folder1",
							Type:      "folder",
							Name:      "folder1",
						}, []iam.IamBinding{
							{
								Member:       "serviceAccount:sa@raito.io",
								Resource:     "folder1",
								ResourceType: "folder",
								Role:         "roles/editor",
							},
							{
								Member:       "user:dieter@raito.io",
								Resource:     "folder1",
								ResourceType: "folder",
								Role:         "roles/viewer",
							},
						},
						)
					})
				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: []iam.IamBinding{},
			},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{},
			},
			expectedAccessProviders: []sync_from_target.AccessProvider{
				{
					ExternalId: "project_project1_roles_owner",
					Name:       "Project project1 - Owner",
					NamingHint: "project_project1_roles_owner",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"ruben@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "project_project1_roles_owner",
					What: []sync_from_target.WhatItem{
						{
							Permissions: []string{"roles/owner"},
							DataObject: &data_source.DataObjectReference{
								FullName: "project1",
								Type:     "project",
							},
						},
					},
				},
				{
					ExternalId: "project_project1_roles_viewer",
					Name:       "Project project1 - Viewer",
					NamingHint: "project_project1_roles_viewer",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"dieter@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "project_project1_roles_viewer",
					What: []sync_from_target.WhatItem{
						{
							Permissions: []string{"roles/viewer"},
							DataObject: &data_source.DataObjectReference{
								FullName: "project1",
								Type:     "project",
							},
						},
					},
				},
				{
					ExternalId: "folder_folder1_roles_editor",
					Name:       "Folder folder1 - Editor",
					NamingHint: "folder_folder1_roles_editor",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"sa@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "folder_folder1_roles_editor",
					What: []sync_from_target.WhatItem{
						{
							Permissions: []string{"roles/editor"},
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
						},
					},
				},
				{
					ExternalId: "folder_folder1_roles_viewer",
					Name:       "Folder folder1 - Viewer",
					NamingHint: "folder_folder1_roles_viewer",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"dieter@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "folder_folder1_roles_viewer",
					What: []sync_from_target.WhatItem{
						{
							Permissions: []string{"roles/viewer"},
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "processing error",
			fields: fields{
				mockSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					gcpRepo.EXPECT().Bindings(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, config *data_source.DataSourceSyncConfig, f func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
						err := f(ctx, &org.GcpOrgEntity{
							EntryName: "projects/project1",
							Id:        "project1",
							Type:      "project",
							Name:      "project1",
						}, []iam.IamBinding{
							{
								Member:       "user:ruben@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
						},
						)
						if err != nil {
							return err
						}

						return errors.New("boom")
					})
				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: []iam.IamBinding{},
			},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{},
			},
			expectedAccessProviders: []sync_from_target.AccessProvider{},
			wantErr:                 assert.Error,
		},
		{
			name: "Import masks",
			fields: fields{
				mockSetup: func(dataIterator *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					dataIterator.EXPECT().Bindings(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, syncConfig *data_source.DataSourceSyncConfig, f func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
						err := f(ctx, &org.GcpOrgEntity{EntryName: "project1", Id: "project1", Type: "project", Name: "project1"}, []iam.IamBinding{{Member: "user:ruben@raito.io", Resource: "project1", ResourceType: "project", Role: "roles/owner"}})
						if err != nil {
							return err
						}

						err = f(ctx, &org.GcpOrgEntity{EntryName: "project1/dataset1", Id: "dataset1", Type: "dataset", Name: "dataset1"}, []iam.IamBinding{{Member: "user:dieter@raito.io", Resource: "project/dataset1", ResourceType: "dataset", Role: "roles/editor"}})
						if err != nil {
							return err
						}

						err = f(ctx, &org.GcpOrgEntity{EntryName: "project1/dataset1/table1", Id: "table1", Type: "table", Name: "table1"}, []iam.IamBinding{{Member: "user:thomas@raito.io", Resource: "project/dataset1/table1", ResourceType: "table", Role: "roles/bigquery.dataviewer"}})
						if err != nil {
							return err
						}

						return f(ctx, &org.GcpOrgEntity{EntryName: "project1/dataset1/table1/column1", Id: "column1", Type: "column", Name: "column1", Location: "eu-west1", PolicyTags: []string{"policytag1"}, FullName: "project1/dataset1/table1/column1"}, []iam.IamBinding{})
					})

					maskingService.EXPECT().ImportMasks(mock.Anything, mock.Anything, set.NewSet("eu-west1"), map[string][]string{"policytag1": {"project1/dataset1/table1/column1"}}, set.NewSet("raitoMask1")).RunAndReturn(func(ctx context.Context, handler wrappers.AccessProviderHandler, s set.Set[string], m map[string][]string, s2 set.Set[string]) error {
						return handler.AddAccessProviders(&sync_from_target.AccessProvider{
							ExternalId: "dataPolicyMask1",
							Name:       "dataPolicyName",
							Type:       ptr.String("maskType"),
							What:       []sync_from_target.WhatItem{{DataObject: &data_source.DataObjectReference{Type: "column", FullName: "project1/dataset1/table1/column1"}}},
							Who:        &sync_from_target.WhoItem{Users: []string{"bart@raito.io"}},
							Action:     sync_from_target.Mask,
							ActualName: "dataPolicyMask1ActualName",
						})
					})

					filteringService.EXPECT().ImportFilters(mock.Anything, mock.Anything, mock.Anything, set.Set[string]{}).Return(nil)
				},
				metadata:             bqMetadata,
				raitoManagedBindings: []iam.IamBinding{},
				raitoMasks:           set.NewSet("raitoMask1"),
			},
			args: args{
				ctx: context.Background(),
				configMap: &config.ConfigMap{
					Parameters: map[string]string{common.BqCatalogEnabled: "true"},
				},
			},
			expectedAccessProviders: []sync_from_target.AccessProvider{
				{
					ExternalId: "project_project1_roles_owner",
					Name:       "Project project1 - Owner",
					NamingHint: "project_project1_roles_owner",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"ruben@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					WhoLocked:    ptr.Bool(false),
					WhatLocked:   ptr.Bool(false),
					NameLocked:   ptr.Bool(false),
					DeleteLocked: ptr.Bool(false),
					ActualName:   "project_project1_roles_owner",
					What: []sync_from_target.WhatItem{
						{
							DataObject:  &data_source.DataObjectReference{Type: "datasource", FullName: "project1"},
							Permissions: []string{"roles/owner"},
						},
					},
				},
				{
					ExternalId: "dataset_project/dataset1_roles_editor",
					Name:       "Dataset project/dataset1 - Editor",
					NamingHint: "dataset_project/dataset1_roles_editor",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"dieter@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					WhoLocked:    ptr.Bool(false),
					WhatLocked:   ptr.Bool(false),
					NameLocked:   ptr.Bool(false),
					DeleteLocked: ptr.Bool(false),
					ActualName:   "dataset_project/dataset1_roles_editor",
					What: []sync_from_target.WhatItem{
						{
							DataObject:  &data_source.DataObjectReference{Type: "dataset", FullName: "project/dataset1"},
							Permissions: []string{"roles/editor"},
						},
					},
				},
				{
					ExternalId: "table_project/dataset1/table1_roles_bigquery.dataviewer",
					Name:       "Table project/dataset1/table1 - Bigquery Dataviewer",
					NamingHint: "table_project/dataset1/table1_roles_bigquery.dataviewer",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"thomas@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					WhoLocked:    ptr.Bool(false),
					WhatLocked:   ptr.Bool(false),
					NameLocked:   ptr.Bool(false),
					DeleteLocked: ptr.Bool(false),
					ActualName:   "table_project/dataset1/table1_roles_bigquery.dataviewer",
					What: []sync_from_target.WhatItem{
						{
							DataObject:  &data_source.DataObjectReference{Type: "table", FullName: "project/dataset1/table1"},
							Permissions: []string{"roles/bigquery.dataviewer"},
						},
					},
				},
				{
					ExternalId: "dataPolicyMask1",
					Name:       "dataPolicyName",
					NamingHint: "",
					Type:       ptr.String("maskType"),
					Action:     sync_from_target.Mask,
					Who: &sync_from_target.WhoItem{
						Users: []string{"bart@raito.io"},
					},
					ActualName: "dataPolicyMask1ActualName",
					What: []sync_from_target.WhatItem{
						{
							DataObject:  &data_source.DataObjectReference{Type: "column", FullName: "project1/dataset1/table1/column1"},
							Permissions: nil,
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Import filters",
			fields: fields{
				mockSetup: func(dataIterator *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					dataIterator.EXPECT().Bindings(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, syncConfig *data_source.DataSourceSyncConfig, f func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
						err := f(ctx, &org.GcpOrgEntity{EntryName: "project1", Id: "project1", Type: "project", Name: "project1"}, []iam.IamBinding{{Member: "user:ruben@raito.io", Resource: "project1", ResourceType: "project", Role: "roles/owner"}})
						if err != nil {
							return err
						}

						err = f(ctx, &org.GcpOrgEntity{EntryName: "project1/dataset1", Id: "dataset1", Type: "dataset", Name: "dataset1"}, []iam.IamBinding{{Member: "user:dieter@raito.io", Resource: "project/dataset1", ResourceType: "dataset", Role: "roles/editor"}})
						if err != nil {
							return err
						}

						err = f(ctx, &org.GcpOrgEntity{EntryName: "project1/dataset1/table1", Id: "table1", Type: "table", Name: "table1"}, []iam.IamBinding{{Member: "user:thomas@raito.io", Resource: "project/dataset1/table1", ResourceType: "table", Role: "roles/bigquery.dataviewer"}})
						if err != nil {
							return err
						}

						return f(ctx, &org.GcpOrgEntity{EntryName: "project1/dataset1/table1/column1", Id: "column1", Type: "column", Name: "column1", Location: "eu-west1", PolicyTags: []string{"policytag1"}, FullName: "project1/dataset1/table1/column1"}, []iam.IamBinding{})
					})

					maskingService.EXPECT().ImportMasks(mock.Anything, mock.Anything, set.NewSet("eu-west1"), map[string][]string{"policytag1": {"project1/dataset1/table1/column1"}}, set.NewSet("raitoMask1")).Return(nil)

					filteringService.EXPECT().ImportFilters(mock.Anything, mock.Anything, mock.Anything, set.Set[string]{}).RunAndReturn(func(ctx context.Context, syncConfig *data_source.DataSourceSyncConfig, handler wrappers.AccessProviderHandler, s set.Set[string]) error {
						return handler.AddAccessProviders(&sync_from_target.AccessProvider{
							ExternalId: "filter1",
							Name:       "filter1",
							NamingHint: "filter1",
							What:       []sync_from_target.WhatItem{{DataObject: &data_source.DataObjectReference{Type: "table", FullName: "project1/dataset1/table1"}}},
							Who:        &sync_from_target.WhoItem{Users: []string{"bart@raito.io"}},
							Action:     sync_from_target.Filtered,
							ActualName: "dataPolicyFilter1ActualName",
							Policy:     "table1 > 10",
						})
					})
				},
				metadata:             bqMetadata,
				raitoManagedBindings: []iam.IamBinding{},
				raitoMasks:           set.NewSet("raitoMask1"),
			},
			args: args{
				ctx: context.Background(),
				configMap: &config.ConfigMap{
					Parameters: map[string]string{common.BqCatalogEnabled: "true"},
				},
			},
			expectedAccessProviders: []sync_from_target.AccessProvider{
				{
					ExternalId: "project_project1_roles_owner",
					Name:       "Project project1 - Owner",
					NamingHint: "project_project1_roles_owner",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"ruben@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					WhoLocked:    ptr.Bool(false),
					WhatLocked:   ptr.Bool(false),
					NameLocked:   ptr.Bool(false),
					DeleteLocked: ptr.Bool(false),
					ActualName:   "project_project1_roles_owner",
					What: []sync_from_target.WhatItem{
						{
							DataObject:  &data_source.DataObjectReference{Type: "datasource", FullName: "project1"},
							Permissions: []string{"roles/owner"},
						},
					},
				},
				{
					ExternalId: "dataset_project/dataset1_roles_editor",
					Name:       "Dataset project/dataset1 - Editor",
					NamingHint: "dataset_project/dataset1_roles_editor",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"dieter@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					WhoLocked:    ptr.Bool(false),
					WhatLocked:   ptr.Bool(false),
					NameLocked:   ptr.Bool(false),
					DeleteLocked: ptr.Bool(false),
					ActualName:   "dataset_project/dataset1_roles_editor",
					What: []sync_from_target.WhatItem{
						{
							DataObject:  &data_source.DataObjectReference{Type: "dataset", FullName: "project/dataset1"},
							Permissions: []string{"roles/editor"},
						},
					},
				},
				{
					ExternalId: "table_project/dataset1/table1_roles_bigquery.dataviewer",
					Name:       "Table project/dataset1/table1 - Bigquery Dataviewer",
					NamingHint: "table_project/dataset1/table1_roles_bigquery.dataviewer",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"thomas@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					WhoLocked:    ptr.Bool(false),
					WhatLocked:   ptr.Bool(false),
					NameLocked:   ptr.Bool(false),
					DeleteLocked: ptr.Bool(false),
					ActualName:   "table_project/dataset1/table1_roles_bigquery.dataviewer",
					What: []sync_from_target.WhatItem{
						{
							DataObject:  &data_source.DataObjectReference{Type: "table", FullName: "project/dataset1/table1"},
							Permissions: []string{"roles/bigquery.dataviewer"},
						},
					},
				},
				{
					ExternalId: "filter1",
					Name:       "filter1",
					NamingHint: "filter1",
					Action:     sync_from_target.Filtered,
					Who: &sync_from_target.WhoItem{
						Users: []string{"bart@raito.io"},
					},
					ActualName: "dataPolicyFilter1ActualName",
					What: []sync_from_target.WhatItem{
						{
							DataObject:  &data_source.DataObjectReference{Type: "table", FullName: "project1/dataset1/table1"},
							Permissions: nil,
						},
					},
					Policy: "table1 > 10",
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer, gcpMock, projectMock, mockMaskingService, filteringService := createAccessSyncer(t, tt.fields.metadata, tt.args.configMap)
			tt.fields.mockSetup(gcpMock, projectMock, mockMaskingService, filteringService)
			syncer.raitoMasks = tt.fields.raitoMasks
			syncer.raitoManagedBindings = set.NewSet(tt.fields.raitoManagedBindings...)

			apHandler := mocks.NewSimpleAccessProviderHandler(t, 4)
			err := syncer.SyncAccessProvidersFromTarget(tt.args.ctx, apHandler, tt.args.configMap)

			tt.wantErr(t, err)
			assert.ElementsMatch(t, tt.expectedAccessProviders, apHandler.AccessProviders)
		})
	}
}

func TestAccessSyncer_ConvertBindingsToAccessProviders(t *testing.T) {
	type fields struct {
		mocksSetup           func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, mockMaskingService *MockMaskingService, filteringService *MockFilteringService)
		metadata             *data_source.MetaData
		raitoManagedBindings set.Set[iam.IamBinding]
	}
	type args struct {
		ctx       context.Context
		configMap *config.ConfigMap
		bindings  []iam.IamBinding
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*sync_from_target.AccessProvider
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Regular bindings to Access Provider and no managed bindings",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {

				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: set.NewSet[iam.IamBinding](),
			},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{},
				bindings: []iam.IamBinding{
					{
						Member:       "user:ruben@raito.io",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/owner",
					},
					{
						Member:       "user:ruben@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/non-raito-managed-binding",
					},
					{
						Member:       "user:dieter@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/viewer",
					},
					{
						Member:       "group:sales@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/viewer",
					},
				},
			},
			want: []*sync_from_target.AccessProvider{
				{
					ExternalId: "project_project1_roles_owner",
					Name:       "Project project1 - Owner",
					NamingHint: "project_project1_roles_owner",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"ruben@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "project_project1_roles_owner",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "project1",
								Type:     "project",
							},
							Permissions: []string{"roles/owner"},
						},
					},
				},
				{
					ExternalId: "folder_folder1_roles_viewer",
					Name:       "Folder folder1 - Viewer",
					NamingHint: "folder_folder1_roles_viewer",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"dieter@raito.io"},
						Groups:          []string{"sales@raito.io"},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "folder_folder1_roles_viewer",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
							Permissions: []string{"roles/viewer"},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Regular bindings to Access Provider and managed bindings",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {

				},
				metadata: gcp.NewDataSourceMetaData(),
				raitoManagedBindings: set.NewSet[iam.IamBinding](
					iam.IamBinding{
						Member:       "user:ruben@raito.io",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/owner",
					},
					iam.IamBinding{
						Member:       "group:sales@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/viewer",
					},
				),
			},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{},
				bindings: []iam.IamBinding{
					{
						Member:       "user:ruben@raito.io",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/owner",
					},
					{
						Member:       "user:ruben@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/editor",
					},
					{
						Member:       "user:dieter@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/viewer",
					},
					{
						Member:       "group:sales@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/viewer",
					},
				},
			},
			want: []*sync_from_target.AccessProvider{
				{
					ExternalId: "folder_folder1_roles_editor",
					Name:       "Folder folder1 - Editor",
					NamingHint: "folder_folder1_roles_editor",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"ruben@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "folder_folder1_roles_editor",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
							Permissions: []string{"roles/editor"},
						},
					},
				},
				{
					ExternalId: "folder_folder1_roles_viewer",
					Name:       "Folder folder1 - Viewer",
					NamingHint: "folder_folder1_roles_viewer",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"dieter@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "folder_folder1_roles_viewer",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
							Permissions: []string{"roles/viewer"},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Regular bindings to Access Provider and include unknown roles",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {

				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: set.NewSet[iam.IamBinding](),
			},
			args: args{
				ctx: context.Background(),
				configMap: &config.ConfigMap{Parameters: map[string]string{
					common.ExcludeNonAplicablePermissions: "false",
				},
				},
				bindings: []iam.IamBinding{
					{
						Member:       "user:ruben@raito.io",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/owner",
					},
					{
						Member:       "user:ruben@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/non-raito-managed-binding",
					},
				},
			},
			want: []*sync_from_target.AccessProvider{
				{
					ExternalId: "project_project1_roles_owner",
					Name:       "Project project1 - Owner",
					NamingHint: "project_project1_roles_owner",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"ruben@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "project_project1_roles_owner",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "project1",
								Type:     "project",
							},
							Permissions: []string{"roles/owner"},
						},
					},
				},
				{
					ExternalId: "folder_folder1_roles_non-raito-managed-binding",
					Name:       "Folder folder1 - Non-Raito-Managed-Binding",
					NamingHint: "folder_folder1_roles_non-raito-managed-binding",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"ruben@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: true,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "folder_folder1_roles_non-raito-managed-binding",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
							Permissions: []string{"roles/non-raito-managed-binding"},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Special bindings for project",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					projectRepo.EXPECT().GetProjectOwner(mock.Anything, mock.Anything).Return([]string{"user:owner@raito.io"}, []string{"user:editor@raito.io"}, []string{"user:viewer@raito.io"}, nil).Once()
				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: set.NewSet[iam.IamBinding](),
			},
			args: args{
				ctx: context.Background(),
				configMap: &config.ConfigMap{Parameters: map[string]string{
					common.GcpProjectId: "projectId",
				},
				},
				bindings: []iam.IamBinding{
					{
						Member:       "special_group:",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/bigquery.dataViewer",
					},
					{
						Member:       "special_group:",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/bigquery.dataViewer",
					},
					{
						Member:       "special_group:",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/bigquery.dataEditor",
					},
					{
						Member:       "special_group:",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/bigquery.dataOwner",
					},
				},
			},
			want: []*sync_from_target.AccessProvider{
				{
					ExternalId: "Project Viewer Mapping",
					Name:       "Project Viewer Mapping",
					NamingHint: "Project Viewer Mapping",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users: []string{"viewer@raito.io"},
					},
					NotInternalizable: true,
					ActualName:        "Project Viewer Mapping",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "project1",
								Type:     "project",
							},
							Permissions: []string{"roles/bigquery.dataViewer"},
						},
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
							Permissions: []string{"roles/bigquery.dataViewer"},
						},
					},
				},
				{
					ExternalId: "Project Editor Mapping",
					Name:       "Project Editor Mapping",
					NamingHint: "Project Editor Mapping",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users: []string{"editor@raito.io"},
					},
					NotInternalizable: true,
					ActualName:        "Project Editor Mapping",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
							Permissions: []string{"roles/bigquery.dataEditor"},
						},
					},
				},
				{
					ExternalId: "Project Owner Mapping",
					Name:       "Project Owner Mapping",
					NamingHint: "Project Owner Mapping",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users: []string{"owner@raito.io"},
					},
					NotInternalizable: true,
					ActualName:        "Project Owner Mapping",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
							Permissions: []string{"roles/bigquery.dataOwner"},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Roles to group by identity",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {

				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: set.NewSet[iam.IamBinding](),
			},
			args: args{
				ctx: context.Background(),
				configMap: &config.ConfigMap{Parameters: map[string]string{
					common.GcpRolesToGroupByIdentity: "roles/viewer,roles/editor",
				},
				},
				bindings: []iam.IamBinding{
					{
						Member:       "user:ruben@raito.io",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/owner",
					},
					{
						Member:       "user:ruben@raito.io",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/viewer",
					},
					{
						Member:       "user:ruben@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/viewer",
					},
					{
						Member:       "user:dieter@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/editor",
					},
					{
						Member:       "user:dieter@raito.io",
						Resource:     "folder2",
						ResourceType: "folder",
						Role:         "roles/viewer",
					},
				},
			},
			want: []*sync_from_target.AccessProvider{
				{
					ExternalId: "project_project1_roles_owner",
					Name:       "Project project1 - Owner",
					NamingHint: "project_project1_roles_owner",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users:           []string{"ruben@raito.io"},
						Groups:          []string{},
						AccessProviders: []string{},
					},
					NotInternalizable: false,
					WhoLocked:         ptr.Bool(false),
					WhatLocked:        ptr.Bool(false),
					NameLocked:        ptr.Bool(false),
					DeleteLocked:      ptr.Bool(false),
					ActualName:        "project_project1_roles_owner",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "project1",
								Type:     "project",
							},
							Permissions: []string{"roles/owner"},
						},
					},
				},
				{
					ExternalId: "Grouped permissions for user ruben@raito.io",
					Name:       "Grouped permissions for user ruben@raito.io",
					NamingHint: "Grouped permissions for user ruben@raito.io",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users: []string{"ruben@raito.io"},
					},
					NotInternalizable: true,
					ActualName:        "Grouped permissions for user ruben@raito.io",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "project1",
								Type:     "project",
							},
							Permissions: []string{"roles/viewer"},
						},
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
							Permissions: []string{"roles/viewer"},
						},
					},
				},
				{
					ExternalId: "Grouped permissions for user dieter@raito.io",
					Name:       "Grouped permissions for user dieter@raito.io",
					NamingHint: "Grouped permissions for user dieter@raito.io",
					Type:       ptr.String(access_provider.AclSet),
					Action:     sync_from_target.Grant,
					Who: &sync_from_target.WhoItem{
						Users: []string{"dieter@raito.io"},
					},
					NotInternalizable: true,
					ActualName:        "Grouped permissions for user dieter@raito.io",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder1",
								Type:     "folder",
							},
							Permissions: []string{"roles/editor"},
						},
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "folder2",
								Type:     "folder",
							},
							Permissions: []string{"roles/viewer"},
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, gcpMock, projectMock, mockMaskingService, filteringService := createAccessSyncer(t, tt.fields.metadata, tt.args.configMap)
			tt.fields.mocksSetup(gcpMock, projectMock, mockMaskingService, filteringService)

			a.raitoManagedBindings = tt.fields.raitoManagedBindings

			got, err := a.ConvertBindingsToAccessProviders(tt.args.ctx, tt.args.configMap, tt.args.bindings)
			if !tt.wantErr(t, err, fmt.Sprintf("ConvertBindingsToAccessProviders(%v, %v, %v)", tt.args.ctx, tt.args.configMap, tt.args.bindings)) {
				return
			}

			assert.ElementsMatchf(t, tt.want, got, "ConvertBindingsToAccessProviders(%v, %v, %v)", tt.args.ctx, tt.args.configMap, tt.args.bindings)
		})
	}
}

func TestAccessSyncer_SyncAccessProviderToTarget(t *testing.T) {
	bqMetadata, err := bigquery.NewDataSourceMetaData(context.Background(), &config.ConfigMap{Parameters: map[string]string{
		common.BqCatalogEnabled: "true",
	}})

	require.NoError(t, err)

	maskInput := importer.AccessProvider{
		Id:          "mask1",
		Name:        "mask1",
		Description: "some mask",
		NamingHint:  "mask1",
		Type:        ptr.String("maskType"),
		ExternalId:  nil,
		Action:      importer.Mask,
		Who: importer.WhoItem{
			Users: []string{
				"bart@raito.io",
			},
			Groups: []string{
				"sales@raito.io",
			},
		},
		Delete: false,
		What: []importer.WhatItem{
			{
				DataObject: &data_source.DataObjectReference{
					FullName: "project1/dataset1/table1/column1",
					Type:     "column",
				},
			},
		},
	}

	filterInput := importer.AccessProvider{
		Id:          "filter1",
		Name:        "filter1",
		Description: "some filter",
		NamingHint:  "filter1",
		ExternalId:  nil,
		Action:      importer.Filtered,
		Who: importer.WhoItem{
			Users: []string{
				"bart@raito.io",
			},
			Groups: []string{
				"sales@raito.io",
			},
		},
		Delete: false,
		What: []importer.WhatItem{
			{
				DataObject: &data_source.DataObjectReference{
					FullName: "project1/dataset1/table1",
					Type:     "table",
				},
			},
		},
		PolicyRule: ptr.String("column1 = 'value1' AND column2 = 'value2'"),
	}

	type fields struct {
		mocksSetup func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService)
		metadata   *data_source.MetaData
	}
	type args struct {
		ctx             context.Context
		accessProviders *importer.AccessProviderImport
		configMap       *config.ConfigMap
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		want                 []importer.AccessProviderSyncFeedback
		expectedBindings     set.Set[iam.IamBinding]
		expectedRaitoMasks   set.Set[string]
		expectedRaitoFilters set.Set[string]
		wantErr              assert.ErrorAssertionFunc
	}{
		{
			name: "No access providers",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {

				},
				metadata: gcp.NewDataSourceMetaData(),
			},
			args: args{
				ctx:             context.Background(),
				accessProviders: &importer.AccessProviderImport{},
				configMap:       &config.ConfigMap{Parameters: map[string]string{}},
			},
			want:                 []importer.AccessProviderSyncFeedback{},
			expectedBindings:     set.NewSet[iam.IamBinding](),
			expectedRaitoMasks:   set.NewSet[string](),
			expectedRaitoFilters: set.NewSet[string](),
			wantErr:              assert.NoError,
		},
		{
			name: "Access provider to binding",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					gcpRepo.EXPECT().UpdateBindings(mock.Anything, &iam.DataObjectReference{
						FullName:   "project1",
						ObjectType: "project",
					}, mock.Anything, []iam.IamBinding{}).Run(func(ctx context.Context, dataObject *iam.DataObjectReference, addBindings []iam.IamBinding, removeBindings []iam.IamBinding) {
						assert.ElementsMatch(t, []iam.IamBinding{
							{
								Member:       "serviceAccount:sa@raito.gserviceaccount.com",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
							{
								Member:       "group:sales@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
							{
								Member:       "user:ruben@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
						}, addBindings)
					}).Return(nil)
					gcpRepo.EXPECT().UpdateBindings(mock.Anything, &iam.DataObjectReference{
						FullName:   "folder1",
						ObjectType: "folder",
					}, mock.Anything, []iam.IamBinding{}).Run(func(ctx context.Context, dataObject *iam.DataObjectReference, addBindings []iam.IamBinding, removeBindings []iam.IamBinding) {
						assert.ElementsMatch(t, []iam.IamBinding{
							{
								Member:       "serviceAccount:sa@raito.gserviceaccount.com",
								Resource:     "folder1",
								ResourceType: "folder",
								Role:         "roles/editor",
							},
							{
								Member:       "user:ruben@raito.io",
								Resource:     "folder1",
								ResourceType: "folder",
								Role:         "roles/editor",
							},
							{
								Member:       "group:sales@raito.io",
								Resource:     "folder1",
								ResourceType: "folder",
								Role:         "roles/editor",
							},
						}, addBindings)
					}).Return(nil)
				},
				metadata: gcp.NewDataSourceMetaData(),
			},
			args: args{
				ctx: context.Background(),
				accessProviders: &importer.AccessProviderImport{AccessProviders: []*importer.AccessProvider{
					{
						Id:          "apId1",
						Name:        "ap1",
						Description: "some description",
						NamingHint:  "ap1",
						Type:        nil,
						ExternalId:  nil,
						Action:      importer.Grant,
						Who: importer.WhoItem{
							Users: []string{
								"ruben@raito.io",
								"sa@raito.gserviceaccount.com",
							},
							Groups: []string{"sales@raito.io"},
						},
						Delete: false,
						What: []importer.WhatItem{
							{
								DataObject: &data_source.DataObjectReference{
									FullName: "project1",
									Type:     "project",
								},
								Permissions: []string{"roles/owner"},
							},
							{
								DataObject: &data_source.DataObjectReference{
									FullName: "folder1",
									Type:     "folder",
								},
								Permissions: []string{"roles/editor"},
							},
						},
						DeleteWhat: nil,
					},
				}},
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			want: []importer.AccessProviderSyncFeedback{
				{
					AccessProvider: "apId1",
					ActualName:     "apId1",
					ExternalId:     nil,
					Type:           ptr.String(access_provider.AclSet),
				},
			},
			expectedBindings: set.NewSet[iam.IamBinding](
				iam.IamBinding{
					Member:       "group:sales@raito.io",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
				iam.IamBinding{
					Member:       "user:ruben@raito.io",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
				iam.IamBinding{
					Member:       "serviceAccount:sa@raito.gserviceaccount.com",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
				iam.IamBinding{
					Member:       "group:sales@raito.io",
					Role:         "roles/editor",
					Resource:     "folder1",
					ResourceType: "folder",
				},
				iam.IamBinding{
					Member:       "user:ruben@raito.io",
					Role:         "roles/editor",
					Resource:     "folder1",
					ResourceType: "folder",
				},
				iam.IamBinding{
					Member:       "serviceAccount:sa@raito.gserviceaccount.com",
					Role:         "roles/editor",
					Resource:     "folder1",
					ResourceType: "folder",
				},
			),
			expectedRaitoMasks:   set.NewSet[string](),
			expectedRaitoFilters: set.NewSet[string](),
			wantErr:              assert.NoError,
		},
		{
			name: "Deleted access provider",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					gcpRepo.EXPECT().UpdateBindings(mock.Anything, &iam.DataObjectReference{
						FullName:   "project1",
						ObjectType: "project",
					}, []iam.IamBinding{}, mock.Anything).Run(func(ctx context.Context, dataObject *iam.DataObjectReference, addBindings []iam.IamBinding, removeBindings []iam.IamBinding) {
						assert.ElementsMatch(t, []iam.IamBinding{
							{
								Member:       "serviceAccount:sa@raito.gserviceaccount.com",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
							{
								Member:       "group:sales@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
							{
								Member:       "user:ruben@raito.io",
								Resource:     "project1",
								ResourceType: "project",
								Role:         "roles/owner",
							},
						}, removeBindings)
					}).Return(nil)
					gcpRepo.EXPECT().UpdateBindings(mock.Anything, &iam.DataObjectReference{
						FullName:   "folder1",
						ObjectType: "folder",
					}, []iam.IamBinding{}, mock.Anything).Run(func(ctx context.Context, dataObject *iam.DataObjectReference, addBindings []iam.IamBinding, removeBindings []iam.IamBinding) {
						assert.ElementsMatch(t, []iam.IamBinding{
							{
								Member:       "serviceAccount:sa@raito.gserviceaccount.com",
								Resource:     "folder1",
								ResourceType: "folder",
								Role:         "roles/editor",
							},
							{
								Member:       "user:ruben@raito.io",
								Resource:     "folder1",
								ResourceType: "folder",
								Role:         "roles/editor",
							},
							{
								Member:       "group:sales@raito.io",
								Resource:     "folder1",
								ResourceType: "folder",
								Role:         "roles/editor",
							},
						}, removeBindings)
					}).Return(nil)
				},
				metadata: gcp.NewDataSourceMetaData(),
			},
			args: args{
				ctx: context.Background(),
				accessProviders: &importer.AccessProviderImport{AccessProviders: []*importer.AccessProvider{
					{
						Id:          "apId1",
						Name:        "ap1",
						Description: "some description",
						NamingHint:  "ap1",
						Type:        nil,
						ExternalId:  nil,
						Action:      importer.Grant,
						Who: importer.WhoItem{
							Users: []string{
								"ruben@raito.io",
								"sa@raito.gserviceaccount.com",
							},
							Groups: []string{"sales@raito.io"},
						},
						Delete: true,
						What: []importer.WhatItem{
							{
								DataObject: &data_source.DataObjectReference{
									FullName: "project1",
									Type:     "project",
								},
								Permissions: []string{"roles/owner"},
							},
							{
								DataObject: &data_source.DataObjectReference{
									FullName: "folder1",
									Type:     "folder",
								},
								Permissions: []string{"roles/editor"},
							},
						},
						DeleteWhat: nil,
					},
				}},
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			want: []importer.AccessProviderSyncFeedback{
				{
					AccessProvider: "apId1",
					ActualName:     "apId1",
					ExternalId:     nil,
					Type:           ptr.String(access_provider.AclSet),
				},
			},
			expectedBindings: set.NewSet[iam.IamBinding](
				iam.IamBinding{
					Member:       "group:sales@raito.io",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
				iam.IamBinding{
					Member:       "user:ruben@raito.io",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
				iam.IamBinding{
					Member:       "serviceAccount:sa@raito.gserviceaccount.com",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
				iam.IamBinding{
					Member:       "group:sales@raito.io",
					Role:         "roles/editor",
					Resource:     "folder1",
					ResourceType: "folder",
				},
				iam.IamBinding{
					Member:       "user:ruben@raito.io",
					Role:         "roles/editor",
					Resource:     "folder1",
					ResourceType: "folder",
				},
				iam.IamBinding{
					Member:       "serviceAccount:sa@raito.gserviceaccount.com",
					Role:         "roles/editor",
					Resource:     "folder1",
					ResourceType: "folder",
				},
			),
			expectedRaitoMasks:   set.NewSet[string](),
			expectedRaitoFilters: set.NewSet[string](),
			wantErr:              assert.NoError,
		},
		{
			name: "Unknown action",
			fields: fields{
				mocksSetup: func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
				},
				metadata: bqMetadata,
			},
			args: args{
				ctx: context.Background(),
				accessProviders: &importer.AccessProviderImport{AccessProviders: []*importer.AccessProvider{
					{
						Id:          "apId1",
						Name:        "ap1",
						Description: "some description",
						NamingHint:  "ap1",
						Type:        nil,
						ExternalId:  nil,
						Action:      importer.Deny,
						Who: importer.WhoItem{
							Users: []string{
								"ruben@raito.io",
							},
						},
						Delete: false,
						What: []importer.WhatItem{
							{
								DataObject: &data_source.DataObjectReference{
									FullName: "project1",
									Type:     "datasource",
								},
								Permissions: []string{"roles/owner"},
							},
						},
						DeleteWhat: nil,
					},
				}},
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			want: []importer.AccessProviderSyncFeedback{
				{
					AccessProvider: "apId1",
					ActualName:     "apId1",
					ExternalId:     nil,
					Errors:         []string{"unsupported action: 2"},
				},
			},
			expectedBindings:     set.NewSet[iam.IamBinding](),
			expectedRaitoMasks:   set.NewSet[string](),
			expectedRaitoFilters: set.NewSet[string](),
			wantErr:              assert.NoError,
		},
		{
			name: "Grants and masks",
			fields: fields{
				mocksSetup: func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					maskingService.EXPECT().ExportMasks(mock.Anything, &maskInput, mock.Anything).RunAndReturn(func(ctx context.Context, provider *importer.AccessProvider, handler wrappers.AccessProviderFeedbackHandler) ([]string, error) {
						err := handler.AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
							AccessProvider: provider.Id,
							ActualName:     provider.Name,
							Type:           provider.Type,
							ExternalId:     &provider.Name,
						})
						if err != nil {
							return nil, err
						}

						return []string{"eu-mask1", "vs-mask1"}, nil
					})

					repo.EXPECT().DataSourceType().Return("project")

					repo.EXPECT().UpdateBindings(mock.Anything, &iam.DataObjectReference{FullName: "project1", ObjectType: "project"}, []iam.IamBinding{{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project1", ResourceType: "project"}}, []iam.IamBinding{}).Return(nil)
				},
				metadata: bqMetadata,
			},
			args: args{
				ctx: context.Background(),
				accessProviders: &importer.AccessProviderImport{AccessProviders: []*importer.AccessProvider{
					{
						Id:          "apId1",
						Name:        "ap1",
						Description: "some description",
						NamingHint:  "ap1",
						Type:        nil,
						ExternalId:  nil,
						Action:      importer.Grant,
						Who: importer.WhoItem{
							Users: []string{
								"ruben@raito.io",
							},
						},
						Delete: false,
						What: []importer.WhatItem{
							{
								DataObject: &data_source.DataObjectReference{
									FullName: "project1",
									Type:     "datasource",
								},
								Permissions: []string{"roles/owner"},
							},
						},
						DeleteWhat: nil,
					},
					&maskInput,
				}},
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			want: []importer.AccessProviderSyncFeedback{
				{
					AccessProvider: "apId1",
					ActualName:     "apId1",
					ExternalId:     nil,
					Type:           ptr.String(access_provider.AclSet),
				},
				{
					AccessProvider: "mask1",
					ActualName:     "mask1",
					ExternalId:     ptr.String("mask1"),
					Type:           ptr.String("maskType"),
				},
			},
			expectedBindings: set.NewSet[iam.IamBinding](
				iam.IamBinding{
					Member:       "user:ruben@raito.io",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
			),
			expectedRaitoMasks:   set.NewSet[string]("eu-mask1", "vs-mask1"),
			expectedRaitoFilters: set.NewSet[string](),
			wantErr:              assert.NoError,
		},
		{
			name: "Grants and filters",
			fields: fields{
				mocksSetup: func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					filteringService.EXPECT().ExportFilter(mock.Anything, &filterInput, mock.Anything).RunAndReturn(func(ctx context.Context, provider *importer.AccessProvider, handler wrappers.AccessProviderFeedbackHandler) (*string, error) {
						err := handler.AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
							AccessProvider: provider.Id,
							ActualName:     provider.Name,
							Type:           provider.Type,
							ExternalId:     &provider.Name,
						})
						if err != nil {
							return nil, err
						}

						return ptr.String("filter1"), nil
					})

					repo.EXPECT().DataSourceType().Return("project")

					repo.EXPECT().UpdateBindings(mock.Anything, &iam.DataObjectReference{FullName: "project1", ObjectType: "project"}, []iam.IamBinding{{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project1", ResourceType: "project"}}, []iam.IamBinding{}).Return(nil)
				},
				metadata: bqMetadata,
			},
			args: args{
				ctx: context.Background(),
				accessProviders: &importer.AccessProviderImport{AccessProviders: []*importer.AccessProvider{
					{
						Id:          "apId1",
						Name:        "ap1",
						Description: "some description",
						NamingHint:  "ap1",
						Type:        nil,
						ExternalId:  nil,
						Action:      importer.Grant,
						Who: importer.WhoItem{
							Users: []string{
								"ruben@raito.io",
							},
						},
						Delete: false,
						What: []importer.WhatItem{
							{
								DataObject: &data_source.DataObjectReference{
									FullName: "project1",
									Type:     "datasource",
								},
								Permissions: []string{"roles/owner"},
							},
						},
						DeleteWhat: nil,
					},
					&filterInput,
				}},
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			want: []importer.AccessProviderSyncFeedback{
				{
					AccessProvider: "apId1",
					ActualName:     "apId1",
					ExternalId:     nil,
					Type:           ptr.String(access_provider.AclSet),
				},
				{
					AccessProvider: "filter1",
					ActualName:     "filter1",
					ExternalId:     ptr.String("filter1"),
				},
			},
			expectedBindings: set.NewSet[iam.IamBinding](
				iam.IamBinding{
					Member:       "user:ruben@raito.io",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
			),
			expectedRaitoMasks:   set.NewSet[string](),
			expectedRaitoFilters: set.NewSet[string]("filter1"),
			wantErr:              assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, gcpMock, projectRepoMock, maskingService, filteringService := createAccessSyncer(t, tt.fields.metadata, tt.args.configMap)
			tt.fields.mocksSetup(gcpMock, projectRepoMock, maskingService, filteringService)

			feedbackHandler := mocks.NewSimpleAccessProviderFeedbackHandler(t)

			if !tt.wantErr(t, a.SyncAccessProviderToTarget(tt.args.ctx, tt.args.accessProviders, feedbackHandler, tt.args.configMap)) {
				return
			}

			assert.ElementsMatch(t, tt.want, feedbackHandler.AccessProviderFeedback)
			assert.Equal(t, tt.expectedBindings, a.raitoManagedBindings)
			assert.Equal(t, tt.expectedRaitoMasks, a.raitoMasks)
			assert.Equal(t, tt.expectedRaitoFilters, a.raitoFilters)
		})
	}
}

func TestAccessSyncer_convertAccessProviderToBindings(t *testing.T) {
	accessProviders := []*importer.AccessProvider{
		{
			Id:          "apId1",
			Name:        "ap1",
			Description: "some description",
			NamingHint:  "ap1",
			Type:        nil,
			ExternalId:  nil,
			Action:      importer.Grant,
			Who: importer.WhoItem{
				Users: []string{
					"ruben@raito.io",
				},
			},
			Delete: false,
			What: []importer.WhatItem{
				{
					DataObject: &data_source.DataObjectReference{
						FullName: "project1",
						Type:     "project",
					},
					Permissions: []string{"roles/owner"},
				},
			},
			DeleteWhat: nil,
		},
		{
			Id:          "apId2",
			Name:        "ap2",
			Description: "some description",
			NamingHint:  "ap2",
			Type:        nil,
			ExternalId:  nil,
			Action:      importer.Grant,
			Who: importer.WhoItem{
				Users: []string{
					"ruben@raito.io",
				},
			},
			Delete: false,
			What: []importer.WhatItem{
				{
					DataObject: &data_source.DataObjectReference{
						FullName: "SomeDataSourceFullName",
						Type:     "datasource",
					},
					Permissions: []string{"roles/owner"},
				},
			},
			DeleteWhat: nil,
		},
		{
			Id:          "apId3",
			Name:        "ap3",
			Description: "some description",
			NamingHint:  "ap3",
			Type:        nil,
			ExternalId:  nil,
			Action:      importer.Grant,
			Who: importer.WhoItem{
				Users: []string{
					"ruben@raito.io",
				},
			},
			DeletedWho: &importer.WhoItem{Users: []string{"michael@raito.io", "oldSa@raito.gserviceaccount.com"},
				Groups: []string{"sales@raito.io"},
			},
			Delete: false,
			What: []importer.WhatItem{
				{
					DataObject: &data_source.DataObjectReference{
						FullName: "project3",
						Type:     "project",
					},
					Permissions: []string{"roles/owner"},
				},
			},
			DeleteWhat: nil,
		},
		{
			Id:          "apId3",
			Name:        "ap3",
			Description: "some description",
			NamingHint:  "ap3",
			Type:        nil,
			ExternalId:  nil,
			Action:      importer.Grant,
			Who: importer.WhoItem{
				Users: []string{
					"ruben@raito.io",
				},
			},
			DeletedWho: &importer.WhoItem{Users: []string{"michael@raito.io"}},
			Delete:     false,
			What: []importer.WhatItem{
				{
					DataObject: &data_source.DataObjectReference{
						FullName: "project4",
						Type:     "project",
					},
					Permissions: []string{"roles/owner"},
				},
			},
			DeleteWhat: []importer.WhatItem{
				{
					DataObject: &data_source.DataObjectReference{
						FullName: "project5",
						Type:     "project",
					},
					Permissions: []string{"roles/editor"},
				},
			},
		},
	}

	type fields struct {
		mocksSetup func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService)
		metadata   *data_source.MetaData
		configMap  *config.ConfigMap
	}
	type args struct {
		ctx             context.Context
		accessProviders []*importer.AccessProvider
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *BindingContainer
	}{
		{
			name: "No accessProviders",
			fields: fields{
				mocksSetup: func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
				},
				metadata:  gcp.NewDataSourceMetaData(),
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			args: args{
				ctx:             context.Background(),
				accessProviders: []*importer.AccessProvider{},
			},
			want: NewBindingContainer(),
		},
		{
			name: "New grant accessProviders",
			fields: fields{
				mocksSetup: func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
				},
				metadata:  gcp.NewDataSourceMetaData(),
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			args: args{
				ctx: context.Background(),
				accessProviders: []*importer.AccessProvider{
					accessProviders[0],
				},
			},
			want: &BindingContainer{
				bindings: map[iam.DataObjectReference]*BindingsForDataObject{
					iam.DataObjectReference{FullName: "project1", ObjectType: "project"}: {
						bindingsToAdd:    set.NewSet(iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project1", ResourceType: "project"}),
						bindingsToDelete: set.NewSet[iam.IamBinding](),
						accessProviders: map[iam.IamBinding][]*importer.AccessProvider{
							iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project1", ResourceType: "project"}: {accessProviders[0]},
						},
					},
				},
			},
		},
		{
			name: "New grant on datasource",
			fields: fields{
				mocksSetup: func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					repo.EXPECT().DataSourceType().Return("datasource_real_type")
				},
				metadata:  gcp.NewDataSourceMetaData(),
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			args: args{
				ctx: context.Background(),
				accessProviders: []*importer.AccessProvider{
					accessProviders[1],
				},
			},
			want: &BindingContainer{
				bindings: map[iam.DataObjectReference]*BindingsForDataObject{
					iam.DataObjectReference{FullName: "SomeDataSourceFullName", ObjectType: "datasource_real_type"}: {
						bindingsToAdd:    set.NewSet(iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "SomeDataSourceFullName", ResourceType: "datasource_real_type"}),
						bindingsToDelete: set.NewSet[iam.IamBinding](),
						accessProviders: map[iam.IamBinding][]*importer.AccessProvider{
							iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "SomeDataSourceFullName", ResourceType: "datasource_real_type"}: {accessProviders[1]},
						},
					},
				},
			},
		},
		{
			name: "Grant with deleted members",
			fields: fields{
				mocksSetup: func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
				},
				metadata:  gcp.NewDataSourceMetaData(),
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			args: args{
				ctx: context.Background(),
				accessProviders: []*importer.AccessProvider{
					accessProviders[2],
				},
			},
			want: &BindingContainer{
				bindings: map[iam.DataObjectReference]*BindingsForDataObject{
					iam.DataObjectReference{FullName: "project3", ObjectType: "project"}: {
						bindingsToAdd:    set.NewSet(iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project3", ResourceType: "project"}),
						bindingsToDelete: set.NewSet(iam.IamBinding{Member: "serviceAccount:oldSa@raito.gserviceaccount.com", Role: "roles/owner", Resource: "project3", ResourceType: "project"}, iam.IamBinding{Member: "user:michael@raito.io", Role: "roles/owner", Resource: "project3", ResourceType: "project"}, iam.IamBinding{Member: "group:sales@raito.io", Role: "roles/owner", Resource: "project3", ResourceType: "project"}),
						accessProviders: map[iam.IamBinding][]*importer.AccessProvider{
							iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project3", ResourceType: "project"}:                            {accessProviders[2]},
							iam.IamBinding{Member: "serviceAccount:oldSa@raito.gserviceaccount.com", Role: "roles/owner", Resource: "project3", ResourceType: "project"}: {accessProviders[2]},
							iam.IamBinding{Member: "user:michael@raito.io", Role: "roles/owner", Resource: "project3", ResourceType: "project"}:                          {accessProviders[2]},
							iam.IamBinding{Member: "group:sales@raito.io", Role: "roles/owner", Resource: "project3", ResourceType: "project"}:                           {accessProviders[2]},
						},
					},
				},
			},
		},
		{
			name: "Grant with deleted what",
			fields: fields{
				mocksSetup: func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
				},
				metadata:  gcp.NewDataSourceMetaData(),
				configMap: &config.ConfigMap{Parameters: map[string]string{}},
			},
			args: args{
				ctx: context.Background(),
				accessProviders: []*importer.AccessProvider{
					accessProviders[3],
				},
			},
			want: &BindingContainer{
				bindings: map[iam.DataObjectReference]*BindingsForDataObject{
					iam.DataObjectReference{FullName: "project4", ObjectType: "project"}: {
						bindingsToAdd:    set.NewSet(iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project4", ResourceType: "project"}),
						bindingsToDelete: set.NewSet(iam.IamBinding{Member: "user:michael@raito.io", Role: "roles/owner", Resource: "project4", ResourceType: "project"}),
						accessProviders: map[iam.IamBinding][]*importer.AccessProvider{
							iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project4", ResourceType: "project"}:   {accessProviders[3]},
							iam.IamBinding{Member: "user:michael@raito.io", Role: "roles/owner", Resource: "project4", ResourceType: "project"}: {accessProviders[3]},
						},
					},
					iam.DataObjectReference{FullName: "project5", ObjectType: "project"}: {
						bindingsToAdd:    set.NewSet[iam.IamBinding](),
						bindingsToDelete: set.NewSet(iam.IamBinding{Member: "user:michael@raito.io", Role: "roles/editor", Resource: "project5", ResourceType: "project"}, iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/editor", Resource: "project5", ResourceType: "project"}),
						accessProviders: map[iam.IamBinding][]*importer.AccessProvider{
							iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/editor", Resource: "project5", ResourceType: "project"}:   {accessProviders[3]},
							iam.IamBinding{Member: "user:michael@raito.io", Role: "roles/editor", Resource: "project5", ResourceType: "project"}: {accessProviders[3]},
						},
					},
				},
			},
		},
		{
			name: "Grant with masked reader",
			fields: fields{
				mocksSetup: func(repo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService, filteringService *MockFilteringService) {
					maskingService.EXPECT().MaskedBinding(mock.Anything, []string{"user:ruben@raito.io"}).Return([]iam.IamBinding{
						{
							Member:       "user:ruben@raito.io",
							Role:         roles.RolesBigQueryMaskedReader.Name,
							Resource:     "project1",
							ResourceType: "project",
						},
					}, nil)
				},
				metadata:  gcp.NewDataSourceMetaData(),
				configMap: &config.ConfigMap{Parameters: map[string]string{common.GcpMaskedReader: "true"}},
			},
			args: args{
				ctx: context.Background(),
				accessProviders: []*importer.AccessProvider{
					accessProviders[0],
				},
			},
			want: &BindingContainer{
				bindings: map[iam.DataObjectReference]*BindingsForDataObject{
					iam.DataObjectReference{FullName: "project1", ObjectType: "project"}: {
						bindingsToAdd: set.NewSet(iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project1", ResourceType: "project"},
							iam.IamBinding{Member: "user:ruben@raito.io", Role: roles.RolesBigQueryMaskedReader.Name, Resource: "project1", ResourceType: "project"}),
						bindingsToDelete: set.NewSet[iam.IamBinding](),
						accessProviders: map[iam.IamBinding][]*importer.AccessProvider{
							iam.IamBinding{Member: "user:ruben@raito.io", Role: "roles/owner", Resource: "project1", ResourceType: "project"}: {accessProviders[0]},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer, repoMock, projectRepoMock, maskingService, filteringService := createAccessSyncer(t, tt.fields.metadata, tt.fields.configMap)
			tt.fields.mocksSetup(repoMock, projectRepoMock, maskingService, filteringService)

			result := syncer.convertAccessProviderToBindings(tt.args.ctx, tt.args.accessProviders)

			assert.Equal(t, len(result.bindings), len(tt.want.bindings))

			for k, v := range tt.want.bindings {
				assert.Equal(t, v.bindingsToAdd, result.bindings[k].bindingsToAdd)
				assert.Equal(t, v.bindingsToDelete, result.bindings[k].bindingsToDelete)

				for binding, aps := range v.accessProviders {
					assert.ElementsMatch(t, aps, result.bindings[k].accessProviders[binding])
				}
			}
		})
	}
}

func createAccessSyncer(t *testing.T, dsMetadata *data_source.MetaData, configMap *config.ConfigMap) (*AccessSyncer, *MockBindingRepository, *MockProjectRepo, *MockMaskingService, *MockFilteringService) {
	t.Helper()

	gcpRepo := NewMockBindingRepository(t)
	projectRepo := NewMockProjectRepo(t)
	maskingService := NewMockMaskingService(t)
	filteringService := NewMockFilteringService(t)

	return NewDataAccessSyncer(gcpRepo, projectRepo, maskingService, filteringService, dsMetadata, configMap), gcpRepo, projectRepo, maskingService, filteringService
}

func Test_handleErrors(t *testing.T) {
	type args struct {
		err        error
		apFeedback map[string]*importer.AccessProviderSyncFeedback
		aps        []*importer.AccessProvider
	}
	type want struct {
		apFeedback map[string]*importer.AccessProviderSyncFeedback
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "No errors",
			args: args{
				err: nil,
				apFeedback: map[string]*importer.AccessProviderSyncFeedback{
					"ap1": {
						AccessProvider: "ap1",
						ActualName:     "ap1",
					},
					"ap2": {
						AccessProvider: "ap2",
						ActualName:     "ap2",
					},
				},
				aps: []*importer.AccessProvider{},
			},
			want: want{apFeedback: map[string]*importer.AccessProviderSyncFeedback{
				"ap1": {
					AccessProvider: "ap1",
					ActualName:     "ap1",
				},
				"ap2": {
					AccessProvider: "ap2",
					ActualName:     "ap2",
				},
			}},
		},
		{
			name: "Error for single ap",
			args: args{
				err: errors.New("some error"),
				apFeedback: map[string]*importer.AccessProviderSyncFeedback{
					"ap1": {
						AccessProvider: "ap1",
						ActualName:     "ap1",
					},
					"ap2": {
						AccessProvider: "ap2",
						ActualName:     "ap2",
					},
				},
				aps: []*importer.AccessProvider{
					{
						Id: "ap1",
					},
				},
			},
			want: want{apFeedback: map[string]*importer.AccessProviderSyncFeedback{
				"ap1": {
					AccessProvider: "ap1",
					ActualName:     "ap1",
					Errors:         []string{"some error"},
				},
				"ap2": {
					AccessProvider: "ap2",
					ActualName:     "ap2",
				},
			}},
		},
		{
			name: "Error for multiple ap",
			args: args{
				err: errors.New("some error"),
				apFeedback: map[string]*importer.AccessProviderSyncFeedback{
					"ap1": {
						AccessProvider: "ap1",
						ActualName:     "ap1",
					},
					"ap2": {
						AccessProvider: "ap2",
						ActualName:     "ap2",
					},
				},
				aps: []*importer.AccessProvider{
					{
						Id: "ap1",
					},
					{
						Id: "ap2",
					},
				},
			},
			want: want{apFeedback: map[string]*importer.AccessProviderSyncFeedback{
				"ap1": {
					AccessProvider: "ap1",
					ActualName:     "ap1",
					Errors:         []string{"some error"},
				},
				"ap2": {
					AccessProvider: "ap2",
					ActualName:     "ap2",
					Errors:         []string{"some error"},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleErrors(tt.args.err, tt.args.apFeedback, tt.args.aps)
		})
	}
}

func Test_generateProjectWhoItem(t *testing.T) {
	type args struct {
		projectOwnerIds []string
	}
	tests := []struct {
		name string
		args args
		want *sync_from_target.WhoItem
	}{
		{
			name: "No project owner",
			args: args{
				projectOwnerIds: []string{},
			},
			want: &sync_from_target.WhoItem{},
		},
		{
			name: "user owner",
			args: args{
				projectOwnerIds: []string{"user:ruben@raito.io"},
			},
			want: &sync_from_target.WhoItem{
				Users: []string{"ruben@raito.io"},
			},
		},
		{
			name: "service account owner",
			args: args{
				projectOwnerIds: []string{"serviceAccount:sa@raito.io"},
			},
			want: &sync_from_target.WhoItem{
				Users: []string{"sa@raito.io"},
			},
		},
		{
			name: "group owner",
			args: args{
				projectOwnerIds: []string{"group:sales@raito.io"},
			},
			want: &sync_from_target.WhoItem{
				Groups: []string{"sales@raito.io"},
			},
		},
		{
			name: "ignore unknown types",
			args: args{
				projectOwnerIds: []string{"deleted_user:michael@raito.io"},
			},
			want: &sync_from_target.WhoItem{},
		},
		{
			name: "Combine all",
			args: args{
				projectOwnerIds: []string{"user:ruben@raito.io", "serviceAccount:sa@raito.io", "group:sales@raito.io"},
			},
			want: &sync_from_target.WhoItem{
				Users:  []string{"ruben@raito.io", "sa@raito.io"},
				Groups: []string{"sales@raito.io"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, generateProjectWhoItem(tt.args.projectOwnerIds), "generateProjectWhoItem(%v)", tt.args.projectOwnerIds)
		})
	}
}
