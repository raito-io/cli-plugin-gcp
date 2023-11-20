package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
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

func (it *DataObjectIterator) DataObjects(ctx context.Context, config *ds.DataSourceSyncConfig, fn func(ctx context.Context, object *org.GcpOrgEntity) error) error {
	return it.sync(ctx, config, false, fn)
}

func (it *DataObjectIterator) sync(ctx context.Context, config *ds.DataSourceSyncConfig, skipColumns bool, fn func(ctx context.Context, object *org.GcpOrgEntity) error) error {
	ds := it.repo.Project()

	if common.ShouldHandle(ds.FullName, config) {
		err := fn(ctx, ds)
		if err != nil {
			return err
		}
	}

	if !common.ShouldGoInto(ds.FullName, config) {
		return nil
	}

	err := it.repo.ListDataSets(ctx, ds, func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery.Dataset) error {
		if common.ShouldHandle(entity.FullName, config) {
			err2 := fn(ctx, entity)
			if err2 != nil {
				return err2
			}
		}

		if !common.ShouldGoInto(entity.FullName, config) {
			return nil
		}

		err2 := it.repo.ListTables(ctx, dataset, entity, func(ctx context.Context, entity *org.GcpOrgEntity, table *bigquery.Table) error {
			if common.ShouldHandle(entity.FullName, config) {
				err2 := fn(ctx, entity)
				if err2 != nil {
					return err2
				}
			}

			if skipColumns || !common.ShouldGoInto(entity.FullName, config) {
				return nil
			}

			err2 := it.repo.ListColumns(ctx, table, entity, func(ctx context.Context, entity *org.GcpOrgEntity) error {
				if common.ShouldHandle(entity.FullName, config) {
					err2 := fn(ctx, entity)
					if err2 != nil {
						return err2
					}
				}

				return nil
			})

			return err2
		})

		return err2
	})

	return err
}

func (it *DataObjectIterator) Bindings(ctx context.Context, config *ds.DataSourceSyncConfig, fn func(ctx context.Context, dataObject *org.GcpOrgEntity, bindings []iam.IamBinding) error) error {
	return it.sync(ctx, config, true, func(ctx context.Context, object *org.GcpOrgEntity) error {
		bindings, err := it.repo.GetBindings(ctx, object)
		if err != nil {
			return fmt.Errorf("get bq bindings: %w", err)
		}

		err = fn(ctx, object, bindings)
		if err != nil {
			return err
		}

		return nil
	})
}

func (it *DataObjectIterator) UpdateBindings(ctx context.Context, dataObject *iam.DataObjectReference, addBindings []iam.IamBinding, removeBindings []iam.IamBinding) error {
	return it.repo.UpdateBindings(ctx, dataObject, addBindings, removeBindings)
}

func (it *DataObjectIterator) DataSourceType() string {
	return "project"
}
