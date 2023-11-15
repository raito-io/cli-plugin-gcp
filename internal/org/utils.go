package org

import (
	"context"
	"fmt"

	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/googleapis/gax-go/v2"

	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

type getPolicyClient interface {
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}

type setPolicyClient interface {
	getPolicyClient
	SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
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

func addBinding(ctx context.Context, policyClient setPolicyClient, binding *iam.IamBinding) error {
	resourceName := _resourceName(binding.ResourceType, binding.Resource)

	resourcePolicy, err := policyClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: resourceName})
	if err != nil {
		return fmt.Errorf("get iam policy for %q: %w", resourceName, err)
	}

	updateExistingBinding := false

	for i := range resourcePolicy.Bindings {
		if resourcePolicy.Bindings[i].Role == binding.Role {
			resourcePolicy.Bindings[i].Members = append(resourcePolicy.Bindings[i].Members, binding.Member)
			updateExistingBinding = true
		}
	}

	if !updateExistingBinding {
		resourcePolicy.Bindings = append(resourcePolicy.Bindings, &iampb.Binding{
			Role:    binding.Role,
			Members: []string{binding.Member},
		})
	}

	_, err = policyClient.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{Resource: resourceName, Policy: resourcePolicy})
	if err != nil {
		return fmt.Errorf("set iam policy for %q: %w", resourceName, err)
	}

	return nil
}

func removeBinding(ctx context.Context, policyClient setPolicyClient, binding *iam.IamBinding) error {
	resourceName := _resourceName(binding.ResourceType, binding.Resource)

	resourcePolicy, err := policyClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: resourceName})
	if err != nil {
		return fmt.Errorf("get iam policy for %q: %w", resourceName, err)
	}

	for i := range resourcePolicy.Bindings {
		if resourcePolicy.Bindings[i].Role == binding.Role {
			updatedMembers := make([]string, 0, len(resourcePolicy.Bindings[i].Members))

			for j := range resourcePolicy.Bindings[i].Members {
				if resourcePolicy.Bindings[i].Members[j] != binding.Member {
					updatedMembers = append(updatedMembers, resourcePolicy.Bindings[i].Members[j])
				}
			}

			resourcePolicy.Bindings[i].Members = updatedMembers
		}
	}

	_, err = policyClient.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{Resource: resourceName, Policy: resourcePolicy})
	if err != nil {
		return fmt.Errorf("set iam policy for %q: %w", resourceName, err)
	}

	return nil
}

func _resourceName(resourceType string, resourceId string) string {
	return fmt.Sprintf("%ss/%s", resourceType, resourceId)
}
