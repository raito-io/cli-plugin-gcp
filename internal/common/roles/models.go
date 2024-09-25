package roles

import (
	"strings"
	"unicode"

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

func splitAndCapitalize(s string) string {
	var words []string
	var start int

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			word := s[start:i]
			words = append(words, TitleCaser.String(word))
			start = i
		}
	}
	// Add the last word
	word := s[start:]
	words = append(words, TitleCaser.String(word))

	return strings.Join(words, " ")
}

// RoleToDisplayName generates a more human readable role name
func RoleToDisplayName(roleName string) string {
	roleName, _ = strings.CutPrefix(roleName, "roles/")
	roleName = strings.ReplaceAll(roleName, ".", " ")

	return splitAndCapitalize(roleName)
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
