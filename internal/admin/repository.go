package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/util/config"
	gcpadmin "google.golang.org/api/admin/directory/v1"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

const maxPageItems = 100

type AdminRepository struct {
	client *gcpadmin.Service

	customerId string
}

func NewAdminRepository(client *gcpadmin.Service, configMap *config.ConfigMap) *AdminRepository {
	return &AdminRepository{
		client: client,

		customerId: configMap.GetString(common.GsuiteCustomerId),
	}
}

func (r *AdminRepository) GetUsers(ctx context.Context, fn func(ctx context.Context, entity *types.UserEntity) error) error {
	nextPageToken := ""

	for {
		usersCall := r.client.Users.List().Customer(r.customerId).MaxResults(maxPageItems)

		if nextPageToken != "" {
			usersCall.PageToken(nextPageToken)
		}

		users, err := usersCall.Do()
		if err != nil {
			return fmt.Errorf("listing users: %s", err.Error())
		}

		for _, u := range users.Users {
			if u.Suspended {
				continue
			}

			err = fn(ctx, &types.UserEntity{ExternalId: fmt.Sprintf("user:%s", u.PrimaryEmail), Name: u.Name.FullName, Email: u.PrimaryEmail})
			if err != nil {
				return err
			}
		}

		if users.NextPageToken != "" {
			nextPageToken = users.NextPageToken
		} else {
			break
		}
	}

	return nil
}

func (r *AdminRepository) GetGroups(ctx context.Context, fn func(ctx context.Context, entity *types.GroupEntity) error) error {
	nextPageToken := ""

	for {
		groupsCall := r.client.Groups.List().Customer(r.customerId).MaxResults(maxPageItems)

		if nextPageToken != "" {
			groupsCall.PageToken(nextPageToken)
		}

		groups, err := groupsCall.Do()
		if err != nil {
			return fmt.Errorf("listing groups: %s", err.Error())
		}

		for _, g := range groups.Groups {
			groupMembers, err2 := r.groupMembers(g.Id)
			if err2 != nil {
				return fmt.Errorf("group members of group %q: %w", g.Id, err2)
			}

			err2 = fn(ctx, &types.GroupEntity{ExternalId: fmt.Sprintf("group:%s", g.Email), Email: g.Email, Members: groupMembers})
			if err2 != nil {
				return err2
			}
		}

		if groups.NextPageToken != "" {
			nextPageToken = groups.NextPageToken
		} else {
			break
		}
	}

	return nil
}

func (r *AdminRepository) groupMembers(groupId string) ([]string, error) {
	nextPageToken := ""

	var memberIds []string

	for {
		membersCall := r.client.Members.List(groupId).MaxResults(maxPageItems)

		if nextPageToken != "" {
			membersCall.PageToken(nextPageToken)
		}

		members, err := membersCall.Do()
		if err != nil {
			return nil, fmt.Errorf("fetching members for group %s: %s", groupId, err.Error())
		}

		for _, m := range members.Members {
			if strings.EqualFold(m.Type, "user") {
				memberIds = append(memberIds, fmt.Sprintf("user:%s", m.Email))
			} else if strings.EqualFold(m.Type, "group") {
				memberIds = append(memberIds, fmt.Sprintf("group:%s", m.Email))
			} else {
				common.Logger.Warn(fmt.Sprintf("Found unknown member type %s for group %s (id: %s; email: %s)", m.Type, groupId, m.Id, m.Email))
			}
		}

		if members.NextPageToken != "" {
			nextPageToken = members.NextPageToken
		} else {
			break
		}
	}

	return memberIds, nil
}
