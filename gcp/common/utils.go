package common

import (
	"context"
	"os"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/raito-io/cli/base"
	"github.com/raito-io/cli/base/util/config"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	admin "google.golang.org/api/admin/directory/v1"
	crmV1 "google.golang.org/api/cloudresourcemanager/v1"
	crmV2 "google.golang.org/api/cloudresourcemanager/v2"
	crmV3 "google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"
)

const CONTEXT_TIMEOUT = 10 * time.Second

var Logger hclog.Logger

func init() {
	Logger = base.Logger()
}

func getConfig(configMap *config.ConfigMap, scopes ...string) (*jwt.Config, error) {
	key := configMap.GetString(GcpSAFileLocation)

	if key == "" {
		key = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	serviceAccountJSON, err := os.ReadFile(key)
	if err != nil {
		return nil, err
	}

	return google.JWTConfigFromJSON(serviceAccountJSON, scopes...)
}

func CrmService(ctx context.Context, configMap *config.ConfigMap) (*crmV1.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, CONTEXT_TIMEOUT)
	defer cancel()

	config, err := getConfig(configMap, admin.CloudPlatformScope)

	if err != nil {
		return nil, err
	}

	return crmV1.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
}

func CrmServiceV2(ctx context.Context, configMap *config.ConfigMap) (*crmV2.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, CONTEXT_TIMEOUT)
	defer cancel()

	config, err := getConfig(configMap, admin.CloudPlatformScope)

	if err != nil {
		return nil, err
	}

	return crmV2.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
}

func CrmServiceV3(ctx context.Context, configMap *config.ConfigMap) (*crmV3.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, CONTEXT_TIMEOUT)
	defer cancel()

	config, err := getConfig(configMap, admin.CloudPlatformScope)

	if err != nil {
		return nil, err
	}

	return crmV3.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
}
