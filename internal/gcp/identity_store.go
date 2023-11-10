package gcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"

	is "github.com/raito-io/cli/base/identity_store"
)

//go:generate go run github.com/vektra/mockery/v2 --name=AdminRepository --with-expecter --inpackage
type AdminRepository interface {
	GetUsers(ctx context.Context, fn func(ctx context.Context, entity *types.UserEntity) error) error
	GetGroups(ctx context.Context, fn func(ctx context.Context, entity *types.GroupEntity) error) error
}

//go:generate go run github.com/vektra/mockery/v2 --name=DataObjectRepository --with-expecter --inpackage
type DataObjectRepository interface {
	UserAndGroups(ctx context.Context, userFn func(ctx context.Context, userId string) error, groupFn func(ctx context.Context, groupId string) error) error
}

type IdentityStoreSyncer struct {
	adminRepository AdminRepository
	dataObjectRepo  DataObjectRepository
}

func NewIdentityStoreSyncer(adminRepo AdminRepository, dataObjectRepo DataObjectRepository) *IdentityStoreSyncer {
	return &IdentityStoreSyncer{
		adminRepository: adminRepo,
		dataObjectRepo:  dataObjectRepo,
	}
}

func (s *IdentityStoreSyncer) GetIdentityStoreMetaData(_ context.Context, _ *config.ConfigMap) (*is.MetaData, error) {
	common.Logger.Debug("Returning meta data for GCP organization identity store")

	return &is.MetaData{
		Type:        "gcp",
		CanBeLinked: true,
		CanBeMaster: true,
	}, nil
}

func (s *IdentityStoreSyncer) SyncIdentityStore(ctx context.Context, identityHandler wrappers.IdentityStoreIdentityHandler, configMap *config.ConfigMap) error {
	groups := map[string]*is.Group{}
	userIds := set.NewSet[string]()

	if configMap.GetBoolWithDefault(common.GsuiteIdentityStoreSync, false) {
		// get groups and make a membership map key: ID of user/group, value array of Group IDs it is member of
		common.Logger.Info("Syncing GCP groups")

		var groupMembership map[string]set.Set[string]
		var err error

		groupMembership, groups, err = s.syncGcpGroups(ctx, identityHandler)
		if err != nil {
			return err
		}

		// get GCP users
		common.Logger.Info("Syncing GCP users")

		userIds, err = s.syncGcpUsers(ctx, identityHandler, groupMembership)
		if err != nil {
			return err
		}
	}
	// Load users and groups from binding
	common.Logger.Info("Syncing groups and users from bindings in gcp")

	err := s.syncBindingUsersAndGroups(ctx, identityHandler, userIds, groups)
	if err != nil {
		return err
	}

	return nil
}

func (s *IdentityStoreSyncer) syncBindingUsersAndGroups(ctx context.Context, identityHandler wrappers.IdentityStoreIdentityHandler, userIds set.Set[string], groups map[string]*is.Group) error {
	err := s.dataObjectRepo.UserAndGroups(ctx, func(ctx context.Context, userId string) error {
		if userIds.Contains(userId) {
			return nil
		}

		userIds.Add(userId)

		common.Logger.Debug(fmt.Sprintf("Found new user %q in bindings", userId))

		email := strings.SplitN(userId, ":", 2)[1]
		user := is.User{
			ExternalId: userId,
			Name:       email,
			UserName:   email,
			Email:      email,
		}

		return identityHandler.AddUsers(&user)
	}, func(ctx context.Context, groupId string) error {
		if _, found := groups[groupId]; found {
			return nil
		}

		groupName := strings.SplitN(groupId, ":", 2)[1]
		group := is.Group{
			ExternalId:  groupId,
			Name:        groupName,
			DisplayName: groupName,
		}

		groups[groupId] = &group

		common.Logger.Debug(fmt.Sprintf("Found new group %q in bindings", groupId))

		return identityHandler.AddGroups(&group)
	})
	if err != nil {
		return fmt.Errorf("load users and groups from binding: %w", err)
	}

	return nil
}

func (s *IdentityStoreSyncer) syncGcpUsers(ctx context.Context, identityHandler wrappers.IdentityStoreIdentityHandler, groupMembership map[string]set.Set[string]) (set.Set[string], error) {
	userIds := set.NewSet[string]()

	err := s.adminRepository.GetUsers(ctx, func(ctx context.Context, entity *types.UserEntity) error {
		common.Logger.Debug(fmt.Sprintf("Found GCP user: %s", entity.ExternalId))

		userIds.Add(entity.ExternalId)

		user := is.User{
			ExternalId: entity.ExternalId,
			Name:       entity.Name,
			UserName:   entity.Email,
			Email:      entity.Email,
		}

		if _, f := groupMembership[entity.ExternalId]; f {
			user.GroupExternalIds = groupMembership[entity.ExternalId].Slice()
		}

		return identityHandler.AddUsers(&user)
	})

	if err != nil {
		return nil, fmt.Errorf("get gcp users: %w", err)
	}

	return userIds, nil
}

func (s *IdentityStoreSyncer) syncGcpGroups(ctx context.Context, identityHandler wrappers.IdentityStoreIdentityHandler) (map[string]set.Set[string], map[string]*is.Group, error) {
	groupMembership := make(map[string]set.Set[string])
	groups := map[string]*is.Group{}

	// Get GCP groups
	err := s.adminRepository.GetGroups(ctx, func(ctx context.Context, entity *types.GroupEntity) error {
		common.Logger.Debug(fmt.Sprintf("Found GCP group: %s", entity.ExternalId))

		groups[entity.ExternalId] = &is.Group{
			ExternalId:  entity.ExternalId,
			Name:        entity.Email,
			DisplayName: entity.Email,
		}

		for _, m := range entity.Members {
			if _, f := groupMembership[m]; !f {
				groupMembership[m] = set.NewSet[string](entity.ExternalId)
			} else {
				groupMembership[m].Add(entity.ExternalId)
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("get gcp groups: %w", err)
	}

	for _, g := range groups {
		if _, f := groupMembership[g.ExternalId]; f {
			g.ParentGroupExternalIds = groupMembership[g.ExternalId].Slice()
		}
	}

	for _, g := range groups {
		err = identityHandler.AddGroups(g)

		if err != nil {
			return nil, nil, err
		}
	}

	return groupMembership, groups, nil
}
