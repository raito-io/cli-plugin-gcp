package gcp

import (
	"context"

	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/gcp/common"
	"github.com/raito-io/cli-plugin-gcp/gcp/iam"

	is "github.com/raito-io/cli/base/identity_store"
)

type IdentityStoreSyncer struct {
	iamServiceProvider func(configMap *config.ConfigMap) iam.IAMService
}

func NewIdentityStoreSyncer() *IdentityStoreSyncer {
	return &IdentityStoreSyncer{iamServiceProvider: newIamServiceProvider}
}

func (s *IdentityStoreSyncer) GetIdentityStoreMetaData(_ context.Context, _ *config.ConfigMap) (*is.MetaData, error) {
	common.Logger.Debug("Returning meta data for GCP organization identity store")

	return &is.MetaData{
		Type:        "gcp",
		CanBeLinked: true,
		CanBeMaster: true,
	}, nil
}

func newIamServiceProvider(configMap *config.ConfigMap) iam.IAMService {
	return iam.NewIAMService(configMap)
}

func (s *IdentityStoreSyncer) WithIAMServiceProvider(provider func(configMap *config.ConfigMap) iam.IAMService) *IdentityStoreSyncer {
	s.iamServiceProvider = provider
	return s
}

func (s *IdentityStoreSyncer) SyncIdentityStore(ctx context.Context, identityHandler wrappers.IdentityStoreIdentityHandler, configMap *config.ConfigMap) error {
	// get groups and make a membership map key: ID of user/group, value array of Group IDs it is member of
	groupMembership := make(map[string]set.Set[string])

	groups, err := s.iamServiceProvider(configMap).GetGroups(ctx, configMap)

	if err != nil {
		return err
	}

	groupList := make([]*is.Group, 0)

	for _, g := range groups {
		// Make sure to always handle the members for all the found groups.
		for _, m := range g.Members {
			if _, f := groupMembership[m]; !f {
				groupMembership[m] = set.NewSet[string](g.ExternalId)
			} else {
				groupMembership[m].Add(g.ExternalId)
			}
		}

		groupList = append(groupList, &is.Group{ExternalId: g.ExternalId, Name: g.Email, DisplayName: g.Email})
	}

	for i, g := range groupList {
		if _, f := groupMembership[g.ExternalId]; f {
			groupList[i].ParentGroupExternalIds = groupMembership[g.ExternalId].Slice()
		}
	}

	for _, g := range groupList {
		err = identityHandler.AddGroups(g)

		if err != nil {
			return err
		}
	}

	// get users
	users, err := s.iamServiceProvider(configMap).GetUsers(ctx, configMap)

	if err != nil {
		return err
	}

	for _, u := range users {
		if _, f := groupMembership[u.ExternalId]; f {
			err2 := identityHandler.AddUsers(&is.User{ExternalId: u.ExternalId, UserName: u.Email, Email: u.Email, Name: u.Name, GroupExternalIds: groupMembership[u.ExternalId].Slice()})

			if err2 != nil {
				return err2
			}
		} else {
			err2 := identityHandler.AddUsers(&is.User{ExternalId: u.ExternalId, UserName: u.Email, Email: u.Email, Name: u.Name})

			if err2 != nil {
				return err2
			}
		}
	}

	// get serviceAccounts
	serviceAcounts, err := s.iamServiceProvider(configMap).GetServiceAccounts(ctx, configMap)

	if err != nil {
		return err
	}

	for _, u := range serviceAcounts {
		err2 := identityHandler.AddUsers(&is.User{ExternalId: u.ExternalId, UserName: u.Email, Email: u.Email, Name: u.Name})

		if err2 != nil {
			return err2
		}
	}

	return nil
}
