//go:build tools
// +build tools

package main

import (
	_ "github.com/vektra/mockery/v2"

	_ "github.com/google/wire/cmd/wire"

	_ "github.com/raito-io/enumer"
)
