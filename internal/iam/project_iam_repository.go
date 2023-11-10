package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/util/config"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

var projectIamPolicyCache map[string]*cloudresourcemanager.Policy = make(map[string]*cloudresourcemanager.Policy)

type projectIamRepository struct {
}

//nolint:dupl
func (r *projectIamRepository) GetUsers(ctx context.Context, configMap *config.ConfigMap, id string) ([]types.UserEntity, error) {
	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return nil, err
	}

	if policy.V1 == nil {
		common.Logger.Warn(fmt.Sprintf("getUsers: Could not retrieve IAM policy for project %s", id))
		return []types.UserEntity{}, nil
	}

	users := make([]types.UserEntity, 0)
	externalIdList := map[string]struct{}{}

	for _, binding := range policy.V1.Bindings {
		for _, member := range binding.Members {
			if _, f := externalIdList[member]; f || !strings.HasPrefix(member, "user:") {
				continue
			}

			user := types.UserEntity{
				ExternalId: member,
				Name:       strings.Replace(member, "user:", "", 1),
				Email:      strings.Split(member, ":")[1],
			}

			users = append(users, user)
			externalIdList[user.ExternalId] = struct{}{}
		}
	}

	return users, nil
}
func (r *projectIamRepository) GetGroups(ctx context.Context, configMap *config.ConfigMap, id string) ([]types.GroupEntity, error) {
	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return nil, err
	}

	if policy.V1 == nil {
		common.Logger.Warn(fmt.Sprintf("getGroups: Could not retrieve IAM policy for project %s", id))
		return []types.GroupEntity{}, nil
	}

	groups := make([]types.GroupEntity, 0)
	externalIdList := map[string]struct{}{}

	for _, binding := range policy.V1.Bindings {
		for _, member := range binding.Members {
			if _, f := externalIdList[member]; f || !strings.HasPrefix(member, "group:") {
				continue
			}

			group := types.GroupEntity{
				ExternalId: member,
				Email:      strings.Split(member, ":")[1],
			}

			groups = append(groups, group)
			externalIdList[group.ExternalId] = struct{}{}
		}
	}

	return groups, nil
}

//nolint:dupl
func (r *projectIamRepository) GetServiceAccounts(ctx context.Context, configMap *config.ConfigMap, id string) ([]types.UserEntity, error) {
	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return nil, err
	}

	if policy.V1 == nil {
		common.Logger.Warn(fmt.Sprintf("getServiceAccounts: Could not retrieve IAM policy for project %s", id))
		return []types.UserEntity{}, nil
	}

	users := make([]types.UserEntity, 0)
	externalIdList := map[string]struct{}{}

	for _, binding := range policy.V1.Bindings {
		for _, member := range binding.Members {
			if _, f := externalIdList[member]; f || !strings.HasPrefix(member, "serviceAccount:") {
				continue
			}

			user := types.UserEntity{
				ExternalId: member,
				Name:       strings.Replace(member, "serviceAccount:", "", 1),
				Email:      strings.Split(member, ":")[1],
			}

			users = append(users, user)
			externalIdList[user.ExternalId] = struct{}{}
		}
	}

	return users, nil
}

func (r *projectIamRepository) GetIamPolicy(ctx context.Context, configMap *config.ConfigMap, id string) (types.IAMPolicyContainer, error) {
	if _, f := projectIamPolicyCache[id]; f {
		return types.IAMPolicyContainer{V1: projectIamPolicyCache[id]}, nil
	}

	common.Logger.Info(fmt.Sprintf("Fetching the IAM policy for project %s", id))

	crmService, err := common.CrmService(ctx, configMap)

	if err != nil {
		return types.IAMPolicyContainer{}, err
	}

	policy, err := crmService.Projects.GetIamPolicy(id, new(cloudresourcemanager.GetIamPolicyRequest)).Do()

	if err != nil {
		if strings.Contains(err.Error(), "403") {
			common.Logger.Warn(fmt.Sprintf("Failed to fetch the IAM policyfor project %s: %s", id, err.Error()))
			return types.IAMPolicyContainer{}, nil
		} else {
			return types.IAMPolicyContainer{}, err
		}
	} else {
		projectIamPolicyCache[id] = policy
	}

	return types.IAMPolicyContainer{V1: projectIamPolicyCache[id]}, nil
}

//nolint:dupl
func (r *projectIamRepository) AddBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	common.Logger.Debug(fmt.Sprintf("Adding IAM binding for the GCP project %s", id))

	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return err
	}

	// Find the policy binding for role. Only one binding can have the role.
	var binding *cloudresourcemanager.Binding

	for _, b := range policy.V1.Bindings {
		if b.Role == role {
			binding = b
			break
		}
	}

	if binding != nil {
		// If the binding exists, adds the member to the binding
		for _, m := range binding.Members {
			if m == member {
				return nil
			}
		}

		binding.Members = append(binding.Members, member)
	} else {
		// If the binding does not exist, adds a new binding to the policy
		binding = &cloudresourcemanager.Binding{
			Role:    role,
			Members: []string{member},
		}

		policy.V1.Bindings = append(policy.V1.Bindings, binding)
	}

	common.Logger.Info(fmt.Sprintf("Adding GCP Project %s Iam Policy Binding: role %q member %q", id, member, role))

	return r.setPolicy(ctx, configMap, id, policy.V1)
}

func (r *projectIamRepository) RemoveBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return err
	}

	// Find the policy binding for role. Only one binding can have the role.
	var binding *cloudresourcemanager.Binding
	var bindingIndex int

	for i, b := range policy.V1.Bindings {
		if b.Role == role {
			binding = b
			bindingIndex = i

			break
		}
	}

	if binding == nil {
		common.Logger.Warn(fmt.Sprintf("Did not find binding for removal; Removing GCP Project %s Iam Policy Binding: role %q member %q", id, member, role))
		return nil
	}

	// Order doesn't matter for bindings or members, so to remove, move the last item
	// into the removed spot and shrink the slice.
	if len(binding.Members) == 1 {
		// If the member is the only member in the binding, removes the binding
		last := len(policy.V1.Bindings) - 1
		policy.V1.Bindings[bindingIndex] = policy.V1.Bindings[last]
		policy.V1.Bindings = policy.V1.Bindings[:last]
	} else {
		// If there is more than one member in the binding, removes the member
		var memberIndex int
		for i, mm := range binding.Members {
			if mm == member {
				memberIndex = i
			}
		}
		last := len(policy.V1.Bindings[bindingIndex].Members) - 1
		binding.Members[memberIndex] = binding.Members[last]
		binding.Members = binding.Members[:last]
	}

	common.Logger.Info(fmt.Sprintf("Removing GCP Project %s Iam Policy Binding: role %q member %q", id, member, role))

	return r.setPolicy(ctx, configMap, id, policy.V1)
}

func (r *projectIamRepository) setPolicy(ctx context.Context, configMap *config.ConfigMap, id string, policy *cloudresourcemanager.Policy) error {
	request := new(cloudresourcemanager.SetIamPolicyRequest)
	request.Policy = policy

	crmService, err := common.CrmService(ctx, configMap)

	if err != nil {
		return err
	}

	policy, err = crmService.Projects.SetIamPolicy(id, request).Do()

	// if no error update IAM policy in cache
	if _, f := projectIamPolicyCache[id]; f && err == nil {
		projectIamPolicyCache[id] = policy
	}

	return err
}
