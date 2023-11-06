package bigquery

import ds "github.com/raito-io/cli/base/data_source"

// Basic Roles
// As defined on https://cloud.google.com/iam/docs/understanding-roles#basic

// rolesOwner
// All Editor permissions, plus permissions for actions like the following:
// - Completing sensitive tasks, like canceling BigQuery jobs
// - Managing roles and permissions for a project and all resources within the project
// - Setting up billing for a project
// No data access is provided by this role
var rolesOwner = ds.DataObjectTypePermission{
	Permission:             "roles/owner",
	Description:            "Full access to most Google Cloud resources. See the list of included permissions.",
	UsageGlobalPermissions: []string{ds.Admin},
}

// rolesEditor
// All viewer permissions, plus permissions for actions that modify state, such as changing existing resources.
//
// The permissions in the Editor role let you create and delete resources for most Google Cloud services.
// However, the Editor role doesn't contain permissions to perform all actions for all services.
// No data acess is provided by this role
var rolesEditor = ds.DataObjectTypePermission{
	Permission:  "roles/editor",
	Description: "View, create, update, and delete most Google Cloud resources. See the list of included permissions.",
}

// rolesViewer
// Permissions for read-only actions that don't affect state, such as viewing (but not modifying) existing resources.
// No data access is provided by this role
var rolesViewer = ds.DataObjectTypePermission{
	Permission:  "roles/viewer",
	Description: "View most Google Cloud resources. See the list of included permissions.",
}

// BigQuery permissions
// As defined on https://cloud.google.com/bigquery/docs/access-control#bq-permissions

// rolesBigQueryAdmin https://cloud.google.com/bigquery/docs/access-control#bigquery.admin
// Provides permissions to manage all resources within the project. Can manage all data within the project, and can cancel jobs from other users running within the project.
// Lowest-level resources where you can grant this role: Datasets, Row access policies, Tables, Views
var rolesBigQueryAdmin = ds.DataObjectTypePermission{
	Permission:             "roles/bigquery.admin",
	Description:            "Provides permissions to manage all resources within the project. Can manage all data within the project, and can cancel jobs from other users running within the project.",
	GlobalPermissions:      []string{ds.Admin},
	UsageGlobalPermissions: []string{ds.Read, ds.Write, ds.Admin},
}

// rolesBigQueryConnectionAdmin https://cloud.google.com/bigquery/docs/access-control#bigquery.connectionAdmin
var rolesBigQueryConnectionAdmin = ds.DataObjectTypePermission{
	Permission: "roles/bigquery.connectionAdmin",
}

// rolesBigQueryConnectionUser https://cloud.google.com/bigquery/docs/access-control#bigquery.connectionUser
var rolesBigQueryConnectionUser = ds.DataObjectTypePermission{
	Permission: "roles/bigquery.connectionUser",
}

// var rolesBigQueryEditor https://cloud.google.com/bigquery/docs/access-control#bigquery.dataEditor
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
var rolesBigQueryEditor = ds.DataObjectTypePermission{
	Permission:             "roles/bigquery.dataEditor",
	Description:            "When applied to a table or view, this role provides permissions to (1) read and update data and metadata for the table or view and (2) delete the table or view. When applied to a dataset, this role provides permissions to (1) read the dataset's metadata and list tables in the dataset and (2) create, update, get, and delete the dataset's tables. When applied at the project or organization level, this role can also create new datasets ",
	GlobalPermissions:      []string{ds.Write},
	UsageGlobalPermissions: []string{ds.Read, ds.Write, ds.Admin},
}

// rolesBigQueryDataOwner https://cloud.google.com/bigquery/docs/access-control#bigquery.dataOwner
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
var rolesBigQueryDataOwner = ds.DataObjectTypePermission{
	Permission:             "roles/bigquery.dataOwner",
	Description:            "When applied to a table or view, this role provides permissions to (1) read and update data and metadata for the table or view, (2) share the table or view and, (3) delete the table or view. When applied to a dataset, this role provides permissions to (1) read, update, and delete the dataset and (2) create, update, get, and delete the dataset's tables. When applied at the project or organization level, this role can also create new datasets.",
	UsageGlobalPermissions: []string{ds.Read, ds.Write, ds.Admin},
}

// rolesBigQueryDataViewer https://cloud.google.com/bigquery/docs/access-control#bigquery.dataViewer
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
var rolesBigQueryDataViewer = ds.DataObjectTypePermission{
	Permission:             "roles/bigquery.dataViewer",
	Description:            "When applied to a table or view, this role provides permissions to read data and metadata from the table or view. When applied to a dataset, this role provides permissions to (1) read the dataset's metadata and list tables in the dataset and (2) Read data and metadata from the dataset's tables.\n",
	GlobalPermissions:      []string{ds.Read},
	UsageGlobalPermissions: []string{ds.Read},
}

// rolesBigQueryFilteredDataViewer https://cloud.google.com/bigquery/docs/access-control#bigquery.filteredDataViewer
// Access to view filtered table data defined by a row access policy
var rolesBigQueryFilteredDataViewer = ds.DataObjectTypePermission{
	Permission:             "roles/bigquery.filteredDataViewer",
	Description:            "Access to view filtered table data defined by a row access policy",
	UsageGlobalPermissions: []string{ds.Read},
}

// rolesBigQueryJobUser https://cloud.google.com/bigquery/docs/access-control#bigquery.jobUser
// Provides permissions to run jobs, including queries, within the project.
// Lowest-level resources where you can grant this role: Project
var rolesBigQueryJobUser = ds.DataObjectTypePermission{
	Permission:  "roles/bigquery.jobUser",
	Description: "Provides permissions to run jobs, including queries, within the project",
}

// rolesBigQueryMetadataViewer https://cloud.google.com/bigquery/docs/access-control#bigquery.metadataViewer
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
var rolesBigQueryMetadataViewer = ds.DataObjectTypePermission{
	Permission:  "roles/bigquery.metadataViewer",
	Description: "Access to view table and dataset metadata",
}

// rolesBigQueryReadSessionUser https://cloud.google.com/bigquery/docs/access-control#bigquery.readSessionUser
// Provides the ability to create and use read sessions.
//
// Lowest-level resources where you can grant this role: Projects
var rolesBigQueryReadSessionUser = ds.DataObjectTypePermission{
	Permission:  "roles/bigquery.readSessionUser",
	Description: "Provides the ability to create and use read sessions.",
}

// rolesBigQueryResourceAdmin https://cloud.google.com/bigquery/docs/access-control#bigquery.resourceAdmin
// Administer all BigQuery resources.
var rolesBigQueryResourceAdmin = ds.DataObjectTypePermission{
	Permission:  "roles/bigquery.resourceAdmin",
	Description: "Administer all BigQuery resources.",
}

// rolesBigQueryResourceEditor https://cloud.google.com/bigquery/docs/access-control#bigquery.resourceEditor
var rolesBigQueryResourceEditor = ds.DataObjectTypePermission{
	Permission:  "roles/bigquery.resourceEditor",
	Description: "Manage all BigQuery resources, but cannot make purchasing decisions.",
}

// rolesBigQueryResourceViewer https://cloud.google.com/bigquery/docs/access-control#bigquery.resourceViewer
// View all BigQuery resources but cannot make changes or purchasing decisions.
var rolesBigQueryResourceViewer = ds.DataObjectTypePermission{
	Permission:  "roles/bigquery.resourceViewer",
	Description: "View all BigQuery resources but cannot make changes or purchasing decisions.",
}

// rolesBigQueryUser https://cloud.google.com/bigquery/docs/access-control#bigquery.user
// When applied to a dataset, this role provides the ability to read the dataset's metadata and list tables in the dataset.
//
// When applied to a project, this role also provides the ability to run jobs, including queries, within the project.
// A principal with this role can enumerate their own jobs, cancel their own jobs, and enumerate datasets within a project.
// Additionally, allows the creation of new datasets within the project; the creator is granted the BigQuery Data Owner role (roles/bigquery.dataOwner) on these new datasets.
//
// Lowest-level resources where you can grant this role: Datasets
var rolesBigQueryUser = ds.DataObjectTypePermission{
	Permission:  "roles/bigquery.user",
	Description: "When applied to a project, access to run queries, create datasets, read dataset metadata, and list tables. When applied to a dataset, access to read dataset metadata and list tables within the dataset.",
}

// rolesBigQueryMaskedReader https://cloud.google.com/bigquery/docs/access-control#bigquerydatapolicy.maskedReader
// Masked read access to sub-resources tagged by the policy tag associated with a data policy, for example, BigQuery columns
var rolesBigQueryMaskedReader = ds.DataObjectTypePermission{
	Permission:        "roles/bigquerydatapolicy.maskedReader",
	Description:       "Masked read access to sub-resources tagged by the policy tag associated with a data policy, for example, BigQuery columns",
	GlobalPermissions: []string{ds.Read, ds.Write, ds.Admin},
}

// BigQuery Column Level Roles
// As defined on https://cloud.google.com/bigquery/docs/column-level-security-intro#roles

// rolesBigQueryCatalogPolicyTagAdmin https://cloud.google.com/bigquery/docs/column-level-security-intro#roles
// The Data Catalog Policy Tag Admin role is required for users who need to create and manage taxonomies and policy tags.
//
// Applies at the project level.
// This role grants the ability to do the following:
//
//	Create, read, update, and delete taxonomies and policy tags.
//
// Get and set IAM policies on policy tags.
//
// rolesBigQueryCatalogPolicyTagAdmin
var rolesBigQueryCatalogPolicyTagAdmin = ds.DataObjectTypePermission{
	Permission:  "roles/datacatalog.categoryAdmin",
	Description: "The Data Catalog Policy Tag Admin role is required for users who need to create and manage taxonomies and policy tags.",
}

// rolesBigQueryCatalogFineGrainedAccess https://cloud.google.com/bigquery/docs/column-level-security-intro#roles
// The Data Catalog Fine-Grained Reader role is required for users who need access to data in secured columns.
//
// Applies at the policy tag level.
// This role grants the ability to access the content of columns restricted by a policy tag.
var rolesBigQueryCatalogFineGrainedAccess = ds.DataObjectTypePermission{
	Permission:  "roles/datacatalog.categoryFineGrainedReader",
	Description: "Read access to sub-resources tagged by a policy tag, for example, BigQuery columns.",
}
