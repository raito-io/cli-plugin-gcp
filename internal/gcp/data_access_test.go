package gcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/raito-io/cli-plugin-gcp/gcp/common"
	"github.com/raito-io/cli-plugin-gcp/gcp/iam"
)

func TestAccessSyncer_SyncAccessProvidersFromTarget(t *testing.T) {
	//Given
	configParams := config.ConfigMap{
		Parameters: map[string]string{"key": "value"},
	}

	fileCreator := mocks.NewSimpleAccessProviderHandler(t, 1)
	iamServiceMock := iam.NewMockIAMService(t)

	iamServiceMock.EXPECT().GetIAMPolicyBindings(mock.Anything, mock.Anything).Return([]iam.IamBinding{
		{
			Member:       "user:u1@example.com",
			Role:         "roles/owner",
			Resource:     "project1",
			ResourceType: "project",
		},
		{
			Member:       "serviceAccount:s1@example.com",
			Role:         "roles/owner",
			Resource:     "project1",
			ResourceType: "project",
		},
		{
			Member:       "group:g1@example.com",
			Role:         "roles/viewer",
			Resource:     "project1",
			ResourceType: "project",
		},
		{
			Member:       "group:g1@example.com",
			Role:         "roles/viewer",
			Resource:     "project2",
			ResourceType: "project",
		},
	}, nil).Once()

	syncer := AccessSyncer{
		iamServiceProvider: func(config *config.ConfigMap) iam.IAMService {
			return iamServiceMock
		},
		getDSMetadata: GetDataSourceMetaData,
	}

	// When
	err := syncer.SyncAccessProvidersFromTarget(context.Background(), fileCreator, &configParams)

	// Then
	assert.NoError(t, err)
	fileCreator.AssertNumberOfCalls(t, "AddAccessProviders", 3)

	n := len(fileCreator.AccessProviders)
	for _, ap := range fileCreator.AccessProviders {
		switch ap.ExternalId {
		case "project_project1_roles_owner":
			n -= 1
			assert.Equal(t, 2, len(ap.Who.Users))
			assert.Equal(t, 0, len(ap.Who.Groups))
			assert.Equal(t, 0, len(ap.Who.AccessProviders))
		case "project_project1_roles_viewer":
			n -= 1
			assert.Equal(t, 0, len(ap.Who.Users))
			assert.Equal(t, 1, len(ap.Who.Groups))
			assert.Equal(t, 0, len(ap.Who.AccessProviders))
		case "project_project2_roles_viewer":
			n -= 1
			assert.Equal(t, 0, len(ap.Who.Users))
			assert.Equal(t, 1, len(ap.Who.Groups))
			assert.Equal(t, 0, len(ap.Who.AccessProviders))
		}
	}

	assert.Equal(t, 0, n)
}

func TestAccessSyncer_SyncAccessProvidersFromTargetError(t *testing.T) {
	//Given
	configParams := config.ConfigMap{
		Parameters: map[string]string{"key": "value"},
	}

	fileCreator := mocks.NewSimpleAccessProviderHandler(t, 1)
	iamServiceMock := iam.NewMockIAMService(t)

	iamServiceMock.EXPECT().GetIAMPolicyBindings(mock.Anything, mock.Anything).Return([]iam.IamBinding{}, fmt.Errorf("error!")).Once()

	syncer := AccessSyncer{
		iamServiceProvider: func(config *config.ConfigMap) iam.IAMService {
			return iamServiceMock
		},
	}

	// When
	err := syncer.SyncAccessProvidersFromTarget(context.Background(), fileCreator, &configParams)

	// Then
	assert.Error(t, err)
	fileCreator.AssertNumberOfCalls(t, "AddAccessProviders", 0)
}

func TestAccessSyncer_SyncAccessProvidersFromTargetIgnoreRaitoBindings(t *testing.T) {
	//Given
	configParams := config.ConfigMap{
		Parameters: map[string]string{"key": "value"},
	}

	fileCreator := mocks.NewSimpleAccessProviderHandler(t, 1)
	iamServiceMock := iam.NewMockIAMService(t)

	iamServiceMock.EXPECT().GetIAMPolicyBindings(mock.Anything, mock.Anything).Return([]iam.IamBinding{
		{
			Member:       "user:u1@example.com",
			Role:         "roles/owner",
			Resource:     "project1",
			ResourceType: "project",
		},
		{
			Member:       "serviceAccount:s1@example.com",
			Role:         "roles/owner",
			Resource:     "project1",
			ResourceType: "project",
		},
		{
			Member:       "group:g1@example.com",
			Role:         "roles/viewer",
			Resource:     "project1",
			ResourceType: "project",
		},
		{
			Member:       "group:g1@example.com",
			Role:         "roles/viewer",
			Resource:     "project2",
			ResourceType: "project",
		},
	}, nil).Once()

	syncer := AccessSyncer{
		iamServiceProvider: func(config *config.ConfigMap) iam.IAMService {
			return iamServiceMock
		},
		getDSMetadata: GetDataSourceMetaData,
		raitoManagedBindings: []iam.IamBinding{
			{
				Member:       "group:g1@example.com",
				Role:         "roles/viewer",
				Resource:     "project1",
				ResourceType: "project",
			},
			{
				Member:       "group:g1@example.com",
				Role:         "roles/viewer",
				Resource:     "project2",
				ResourceType: "project",
			},
		},
	}

	// When
	err := syncer.SyncAccessProvidersFromTarget(context.Background(), fileCreator, &configParams)

	// Then
	assert.NoError(t, err)
	fileCreator.AssertNumberOfCalls(t, "AddAccessProviders", 1)

	ap := fileCreator.AccessProviders[0]
	assert.Equal(t, 2, len(ap.Who.Users))
	assert.Equal(t, 0, len(ap.Who.Groups))
	assert.Equal(t, 0, len(ap.Who.AccessProviders))

}

func TestAccessSyncer_SyncAccessProvidersFromTargetIgnoreNonApplicablePermissions(t *testing.T) {
	//Given
	configParams := config.ConfigMap{
		Parameters: map[string]string{"key": "value"},
	}

	fileCreator := mocks.NewSimpleAccessProviderHandler(t, 1)
	iamServiceMock := iam.NewMockIAMService(t)

	iamServiceMock.EXPECT().GetIAMPolicyBindings(mock.Anything, mock.Anything).Return([]iam.IamBinding{
		{
			Member:       "user:u1@example.com",
			Role:         "roles/random",
			Resource:     "project1",
			ResourceType: "project",
		},
	}, nil).Times(2)

	syncer := AccessSyncer{
		iamServiceProvider: func(config *config.ConfigMap) iam.IAMService {
			return iamServiceMock
		},
		getDSMetadata: GetDataSourceMetaData,
	}

	// When
	err := syncer.SyncAccessProvidersFromTarget(context.Background(), fileCreator, &configParams)

	// Then: by default non applicable permissions are skipped
	assert.NoError(t, err)
	fileCreator.AssertNumberOfCalls(t, "AddAccessProviders", 0)

	// allow non-applicable permissions to be synced
	configParams.Parameters[common.ExcludeNonAplicablePermissions] = "false"

	err = syncer.SyncAccessProvidersFromTarget(context.Background(), fileCreator, &configParams)

	assert.NoError(t, err)
	fileCreator.AssertNumberOfCalls(t, "AddAccessProviders", 1)

	// but it should not be internalizable
	ap := fileCreator.AccessProviders[0]
	assert.Equal(t, true, ap.NotInternalizable)
}

func TestAccessSyncer_SyncAccessProvidersToTarget(t *testing.T) {
	//Given
	configParams := config.ConfigMap{
		Parameters: map[string]string{"key": "value"},
	}

	fileCreator := mocks.NewSimpleAccessProviderFeedbackHandler(t)
	iamServiceMock := iam.NewMockIAMService(t)

	iamServiceMock.EXPECT().AddIamBinding(mock.Anything, mock.Anything, iam.IamBinding{
		Member:       "user:user1@example.com",
		Role:         "role/owner",
		Resource:     "project1",
		ResourceType: "project",
	}).Return(nil).Once()

	iamServiceMock.EXPECT().AccessProviderBindingHooks().Return(nil)

	iamServiceMock.EXPECT().AddIamBinding(mock.Anything, mock.Anything, iam.IamBinding{
		Member:       "group:group1@example.com",
		Role:         "role/owner",
		Resource:     "project1",
		ResourceType: "project",
	}).Return(nil).Once()

	iamServiceMock.EXPECT().AddIamBinding(mock.Anything, mock.Anything, iam.IamBinding{
		Member:       "serviceAccount:sa@gserviceaccount.com",
		Role:         "role/owner",
		Resource:     "project1",
		ResourceType: "project",
	}).Return(nil).Once()

	iamServiceMock.EXPECT().RemoveIamBinding(mock.Anything, mock.Anything, iam.IamBinding{
		Member:       "user:user1@example.com",
		Role:         "role/owner",
		Resource:     "project1",
		ResourceType: "project",
	}).Return(nil).Once()

	iamServiceMock.EXPECT().RemoveIamBinding(mock.Anything, mock.Anything, iam.IamBinding{
		Member:       "group:group1@example.com",
		Role:         "role/owner",
		Resource:     "project1",
		ResourceType: "project",
	}).Return(nil).Once()

	iamServiceMock.EXPECT().RemoveIamBinding(mock.Anything, mock.Anything, iam.IamBinding{
		Member:       "serviceAccount:sa@gserviceaccount.com",
		Role:         "role/owner",
		Resource:     "project1",
		ResourceType: "project",
	}).Return(nil).Once()

	syncer := AccessSyncer{
		iamServiceProvider: func(config *config.ConfigMap) iam.IAMService {
			return iamServiceMock
		},
	}

	// When
	err := syncer.SyncAccessProviderToTarget(context.Background(), &sync_to_target.AccessProviderImport{
		AccessProviders: []*sync_to_target.AccessProvider{
			{
				Id:          "ap1",
				Name:        "ap1",
				NamingHint:  "ap1",
				Description: "ap1",
				Action:      sync_to_target.Grant,
				Delete:      false,
				Who: sync_to_target.WhoItem{
					Users:  []string{"user1@example.com", "sa@gserviceaccount.com"},
					Groups: []string{"group1@example.com"},
				},
				ActualName: ptr.String("a1"),
				What: []sync_to_target.WhatItem{
					{
						DataObject: &data_source.DataObjectReference{
							FullName: "project1",
							Type:     "project",
						},
						Permissions: []string{"role/owner"},
					},
				},
			},
		},
	}, fileCreator, &configParams)

	// Then
	assert.NoError(t, err)

	// When (delete)
	err = syncer.SyncAccessProviderToTarget(context.Background(), &sync_to_target.AccessProviderImport{
		AccessProviders: []*sync_to_target.AccessProvider{
			{
				Id:          "ap1",
				Name:        "ap1",
				NamingHint:  "ap1",
				Description: "ap1",
				Action:      sync_to_target.Grant,
				Delete:      true,
				Who: sync_to_target.WhoItem{
					Users:  []string{"user1@example.com", "sa@gserviceaccount.com"},
					Groups: []string{"group1@example.com"},
				},
				ActualName: ptr.String("a1"),
				What: []sync_to_target.WhatItem{
					{
						DataObject: &data_source.DataObjectReference{
							FullName: "project1",
							Type:     "project",
						},
						Permissions: []string{"role/owner"},
					},
				},
			},
		},
	}, fileCreator, &configParams)

	// Then
	assert.NoError(t, err)
}
