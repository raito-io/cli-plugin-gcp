package syncer

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/org"

	is "github.com/raito-io/cli/base/identity_store"
)

//go:generate go run github.com/vektra/mockery/v2 --name=AdminRepository --with-expecter --inpackage
type AdminRepository interface {
	GetUsers(ctx context.Context, fn func(ctx context.Context, entity *iam.UserEntity) error) error
	GetGroups(ctx context.Context, fn func(ctx context.Context, entity *iam.GroupEntity) error) error
}

//go:generate go run github.com/vektra/mockery/v2 --name=DataObjectRepository --with-expecter --inpackage
type DataObjectRepository interface {
	Bindings(ctx context.Context, fn func(ctx context.Context, dataObject *org.GcpOrgEntity, bindings []iam.IamBinding) error) error
}

type IdentityStoreSyncer struct {
	adminRepository AdminRepository
	dataObjectRepo  DataObjectRepository
	metadata        *is.MetaData
}

func NewIdentityStoreSyncer(adminRepo AdminRepository, dataObjectRepo DataObjectRepository, isMetadata *is.MetaData) *IdentityStoreSyncer {
	return &IdentityStoreSyncer{
		adminRepository: adminRepo,
		dataObjectRepo:  dataObjectRepo,
		metadata:        isMetadata,
	}
}

func (s *IdentityStoreSyncer) GetIdentityStoreMetaData(_ context.Context, _ *config.ConfigMap) (*is.MetaData, error) {
	common.Logger.Debug("Returning meta data for GCP organization identity store")

	return s.metadata, nil
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
	err := s.dataObjectRepo.Bindings(ctx, func(ctx context.Context, dataObject *org.GcpOrgEntity, bindings []iam.IamBinding) error {
		for _, binding := range bindings {
			memberParts := strings.SplitN(binding.Member, ":", 2)
			email := memberParts[1]
			memberPrefix := memberParts[0]

			switch memberPrefix {
			case "user", "serviceAccount":
				if userIds.Contains(binding.Member) {
					continue
				}

				userIds.Add(binding.Member)

				common.Logger.Debug(fmt.Sprintf("Found new user %q in bindings", binding.Member))

				user := is.User{
					ExternalId: binding.Member,
					Name:       email,
					UserName:   email,
					Email:      email,
				}

				err := identityHandler.AddUsers(&user)
				if err != nil {
					common.Logger.Error(fmt.Sprintf("Failed to add user %q: %s", binding.Member, err.Error()))
				}
			case "group":
				if _, found := groups[binding.Member]; found {
					continue
				}

				group := is.Group{
					ExternalId:  binding.Member,
					Name:        email,
					DisplayName: email,
				}

				groups[binding.Member] = &group

				common.Logger.Debug(fmt.Sprintf("Found new group %q in bindings", binding.Member))

				err := identityHandler.AddGroups(&group)
				if err != nil {
					common.Logger.Error(fmt.Sprintf("Failed to add group %q: %s", binding.Member, err.Error()))
				}
			case "special_group":
				common.Logger.Info(fmt.Sprintf("Ignore special group %q", binding.Member))
			default:
				common.Logger.Warn(fmt.Sprintf("Ignore unknown member type: %s", binding.Member))
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("load users and groups from binding: %w", err)
	}

	return nil
}

func (s *IdentityStoreSyncer) syncGcpUsers(ctx context.Context, identityHandler wrappers.IdentityStoreIdentityHandler, groupMembership map[string]set.Set[string]) (set.Set[string], error) {
	userIds := set.NewSet[string]()

	err := s.adminRepository.GetUsers(ctx, func(ctx context.Context, entity *iam.UserEntity) error {
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

		err := identityHandler.AddUsers(&user)
		if err != nil {
			return fmt.Errorf("add user to handler: %w", err)
		}

		return nil
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
	err := s.adminRepository.GetGroups(ctx, func(ctx context.Context, entity *iam.GroupEntity) error {
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
			return nil, nil, fmt.Errorf("add group to handler: %w", err)
		}
	}

	return groupMembership, groups, nil
}
