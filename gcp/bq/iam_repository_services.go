package bigquery

import (
	"context"

	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/gcp/iam"
)

var _ iam.IAMRepository = (*IamRepositories)(nil)

type IamRepositories struct {
	services map[string]iam.IAMRepository
}

func (i IamRepositories) GetUsers(ctx context.Context, configMap *config.ConfigMap, id string) ([]iam.UserEntity, error) {
	var result []iam.UserEntity

	for _, service := range i.services {
		serviceResult, err := service.GetUsers(ctx, configMap, id)
		if err != nil {
			return result, err
		}

		result = append(result, serviceResult...)
	}

	return result, nil
}

func (i IamRepositories) GetGroups(ctx context.Context, configMap *config.ConfigMap, id string) ([]iam.GroupEntity, error) {
	var result []iam.GroupEntity

	for _, service := range i.services {
		serviceResult, err := service.GetGroups(ctx, configMap, id)
		if err != nil {
			return result, err
		}

		result = append(result, serviceResult...)
	}

	return result, nil
}

func (i IamRepositories) GetServiceAccounts(ctx context.Context, configMap *config.ConfigMap, id string) ([]iam.UserEntity, error) {
	var result []iam.UserEntity

	for _, service := range i.services {
		serviceResult, err := service.GetServiceAccounts(ctx, configMap, id)
		if err != nil {
			return result, err
		}

		result = append(result, serviceResult...)
	}

	return result, nil
}

func (i IamRepositories) GetIamPolicy(ctx context.Context, configMap *config.ConfigMap, id string) (iam.IAMPolicyContainer, error) {
	//TODO implement me
	panic("implement me")
}

func (i IamRepositories) AddBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	//TODO implement me
	panic("implement me")
}

func (i IamRepositories) RemoveBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	//TODO implement me
	panic("implement me")
}
