package common

const (
	GcpSAFileLocation                       = "gcp-serviceaccount-json-location"
	GcpOrgId                                = "gcp-organization-id"
	GcpProjectId                            = "gcp-project-id"
	GsuiteIdentityStoreSync                 = "gsuite-identity-store-sync"
	GsuiteImpersonateSubject                = "gsuite-impersonate-subject"
	GsuiteCustomerId                        = "gsuite-customer-id"
	ExcludeNonAplicablePermissions          = "skip-non-applicable-permissions"
	GcpRolesToGroupByIdentity               = "gcp-roles-to-group-by-identity"
	GcpMaskedReader                         = "gcp-masked-reader"
	GcpIncludePaths                         = "gcp-include-paths"
	GcpExcludePaths                         = "gcp-exclude-paths"
	GcpServiceAccountsInIdentitySyncEnabled = "gcp-service-accounts-in-identity-sync-enabled"

	BqExcludedDatasets      = "bq-excluded-datasets"
	BqIncludeHiddenDatasets = "bq-include-hidden-datasets"
	BqDataUsageWindow       = "bq-data-usage-window"
	BqCatalogEnabled        = "bq-catalog-enabled"

	TagSource = "gcp"
)
