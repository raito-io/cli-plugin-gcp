package iam

import (
	"strings"

	crmV1 "google.golang.org/api/cloudresourcemanager/v1"
	crmV2 "google.golang.org/api/cloudresourcemanager/v2"
)

type GroupEntity struct {
	ExternalId string
	Email      string
	Members    []string
}

type UserEntity struct {
	ExternalId string
	Name       string
	Email      string
}

type IamBinding struct {
	Member       string
	Role         string
	Resource     string
	ResourceType string
}

//go:generate go run github.com/raito-io/enumer -gqlgen -type=IamType
type IamType int

const (
	Project IamType = iota
	Folder
	Organization
	GSuite
	Service
)

type IAMPolicyContainer struct {
	V1      *crmV1.Policy
	V2      *crmV2.Policy
	Service []IamBinding // Service IAM repo should immediately fill bindings as []IamBinding
}

type IAMBindings struct {
	Type IamType
}

func (a IamBinding) Equals(b IamBinding) bool {
	return strings.EqualFold(a.Member, b.Member) && strings.EqualFold(a.Role, b.Role) && strings.EqualFold(a.Resource, b.Resource) && strings.EqualFold(a.ResourceType, b.ResourceType)
}
