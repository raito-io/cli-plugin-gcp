package org

import (
	"context"
	"fmt"

	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

type organizationClient interface {
	GetOrganization(ctx context.Context, req *resourcemanagerpb.GetOrganizationRequest, opts ...gax.CallOption) (*resourcemanagerpb.Organization, error)
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
	SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
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
		Name:      organization.DisplayName,
		Id:        name,
		FullName:  name,
		Type:      "organization",
		Parent:    nil,
	}, nil
}

func (r *OrganizationRepository) GetIamPolicy(ctx context.Context, _ string) ([]iam.IamBinding, error) {
	return getAndParseBindings(ctx, r.organizationClient, "organization", r.organizationId)
}

func (r *OrganizationRepository) AddBinding(ctx context.Context, binding *iam.IamBinding) error {
	binding.Resource = r.organizationId

	return addBinding(ctx, r.organizationClient, binding)
}

func (r *OrganizationRepository) RemoveBinding(ctx context.Context, binding *iam.IamBinding) error {
	binding.Resource = r.organizationId

	return removeBinding(ctx, r.organizationClient, binding)
}
