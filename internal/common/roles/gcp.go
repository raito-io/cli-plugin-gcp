package roles

import ds "github.com/raito-io/cli/base/data_source"

// Basic Roles
// As defined on https://cloud.google.com/iam/docs/understanding-roles#basic

// RolesOwner All Editor permissions, plus permissions for actions like the following:
// - Completing sensitive tasks, like canceling BigQuery jobs
// - Managing roles and permissions for a project and all resources within the project
// - Setting up billing for a project
// No data access is provided by this role
var RolesOwner = GcpRole{
	Name:                   "roles/owner",
	Description:            "Full access to most Google Cloud resources. See the list of included permissions.",
	GlobalPermissions:      map[Service][]string{ServiceGcp: {ds.Admin}},
	UsageGlobalPermissions: map[Service][]string{ServiceGcp: {ds.Read, ds.Write, ds.Admin}, ServiceBigQuery: {ds.Admin}},
}

// RolesEditor All viewer permissions, plus permissions for actions that modify state, such as changing existing resources.
//
// The permissions in the Editor role let you create and delete resources for most Google Cloud services.
// However, the Editor role doesn't contain permissions to perform all actions for all services.
// No data acess is provided by this role
var RolesEditor = GcpRole{
	Name:                   "roles/editor",
	Description:            "View, create, update, and delete most Google Cloud resources. See the list of included permissions.",
	GlobalPermissions:      map[Service][]string{ServiceGcp: {ds.Read, ds.Write}},
	UsageGlobalPermissions: map[Service][]string{ServiceGcp: {ds.Write}},
}

// RolesViewer Permissions for read-only actions that don't affect state, such as viewing (but not modifying) existing resources.
// No data access is provided by this role
var RolesViewer = GcpRole{
	Name:                   "roles/viewer",
	Description:            "View most Google Cloud resources. See the list of included permissions.",
	GlobalPermissions:      map[Service][]string{ServiceGcp: {ds.Read}},
	UsageGlobalPermissions: map[Service][]string{ServiceGcp: {ds.Read}},
}
