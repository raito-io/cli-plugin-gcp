package roles

import ds "github.com/raito-io/cli/base/data_source"

type Service string

const (
	ServiceGcp      = Service("GCP")
	ServiceBigQuery = Service("BQ")
)

type GcpRole struct {
	Name        string
	Description string

	GlobalPermissions      map[Service][]string // global permissions per service
	UsageGlobalPermissions map[Service][]string // usage permissions per service
}

func (r *GcpRole) ToDataObjectTypePermission(service Service) *ds.DataObjectTypePermission {
	var globalPermissions []string
	var usageGlobalPermissions []string

	if r.GlobalPermissions != nil {
		globalPermissions = r.GlobalPermissions[service]
	}

	if r.UsageGlobalPermissions != nil {
		usageGlobalPermissions = r.UsageGlobalPermissions[service]
	}

	return &ds.DataObjectTypePermission{
		Permission:  r.Name,
		Description: r.Description,

		GlobalPermissions:      globalPermissions,
		UsageGlobalPermissions: usageGlobalPermissions,
	}
}
