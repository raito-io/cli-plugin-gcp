package bigquery

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raito-io/cli/base/access_provider/sync_from_target"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/data_usage"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
)

//go:generate go run github.com/vektra/mockery/v2 --name=dataUsageRepository --with-expecter --inpackage
type dataUsageRepository interface {
	GetDataUsage(ctx context.Context, configMap *config.ConfigMap) ([]BQInformationSchemaEntity, error)
}

type DataUsageSyncer struct {
	repoProvider func() dataUsageRepository
}

func NewDataUsageSyncer() *DataUsageSyncer {
	return &DataUsageSyncer{repoProvider: newDatausageRepo}
}

func newDatausageRepo() dataUsageRepository {
	return &BigQueryRepository{}
}

func (s *DataUsageSyncer) SyncDataUsage(ctx context.Context, fileCreator wrappers.DataUsageStatementHandler, configParams *config.ConfigMap) error {
	dataUsageEntities, err := s.repoProvider().GetDataUsage(ctx, configParams)

	startDateTime, usageFirstUsed, usageLastUsed := GetDataUsageStartDate(ctx, configParams)
	startDate := startDateTime.Unix()

	if err != nil {
		return err
	}

	loggingThreshold := uint64(1 * 1024 * 1024)
	maximumFileSize := uint64(2 * 1024 * 1024 * 1024) // TODO: temporary limit of ~2Gb for debugging

	numSkippedNoCachedQuery := 0
	numSkippedNoResources := 0
	numStatements := 0

	for _, du := range dataUsageEntities {
		// if the query is not a cache hit, and from before the configured start time, ignore it
		if !du.CachedQuery && (du.EndTime < startDate ||
			// if the query is a cache hit, and falls in the configured window, make sure it has not been synced before
			(usageFirstUsed != nil && usageLastUsed != nil && (du.EndTime < usageFirstUsed.Unix() || du.EndTime > usageLastUsed.Unix()))) {
			continue
		}

		if du.Tables == nil || len(du.Tables) == 0 {
			numSkippedNoCachedQuery += 1
			continue
		}

		accessedResources := []sync_from_target.WhatItem{}

		for _, rt := range du.Tables {
			accessedResources = append(accessedResources, sync_from_target.WhatItem{
				DataObject: &data_source.DataObjectReference{
					FullName: fmt.Sprintf("%s.%s.%s", rt.Project, rt.Dataset, rt.Table),
					Type:     data_source.Table,
				},
				Permissions: []string{"SELECT"},
			})
		}

		if len(accessedResources) == 0 {
			numSkippedNoResources += 1
			continue
		}

		err := fileCreator.AddStatements([]data_usage.Statement{
			{
				ExternalId:          uuid.NewString(),
				User:                du.User,
				StartTime:           du.StartTime,
				EndTime:             du.EndTime,
				AccessedDataObjects: accessedResources,
				Success:             true,
			},
		})
		numStatements += 1

		fileSize := fileCreator.GetImportFileSize()
		if fileSize > loggingThreshold {
			logger.Info(fmt.Sprintf("Import file size larger than %d bytes after %d statements => ~%.1f bytes/statement", fileSize, numStatements, float32(fileSize)/float32(numStatements)))
			loggingThreshold = 10 * loggingThreshold
		}

		if fileSize > maximumFileSize {
			logger.Warn(fmt.Sprintf("Current data usage file size larger than %d bytes(%d statements), not adding any more data to import", maximumFileSize, numStatements))
			break
		}

		if err != nil {
			return err
		}
	}

	logger.Info(fmt.Sprintf("%d statements skipped due to no cached query available, %d statements skipped due to no data objects in statement", numSkippedNoCachedQuery, numSkippedNoResources))

	return nil
}
