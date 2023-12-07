package org

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/raito-io/golang-set/set"
	"google.golang.org/api/iterator"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

type getPolicyClient interface {
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}

type setPolicyClient interface {
	getPolicyClient
	SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}
type tagBindingsClient interface {
	ListEffectiveTags(ctx context.Context, req *resourcemanagerpb.ListEffectiveTagsRequest, opts ...gax.CallOption) *resourcemanager.EffectiveTagIterator
}

func getAndParseBindings(ctx context.Context, policyClient getPolicyClient, resourceType string, resourceId string) ([]iam.IamBinding, error) {
	resourceName := _resourceName(resourceType, resourceId)

	policy, err := policyClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: resourceName})
	if err != nil {
		return nil, fmt.Errorf("get %s iam policy: %w", resourceType, err)
	}

	var result []iam.IamBinding

	for _, binding := range policy.Bindings {
		for _, member := range binding.Members {
			result = append(result, iam.IamBinding{
				Role:         binding.Role,
				Member:       member,
				Resource:     resourceId,
				ResourceType: resourceType,
			})
		}
	}

	return result, nil
}

func getTagsForResource(ctx context.Context, tagBindingsClient tagBindingsClient, req *resourcemanagerpb.ListEffectiveTagsRequest) map[string]string {
	tags := make(map[string]string)

	tagsIterator := tagBindingsClient.ListEffectiveTags(ctx, req)

	for {
		tag, err := tagsIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while listing tags: %s", err.Error()))

			continue
		} else if err != nil {
			common.Logger.Error(fmt.Sprintf("Issue when retrieving tags for resource: %s", req.Parent), err.Error())
			return nil
		}

		if tag.Inherited == false {
			tags[tag.TagKey] = tag.TagValue
		}
	}

	return tags
}

func updateBindings(ctx context.Context, policyClient setPolicyClient, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding) error {
	membersToRemoveFromRole := map[string]set.Set[string]{} // Role -> Member
	membersToAddToRole := map[string][]string{}             // Role -> Member
	rolesToAdd := set.Set[string]{}

	for _, binding := range bindingsToAdd {
		membersToAddToRole[binding.Role] = append(membersToAddToRole[binding.Role], binding.Member)
		rolesToAdd.Add(binding.Role)
	}

	for _, binding := range bindingsToDelete {
		if _, found := membersToRemoveFromRole[binding.Role]; !found {
			membersToRemoveFromRole[binding.Role] = set.Set[string]{}
		}

		membersToRemoveFromRole[binding.Role].Add(binding.Member)
	}

	resourceName := _resourceName(dataObject.ObjectType, dataObject.FullName)

	resourcePolicy, err := policyClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: resourceName})
	if err != nil {
		return fmt.Errorf("get iam policy for %q: %w", resourceName, err)
	}

	for i := range resourcePolicy.Bindings {
		// Remove old assignees
		if membersToRemove, found := membersToRemoveFromRole[resourcePolicy.Bindings[i].Role]; found {
			updatedMembers := make([]string, 0, len(resourcePolicy.Bindings[i].Members))

			for _, m := range resourcePolicy.Bindings[i].Members {
				if !membersToRemove.Contains(m) {
					updatedMembers = append(updatedMembers, m)
				}
			}

			resourcePolicy.Bindings[i].Members = updatedMembers
		}

		// Add new assignees
		if members, found := membersToAddToRole[resourcePolicy.Bindings[i].Role]; found {
			resourcePolicy.Bindings[i].Members = append(resourcePolicy.Bindings[i].Members, members...)
			rolesToAdd.Remove(resourcePolicy.Bindings[i].Role)
		}
	}

	for role := range rolesToAdd {
		resourcePolicy.Bindings = append(resourcePolicy.Bindings, &iampb.Binding{
			Role:    role,
			Members: membersToAddToRole[role],
		})
	}

	_, err = policyClient.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{Resource: resourceName, Policy: resourcePolicy})
	if err != nil {
		return fmt.Errorf("update iam policy for %q: %w", resourceName, err)
	}

	return nil
}

func _resourceName(resourceType string, resourceId string) string {
	return fmt.Sprintf("%ss/%s", resourceType, resourceId)
}
