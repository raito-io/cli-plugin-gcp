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
	tagBindingsClient  tagBindingsClient
}

func NewOrganizationRepository(organizationClient organizationClient, tagBindingsClient tagBindingsClient, configMap *config.ConfigMap) *OrganizationRepository {
	return &OrganizationRepository{
		organizationClient: organizationClient,
		organizationId:     configMap.GetString(common.GcpOrgId),
		tagBindingsClient:  tagBindingsClient,
	}
}

func (r *OrganizationRepository) GetOrganization(ctx context.Context) (*GcpOrgEntity, error) {
	organization, err := r.organizationClient.GetOrganization(ctx, &resourcemanagerpb.GetOrganizationRequest{
		Name: fmt.Sprintf("organizations/%s", r.organizationId),
	})

	name := r.raitoOrgId()
	entryName := r.organizationId
	displayname := name

	if common.IsGoogle400Error(err) {
		common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching organisation information: %s", err.Error()))

		return nil, nil
	} else if err != nil {
		common.Logger.Warn(fmt.Sprintf("Encountered error while fetching organisation information: %s", err.Error()))
	} else {
		entryName = organization.Name
		displayname = organization.DisplayName
	}

	tags := getTagsForResource(ctx, r.tagBindingsClient, &resourcemanagerpb.ListEffectiveTagsRequest{
		Parent: fmt.Sprintf("//cloudresourcemanager.googleapis.com/%s", entryName)})

	return &GcpOrgEntity{
		EntryName: entryName,
		Name:      displayname,
		Id:        name,
		FullName:  name,
		Type:      "organization",
		Parent:    nil,
		Tags:      tags,
	}, nil
}

func (r *OrganizationRepository) GetIamPolicy(ctx context.Context, _ string) ([]iam.IamBinding, error) {
	bindings, err := getAndParseBindings(ctx, r.organizationClient, "organization", r.organizationId)
	if err != nil {
		return nil, err
	}

	for i := range bindings {
		bindings[i].Resource = r.raitoOrgId()
	}

	return bindings, nil
}

func (r *OrganizationRepository) UpdateBinding(ctx context.Context, dataObject *iam.DataObjectReference, bindingsToAdd []iam.IamBinding, bindingsToDelete []iam.IamBinding) error {
	updatedBindingsToAdd := make([]iam.IamBinding, len(bindingsToAdd))
	for i := range bindingsToAdd {
		updatedBindingsToAdd[i] = bindingsToAdd[i]
		updatedBindingsToAdd[i].Resource = r.organizationId
	}

	updatedBindingsToRemove := make([]iam.IamBinding, len(bindingsToDelete))
	for i := range bindingsToDelete {
		updatedBindingsToRemove[i] = bindingsToDelete[i]
		updatedBindingsToRemove[i].Resource = r.organizationId
	}

	do := *dataObject
	do.FullName = r.organizationId

	return updateBindings(ctx, r.organizationClient, &do, updatedBindingsToAdd, updatedBindingsToRemove)
}

func (r *OrganizationRepository) raitoOrgId() string {
	return fmt.Sprintf("gcp-org-%s", r.organizationId)
}
