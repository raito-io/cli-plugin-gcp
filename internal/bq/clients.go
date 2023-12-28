package bigquery

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/bigquery"
	datapolicies "cloud.google.com/go/bigquery/datapolicies/apiv1"
	datacatalog "cloud.google.com/go/datacatalog/apiv1"
	"github.com/raito-io/cli/base/util/config"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	admin "google.golang.org/api/admin/directory/v1"
	bigquery2 "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/option"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

func NewBiqQueryClient(ctx context.Context, configMap *config.ConfigMap) (*bigquery.Client, func(), error) {
	gcpProjectId := configMap.GetString(common.GcpProjectId)

	config, err := getConfig(configMap, admin.CloudPlatformScope)

	if err != nil {
		return nil, nil, err
	}

	client, err := bigquery.NewClient(ctx, gcpProjectId, option.WithHTTPClient(config.Client(ctx)))
	if err != nil {
		return nil, nil, fmt.Errorf("new bq client: %w", err)
	}

	return client, func() {
		client.Close()
	}, nil
}

func NewPolicyTagClient(ctx context.Context, configMap *config.ConfigMap) (*datacatalog.PolicyTagManagerClient, func(), error) {
	client, err := datacatalog.NewPolicyTagManagerClient(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)), option.WithScopes(admin.CloudPlatformScope))
	if err != nil {
		return nil, nil, fmt.Errorf("new policy tag manager client: %w", err)
	}

	return client, func() {
		client.Close()
	}, nil
}

func NewDataPolicyClient(ctx context.Context, configMap *config.ConfigMap) (*datapolicies.DataPolicyClient, func(), error) {
	client, err := datapolicies.NewDataPolicyClient(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)), option.WithScopes(admin.CloudPlatformScope))
	if err != nil {
		return nil, nil, fmt.Errorf("new data policy client: %w", err)
	}

	return client, func() {
		client.Close()
	}, nil
}

func NewServiceClient(ctx context.Context, configMap *config.ConfigMap) (*bigquery2.Service, error) {
	service, err := bigquery2.NewService(ctx, option.WithCredentialsFile(configMap.GetString(common.GcpSAFileLocation)), option.WithScopes(admin.CloudPlatformScope, bigquery2.BigqueryScope))
	if err != nil {
		return nil, fmt.Errorf("new service client: %w", err)
	}

	return service, nil
}

func NewRowAccessClient(service *bigquery2.Service) *bigquery2.RowAccessPoliciesService {
	return service.RowAccessPolicies
}

func getConfig(configMap *config.ConfigMap, scopes ...string) (*jwt.Config, error) {
	key := configMap.GetString(common.GcpSAFileLocation)

	if key == "" {
		key = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	serviceAccountJSON, err := os.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", key, err)
	}

	jwtConfig, err := google.JWTConfigFromJSON(serviceAccountJSON, scopes...)
	if err != nil {
		return nil, fmt.Errorf("create jwt config from file %q: %w", key, err)
	}

	return jwtConfig, nil
}
