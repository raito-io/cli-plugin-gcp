package admin

import (
	"context"
	"fmt"
	"os"

	"github.com/raito-io/cli/base/util/config"
	"golang.org/x/oauth2/google"
	gcpadmin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

func NewGcpAdminService(ctx context.Context, configMap *config.ConfigMap) (*gcpadmin.Service, error) {
	key := configMap.GetString(common.GcpSAFileLocation)

	if key == "" {
		key = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	serviceAccountJSON, err := os.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", key, err)
	}

	config, err := google.JWTConfigFromJSON(serviceAccountJSON, gcpadmin.AdminDirectoryGroupReadonlyScope, gcpadmin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("create jwt config from file %q: %w", key, err)
	}

	config.Subject = configMap.GetString(common.GsuiteImpersonateSubject)

	customerId := configMap.GetString(common.GsuiteCustomerId)

	if customerId == "" || config.Subject == "" {
		return nil, fmt.Errorf("for GSuite identity store sync please configure %s and %s", common.GsuiteCustomerId, common.GsuiteImpersonateSubject)
	}

	service, err := gcpadmin.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
	if err != nil {
		return nil, fmt.Errorf("create gcp admin service: %w", err)
	}

	return service, nil
}
