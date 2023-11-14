package bigquery

import (
	"context"
	"os"

	"cloud.google.com/go/bigquery"
	"github.com/raito-io/cli/base/util/config"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	admin "google.golang.org/api/admin/directory/v1"
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
		return nil, nil, err
	}

	return client, func() {
		client.Close()
	}, nil
}

func getConfig(configMap *config.ConfigMap, scopes ...string) (*jwt.Config, error) {
	key := configMap.GetString(common.GcpSAFileLocation)

	if key == "" {
		key = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	serviceAccountJSON, err := os.ReadFile(key)
	if err != nil {
		return nil, err
	}

	return google.JWTConfigFromJSON(serviceAccountJSON, scopes...)
}
