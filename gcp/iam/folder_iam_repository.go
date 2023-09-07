package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/util/config"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v2"

	"github.com/raito-io/cli-plugin-gcp/gcp/common"
)

var folderIamPolicyCache map[string]*cloudresourcemanager.Policy = make(map[string]*cloudresourcemanager.Policy)

type folderIamRepository struct {
}

//nolint:dupl
func (r *folderIamRepository) GetUsers(ctx context.Context, configMap *config.ConfigMap, id string) ([]UserEntity, error) {
	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return nil, err
	}

	if policy.V2 == nil {
		common.Logger.Warn(fmt.Sprintf("getUsers: Could not retrieve IAM policy for folder %s", id))
		return []UserEntity{}, nil
	}

	users := make([]UserEntity, 0)
	externalIdList := map[string]struct{}{}

	for _, binding := range policy.V2.Bindings {
		for _, member := range binding.Members {
			if _, f := externalIdList[member]; f || !strings.HasPrefix(member, "user:") {
				continue
			}

			user := UserEntity{
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

func (r *folderIamRepository) GetGroups(ctx context.Context, configMap *config.ConfigMap, id string) ([]GroupEntity, error) {
	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return nil, err
	}

	if policy.V2 == nil {
		common.Logger.Warn(fmt.Sprintf("getGroups: Could not retrieve IAM policy for folder %s", id))
		return []GroupEntity{}, nil
	}

	groups := make([]GroupEntity, 0)
	externalIdList := map[string]struct{}{}

	for _, binding := range policy.V2.Bindings {
		for _, member := range binding.Members {
			if _, f := externalIdList[member]; f || !strings.HasPrefix(member, "group:") {
				continue
			}

			group := GroupEntity{
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
func (r *folderIamRepository) GetServiceAccounts(ctx context.Context, configMap *config.ConfigMap, id string) ([]UserEntity, error) {
	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return nil, err
	}

	if policy.V2 == nil {
		common.Logger.Warn(fmt.Sprintf("getServiceAccounts: Could not retrieve IAM policy for folder %s", id))
		return []UserEntity{}, nil
	}

	users := make([]UserEntity, 0)
	externalIdList := map[string]struct{}{}

	for _, binding := range policy.V2.Bindings {
		for _, member := range binding.Members {
			if _, f := externalIdList[member]; f || !strings.HasPrefix(member, "serviceAccount:") {
				continue
			}

			user := UserEntity{
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

func (r *folderIamRepository) GetIamPolicy(ctx context.Context, configMap *config.ConfigMap, id string) (IAMPolicyContainer, error) {
	if !strings.HasPrefix(id, "folders/") {
		id = fmt.Sprintf("folders/%s", id)
	}

	if _, f := folderIamPolicyCache[id]; f {
		return IAMPolicyContainer{V2: folderIamPolicyCache[id]}, nil
	}

	common.Logger.Info(fmt.Sprintf("Fetching the IAM policy for folder %s", id))

	crmService, err := common.CrmServiceV2(ctx, configMap)

	if err != nil {
		return IAMPolicyContainer{}, nil
	}

	policy, err := crmService.Folders.GetIamPolicy(id, new(cloudresourcemanager.GetIamPolicyRequest)).Do()

	if err != nil {
		if strings.Contains(err.Error(), "403") {
			common.Logger.Warn(fmt.Sprintf("Failed to fetch the IAM policyfor folder %s: %s", id, err.Error()))
			return IAMPolicyContainer{V2: &cloudresourcemanager.Policy{}}, nil
		} else {
			return IAMPolicyContainer{}, err
		}
	} else {
		folderIamPolicyCache[id] = policy
	}

	return IAMPolicyContainer{V2: folderIamPolicyCache[id]}, nil
}

//nolint:dupl
func (r *folderIamRepository) AddBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return err
	}

	// Find the policy binding for role. Only one binding can have the role.
	var binding *cloudresourcemanager.Binding

	for _, b := range policy.V2.Bindings {
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

		policy.V2.Bindings = append(policy.V2.Bindings, binding)
	}

	common.Logger.Info(fmt.Sprintf("Adding GCP Folder %s Iam Policy Binding: role %q member %q", id, member, role))

	return r.setPolicy(ctx, configMap, id, policy.V2)
}

func (r *folderIamRepository) RemoveBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	policy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return err
	}

	// Find the policy binding for role. Only one binding can have the role.
	var binding *cloudresourcemanager.Binding
	var bindingIndex int

	for i, b := range policy.V2.Bindings {
		if b.Role == role {
			binding = b
			bindingIndex = i

			break
		}
	}

	if binding == nil {
		common.Logger.Warn(fmt.Sprintf("Did not find binding for removal; GCP Folder %s Iam Policy Binding: role %q member %q", id, member, role))
		return nil
	}

	// Order doesn't matter for bindings or members, so to remove, move the last item
	// into the removed spot and shrink the slice.
	if len(binding.Members) == 1 {
		// If the member is the only member in the binding, removes the binding
		last := len(policy.V2.Bindings) - 1
		policy.V2.Bindings[bindingIndex] = policy.V2.Bindings[last]
		policy.V2.Bindings = policy.V2.Bindings[:last]
	} else {
		// If there is more than one member in the binding, removes the member
		var memberIndex int
		for i, mm := range binding.Members {
			if mm == member {
				memberIndex = i
			}
		}
		last := len(policy.V2.Bindings[bindingIndex].Members) - 1
		binding.Members[memberIndex] = binding.Members[last]
		binding.Members = binding.Members[:last]
	}

	common.Logger.Info(fmt.Sprintf("Removing GCP Folder %s Iam Policy Binding: role %q member %q", id, member, role))

	return r.setPolicy(ctx, configMap, id, policy.V2)
}

func (r *folderIamRepository) setPolicy(ctx context.Context, configMap *config.ConfigMap, id string, policy *cloudresourcemanager.Policy) error {
	request := new(cloudresourcemanager.SetIamPolicyRequest)
	request.Policy = policy

	crmService, err := common.CrmServiceV2(ctx, configMap)

	if err != nil {
		return err
	}

	if !strings.HasPrefix(id, "folders/") {
		id = fmt.Sprintf("folders/%s", id)
	}

	policy, err = crmService.Folders.SetIamPolicy(id, request).Do()

	// if no error update IAM policy in cache
	if _, f := folderIamPolicyCache[id]; f && err == nil {
		folderIamPolicyCache[id] = policy
	}

	return err
}
