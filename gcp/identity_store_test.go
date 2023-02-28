package gcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/raito-io/cli-plugin-gcp/gcp/iam"
	is "github.com/raito-io/cli/base/identity_store"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIdentityStoreSyncer_SyncIdentityStore(t *testing.T) {

	identityHandlerMock := mocks.NewSimpleIdentityStoreIdentityHandler(t, 1)

	iamServiceMock := iam.NewMockIAMService(t)
	iamServiceMock.EXPECT().GetGroups(mock.Anything, mock.Anything).Return([]iam.GroupEntity{
		{
			ExternalId: "g1",
			Email:      "g1@example.com",
			Members:    []string{"user1", "g2"},
		},
		{
			ExternalId: "g2",
			Email:      "g2@example.com",
			Members:    []string{"user2"},
		},
	}, nil).Once()
	iamServiceMock.EXPECT().GetUsers(mock.Anything, mock.Anything).Return([]iam.UserEntity{
		{
			ExternalId: "user1",
			Name:       "user1",
			Email:      "user1@example.com",
		},
		{
			ExternalId: "user2",
			Name:       "user2",
			Email:      "user2@example.com",
		},
	}, nil).Once()
	iamServiceMock.EXPECT().GetServiceAccounts(mock.Anything, mock.Anything).Return([]iam.UserEntity{
		{
			ExternalId: "sa1",
			Name:       "sa1",
			Email:      "sa1@example.com",
		},
	}, nil).Once()

	// Given
	configMap := &config.ConfigMap{
		Parameters: map[string]string{},
	}

	syncer := IdentityStoreSyncer{
		iamServiceProvider: func(config *config.ConfigMap) iam.IAMService {
			return iamServiceMock
		},
	}

	// When
	err := syncer.SyncIdentityStore(context.Background(), identityHandlerMock, configMap)

	// Then
	assert.NoError(t, err)

	identityHandlerMock.AssertNumberOfCalls(t, "AddUsers", 3)
	identityHandlerMock.AssertNumberOfCalls(t, "AddGroups", 2)

	identityHandlerMock.AssertCalled(t, "AddUsers", &is.User{ExternalId: "user1", UserName: "user1@example.com", Email: "user1@example.com", Name: "user1", GroupExternalIds: []string{"g1"}})
	identityHandlerMock.AssertCalled(t, "AddUsers", &is.User{ExternalId: "user2", UserName: "user2@example.com", Email: "user2@example.com", Name: "user2", GroupExternalIds: []string{"g2"}})

	identityHandlerMock.AssertCalled(t, "AddGroups", &is.Group{ExternalId: "g1", Name: "g1@example.com", DisplayName: "g1@example.com"})
	identityHandlerMock.AssertCalled(t, "AddGroups", &is.Group{ExternalId: "g2", Name: "g2@example.com", DisplayName: "g2@example.com", ParentGroupExternalIds: []string{"g1"}})
}

func TestIdentityStoreSyncer_SAError(t *testing.T) {

	identityHandlerMock := mocks.NewSimpleIdentityStoreIdentityHandler(t, 1)

	iamServiceMock := iam.NewMockIAMService(t)
	iamServiceMock.EXPECT().GetGroups(mock.Anything, mock.Anything).Return([]iam.GroupEntity{
		{
			ExternalId: "g1",
			Email:      "g1@example.com",
			Members:    []string{"user1", "g2"},
		},
		{
			ExternalId: "g2",
			Email:      "g2@example.com",
			Members:    []string{"user2"},
		},
	}, nil).Once()
	iamServiceMock.EXPECT().GetUsers(mock.Anything, mock.Anything).Return([]iam.UserEntity{
		{
			ExternalId: "user1",
			Name:       "user1",
			Email:      "user1@example.com",
		},
		{
			ExternalId: "user2",
			Name:       "user2",
			Email:      "user2@example.com",
		},
	}, nil).Once()
	iamServiceMock.EXPECT().GetServiceAccounts(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error!")).Once()

	// Given
	configMap := &config.ConfigMap{
		Parameters: map[string]string{},
	}

	syncer := IdentityStoreSyncer{
		iamServiceProvider: func(config *config.ConfigMap) iam.IAMService {
			return iamServiceMock
		},
	}

	// When
	err := syncer.SyncIdentityStore(context.Background(), identityHandlerMock, configMap)

	// Then
	assert.Error(t, err)
}

func TestIdentityStoreSyncer_UserError(t *testing.T) {

	identityHandlerMock := mocks.NewSimpleIdentityStoreIdentityHandler(t, 1)

	iamServiceMock := iam.NewMockIAMService(t)
	iamServiceMock.EXPECT().GetGroups(mock.Anything, mock.Anything).Return([]iam.GroupEntity{
		{
			ExternalId: "g1",
			Email:      "g1@example.com",
			Members:    []string{"user1", "g2"},
		},
		{
			ExternalId: "g2",
			Email:      "g2@example.com",
			Members:    []string{"user2"},
		},
	}, nil).Once()
	iamServiceMock.EXPECT().GetUsers(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error!")).Once()

	// Given
	configMap := &config.ConfigMap{
		Parameters: map[string]string{},
	}

	syncer := IdentityStoreSyncer{
		iamServiceProvider: func(config *config.ConfigMap) iam.IAMService {
			return iamServiceMock
		},
	}

	// When
	err := syncer.SyncIdentityStore(context.Background(), identityHandlerMock, configMap)

	// Then
	assert.Error(t, err)
}

func TestIdentityStoreSyncer_GroupError(t *testing.T) {

	identityHandlerMock := mocks.NewSimpleIdentityStoreIdentityHandler(t, 1)

	iamServiceMock := iam.NewMockIAMService(t)
	iamServiceMock.EXPECT().GetGroups(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error!")).Once()

	// Given
	configMap := &config.ConfigMap{
		Parameters: map[string]string{},
	}

	syncer := IdentityStoreSyncer{
		iamServiceProvider: func(config *config.ConfigMap) iam.IAMService {
			return iamServiceMock
		},
	}

	// When
	err := syncer.SyncIdentityStore(context.Background(), identityHandlerMock, configMap)

	// Then
	assert.Error(t, err)
}
