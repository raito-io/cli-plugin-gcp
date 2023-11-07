package roles

import ds "github.com/raito-io/cli/base/data_source"

// BigQuery permissions
// As defined on https://cloud.google.com/bigquery/docs/access-control#bq-permissions

// RolesBigQueryAdmin https://cloud.google.com/bigquery/docs/access-control#bigquery.admin
// Provides permissions to manage all resources within the project. Can manage all data within the project, and can cancel jobs from other users running within the project.
// Lowest-level resources where you can grant this role: Datasets, Row access policies, Tables, Views
var RolesBigQueryAdmin = GcpRole{
	Name:                   "roles/bigquery.admin",
	Description:            "Provides permissions to manage all resources within the project. Can manage all data within the project, and can cancel jobs from other users running within the project.",
	GlobalPermissions:      map[Service][]string{ServiceBigQuery: {ds.Admin}},
	UsageGlobalPermissions: map[Service][]string{ServiceBigQuery: {ds.Read, ds.Write, ds.Admin}},
}

// RolesBigQueryConnectionAdmin https://cloud.google.com/bigquery/docs/access-control#bigquery.connectionAdmin
var RolesBigQueryConnectionAdmin = GcpRole{
	Name: "roles/bigquery.connectionAdmin",
}

// RolesBigQueryConnectionUser https://cloud.google.com/bigquery/docs/access-control#bigquery.connectionUser
var RolesBigQueryConnectionUser = GcpRole{
	Name: "roles/bigquery.connectionUser",
}

// RolesBigQueryEditor https://cloud.google.com/bigquery/docs/access-control#bigquery.dataEditor
// When applied to a table or view, this role provides permissions to:
//   - Read and update data and metadata for the table or view.
//   - Delete the table or view.
//
// This role cannot be applied to individual models or routines.
//
// When applied to a dataset, this role provides permissions to:
//   - Read the dataset's metadata and list tables in the dataset.
//   - Create, update, get, and delete the dataset's tables.
//
// When applied at the project or organization level, this role can also create new datasets.
//
// Lowest-level resources where you can grant this role: Tables, Views
var RolesBigQueryEditor = GcpRole{
	Name:        "roles/bigquery.dataEditor",
	Description: "When applied to a table or view, this role provides permissions to (1) read and update data and metadata for the table or view and (2) delete the table or view. When applied to a dataset, this role provides permissions to (1) read the dataset's metadata and list tables in the dataset and (2) create, update, get, and delete the dataset's tables. When applied at the project or organization level, this role can also create new datasets ",

	GlobalPermissions:      map[Service][]string{ServiceBigQuery: {ds.Write}},
	UsageGlobalPermissions: map[Service][]string{ServiceBigQuery: {ds.Read, ds.Write, ds.Admin}},
}

// RolesBigQueryDataOwner https://cloud.google.com/bigquery/docs/access-control#bigquery.dataOwner
// When applied to a table or view, this role provides permissions to:
//   - Read and update data and metadata for the table or view.
//   - Share the table or view.
//   - Delete the table or view.
//
// This role cannot be applied to individual models or routines.
//
// When applied to a dataset, this role provides permissions to:
//   - Read, update, and delete the dataset.
//   - Create, update, get, and delete the dataset's tables.
//
// When applied at the project or organization level, this role can also create new datasets.
//
// Lowest-level resources where you can grant this role: Tables, Views
var RolesBigQueryDataOwner = GcpRole{
	Name:                   "roles/bigquery.dataOwner",
	Description:            "When applied to a table or view, this role provides permissions to (1) read and update data and metadata for the table or view, (2) share the table or view and, (3) delete the table or view. When applied to a dataset, this role provides permissions to (1) read, update, and delete the dataset and (2) create, update, get, and delete the dataset's tables. When applied at the project or organization level, this role can also create new datasets.",
	UsageGlobalPermissions: map[Service][]string{ServiceBigQuery: {ds.Read, ds.Write, ds.Admin}},
}

// RolesBigQueryDataViewer https://cloud.google.com/bigquery/docs/access-control#bigquery.dataViewer
// When applied to a table or view, this role provides permissions to:
//   - Read data and metadata from the table or view.
//
// This role cannot be applied to individual models or routines.
//
// When applied to a dataset, this role provides permissions to:
//   - Read the dataset's metadata and list tables in the dataset.
//   - Read data and metadata from the dataset's tables.
//
// When applied at the project or organization level, this role can also enumerate all datasets in the project. Additional roles, however, are necessary to allow the running of jobs.
//
// Lowest-level resources where you can grant this role: Tables, Views
var RolesBigQueryDataViewer = GcpRole{
	Name:                   "roles/bigquery.dataViewer",
	Description:            "When applied to a table or view, this role provides permissions to read data and metadata from the table or view. When applied to a dataset, this role provides permissions to (1) read the dataset's metadata and list tables in the dataset and (2) Read data and metadata from the dataset's tables.\n",
	GlobalPermissions:      map[Service][]string{ServiceBigQuery: {ds.Read}},
	UsageGlobalPermissions: map[Service][]string{ServiceBigQuery: {ds.Read}},
}

// RolesBigQueryFilteredDataViewer https://cloud.google.com/bigquery/docs/access-control#bigquery.filteredDataViewer
// Access to view filtered table data defined by a row access policy
var RolesBigQueryFilteredDataViewer = GcpRole{
	Name:                   "roles/bigquery.filteredDataViewer",
	Description:            "Access to view filtered table data defined by a row access policy",
	UsageGlobalPermissions: map[Service][]string{ServiceBigQuery: {ds.Read}},
}

// RolesBigQueryJobUser https://cloud.google.com/bigquery/docs/access-control#bigquery.jobUser
// Provides permissions to run jobs, including queries, within the project.
// Lowest-level resources where you can grant this role: Project
var RolesBigQueryJobUser = GcpRole{
	Name:        "roles/bigquery.jobUser",
	Description: "Provides permissions to run jobs, including queries, within the project",
}

// RolesBigQueryMetadataViewer https://cloud.google.com/bigquery/docs/access-control#bigquery.metadataViewer
// When applied to a table or view, this role provides permissions to:
//   - Read metadata from the table or view.
//
// This role cannot be applied to individual models or routines.
//
// When applied to a dataset, this role provides permissions to:
//   - List tables and views in the dataset.
//   - Read metadata from the dataset's tables and views.
//
// When applied at the project or organization level, this role provides permissions to:
//   - List all datasets and read metadata for all datasets in the project.
//   - List all tables and views and read metadata for all tables and views in the project.
//
// Additional roles are necessary to allow the running of jobs.
//
// Lowest-level resources where you can grant this role: Tables, Views
var RolesBigQueryMetadataViewer = GcpRole{
	Name:        "roles/bigquery.metadataViewer",
	Description: "Access to view table and dataset metadata",
}

// RolesBigQueryReadSessionUser https://cloud.google.com/bigquery/docs/access-control#bigquery.readSessionUser
// Provides the ability to create and use read sessions.
//
// Lowest-level resources where you can grant this role: Projects
var RolesBigQueryReadSessionUser = GcpRole{
	Name:        "roles/bigquery.readSessionUser",
	Description: "Provides the ability to create and use read sessions.",
}

// RolesBigQueryResourceAdmin https://cloud.google.com/bigquery/docs/access-control#bigquery.resourceAdmin
// Administer all BigQuery resources.
var RolesBigQueryResourceAdmin = GcpRole{
	Name:        "roles/bigquery.resourceAdmin",
	Description: "Administer all BigQuery resources.",
}

// RolesBigQueryResourceEditor https://cloud.google.com/bigquery/docs/access-control#bigquery.resourceEditor
var RolesBigQueryResourceEditor = GcpRole{
	Name:        "roles/bigquery.resourceEditor",
	Description: "Manage all BigQuery resources, but cannot make purchasing decisions.",
}

// RolesBigQueryResourceViewer https://cloud.google.com/bigquery/docs/access-control#bigquery.resourceViewer
// View all BigQuery resources but cannot make changes or purchasing decisions.
var RolesBigQueryResourceViewer = GcpRole{
	Name:        "roles/bigquery.resourceViewer",
	Description: "View all BigQuery resources but cannot make changes or purchasing decisions.",
}

// RolesBigQueryUser https://cloud.google.com/bigquery/docs/access-control#bigquery.user
// When applied to a dataset, this role provides the ability to read the dataset's metadata and list tables in the dataset.
//
// When applied to a project, this role also provides the ability to run jobs, including queries, within the project.
// A principal with this role can enumerate their own jobs, cancel their own jobs, and enumerate datasets within a project.
// Additionally, allows the creation of new datasets within the project; the creator is granted the BigQuery Data Owner role (roles/bigquery.dataOwner) on these new datasets.
//
// Lowest-level resources where you can grant this role: Datasets
var RolesBigQueryUser = GcpRole{
	Name:        "roles/bigquery.user",
	Description: "When applied to a project, access to run queries, create datasets, read dataset metadata, and list tables. When applied to a dataset, access to read dataset metadata and list tables within the dataset.",
}

// RolesBigQueryMaskedReader https://cloud.google.com/bigquery/docs/access-control#bigquerydatapolicy.maskedReader
// Masked read access to sub-resources tagged by the policy tag associated with a data policy, for example, BigQuery columns
var RolesBigQueryMaskedReader = GcpRole{
	Name:              "roles/bigquerydatapolicy.maskedReader",
	Description:       "Masked read access to sub-resources tagged by the policy tag associated with a data policy, for example, BigQuery columns",
	GlobalPermissions: map[Service][]string{ServiceBigQuery: {ds.Read, ds.Write, ds.Admin}},
}

// BigQuery Column Level Roles
// As defined on https://cloud.google.com/bigquery/docs/column-level-security-intro#roles

// RolesBigQueryCatalogPolicyTagAdmin https://cloud.google.com/bigquery/docs/column-level-security-intro#roles
// The Data Catalog Policy Tag Admin role is required for users who need to create and manage taxonomies and policy tags.
//
// Applies at the project level.
// This role grants the ability to do the following:
//
//	Create, read, update, and delete taxonomies and policy tags.
//
// Get and set IAM policies on policy tags.
var RolesBigQueryCatalogPolicyTagAdmin = GcpRole{
	Name:        "roles/datacatalog.categoryAdmin",
	Description: "The Data Catalog Policy Tag Admin role is required for users who need to create and manage taxonomies and policy tags.",
}

// RolesBigQueryCatalogFineGrainedAccess https://cloud.google.com/bigquery/docs/column-level-security-intro#roles
// The Data Catalog Fine-Grained Reader role is required for users who need access to data in secured columns.
//
// Applies at the policy tag level.
// This role grants the ability to access the content of columns restricted by a policy tag.
var RolesBigQueryCatalogFineGrainedAccess = GcpRole{
	Name:        "roles/datacatalog.categoryFineGrainedReader",
	Description: "Read access to sub-resources tagged by a policy tag, for example, BigQuery columns.",
}
