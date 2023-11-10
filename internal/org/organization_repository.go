package org

import (
	"context"
	"fmt"

	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
)

type organizationClient interface {
	GetOrganization(ctx context.Context, req *resourcemanagerpb.GetOrganizationRequest, opts ...gax.CallOption) (*resourcemanagerpb.Organization, error)
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}

type OrganizationRepository struct {
	organizationClient organizationClient
	organizationId     string
}

func NewOrganizationRepository(organizationClient organizationClient, configMap *config.ConfigMap) *OrganizationRepository {
	return &OrganizationRepository{
		organizationClient: organizationClient,
		organizationId:     configMap.GetString(common.GcpOrgId),
	}
}

func (r *OrganizationRepository) GetOrganization(ctx context.Context) (*GcpOrgEntity, error) {
	organization, err := r.organizationClient.GetOrganization(ctx, &resourcemanagerpb.GetOrganizationRequest{
		Name: fmt.Sprintf("organizations/%s", r.organizationId),
	})

	name := fmt.Sprintf("gcp-org-%s", r.organizationId)

	if err != nil {
		return nil, fmt.Errorf("get organization %q: %w", r.organizationId, err)
	}

	return &GcpOrgEntity{
		EntryName: organization.Name,
		Name:      name,
		Id:        name,
		Type:      "organization",
		Parent:    nil,
	}, nil
}

func (r *OrganizationRepository) GetIamPolicies(ctx context.Context) ([]types.IamBinding, error) {
	return parseBindings(ctx, r.organizationClient, "organization", r.organizationId)
}
