package gcp

import (
	"context"

	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
)

type DataUsageSyncer struct {
}

func NewDataUsageSyncer() *DataUsageSyncer {
	return &DataUsageSyncer{}
}

func (s *DataUsageSyncer) SyncDataUsage(ctx context.Context, fileCreator wrappers.DataUsageStatementHandler, configParams *config.ConfigMap) error {
	return nil
}
