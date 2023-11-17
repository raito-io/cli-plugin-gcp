package roles

import (
	"github.com/gogo/protobuf/proto"
	ds "github.com/raito-io/cli/base/data_source"
)

func SetGlobalPermissions(role *ds.DataObjectTypePermission, globalPermissions ...string) *ds.DataObjectTypePermission {
	newRole := proto.Clone(role).(*ds.DataObjectTypePermission)
	newRole.GlobalPermissions = append(newRole.GlobalPermissions, globalPermissions...)

	return newRole
}
