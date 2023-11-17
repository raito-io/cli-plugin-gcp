package common

import (
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/data_source"
)

// ShouldHandle determines if this data object needs to be handled by the syncer or not. It does this by looking at the configuration options to only sync a part.
func ShouldHandle(fullName string, config *data_source.DataSourceSyncConfig) (ret bool) {
	defer func() {
		if config.DataObjectParent != "" {
			Logger.Debug(fmt.Sprintf("shouldHandle %s: %t", fullName, ret))
		}
	}()

	// No partial sync specified, so do everything
	if config.DataObjectParent == "" {
		return true
	}

	// Check if the data object is under the data object to start from
	if !strings.HasPrefix(fullName, config.DataObjectParent) || config.DataObjectParent == fullName {
		return false
	}

	// Check if we hit any excludes
	for _, exclude := range config.DataObjectExcludes {
		if strings.HasPrefix(fullName, config.DataObjectParent+"."+exclude) {
			return false
		}
	}

	return true
}

// ShouldGoInto checks if we need to go deeper into this data object or not.
func ShouldGoInto(fullName string, config *data_source.DataSourceSyncConfig) (ret bool) {
	defer func() {
		if config.DataObjectParent != "" {
			Logger.Debug(fmt.Sprintf("shouldGoInto %s: %t", fullName, ret))
		}
	}()

	// No partial sync specified, so do everything
	if config.DataObjectParent == "" || strings.HasPrefix(config.DataObjectParent, fullName) || strings.HasPrefix(fullName, config.DataObjectParent) {
		return true
	}

	return false
}
