package org

import (
	"context"
	"testing"

	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

func TestOrgIdenityStoreSyncer_GetUsers(t *testing.T) {
	type fields struct {
		mockSetup func(configMap *config.ConfigMap, adminRepo *MockAdminRepository, projectRepository *mockProjectRepository, gcpDataIterator *mockGcpDataIterator)
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedUsers []*iam.UserEntity
		wantErr       require.ErrorAssertionFunc
	}{
		{
			name: "no users",
			fields: fields{
				mockSetup: func(_ *config.ConfigMap, adminRepo *MockAdminRepository, projectRepository *mockProjectRepository, gcpDataIterator *mockGcpDataIterator) {
					adminRepo.EXPECT().GetUsers(mock.Anything, mock.Anything).Return(nil)
					gcpDataIterator.EXPECT().DataObjects(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				},
			},
			args: args{
				ctx: context.Background(),
			},
			expectedUsers: []*iam.UserEntity{},
			wantErr:       require.NoError,
		},
		{
			name: "Only admin users, no projects",
			fields: fields{
				mockSetup: func(_ *config.ConfigMap, adminRepo *MockAdminRepository, projectRepository *mockProjectRepository, gcpDataIterator *mockGcpDataIterator) {
					adminRepo.EXPECT().GetUsers(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, f func(context.Context, *iam.UserEntity) error) error {
						err := f(ctx, &iam.UserEntity{ExternalId: "1", Name: "user1", Email: "user1@raito.io"})
						if err != nil {
							return err
						}

						return f(ctx, &iam.UserEntity{ExternalId: "2", Name: "user2", Email: "user2@raito.io"})
					})
					gcpDataIterator.EXPECT().DataObjects(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				},
			},
			args: args{
				ctx: context.Background(),
			},
			expectedUsers: []*iam.UserEntity{
				{ExternalId: "1", Name: "user1", Email: "user1@raito.io"},
				{ExternalId: "2", Name: "user2", Email: "user2@raito.io"},
			},
			wantErr: require.NoError,
		},
		{
			name: "Admin and project users",
			fields: fields{
				mockSetup: func(_ *config.ConfigMap, adminRepo *MockAdminRepository, projectRepository *mockProjectRepository, gcpDataIterator *mockGcpDataIterator) {
					adminRepo.EXPECT().GetUsers(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, f func(context.Context, *iam.UserEntity) error) error {
						err := f(ctx, &iam.UserEntity{ExternalId: "1", Name: "user1", Email: "user1@raito.io"})
						if err != nil {
							return err
						}

						return f(ctx, &iam.UserEntity{ExternalId: "2", Name: "user2", Email: "user2@raito.io"})
					})
					gcpDataIterator.EXPECT().DataObjects(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, syncConfig *data_source.DataSourceSyncConfig, f func(context.Context, *GcpOrgEntity) error) error {
						err := f(ctx, &GcpOrgEntity{Id: "1", Type: TypeProject, Name: "project1"})
						if err != nil {
							return err
						}

						return f(ctx, &GcpOrgEntity{Id: "2", Type: TypeProject, Name: "project2"})
					})

					projectRepository.EXPECT().GetUsers(mock.Anything, "projects/1", mock.Anything).RunAndReturn(func(ctx context.Context, s string, f func(context.Context, *iam.UserEntity) error) error {
						return f(ctx, &iam.UserEntity{ExternalId: "3", Name: "user3"})
					})

					projectRepository.EXPECT().GetUsers(mock.Anything, "projects/2", mock.Anything).RunAndReturn(func(ctx context.Context, s string, f func(context.Context, *iam.UserEntity) error) error {
						err := f(ctx, &iam.UserEntity{ExternalId: "4", Name: "user4"})
						if err != nil {
							return err
						}

						return f(ctx, &iam.UserEntity{ExternalId: "5", Name: "user5"})
					})
				},
			},
			args: args{
				ctx: context.Background(),
			},
			expectedUsers: []*iam.UserEntity{
				{ExternalId: "1", Name: "user1", Email: "user1@raito.io"},
				{ExternalId: "2", Name: "user2", Email: "user2@raito.io"},
				{ExternalId: "3", Name: "user3"},
				{ExternalId: "4", Name: "user4"},
				{ExternalId: "5", Name: "user5"},
			},
			wantErr: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, configMap, adminRepo, projectRepository, gcpDataIterator := createOrgIdentityStoreSyncer(t)
			tt.fields.mockSetup(configMap, adminRepo, projectRepository, gcpDataIterator)

			var result []*iam.UserEntity

			err := r.GetUsers(tt.args.ctx, func(ctx context.Context, entity *iam.UserEntity) error {
				result = append(result, entity)

				return nil
			})

			tt.wantErr(t, err)

			if err != nil {
				return
			}

			assert.ElementsMatch(t, tt.expectedUsers, result)
		})
	}
}

func TestOrgIdenityStoreSyncer_GetGroups(t *testing.T) {
	type fields struct {
		mockSetup func(configMap *config.ConfigMap, adminRepo *MockAdminRepository)
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedGroups []*iam.GroupEntity
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "no groups",
			fields: fields{
				mockSetup: func(_ *config.ConfigMap, adminRepo *MockAdminRepository) {
					adminRepo.EXPECT().GetGroups(mock.Anything, mock.Anything).Return(nil)
				},
			},
			args: args{
				ctx: context.Background(),
			},
			expectedGroups: []*iam.GroupEntity{},
			wantErr:        require.NoError,
		},
		{
			name: "Basic groups",
			fields: fields{
				mockSetup: func(_ *config.ConfigMap, adminRepo *MockAdminRepository) {
					adminRepo.EXPECT().GetGroups(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, f func(context.Context, *iam.GroupEntity) error) error {
						err := f(ctx, &iam.GroupEntity{ExternalId: "group1", Email: "group1@raito.io"})
						if err != nil {
							return err
						}

						return f(ctx, &iam.GroupEntity{ExternalId: "group2", Email: "group2@raito.io"})
					})
				},
			},
			args: args{
				ctx: context.Background(),
			},
			expectedGroups: []*iam.GroupEntity{
				{ExternalId: "group1", Email: "group1@raito.io"},
				{ExternalId: "group2", Email: "group2@raito.io"},
			},
			wantErr: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, configMap, adminRepo, _, _ := createOrgIdentityStoreSyncer(t)
			tt.fields.mockSetup(configMap, adminRepo)

			var result []*iam.GroupEntity

			err := r.GetGroups(tt.args.ctx, func(ctx context.Context, entity *iam.GroupEntity) error {
				result = append(result, entity)

				return nil
			})

			tt.wantErr(t, err)

			if err != nil {
				return
			}

			assert.ElementsMatch(t, tt.expectedGroups, result)
		})
	}
}

func createOrgIdentityStoreSyncer(t *testing.T) (*OrgIdenityStoreSyncer, *config.ConfigMap, *MockAdminRepository, *mockProjectRepository, *mockGcpDataIterator) {
	t.Helper()

	configMap := &config.ConfigMap{Parameters: map[string]string{
		common.GsuiteCustomerId:         "GSUITE_CUSTOMER_ID",
		common.GsuiteImpersonateSubject: "GSUITE_IMPERSONATE_SUBJECT",
		common.GcpProjectId:             "GCP_PROJECT_ID",
		common.GcpSAFileLocation:        "GOOGLE_APPLICATION_CREDENTIALS",
		common.GcpOrgId:                 "GCP_ORGANIZATION_ID",
	}}

	adminRepo := &MockAdminRepository{}
	projectRepo := &mockProjectRepository{}
	gcpDataIterator := &mockGcpDataIterator{}

	return NewOrgIdentityStoreSyncer(configMap, adminRepo, projectRepo, gcpDataIterator), configMap, adminRepo, projectRepo, gcpDataIterator
}
