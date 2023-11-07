package bigquery

import (
	"context"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/internal/gcp"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"

	is "github.com/raito-io/cli/base/identity_store"
)

type IdentityStoreSyncer struct {
	iamServiceProvider func(configMap *config.ConfigMap) iam.IAMService
}

func NewIdentityStoreSyncer() *IdentityStoreSyncer {
	return &IdentityStoreSyncer{iamServiceProvider: newIamServiceProvider}
}

func newIamServiceProvider(configMap *config.ConfigMap) iam.IAMService {
	return iam.NewIAMService(configMap)
}

func (s *IdentityStoreSyncer) GetIdentityStoreMetaData(_ context.Context, _ *config.ConfigMap) (*is.MetaData, error) {
	logger.Debug("Returning meta data for BigQuery identity store")

	return &is.MetaData{
		Type:        "bigquery",
		CanBeLinked: false,
		CanBeMaster: false,
	}, nil
}

func (s *IdentityStoreSyncer) SyncIdentityStore(ctx context.Context, identityHandler wrappers.IdentityStoreIdentityHandler, configMap *config.ConfigMap) error {
	syncer := gcp.NewIdentityStoreSyncer().WithIAMServiceProvider(func(configMap *config.ConfigMap) iam.IAMService {
		return s.iamServiceProvider(configMap).WithServiceIamRepo([]string{}, &bigQueryIamRepository{}, GetResourceIds)
	})

	return syncer.SyncIdentityStore(ctx, identityHandler, configMap)
}
