package bigquery

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"google.golang.org/api/iterator"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

type BigQueryRepository struct {
}

func (r *BigQueryRepository) GetDataSets(ctx context.Context, configMap *config.ConfigMap) ([]BQEntity, error) {
	gcpProject := configMap.GetString(common.GcpProjectId)

	conn, err := ConnectToBigQuery(configMap, ctx)

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	dsIterator := conn.Datasets(ctx)
	dsIterator.ListHidden = configMap.GetBool(BqIncludeHiddenDatasets)

	entities := make([]BQEntity, 0)

	for {
		ds, err := dsIterator.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, err
		}

		md, err := ds.Metadata(ctx)
		if err != nil {
			break
		}

		entities = append(entities, BQEntity{
			Type:     data_source.Dataset,
			Name:     ds.DatasetID,
			ID:       ds.DatasetID,
			FullName: fmt.Sprintf("%s.%s", gcpProject, ds.DatasetID),
			ParentId: gcpProject,
			Location: md.Location,
		})
	}

	return entities, nil
}

func (r *BigQueryRepository) GetTables(ctx context.Context, configMap *config.ConfigMap, parent BQEntity) ([]BQEntity, error) {
	gcpProject := configMap.GetString(common.GcpProjectId)
	conn, err := ConnectToBigQuery(configMap, ctx)

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ds := conn.Dataset(parent.ID)
	tIterator := ds.Tables(ctx)

	entities := make([]BQEntity, 0)

	for {
		tab, err := tIterator.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, err
		}

		entityType := data_source.Table

		meta, err := tab.Metadata(ctx)
		if err != nil {
			return nil, err
		}

		if meta.Type == bigquery.ViewTable || meta.Type == bigquery.MaterializedView {
			entityType = data_source.View
		}

		entities = append(entities, BQEntity{
			Type:     entityType,
			Name:     tab.TableID,
			ID:       tab.TableID,
			FullName: fmt.Sprintf("%s.%s.%s", gcpProject, ds.DatasetID, tab.TableID),
			ParentId: fmt.Sprintf("%s.%s", gcpProject, ds.DatasetID),
			Location: meta.Location,
		})

		// add columns to importer
		tMeta, err := tab.Metadata(ctx)

		if err != nil {
			return nil, err
		}

		for _, col := range tMeta.Schema {
			var policyTags []string
			if col.PolicyTags != nil {
				policyTags = col.PolicyTags.Names
			}

			entities = append(entities, BQEntity{
				Type:       "column",
				Name:       col.Name,
				ID:         col.Name,
				FullName:   fmt.Sprintf("%s.%s.%s.%s", gcpProject, ds.DatasetID, tab.TableID, col.Name),
				ParentId:   fmt.Sprintf("%s.%s.%s", gcpProject, ds.DatasetID, tab.TableID),
				Location:   tMeta.Location,
				PolicyTags: policyTags,
			})
		}
	}

	return entities, nil
}

func (r *BigQueryRepository) getViews(ctx context.Context, configMap *config.ConfigMap, parent BQEntity) ([]BQEntity, error) {
	gcpProject := configMap.GetString(common.GcpProjectId)
	conn, err := ConnectToBigQuery(configMap, ctx)

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ds := conn.Dataset(parent.ID)
	tIterator := ds.Tables(ctx)

	entities := make([]BQEntity, 0)

	for {
		tab, err := tIterator.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, err
		}

		meta, err := tab.Metadata(ctx)
		if err != nil {
			return nil, err
		}

		if meta.Type != bigquery.ViewTable && meta.Type != bigquery.MaterializedView {
			continue
		}

		entities = append(entities, BQEntity{
			Type:     data_source.View,
			Name:     tab.TableID,
			ID:       tab.TableID,
			FullName: fmt.Sprintf("%s.%s.%s", gcpProject, ds.DatasetID, tab.TableID),
			ParentId: fmt.Sprintf("%s.%s", gcpProject, ds.DatasetID),
			Location: meta.Location,
		})
	}

	return entities, nil
}

func (r *BigQueryRepository) getAllViews(ctx context.Context, configMap *config.ConfigMap) ([]BQEntity, error) {
	ds, err := r.GetDataSets(ctx, configMap)

	if err != nil {
		return nil, err
	}

	entities := make([]BQEntity, 0)

	for _, d := range ds {
		e, err2 := r.getViews(ctx, configMap, d)

		if err2 != nil {
			return entities, err2
		}

		entities = append(entities, e...)
	}

	return entities, nil
}

func (r *BigQueryRepository) GetDataUsage(ctx context.Context, configMap *config.ConfigMap) ([]BQInformationSchemaEntity, error) {
	conn, err := ConnectToBigQuery(configMap, ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	dsIterator := conn.Datasets(ctx)
	dsIterator.ListHidden = configMap.GetBool(BqIncludeHiddenDatasets)

	regions := make(map[string]interface{}, 0)

	for {
		ds, err2 := dsIterator.Next()
		if err2 == iterator.Done {
			break
		} else if err2 != nil {
			return nil, err2
		}

		md, err2 := ds.Metadata(ctx)

		if err2 != nil {
			return nil, err2
		}

		regions[md.Location] = struct{}{}
	}

	allViews, err := r.getAllViews(ctx, configMap)

	if err != nil {
		return nil, err
	}

	records := make([]BQInformationSchemaEntity, 0)

	for r := range regions {
		logger.Error("querying INFORMATION_SCHEMA in BigQuery region " + r)

		recs, err := getDataUsage(ctx, configMap, strings.ToLower(r))

		if err != nil {
			return nil, err
		}

		// check if query contains a view and if so add the view itself to referenced tables
		for i := range recs {
			for _, view := range allViews {
				if strings.Contains(recs[i].Query, view.FullName) {
					recs[i].Tables = append(recs[i].Tables, BQReferencedTable{
						Project: configMap.GetString(common.GcpProjectId),
						Dataset: strings.Split(view.ParentId, ".")[1],
						Table:   view.Name,
					})

					logger.Debug(fmt.Sprintf("Query %q contains view %q, adding a reference to it for usage", recs[i].Query, view.FullName))

					break
				}
			}
		}

		records = append(records, recs...)
	}

	logger.Debug(fmt.Sprintf("Retrieved %d records from all regions (%s)", len(records), regions))

	return records, nil
}

func getDataUsage(ctx context.Context, configMap *config.ConfigMap, region string) ([]BQInformationSchemaEntity, error) {
	conn, err := ConnectToBigQuery(configMap, ctx)

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	windowStart, usageFirstUsed, usageLastUsed := GetDataUsageStartDate(ctx, configMap)
	if usageFirstUsed != nil && usageLastUsed != nil {
		logger.Info(fmt.Sprintf("Using start date %s, excluding [%s, %s]", windowStart.Format(time.RFC3339), usageFirstUsed.Format(time.RFC3339), usageLastUsed.Format(time.RFC3339)))
	} else {
		logger.Info(fmt.Sprintf("Using start date %s", windowStart.Format(time.RFC3339)))
	}

	timeQueryFragment := fmt.Sprintf(`end_time >= %d`, windowStart.Unix())
	if usageFirstUsed != nil && usageLastUsed != nil {
		timeQueryFragment = fmt.Sprintf(`((end_time >= %[1]d AND end_time < %[2]d) OR end_time > %[3]d)`, windowStart.Unix(), usageFirstUsed.Unix(), usageLastUsed.Unix())
	}

	logger.Info(fmt.Sprintf("time fragment query: %s", timeQueryFragment))

	query := conn.Query(fmt.Sprintf(`
		WITH hits as (
			SELECT
				CASE WHEN cache_hit IS NOT NULL THEN cache_hit ELSE FALSE END AS cache_hit,
				user_email,
				REGEXP_REPLACE(query, r"[lL][iI][mM][iI][tT]\s+\d+.*", "") AS query,
				statement_type,
				referenced_tables,
				UNIX_SECONDS(start_time) AS start_time,
				UNIX_SECONDS(end_time) AS end_time
			FROM
				%[1]s.INFORMATION_SCHEMA.JOBS AS cache_hits
			WHERE
				state = "DONE"
				AND statement_type in ("SELECT", "INSERT", "UPDATE", "DELETE", "MERGE", "TRUNCATE_TABLE")"
				AND NOT CONTAINS_SUBSTR(query,"INFORMATION_SCHEMA")
		), cache_hits as (
			SELECT cache_hit,user_email,query,statement_type,start_time,end_time from hits WHERE %[2]s AND cache_hit
		),non_cache_hits as (
			SELECT * from hits WHERE %[2]s AND NOT cache_hit
		),  query_lookup_distinct as (
			SELECT DISTINCT query,project_id,table_id,dataset_id from hits t, t.referenced_tables WHERE NOT cache_hit
		), query_lookup as (
			SELECT query, ARRAY_AGG(struct(project_id as project_id,dataset_id as dataset_id,table_id as table_id)) as referenced_tables from query_lookup_distinct GROUP by query
		)
		
		SELECT cache_hit,user_email,cache_hits.query,statement_type,referenced_tables,start_time,end_time FROM cache_hits LEFT JOIN query_lookup ON cache_hits.query = query_lookup.query 
		UNION ALL SELECT * FROM non_cache_hits
		ORDER BY
			end_time ASC`, fmt.Sprintf("`region-%s`", region), timeQueryFragment))

	start := time.Now()
	rows, err := query.Read(ctx)

	if err != nil {
		return nil, err
	}

	logger.Debug("BigQuery Query finished, processing results")

	entities := []BQInformationSchemaEntity{}

	maxTime := int64(0)
	minTime := int64(math.MaxInt64)
	minNotCachedTime := int64(math.MaxInt64)

	for {
		var row BQInformationSchemaEntity
		err := rows.Next(&row)

		if len(entities)%100 == 0 {
			logger.Debug(fmt.Sprintf("processing record %d", len(entities)))
		}

		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		if row.StartTime > maxTime {
			maxTime = row.StartTime
		}

		if row.CachedQuery && row.StartTime < minNotCachedTime {
			minNotCachedTime = row.StartTime
		}

		if row.StartTime < minTime {
			minTime = row.StartTime
		}

		entities = append(entities, row)
	}

	logger.Debug(fmt.Sprintf("Retrieved %d records in %.2f s; minimum timestamp: %d (cache min timestamp: %d), maximum: %d ", len(entities), time.Since(start).Seconds(), minNotCachedTime, minTime, maxTime))

	return entities, nil
}

func GetDataUsageStartDate(_ context.Context, configMap *config.ConfigMap) (time.Time, *time.Time, *time.Time) {
	numberOfDays := configMap.GetIntWithDefault(BqDataUsageWindow, 90)
	if numberOfDays > 90 {
		logger.Info(fmt.Sprintf("Capping data usage window to 90 days (from %d days)", numberOfDays))
		numberOfDays = 90
	}

	if numberOfDays <= 0 {
		logger.Info(fmt.Sprintf("Invalid input for data usage window (%d), setting to default 90 days", numberOfDays))
		numberOfDays = 90
	}

	syncStart := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -numberOfDays)

	var earliestTime *time.Time

	if _, found := configMap.Parameters["firstUsed"]; found {
		dateRaw, errLocal := time.Parse(time.RFC3339, configMap.Parameters["firstUsed"])
		logger.Debug(fmt.Sprintf("firstUsed parameter: %s", dateRaw.Format(time.RFC3339)))

		// 12-hour fudge factor; earliest usage data doesn't usually coincide with the start of the window
		if errLocal == nil && dateRaw.Add(-time.Hour*12).After(syncStart) {
			earliestTime = &dateRaw
		}
	}

	var latestTime *time.Time

	if _, found := configMap.Parameters["lastUsed"]; found {
		latestUsageRaw, errLocal := time.Parse(time.RFC3339, configMap.Parameters["lastUsed"])
		logger.Debug(fmt.Sprintf("lastUsed parameter: %s", latestUsageRaw.Format(time.RFC3339)))

		if errLocal == nil && latestUsageRaw.After(syncStart) {
			latestTime = &latestUsageRaw
		}
	}

	if earliestTime == nil && latestTime != nil {
		return *latestTime, nil, nil
	}

	return syncStart, earliestTime, latestTime
}
