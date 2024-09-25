package roles

import (
	"strings"

	ds "github.com/raito-io/cli/base/data_source"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Service string

const (
	ServiceGcp      = Service("GCP")
	ServiceBigQuery = Service("BQ")
)

var TitleCaser = cases.Title(language.English)

// RoleToDisplayName generates a more human readable role name
func RoleToDisplayName(roleName string) string {
	roleName, _ = strings.CutPrefix(roleName, "roles/")
	roleName = strings.ReplaceAll(roleName, ".", " ")
	return TitleCaser.String(roleName)
}

type GcpRole struct {
	Name        string
	DisplayName string
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
