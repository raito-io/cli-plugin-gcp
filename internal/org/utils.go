package org

import (
	"context"
	"fmt"

	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/googleapis/gax-go/v2"

	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

type policyClient interface {
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}

func parseBindings(ctx context.Context, policyClient policyClient, resourceType string, resourceId string) ([]types.IamBinding, error) {
	resourceName := fmt.Sprintf("%ss/%s", resourceType, resourceId)

	policy, err := policyClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: resourceName})
	if err != nil {
		return nil, fmt.Errorf("get %s iam policy: %w", resourceType, err)
	}

	var result []types.IamBinding

	for _, binding := range policy.Bindings {
		for _, member := range binding.Members {
			result = append(result, types.IamBinding{
				Role:         binding.Role,
				Member:       member,
				Resource:     resourceId,
				ResourceType: resourceType,
			})
		}
	}

	return result, nil
}
