//go:build wireinject
// +build wireinject

package syncer

import (
	"github.com/google/wire"
)

var Wired = wire.NewSet(
	NewDataSourceSyncer,
)
