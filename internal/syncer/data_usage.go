package syncer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/data_usage"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"

	bigquery "github.com/raito-io/cli-plugin-gcp/internal/bq"
	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

//go:generate go run github.com/vektra/mockery/v2 --name=IdGen --with-expecter --inpackage
type IdGen interface {
	New() string
}

//go:generate go run github.com/vektra/mockery/v2 --name=DataUsageRepository --with-expecter --inpackage
type DataUsageRepository interface {
	GetDataUsage(ctx context.Context, windowStart *time.Time, usageFirstUsed *time.Time, usageLastUsed *time.Time, fn func(ctx context.Context, entity *bigquery.BQInformationSchemaEntity) error) error
}

type DataUsageSyncer struct {
	repo        DataUsageRepository
	idGenerator IdGen
	usageWindow int
}

func NewDataUsageSyncer(repo DataUsageRepository, idGen IdGen, configMap *config.ConfigMap) *DataUsageSyncer {
	return &DataUsageSyncer{
		repo:        repo,
		idGenerator: idGen,
		usageWindow: configMap.GetIntWithDefault(common.BqDataUsageWindow, 90),
	}
}

func (s *DataUsageSyncer) SyncDataUsage(ctx context.Context, fileCreator wrappers.DataUsageStatementHandler, configParams *config.ConfigMap) error {
	windowStart, usageFirstUsed, usageLastUsed := s.getDataUsageStartDate(configParams)

	startDate := windowStart.Unix()

	loggingThreshold := uint64(1 * 1024 * 1024)
	maximumFileSize := uint64(2 * 1024 * 1024 * 1024) // TODO: temporary limit of ~2Gb for debugging

	numSkippedNoCachedQuery := 0
	numSkippedNoResources := 0
	numStatements := 0

	err := s.repo.GetDataUsage(ctx, &windowStart, usageFirstUsed, usageLastUsed, func(ctx context.Context, du *bigquery.BQInformationSchemaEntity) error {
		if !du.CachedQuery && (du.EndTime < startDate ||
			// if the query is a cache hit, and falls in the configured window, make sure it has not been synced before
			(usageFirstUsed != nil && usageLastUsed != nil && (du.EndTime < usageFirstUsed.Unix() || du.EndTime > usageLastUsed.Unix()))) {
			return nil
		}

		if du.Tables == nil || len(du.Tables) == 0 {
			numSkippedNoCachedQuery += 1
			return nil
		}

		accessedResources := []data_usage.UsageDataObjectItem{}

		for _, rt := range du.Tables {
			fullnameParts := make([]string, 0, 3)

			if rt.Project.Valid {
				fullnameParts = append(fullnameParts, rt.Project.String())

				if rt.Dataset.Valid {
					fullnameParts = append(fullnameParts, rt.Dataset.String())

					if rt.Table.Valid {
						fullnameParts = append(fullnameParts, rt.Table.String())
					}
				}
			} else {
				continue
			}

			globalPermission, found := bigquery.QueryStatementTypeMap[du.StatementType]
			if !found {
				common.Logger.Warn(fmt.Sprintf("Unknown statement type %s", du.StatementType))

				continue
			}

			accessedResources = append(accessedResources, data_usage.UsageDataObjectItem{
				DataObject: data_usage.UsageDataObjectReference{
					FullName: strings.Join(fullnameParts, "."),
					Type:     data_source.Table,
				},
				GlobalPermission: globalPermission,
			})
		}

		if len(accessedResources) == 0 {
			numSkippedNoResources += 1
			return nil
		}

		err := fileCreator.AddStatements([]data_usage.Statement{
			{
				ExternalId:          s.idGenerator.New(),
				User:                du.User,
				StartTime:           du.StartTime,
				EndTime:             du.EndTime,
				Query:               du.Query,
				AccessedDataObjects: accessedResources,
				Success:             true,
			},
		})

		if err != nil {
			return fmt.Errorf("add statement: %w", err)
		}

		numStatements += 1

		fileSize := fileCreator.GetImportFileSize()
		if fileSize > loggingThreshold {
			common.Logger.Info(fmt.Sprintf("Import file size larger than %d bytes after %d statements => ~%.1f bytes/statement", fileSize, numStatements, float32(fileSize)/float32(numStatements)))
			loggingThreshold = 10 * loggingThreshold
		}

		if fileSize > maximumFileSize {
			common.Logger.Warn(fmt.Sprintf("Current data usage file size larger than %d bytes(%d statements), not adding any more data to import", maximumFileSize, numStatements))
			return nil
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("get data usage: %w", err)
	}

	common.Logger.Info(fmt.Sprintf("%d statements skipped due to no cached query available, %d statements skipped due to no data objects in statement", numSkippedNoCachedQuery, numSkippedNoResources))

	return nil
}

func (s *DataUsageSyncer) getDataUsageStartDate(configMap *config.ConfigMap) (time.Time, *time.Time, *time.Time) {
	numberOfDays := s.usageWindow
	if numberOfDays > 90 {
		common.Logger.Info(fmt.Sprintf("Capping data usage window to 90 days (from %d days)", numberOfDays))
		numberOfDays = 90
	}

	if numberOfDays <= 0 {
		common.Logger.Info(fmt.Sprintf("Invalid input for data usage window (%d), setting to default 90 days", numberOfDays))
		numberOfDays = 90
	}

	syncStart := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -numberOfDays)

	var earliestTime *time.Time

	if _, found := configMap.Parameters["firstUsed"]; found {
		dateRaw, errLocal := time.Parse(time.RFC3339, configMap.Parameters["firstUsed"])
		common.Logger.Debug(fmt.Sprintf("firstUsed parameter: %s", dateRaw.Format(time.RFC3339)))

		// 12-hour fudge factor; earliest usage data doesn't usually coincide with the start of the window
		if errLocal == nil && dateRaw.Add(-time.Hour*12).After(syncStart) {
			earliestTime = &dateRaw
		}
	}

	var latestTime *time.Time

	if _, found := configMap.Parameters["lastUsed"]; found {
		latestUsageRaw, errLocal := time.Parse(time.RFC3339, configMap.Parameters["lastUsed"])
		common.Logger.Debug(fmt.Sprintf("lastUsed parameter: %s", latestUsageRaw.Format(time.RFC3339)))

		if errLocal == nil && latestUsageRaw.After(syncStart) {
			latestTime = &latestUsageRaw
		}
	}

	if earliestTime == nil && latestTime != nil {
		return *latestTime, nil, nil
	}

	return syncStart, earliestTime, latestTime
}
