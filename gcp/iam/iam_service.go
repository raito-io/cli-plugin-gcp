package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli-plugin-gcp/gcp/common"
	"github.com/raito-io/cli-plugin-gcp/gcp/org"
)

const ownerRole = "roles/owner"
const editorRole = "roles/editor"
const viewerRole = "roles/viewer"

//go:generate go run github.com/vektra/mockery/v2 --name=IAMService --with-expecter --inpackage
type IAMService interface {
	WithServiceIamRepo(resourceTypes []string, localRepo IAMRepository, ids func(ctx context.Context, configMap *config.ConfigMap) ([]string, error)) IAMService
	GetUsers(ctx context.Context, configMap *config.ConfigMap) ([]UserEntity, error)
	GetGroups(ctx context.Context, configMap *config.ConfigMap) ([]GroupEntity, error)
	GetServiceAccounts(ctx context.Context, configMap *config.ConfigMap) ([]UserEntity, error)
	GetIAMPolicyBindings(ctx context.Context, configMap *config.ConfigMap) ([]IamBinding, error)
	AddIamBinding(ctx context.Context, configMap *config.ConfigMap, binding IamBinding) error
	RemoveIamBinding(ctx context.Context, configMap *config.ConfigMap, binding IamBinding) error
	GetProjectOwners(ctx context.Context, configMap *config.ConfigMap, projectId string) (owner []string, editor []string, viewer []string, err error)
}

//go:generate go run github.com/vektra/mockery/v2 --name=IAMRepository --with-expecter --inpackage
type IAMRepository interface {
	GetUsers(ctx context.Context, configMap *config.ConfigMap, id string) ([]UserEntity, error)
	GetGroups(ctx context.Context, configMap *config.ConfigMap, id string) ([]GroupEntity, error)
	GetServiceAccounts(ctx context.Context, configMap *config.ConfigMap, id string) ([]UserEntity, error)
	GetIamPolicy(ctx context.Context, configMap *config.ConfigMap, id string) (IAMPolicyContainer, error)
	AddBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error
	RemoveBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error
}

type dataSourceRepository interface {
	GetProjects(ctx context.Context, configMap *config.ConfigMap) ([]org.GcpOrgEntity, error)
	GetFolders(ctx context.Context, configMap *config.ConfigMap) ([]org.GcpOrgEntity, error)
}

type iamService struct {
	repos            map[IamType]IAMRepository
	gcpRepo          dataSourceRepository
	serviceRepoIds   func(ctx context.Context, configMap *config.ConfigMap) ([]string, error)
	serviceRepoTypes set.Set[string]
}

func NewIAMService(configMap *config.ConfigMap) *iamService {
	repos := map[IamType]IAMRepository{
		Organization: &organizationIamRepository{},
		Folder:       &folderIamRepository{},
		Project:      &projectIamRepository{},
	}

	if configMap.GetStringWithDefault(common.GcpProjectId, "") != "" {
		repos = map[IamType]IAMRepository{
			Project: &projectIamRepository{},
		}
	}

	if configMap.GetBool(common.GsuiteIdentityStoreSync) {
		repos[GSuite] = &gsuiteIamRepository{}
	}

	return &iamService{
		repos:   repos,
		gcpRepo: org.NewGCPRepository(),
	}
}

func (s *iamService) WithServiceIamRepo(resourceTypes []string, serviceRepo IAMRepository, ids func(ctx context.Context, configMap *config.ConfigMap) ([]string, error)) IAMService {
	s.repos[Service] = serviceRepo
	s.serviceRepoIds = ids
	s.serviceRepoTypes = set.NewSet[string](resourceTypes...)

	return s
}

func (s *iamService) GetUsers(ctx context.Context, configMap *config.ConfigMap) ([]UserEntity, error) {
	users := make([]UserEntity, 0)

	ids := set.NewSet[string]()

	typeToIdsMap, err := s.getIdsByRepoType(ctx, configMap)

	if err != nil {
		return users, err
	}

	for t, repo := range s.repos {
		for _, id := range typeToIdsMap[t] {
			items, err2 := repo.GetUsers(ctx, configMap, id)

			if err2 != nil {
				return nil, err2
			}

			for _, item := range items {
				if !ids.Contains(item.ExternalId) || t == GSuite {
					ids.Add(item.ExternalId)
					users = append(users, item)
				}
			}
		}
	}

	return users, nil
}

func (s *iamService) GetGroups(ctx context.Context, configMap *config.ConfigMap) ([]GroupEntity, error) {
	groups := make([]GroupEntity, 0)

	ids := set.NewSet[string]()

	typeToIdsMap, err := s.getIdsByRepoType(ctx, configMap)

	if err != nil {
		return groups, err
	}

	for t, repo := range s.repos {
		for _, id := range typeToIdsMap[t] {
			items, err2 := repo.GetGroups(ctx, configMap, id)

			if err2 != nil {
				return nil, err2
			}

			for _, item := range items {
				if !ids.Contains(item.ExternalId) || t == GSuite {
					ids.Add(item.ExternalId)
					groups = append(groups, item)
				}
			}
		}
	}

	return groups, nil
}

func (s *iamService) GetServiceAccounts(ctx context.Context, configMap *config.ConfigMap) ([]UserEntity, error) {
	serviceAccounts := make([]UserEntity, 0)

	typeToIdsMap, err := s.getIdsByRepoType(ctx, configMap)

	if err != nil {
		return serviceAccounts, err
	}

	for t, repo := range s.repos {
		for _, id := range typeToIdsMap[t] {
			u, err2 := repo.GetServiceAccounts(ctx, configMap, id)

			if err2 != nil {
				return nil, err2
			}

			serviceAccounts = append(serviceAccounts, u...)
		}
	}

	return serviceAccounts, nil
}

func (s *iamService) GetIAMPolicyBindings(ctx context.Context, configMap *config.ConfigMap) ([]IamBinding, error) {
	bindings := []IamBinding{}
	typeToIdsMap, err := s.getIdsByRepoType(ctx, configMap)

	if err != nil {
		return nil, err
	}

	for t, repo := range s.repos {
		for _, id := range typeToIdsMap[t] {
			policyContainer, err2 := repo.GetIamPolicy(ctx, configMap, id)

			if err2 != nil {
				return nil, err2
			}

			if policyContainer.Service != nil {
				bindings = append(bindings, policyContainer.Service...)
			} else if policyContainer.V1 != nil {
				for _, binding := range policyContainer.V1.Bindings {
					for _, member := range binding.Members {
						bindings = append(bindings, IamBinding{
							Member:       member,
							Role:         binding.Role,
							Resource:     id,
							ResourceType: strings.ToLower(t.String()),
						})
					}
				}
			} else if policyContainer.V2 != nil {
				for _, binding := range policyContainer.V2.Bindings {
					for _, member := range binding.Members {
						bindings = append(bindings, IamBinding{
							Member:       member,
							Role:         binding.Role,
							Resource:     id,
							ResourceType: strings.ToLower(t.String()),
						})
					}
				}
			}
		}
	}

	return bindings, nil
}

func (s *iamService) GetProjectOwners(ctx context.Context, configMap *config.ConfigMap, projectId string) (owner []string, editor []string, viewer []string, err error) {
	repo := s.repos[Project]

	policyContainer, err := repo.GetIamPolicy(ctx, configMap, projectId)
	if err != nil {
		return nil, nil, nil, err
	}

	if policyContainer.Service != nil {
		for _, binding := range policyContainer.Service {
			if binding.Role == ownerRole {
				owner = append(owner, binding.Member)
			} else if binding.Role == editorRole {
				editor = append(editor, binding.Member)
			} else if binding.Role == viewerRole {
				viewer = append(viewer, binding.Member)
			}
		}
	} else if policyContainer.V1 != nil {
		for _, binding := range policyContainer.V1.Bindings {
			if binding.Role == ownerRole {
				owner = append(owner, binding.Members...)
			} else if binding.Role == editorRole {
				editor = append(editor, binding.Members...)
			} else if binding.Role == viewerRole {
				viewer = append(viewer, binding.Members...)
			}
		}
	} else if policyContainer.V2 != nil {
		for _, binding := range policyContainer.V2.Bindings {
			if binding.Role == ownerRole {
				owner = append(owner, binding.Members...)
			} else if binding.Role == editorRole {
				editor = append(editor, binding.Members...)
			} else if binding.Role == viewerRole {
				viewer = append(viewer, binding.Members...)
			}
		}
	}

	return owner, editor, viewer, nil
}

func (s *iamService) AddIamBinding(ctx context.Context, configMap *config.ConfigMap, binding IamBinding) error {
	if s.serviceRepoTypes.Contains(binding.ResourceType) {
		binding.ResourceType = Service.String()
	}

	for t, repo := range s.repos {
		if strings.EqualFold(t.String(), binding.ResourceType) {
			return repo.AddBinding(ctx, configMap, binding.Resource, binding.Member, binding.Role)
		}
	}

	return fmt.Errorf("adding IAM bindings for resource type %s is not supported", binding.ResourceType)
}

func (s *iamService) RemoveIamBinding(ctx context.Context, configMap *config.ConfigMap, binding IamBinding) error {
	if s.serviceRepoTypes.Contains(binding.ResourceType) {
		binding.ResourceType = Service.String()
	}

	for t, repo := range s.repos {
		if strings.EqualFold(t.String(), binding.ResourceType) {
			return repo.RemoveBinding(ctx, configMap, binding.Resource, binding.Member, binding.Role)
		}
	}

	return fmt.Errorf("removing IAM bindings for resource type %s is not supported", binding.ResourceType)
}

func (s *iamService) getIdsByRepoType(ctx context.Context, configMap *config.ConfigMap) (map[IamType][]string, error) {
	out := map[IamType][]string{}

	for t := range s.repos {
		out[t] = make([]string, 0)
	}

	// add gcp org id
	if _, f := out[Organization]; f {
		out[Organization] = append(out[Organization], configMap.GetString(common.GcpOrgId))
	}

	// if we have a GCP Service repo (e.g. BigQuery) we add run the serviceRepoIds method to acquire resource ids
	if _, f := out[Service]; f {
		ids, err := s.serviceRepoIds(ctx, configMap)

		if err != nil {
			return nil, err
		}

		out[Service] = ids
	}

	// GSuite we add empty string as id to get it in the loop as it does not have resource Ids to loop over
	if _, f := out[GSuite]; f {
		out[GSuite] = append(out[GSuite], "")
	}

	// get project ids
	if _, f := out[Project]; f {
		gcpProjectId := configMap.GetString(common.GcpProjectId)

		if gcpProjectId != "" {
			out[Project] = append(out[Project], gcpProjectId)
			return out, nil
		}

		gcpProjectIdList, err := s.gcpRepo.GetProjects(ctx, configMap)

		if err != nil {
			return nil, err
		}

		for _, project := range gcpProjectIdList {
			out[Project] = append(out[Project], project.Id)
		}
	}

	// get project ids
	if _, f := out[Folder]; f {
		gcpFolderIds, err := s.gcpRepo.GetFolders(ctx, configMap)

		if err != nil {
			return nil, err
		}

		for _, folder := range gcpFolderIds {
			out[Folder] = append(out[Folder], folder.Id)
		}
	}

	return out, nil
}
