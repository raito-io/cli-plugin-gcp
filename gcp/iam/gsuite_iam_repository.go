package iam

import (
	"context"
	"fmt"
	"os"

	"github.com/raito-io/cli-plugin-gcp/gcp/common"
	"github.com/raito-io/cli/base/util/config"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

const MAX_PAGE_ITEMS = 100

type gsuiteIamRepository struct {
	customerId string
}

func (r *gsuiteIamRepository) client(ctx context.Context, configMap *config.ConfigMap, scopes ...string) (*admin.Service, error) {
	key := configMap.GetString(common.GcpSAFileLocation)

	if key == "" {
		key = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	serviceAccountJSON, err := os.ReadFile(key)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(serviceAccountJSON, scopes...)

	config.Subject = configMap.GetString(common.GsuiteImpersonateSubject)

	if err != nil {
		return nil, err
	}

	r.customerId = configMap.GetString(common.GsuiteCustomerId)

	if r.customerId == "" || config.Subject == "" {
		return nil, fmt.Errorf("for GSuite identity store sync please configure %s and %s", common.GsuiteCustomerId, common.GsuiteImpersonateSubject)
	}

	return admin.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
}

func (r *gsuiteIamRepository) GetUsers(ctx context.Context, configMap *config.ConfigMap, id string) ([]UserEntity, error) {
	res := make([]UserEntity, 0)

	client, err := r.client(ctx, configMap, admin.AdminDirectoryUserReadonlyScope)

	if err != nil {
		return nil, err
	}

	nextPageToken := ""

	for {
		usersCall := client.Users.List().Customer(r.customerId).MaxResults(MAX_PAGE_ITEMS)

		if nextPageToken != "" {
			usersCall = usersCall.PageToken(nextPageToken)
		}

		users, err2 := usersCall.Do()

		if err2 != nil {
			return nil, err2
		}

		for _, u := range users.Users {
			if u.Suspended {
				continue
			}

			res = append(res, UserEntity{ExternalId: fmt.Sprintf("user:%s", u.PrimaryEmail), Name: u.Name.FullName, Email: u.PrimaryEmail})
		}

		if users.NextPageToken != "" {
			nextPageToken = users.NextPageToken
		} else {
			break
		}
	}

	return res, nil
}

func (r *gsuiteIamRepository) GetGroups(ctx context.Context, configMap *config.ConfigMap, id string) ([]GroupEntity, error) {
	res := make([]GroupEntity, 0)

	client, err := r.client(ctx, configMap, admin.AdminDirectoryGroupReadonlyScope)

	if err != nil {
		return nil, err
	}

	nextPageToken := ""

	for {
		groupsCall := client.Groups.List().Customer(r.customerId).MaxResults(MAX_PAGE_ITEMS)

		if nextPageToken != "" {
			groupsCall = groupsCall.PageToken(nextPageToken)
		}

		groups, err2 := groupsCall.Do()

		if err2 != nil {
			return nil, err2
		}

		for _, g := range groups.Groups {
			members, err3 := r.groupMembers(ctx, configMap, g.Id)

			if err3 != nil {
				return nil, err3
			}

			res = append(res, GroupEntity{ExternalId: fmt.Sprintf("group:%s", g.Email), Email: g.Email, Members: members})
		}

		if groups.NextPageToken != "" {
			nextPageToken = groups.NextPageToken
		} else {
			break
		}
	}

	return res, nil
}

func (r *gsuiteIamRepository) groupMembers(ctx context.Context, configMap *config.ConfigMap, groupId string) ([]string, error) {
	res := make([]string, 0)

	client, err := r.client(ctx, configMap, admin.AdminDirectoryGroupReadonlyScope)

	if err != nil {
		return nil, err
	}

	nextPageToken := ""

	for {
		membersCall := client.Members.List(groupId).MaxResults(MAX_PAGE_ITEMS)

		if nextPageToken != "" {
			membersCall = membersCall.PageToken(nextPageToken)
		}

		members, err2 := membersCall.Do()

		if err2 != nil {
			return nil, err2
		}

		for _, m := range members.Members {
			res = append(res, fmt.Sprintf("user:%s", m.Email))
		}

		if members.NextPageToken != "" {
			nextPageToken = members.NextPageToken
		} else {
			break
		}
	}

	return res, nil
}

// below interface methods do not apply to GSuite so they return nil and/or no error
func (r *gsuiteIamRepository) GetServiceAccounts(ctx context.Context, configMap *config.ConfigMap, id string) ([]UserEntity, error) {
	return nil, nil
}

func (r *gsuiteIamRepository) GetIamPolicy(ctx context.Context, configMap *config.ConfigMap, id string) (IAMPolicyContainer, error) {
	return IAMPolicyContainer{}, nil
}

func (r *gsuiteIamRepository) AddBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	return nil
}

func (r *gsuiteIamRepository) RemoveBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	return nil
}
