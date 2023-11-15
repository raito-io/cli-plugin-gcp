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
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/gcp"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func TestAccessSyncer_SyncAccessProvidersFromTarget(t *testing.T) {
	type fields struct {
		mockSetup            func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService)
		metadata             *data_source.MetaData
		raitoManagedBindings []iam.IamBinding
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
				mockSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {
					gcpRepo.EXPECT().Bindings(mock.Anything, mock.Anything).Return(nil)
				},
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
				mockSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {
					gcpRepo.EXPECT().Bindings(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, f func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
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
				configMap: &config.ConfigMap{},
			},
			expectedAccessProviders: []sync_from_target.AccessProvider{
				{
					ExternalId: "project_project1_roles_owner",
					Name:       "project_project1_roles_owner",
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
				mockSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {
					gcpRepo.EXPECT().Bindings(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, f func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
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
					Name:       "project_project1_roles_owner",
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
					Name:       "project_project1_roles_viewer",
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
					Name:       "folder_folder1_roles_editor",
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
					Name:       "folder_folder1_roles_viewer",
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
				mockSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {
					gcpRepo.EXPECT().Bindings(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, f func(context.Context, *org.GcpOrgEntity, []iam.IamBinding) error) error {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer, gcpMock, projectMock, mockMaskingService := createAccessSyncer(t, tt.fields.metadata, tt.args.configMap)
			tt.fields.mockSetup(gcpMock, projectMock, mockMaskingService)

			apHandler := mocks.NewSimpleAccessProviderHandler(t, 4)
			err := syncer.SyncAccessProvidersFromTarget(tt.args.ctx, apHandler, tt.args.configMap)

			tt.wantErr(t, err)
			assert.ElementsMatch(t, tt.expectedAccessProviders, apHandler.AccessProviders)
		})
	}
}

func TestAccessSyncer_ConvertBindingsToAccessProviders(t *testing.T) {
	type fields struct {
		mocksSetup           func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, mockMaskingService *MockMaskingService)
		metadata             *data_source.MetaData
		raitoManagedBindings []iam.IamBinding
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
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {

				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: []iam.IamBinding{},
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
					Name:       "project_project1_roles_owner",
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
					Name:       "folder_folder1_roles_viewer",
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
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {

				},
				metadata: gcp.NewDataSourceMetaData(),
				raitoManagedBindings: []iam.IamBinding{
					{
						Member:       "user:ruben@raito.io",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/owner",
					},
					{
						Member:       "group:sales@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/viewer",
					},
				},
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
					Name:       "folder_folder1_roles_editor",
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
					Name:       "folder_folder1_roles_viewer",
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
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {

				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: []iam.IamBinding{},
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
					Name:       "project_project1_roles_owner",
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
					Name:       "folder_folder1_roles_non-raito-managed-binding",
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
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {
					projectRepo.EXPECT().GetProjectOwner(mock.Anything, mock.Anything).Return([]string{"user:owner@raito.io"}, []string{"user:editor@raito.io"}, []string{"user:viewer@raito.io"}, nil).Once()
				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: []iam.IamBinding{},
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
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {

				},
				metadata:             gcp.NewDataSourceMetaData(),
				raitoManagedBindings: []iam.IamBinding{},
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
					Name:       "project_project1_roles_owner",
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
			a, gcpMock, projectMock, mockMaskingService := createAccessSyncer(t, tt.fields.metadata, tt.args.configMap)
			tt.fields.mocksSetup(gcpMock, projectMock, mockMaskingService)

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
	type fields struct {
		mocksSetup func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService)
		metadata   *data_source.MetaData
	}
	type args struct {
		ctx             context.Context
		accessProviders *importer.AccessProviderImport
		configMap       *config.ConfigMap
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		want             []importer.AccessProviderSyncFeedback
		expectedBindings []iam.IamBinding
		wantErr          assert.ErrorAssertionFunc
	}{
		{
			name: "No access providers",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {

				},
				metadata: gcp.NewDataSourceMetaData(),
			},
			args: args{
				ctx:             context.Background(),
				accessProviders: &importer.AccessProviderImport{},
				configMap:       &config.ConfigMap{},
			},
			want:             []importer.AccessProviderSyncFeedback{},
			expectedBindings: []iam.IamBinding{},
			wantErr:          assert.NoError,
		},
		{
			name: "Access provider to binding",
			fields: fields{
				mocksSetup: func(gcpRepo *MockBindingRepository, projectRepo *MockProjectRepo, maskingService *MockMaskingService) {
					gcpRepo.EXPECT().AddBinding(mock.Anything, iam.IamBinding{
						Member:       "serviceAccount:sa@raito.gserviceaccount.com",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/owner",
					}).Return(nil).Once()

					gcpRepo.EXPECT().AddBinding(mock.Anything, iam.IamBinding{
						Member:       "user:ruben@raito.io",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/owner",
					}).Return(nil).Once()

					gcpRepo.EXPECT().AddBinding(mock.Anything, iam.IamBinding{
						Member:       "group:sales@raito.io",
						Resource:     "project1",
						ResourceType: "project",
						Role:         "roles/owner",
					}).Return(nil).Once()

					gcpRepo.EXPECT().AddBinding(mock.Anything, iam.IamBinding{
						Member:       "serviceAccount:sa@raito.gserviceaccount.com",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/editor",
					}).Return(nil).Once()

					gcpRepo.EXPECT().AddBinding(mock.Anything, iam.IamBinding{
						Member:       "user:ruben@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/editor",
					}).Return(nil).Once()

					gcpRepo.EXPECT().AddBinding(mock.Anything, iam.IamBinding{
						Member:       "group:sales@raito.io",
						Resource:     "folder1",
						ResourceType: "folder",
						Role:         "roles/editor",
					}).Return(nil).Once()
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
			},
			want: []importer.AccessProviderSyncFeedback{
				{
					AccessProvider: "apId1",
					ActualName:     "apId1",
					ExternalId:     nil,
					Type:           ptr.String(access_provider.AclSet),
				},
			},
			expectedBindings: []iam.IamBinding{
				{
					Member:       "group:sales@raito.io",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
				{
					Member:       "user:ruben@raito.io",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
				{
					Member:       "serviceAccount:sa@raito.gserviceaccount.com",
					Role:         "roles/owner",
					Resource:     "project1",
					ResourceType: "project",
				},
				{
					Member:       "group:sales@raito.io",
					Role:         "roles/editor",
					Resource:     "folder1",
					ResourceType: "folder",
				},
				{
					Member:       "user:ruben@raito.io",
					Role:         "roles/editor",
					Resource:     "folder1",
					ResourceType: "folder",
				},
				{
					Member:       "serviceAccount:sa@raito.gserviceaccount.com",
					Role:         "roles/editor",
					Resource:     "folder1",
					ResourceType: "folder",
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, gcpMock, projectRepoMock, maskingService := createAccessSyncer(t, tt.fields.metadata, tt.args.configMap)
			tt.fields.mocksSetup(gcpMock, projectRepoMock, maskingService)

			feedbackHandler := mocks.NewSimpleAccessProviderFeedbackHandler(t)

			if !tt.wantErr(t, a.SyncAccessProviderToTarget(tt.args.ctx, tt.args.accessProviders, feedbackHandler, tt.args.configMap)) {
				return
			}

			assert.ElementsMatch(t, tt.want, feedbackHandler.AccessProviderFeedback)
			assert.ElementsMatch(t, tt.expectedBindings, a.raitoManagedBindings)
		})
	}
}

func createAccessSyncer(t *testing.T, dsMetadata *data_source.MetaData, configMap *config.ConfigMap) (*AccessSyncer, *MockBindingRepository, *MockProjectRepo, *MockMaskingService) {
	t.Helper()

	gcpRepo := NewMockBindingRepository(t)
	projectRepo := NewMockProjectRepo(t)
	maskingService := NewMockMaskingService(t)

	return NewDataAccessSyncer(gcpRepo, projectRepo, maskingService, dsMetadata, configMap), gcpRepo, projectRepo, maskingService
}
