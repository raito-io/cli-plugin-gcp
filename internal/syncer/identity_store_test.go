package syncer

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/identity_store"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/gcp"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

func TestIdentityStoreSyncer_SyncIdentityStore(t *testing.T) {
	type fields struct {
		mockSetup func(adminRepoMock *MockAdminRepository)
	}
	type args struct {
		ctx       context.Context
		configMap *config.ConfigMap
	}
	type expected struct {
		groups []identity_store.Group
		users  []identity_store.User
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected expected
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "No users and groups",
			fields: fields{mockSetup: func(adminRepoMock *MockAdminRepository) {
				adminRepoMock.EXPECT().GetGroups(mock.Anything, mock.Anything).Return(nil)
				adminRepoMock.EXPECT().GetUsers(mock.Anything, mock.Anything).Return(nil)
			}},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{Parameters: map[string]string{common.GsuiteIdentityStoreSync: "true"}},
			},
			expected: expected{groups: []identity_store.Group{}, users: []identity_store.User{}},
			wantErr:  assert.NoError,
		},
		{
			name: "Users in gcp",
			fields: fields{mockSetup: func(adminRepoMock *MockAdminRepository) {
				adminRepoMock.EXPECT().GetGroups(mock.Anything, mock.Anything).Return(nil)
				adminRepoMock.EXPECT().GetUsers(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context, *iam.UserEntity) error) error {
					err := fn(ctx, &iam.UserEntity{ExternalId: "user:dieter@raitio.io", Email: "dieter@raito.io", Name: "Dieter Wachters"})
					if err != nil {
						return err
					}

					return fn(ctx, &iam.UserEntity{ExternalId: "user:ruben@raitio.io", Email: "ruben@raito.io", Name: "Ruben Mennes"})
				})
			}},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{Parameters: map[string]string{common.GsuiteIdentityStoreSync: "true"}},
			},
			expected: expected{
				groups: []identity_store.Group{},
				users: []identity_store.User{
					{
						ExternalId: "user:dieter@raitio.io",
						Email:      "dieter@raito.io",
						Name:       "Dieter Wachters",
						UserName:   "dieter@raito.io",
					},
					{
						ExternalId: "user:ruben@raitio.io",
						Email:      "ruben@raito.io",
						Name:       "Ruben Mennes",
						UserName:   "ruben@raito.io",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Service Account users in gcp",
			fields: fields{mockSetup: func(adminRepoMock *MockAdminRepository) {
				adminRepoMock.EXPECT().GetGroups(mock.Anything, mock.Anything).Return(nil)
				adminRepoMock.EXPECT().GetUsers(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context, *iam.UserEntity) error) error {
					err := fn(ctx, &iam.UserEntity{ExternalId: "user:fake_service_account", Email: "service-account@raito.io", Name: "Fake service account"})
					if err != nil {
						return err
					}

					err = fn(ctx, &iam.UserEntity{ExternalId: "user:service_account_1", Email: "113061545108@cloudservices.gserviceaccount.com", Name: "Service account 1"})
					if err != nil {
						return err
					}

					return fn(ctx, &iam.UserEntity{ExternalId: "user:service_account_2", Email: "data-engineer-sync@bq-demodata.iam.gserviceaccount.com", Name: "Service account 2"})
				})
			}},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{Parameters: map[string]string{common.GsuiteIdentityStoreSync: "true"}},
			},
			expected: expected{
				groups: []identity_store.Group{},
				users: []identity_store.User{
					{
						ExternalId: "user:fake_service_account",
						Email:      "service-account@raito.io",
						Name:       "Fake service account",
						UserName:   "service-account@raito.io",
					},
					{
						ExternalId: "user:service_account_1",
						Email:      "113061545108@cloudservices.gserviceaccount.com",
						Name:       "Service account 1",
						UserName:   "113061545108@cloudservices.gserviceaccount.com",
						IsMachine:  ptr.Bool(true),
					},
					{
						ExternalId: "user:service_account_2",
						Email:      "data-engineer-sync@bq-demodata.iam.gserviceaccount.com",
						Name:       "Service account 2",
						UserName:   "data-engineer-sync@bq-demodata.iam.gserviceaccount.com",
						IsMachine:  ptr.Bool(true),
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Groups and users in gcp and bindings",
			fields: fields{mockSetup: func(adminRepoMock *MockAdminRepository) {
				adminRepoMock.EXPECT().GetGroups(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context, *iam.GroupEntity) error) error {
					err := fn(ctx, &iam.GroupEntity{ExternalId: "group:admin@raito.io", Email: "administrators@raito.io", Members: []string{"user:dieter@raito.io", "serviceAccount:sa@raito.io"}})
					if err != nil {
						return err
					}

					return fn(ctx, &iam.GroupEntity{ExternalId: "group:engineers@raito.io", Email: "engineers@raito.io", Members: []string{"user:ruben@raito.io", "group:admin@raito.io", "user:dieter@raito.io"}})
				})
				adminRepoMock.EXPECT().GetUsers(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context, *iam.UserEntity) error) error {
					err := fn(ctx, &iam.UserEntity{ExternalId: "user:dieter@raito.io", Email: "dieter@raito.io", Name: "Dieter Wachters"})
					if err != nil {
						return err
					}

					err = fn(ctx, &iam.UserEntity{ExternalId: "serviceAccount:sa@raito.io", Email: "sa@raito.io", Name: "sa@raito.io"})
					if err != nil {
						return err
					}

					return fn(ctx, &iam.UserEntity{ExternalId: "user:ruben@raito.io", Email: "ruben@raito.io", Name: "Ruben Mennes"})
				})
			}},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{Parameters: map[string]string{common.GsuiteIdentityStoreSync: "true"}},
			},
			expected: expected{
				groups: []identity_store.Group{
					{
						ExternalId:  "group:engineers@raito.io",
						Name:        "engineers@raito.io",
						DisplayName: "engineers@raito.io",
					},
					{
						ExternalId:             "group:admin@raito.io",
						Name:                   "administrators@raito.io",
						DisplayName:            "administrators@raito.io",
						ParentGroupExternalIds: []string{"group:engineers@raito.io"},
					},
				},
				users: []identity_store.User{
					{
						ExternalId:       "user:dieter@raito.io",
						Email:            "dieter@raito.io",
						Name:             "Dieter Wachters",
						UserName:         "dieter@raito.io",
						GroupExternalIds: []string{"group:admin@raito.io", "group:engineers@raito.io"},
					},
					{
						ExternalId:       "user:ruben@raito.io",
						Email:            "ruben@raito.io",
						Name:             "Ruben Mennes",
						UserName:         "ruben@raito.io",
						GroupExternalIds: []string{"group:engineers@raito.io"},
					},
					{
						ExternalId:       "serviceAccount:sa@raito.io",
						Name:             "sa@raito.io",
						UserName:         "sa@raito.io",
						Email:            "sa@raito.io",
						GroupExternalIds: []string{"group:admin@raito.io"},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error during processing",
			fields: fields{mockSetup: func(adminRepoMock *MockAdminRepository) {
				adminRepoMock.EXPECT().GetGroups(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context, *iam.GroupEntity) error) error {
					err := fn(ctx, &iam.GroupEntity{ExternalId: "group:admin@raito.io", Email: "administrators@raito.io", Members: []string{"user:dieter@raito.io", "serviceAccount:sa@raito.io"}})
					if err != nil {
						return err
					}

					return fn(ctx, &iam.GroupEntity{ExternalId: "group:engineers@raito.io", Email: "engineers@raito.io", Members: []string{"user:ruben@raito.io", "group:admin@raito.io", "user:dieter@raito.io"}})
				})
				adminRepoMock.EXPECT().GetUsers(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context, *iam.UserEntity) error) error {
					err := fn(ctx, &iam.UserEntity{ExternalId: "user:dieter@raito.io", Email: "dieter@raito.io", Name: "Dieter Wachters"})
					if err != nil {
						return err
					}

					return errors.New("boom")
				})
			}},
			args: args{
				ctx:       context.Background(),
				configMap: &config.ConfigMap{Parameters: map[string]string{common.GsuiteIdentityStoreSync: "true"}},
			},
			expected: expected{
				groups: []identity_store.Group{
					{
						ExternalId:  "group:engineers@raito.io",
						Name:        "engineers@raito.io",
						DisplayName: "engineers@raito.io",
					},
					{
						ExternalId:             "group:admin@raito.io",
						Name:                   "administrators@raito.io",
						DisplayName:            "administrators@raito.io",
						ParentGroupExternalIds: []string{"group:engineers@raito.io"},
					},
				},
				users: []identity_store.User{
					{
						ExternalId:       "user:dieter@raito.io",
						Email:            "dieter@raito.io",
						Name:             "Dieter Wachters",
						UserName:         "dieter@raito.io",
						GroupExternalIds: []string{"group:admin@raito.io", "group:engineers@raito.io"},
					},
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, adminRepoMock := createIdentityStoreSyncer(t, gcp.NewIdentityStoreMetadata())
			tt.fields.mockSetup(adminRepoMock)

			isHandlerMock := mocks.NewSimpleIdentityStoreIdentityHandler(t, 1)

			tt.wantErr(t, s.SyncIdentityStore(tt.args.ctx, isHandlerMock, tt.args.configMap))
			assert.ElementsMatch(t, tt.expected.groups, isHandlerMock.Groups)

			for i := range isHandlerMock.Users {
				sort.Slice(isHandlerMock.Users[i].GroupExternalIds, func(j, k int) bool {
					return isHandlerMock.Users[i].GroupExternalIds[j] < isHandlerMock.Users[i].GroupExternalIds[k]
				})
			}

			assert.ElementsMatch(t, tt.expected.users, isHandlerMock.Users)
		})
	}
}

func createIdentityStoreSyncer(t *testing.T, metadata *identity_store.MetaData) (*IdentityStoreSyncer, *MockAdminRepository) {
	t.Helper()

	adminRepoMock := NewMockAdminRepository(t)

	return NewIdentityStoreSyncer(adminRepoMock, metadata), adminRepoMock
}
