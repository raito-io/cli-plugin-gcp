//go:build wireinject
// +build wireinject

package admin

import (
	"github.com/google/wire"
)

var Wired = wire.NewSet(
	NewAdminRepository,

	NewGcpAdminService,
)
