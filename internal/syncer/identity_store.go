package syncer

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/smithy-go/ptr"
	is "github.com/raito-io/cli/base/identity_store"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

//go:generate go run github.com/vektra/mockery/v2 --name=AdminRepository --with-expecter --inpackage
type AdminRepository interface {
	GetUsers(ctx context.Context, fn func(ctx context.Context, entity *iam.UserEntity) error) error
	GetGroups(ctx context.Context, fn func(ctx context.Context, entity *iam.GroupEntity) error) error
}

type IdentityStoreSyncer struct {
	adminRepository AdminRepository
	metadata        *is.MetaData
}

func NewIdentityStoreSyncer(adminRepo AdminRepository, isMetadata *is.MetaData) *IdentityStoreSyncer {
	return &IdentityStoreSyncer{
		adminRepository: adminRepo,
		metadata:        isMetadata,
	}
}

func (s *IdentityStoreSyncer) GetIdentityStoreMetaData(_ context.Context, _ *config.ConfigMap) (*is.MetaData, error) {
	common.Logger.Debug("Returning meta data for GCP organization identity store")

	return s.metadata, nil
}

func (s *IdentityStoreSyncer) SyncIdentityStore(ctx context.Context, identityHandler wrappers.IdentityStoreIdentityHandler, configMap *config.ConfigMap) error {
	if configMap.GetBoolWithDefault(common.GsuiteIdentityStoreSync, false) {
		// get groups and make a membership map key: ID of user/group, value array of Group IDs it is member of
		common.Logger.Info("Syncing GCP groups")

		var groupMembership map[string]set.Set[string]
		var err error

		groupMembership, _, err = s.syncGcpGroups(ctx, identityHandler)
		if err != nil {
			return err
		}

		// get GCP users
		common.Logger.Info("Syncing GCP users")

		_, err = s.syncGcpUsers(ctx, identityHandler, groupMembership)
		if err != nil {
			return err
		}
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

		if s.userIsServiceAccount(entity) {
			user.IsMachine = ptr.Bool(true)
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

func (s *IdentityStoreSyncer) userIsServiceAccount(userEntity *iam.UserEntity) bool {
	serviceAccountEmailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.gserviceaccount\.com$`)

	return serviceAccountEmailRegex.MatchString(userEntity.Email)
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
