//go:build bigquery

package main

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/raito-io/cli/base"
	"github.com/raito-io/cli/base/info"
	"github.com/raito-io/cli/base/util/plugin"
	"github.com/raito-io/cli/base/wrappers"

	bigquery "github.com/raito-io/cli-plugin-gcp/gcp/bq"
	"github.com/raito-io/cli-plugin-gcp/gcp/common"
)

func main() {
	logger = base.Logger()
	logger.SetLevel(hclog.Debug)

	err := base.RegisterPlugins(
		wrappers.IdentityStoreSync(bigquery.NewIdentityStoreSyncer()),
		wrappers.DataSourceSync(bigquery.NewDataSourceSyncer()),
		wrappers.DataAccessSync(bigquery.NewDataAccessSyncer()),
		wrappers.DataUsageSync(bigquery.NewDataUsageSyncer()), &info.InfoImpl{
			Info: &plugin.PluginInfo{
				Name:    "BigQuery",
				Version: plugin.ParseVersion(version),
				Parameters: []*plugin.ParameterInfo{
					{Name: common.GcpSAFileLocation, Description: "The location of the GCP Service Account Key JSON (if not set GOOGLE_APPLICATION_CREDENTIALS env var is used)", Mandatory: false},
					{Name: common.GcpProjectId, Description: "The ID of the Google Cloud Platform Project", Mandatory: true},
					{Name: common.GsuiteIdentityStoreSync, Description: "If set to true, users and groups are synced from GSuite, if set to false only users and groups from the GCP project IAM scope are retrieved. GSuite requires a service account with domain wide delegation set up", Mandatory: true},
					{Name: common.GsuiteImpersonateSubject, Description: "The Subject email to impersonate when syncing from GSuite", Mandatory: false},
					{Name: common.GsuiteCustomerId, Description: "The Customer ID for the GSuite account", Mandatory: false},
					{Name: bigquery.BqExcludedDatasets, Description: "The optional comma-separated list of datasets that should be skipped.", Mandatory: false},
					{Name: bigquery.BqIncludeHiddenDatasets, Description: "The optional boolean indicating wether the CLI retrieves hidden BQ datasets.", Mandatory: false},
					{Name: bigquery.BqDataUsageWindow, Description: "The maximum number of days of BQ usage data to retrieve. Default and maximum is 90 days. ", Mandatory: false},
					{Name: bigquery.GcpRolesToGroupByIdentity, Description: "The optional comma-separate list of role names. When set, the bindings with these roles will be grouped by identity (user or group) instead of by resource. Note that the resulting Access Controls will not be editable from Raito Cloud. This can be used to lower the amount of imported Access Controls for roles like 'roles/bigquery.dataOwner'.", Mandatory: false},
				},
			},
		})

	if err != nil {
		logger.Error(fmt.Sprintf("error while registering plugins: %s", err.Error()))
	}
}
