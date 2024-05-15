package main

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/iterator"
)

func getDatasetsWithTablesFromGCP(usageConfig *UsageConfig, datasetIds []string) (map[string][]string, error) {
	tablesClient, closeFn, err := connectToBigQuery(context.Background(), usageConfig,nil, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting to bigquery: %w", err)
	}
	defer closeFn()

	logger.Info(fmt.Sprintf("Find tables for datasets: %v", datasetIds))

	datasetTableMap := map[string][]string{}

	for _, datasetId := range datasetIds {

		logger.Debug(fmt.Sprintf("Looking for tables in dataset: %s", datasetId))
		tableIds := tablesClient.Dataset(datasetId).Tables(context.Background())
		tableNames := []string{}

		for {
			t, err := tableIds.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("tables in dataset: %w", err)
			}
			tableNames = append(tableNames, t.TableID)
		}
		datasetTableMap[datasetId] = tableNames

	}
	return datasetTableMap, nil
}

func resolveAccessibleTables(configuredDataObjects []DatasetTables, actualDatasetsWithTables map[string][]string) ([]string, []DatasetTable) {
	queryableTables := []string{}
	queryableTablesStructured := []DatasetTable{}
	for _, doSet := range configuredDataObjects {
		dataset := doSet.Dataset
		for _, table := range doSet.Tables {
			if datasetTables, found := actualDatasetsWithTables[dataset]; found {
				for _, existingTable := range datasetTables {

					addTable := table == "*"
					addTable = addTable || strings.HasSuffix(table, "*") && strings.HasPrefix(existingTable, table[:len(table)-1])
					addTable = addTable || strings.EqualFold(table, existingTable)

					if addTable {
						queryableTables = append(queryableTables, fmt.Sprintf("%s.%s", dataset, existingTable))
						queryableTablesStructured = append(queryableTablesStructured, DatasetTable{dataset, existingTable})
					}
				}
			}
		}
	}
	return queryableTables, queryableTablesStructured
}
