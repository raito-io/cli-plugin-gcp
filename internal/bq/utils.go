package bigquery

import (
	"context"

	"cloud.google.com/go/bigquery"
	"github.com/hashicorp/go-hclog"
	"github.com/raito-io/cli/base"
	"github.com/raito-io/cli/base/util/config"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

var logger hclog.Logger

func init() {
	logger = base.Logger()
}

func ConnectToBigQuery(configMap *config.ConfigMap, ctx context.Context) (*bigquery.Client, error) {
	gcpProjectId := configMap.GetString(common.GcpProjectId)

	config, err := getConfig(configMap, admin.CloudPlatformScope)

	if err != nil {
		return nil, err
	}

	return bigquery.NewClient(ctx, gcpProjectId, option.WithHTTPClient(config.Client(ctx)))
}

func getRoleForBQEntity(t bigquery.AccessRole) string {
	if t == bigquery.OwnerRole {
		return "roles/bigquery.dataOwner"
	} else if t == bigquery.ReaderRole {
		return "roles/bigquery.dataViewer"
	} else if t == bigquery.WriterRole {
		return "roles/bigquery.dataEditor"
	}

	return string(t)
}

func getBQEntityForRole(t string) bigquery.AccessRole {
	switch t {
	case "roles/bigquery.dataOwner":
		return bigquery.OwnerRole
	case "roles/bigquery.dataEditor":
		return bigquery.WriterRole
	case "roles/bigquery.dataViewer":
		return bigquery.ReaderRole
	}

	return bigquery.AccessRole(t)
}

func GetValueIfExists[T any](p *T, defaultValue T) T {
	if p == nil {
		return defaultValue
	}

	return *p
}
