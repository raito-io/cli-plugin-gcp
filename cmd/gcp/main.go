package main

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/raito-io/cli/base"
	"github.com/raito-io/cli/base/access_provider"
	"github.com/raito-io/cli/base/info"
	"github.com/raito-io/cli/base/util/plugin"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/version"
)

func main() {
	logger := base.Logger()
	logger.SetLevel(hclog.Debug)

	err := base.RegisterPlugins(
		wrappers.IdentityStoreSyncFactory(InitializeIdentityStoreSyncer),
		wrappers.DataSourceSyncFactory(InitializeDataSourceSyncer),
		wrappers.DataAccessSyncFactory(InitializeDataAccessSyncer, access_provider.WithSupportPartialSync()),
		&info.InfoImpl{
			Info: &plugin.PluginInfo{
				Name:    "gcp",
				Version: plugin.ParseVersion(version.Version),
				Parameters: []*plugin.ParameterInfo{
					{Name: common.GcpSAFileLocation, Description: "The location of the GCP Service Account Key JSON (if not set GOOGLE_APPLICATION_CREDENTIALS env var is used)", Mandatory: false},
					{Name: common.GcpOrgId, Description: "The ID of the GCP Organization to synchronise", Mandatory: true},
					{Name: common.GsuiteIdentityStoreSync, Description: "If set to true, users and groups are synced from GSuite, if set to false only users and groups from the GCP project IAM scope are retrieved. Gsuite requires a service account with domain wide delegation set up", Mandatory: false},
					{Name: common.GsuiteImpersonateSubject, Description: "The Subject email to impersonate when syncing from GSuite", Mandatory: false},
					{Name: common.GsuiteCustomerId, Description: "The Customer ID for the GSuite account", Mandatory: false},
					{Name: common.GcpRolesToGroupByIdentity, Description: "The optional comma-separate list of role names. When set, the bindings with these roles will be grouped by identity (user or group) instead of by resource. Note that the resulting Access Controls will not be editable from Raito Cloud. This can be used to lower the amount of imported Access Controls for roles like 'roles/owner' and 'roles/bigquery.dataOwner'.", Mandatory: false},
					{Name: common.GcpIncludePaths, Description: "Optional comma-separated list of paths to include. If specified, only these paths will be handled. For example: /folder1/subfolder,/folder2", Mandatory: false},
					{Name: common.GcpExcludePaths, Description: "Optional comma-separated list of paths to exclude. If specified, these paths will not be handled. Excludes have preference over includes. For example: /folder2/subfolder", Mandatory: false},
					{Name: common.GcpServiceAccountsInIdentitySyncEnabled, Description: "Optional flag to enable/disable the retrieving of service accounts during the identity-store sync. By default this will be enabled", Mandatory: false},
				},
				TagSource: common.TagSource,
			},
		})

	if err != nil {
		logger.Error(fmt.Sprintf("error while registering plugins: %s", err.Error()))
	}
}
