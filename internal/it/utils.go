//go:build integration || syncintegration

package it

import (
	"os"

	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

func IntegrationTestConfigMap() *config.ConfigMap {
	return &config.ConfigMap{Parameters: map[string]string{
		common.GsuiteCustomerId:         os.Getenv("GSUITE_CUSTOMER_ID"),
		common.GsuiteImpersonateSubject: os.Getenv("GSUITE_IMPERSONATE_SUBJECT"),
		common.GcpProjectId:             os.Getenv("GCP_PROJECT_ID"),
		common.GcpSAFileLocation:        os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		common.GcpOrgId:                 os.Getenv("GCP_ORGANIZATION_ID"),
	}}
}
