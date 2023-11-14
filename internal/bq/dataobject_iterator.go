package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

type DataObjectIterator struct {
	repo      *Repository
	projectId string
}

func NewDataObjectIterator(repo *Repository, configMap *config.ConfigMap) *DataObjectIterator {
	return &DataObjectIterator{
		repo:      repo,
		projectId: configMap.GetString(common.GcpProjectId),
	}
}

func (it *DataObjectIterator) DataObjects(ctx context.Context, fn func(ctx context.Context, object *org.GcpOrgEntity) error) error {
	ds := org.GcpOrgEntity{
		EntryName:   it.projectId,
		Id:          it.projectId,
		Name:        it.projectId,
		FullName:    it.projectId,
		Type:        data_source.Datasource,
		Description: fmt.Sprintf("BigQuery DataSource for GCP project %s", it.projectId),
		Location:    "",
		PolicyTags:  nil,
		Parent:      nil,
	}

	err := fn(ctx, &ds)
	if err != nil {
		return err
	}

	err = it.repo.ListDataSets(ctx, &ds, func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery.Dataset) error {
		err := fn(ctx, entity)
		if err != nil {
			return err
		}

		err = it.repo.ListTables(ctx, dataset, entity, func(ctx context.Context, entity *org.GcpOrgEntity, table *bigquery.Table) error {
			err = fn(ctx, entity)
			if err != nil {
				return err
			}

			err = it.repo.ListColumns(ctx, table, entity, func(ctx context.Context, entity *org.GcpOrgEntity) error {
				err = fn(ctx, entity)
				if err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
