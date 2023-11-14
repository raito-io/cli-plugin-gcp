package bigquery

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"google.golang.org/api/iterator"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

type Repository struct {
	client     *bigquery.Client
	projectId  string
	listHidden bool
}

func NewRepository(client *bigquery.Client, configMap *config.ConfigMap) *Repository {
	return &Repository{
		client:     client,
		projectId:  configMap.GetString(common.GcpProjectId),
		listHidden: configMap.GetBool(BqIncludeHiddenDatasets),
	}
}

func (c *Repository) ListDataSets(ctx context.Context, parent *org.GcpOrgEntity, fn func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery.Dataset) error) error {
	dsIterator := c.client.Datasets(ctx)
	dsIterator.ListHidden = c.listHidden

	for {
		ds, err := dsIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return err
		}

		md, err := ds.Metadata(ctx)
		if err != nil {
			logger.Error(fmt.Sprintf("Error getting metadata for dataset %s: %s", ds.DatasetID, err))
		}

		id := fmt.Sprintf("%s.%s", parent.Id, ds.DatasetID)

		entity := org.GcpOrgEntity{
			Type:        data_source.Dataset,
			Name:        ds.DatasetID,
			Id:          id,
			FullName:    id,
			Description: c.description(data_source.Dataset),
			Parent:      parent,
			Location:    md.Location,
		}

		err = fn(ctx, &entity, ds)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Repository) ListTables(ctx context.Context, ds *bigquery.Dataset, parent *org.GcpOrgEntity, fn func(ctx context.Context, entity *org.GcpOrgEntity, tab *bigquery.Table) error) error {
	tIterator := ds.Tables(ctx)

	for {
		tab, err := tIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return err
		}

		entityType := data_source.Table

		meta, err := tab.Metadata(ctx)
		if err != nil {
			return err
		}

		if meta.Type == bigquery.ViewTable || meta.Type == bigquery.MaterializedView {
			entityType = data_source.View
		}

		id := fmt.Sprintf("%s.%s", parent.Id, tab.TableID)

		entity := org.GcpOrgEntity{
			Type:        entityType,
			Name:        tab.TableID,
			Id:          id,
			FullName:    id,
			Description: c.description(entityType),
			Parent:      parent,
			Location:    meta.Location,
		}

		err = fn(ctx, &entity, tab)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Repository) ListColumns(ctx context.Context, tab *bigquery.Table, parent *org.GcpOrgEntity, fn func(ctx context.Context, entity *org.GcpOrgEntity) error) error {
	tMeta, err := tab.Metadata(ctx)
	if err != nil {
		return err
	}

	for _, col := range tMeta.Schema {
		var policyTags []string
		if col.PolicyTags != nil {
			policyTags = col.PolicyTags.Names
		}

		id := fmt.Sprintf("%s.%s", parent.Id, col.Name)

		entity := org.GcpOrgEntity{
			Type:        "column",
			Name:        col.Name,
			Id:          id,
			FullName:    id,
			Parent:      parent,
			Description: c.description("column"),
			Location:    tMeta.Location,
			PolicyTags:  policyTags,
		}

		err = fn(ctx, &entity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Repository) ListViews(ctx context.Context, ds *bigquery.Dataset, parent *org.GcpOrgEntity, fn func(ctx context.Context, entity *org.GcpOrgEntity) error) error {
	tIterator := ds.Tables(ctx)

	for {
		tab, err := tIterator.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}

		meta, err := tab.Metadata(ctx)
		if err != nil {
			return err
		}

		if meta.Type != bigquery.ViewTable && meta.Type != bigquery.MaterializedView {
			continue
		}

		id := fmt.Sprintf("%s.%s", parent.Id, tab.TableID)

		entity := org.GcpOrgEntity{
			Type:        data_source.View,
			Name:        tab.TableID,
			Id:          id,
			FullName:    id,
			Description: c.description(data_source.View),
			Parent:      parent,
			Location:    meta.Location,
		}

		err = fn(ctx, &entity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Repository) GetBindings(ctx context.Context, entity *org.GcpOrgEntity) ([]types.IamBinding, error) {
	entityIdParts := strings.Split(entity.Id, ".")

	switch entity.Type {
	case data_source.Dataset:
		return c.getDataSetBindings(ctx, entity, entityIdParts)
	case data_source.Table:
		return c.getTableBindings(ctx, entity, entityIdParts)
	}

	return nil, nil
}

func (c *Repository) getDataSetBindings(ctx context.Context, entity *org.GcpOrgEntity, entityIdParts []string) ([]types.IamBinding, error) {
	ds := c.client.Dataset(entityIdParts[1])

	dsMeta, err := ds.Metadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("metadata of dataset %q: %w", entityIdParts[1], err)
	}

	var resultBindings []types.IamBinding

	for _, a := range dsMeta.Access {
		if a.EntityType == bigquery.UserEmailEntity || a.EntityType == bigquery.GroupEmailEntity || a.EntityType == bigquery.SpecialGroupEntity {
			prefix := userPrefix

			if a.EntityType == bigquery.GroupEmailEntity {
				prefix = groupPrefix
			} else if a.EntityType == bigquery.SpecialGroupEntity {
				prefix = specialGroupPrefix
			} else if strings.Contains(a.Entity, "gserviceaccount") {
				prefix = serviceAccountPrefix
			}

			resultBindings = append(resultBindings, types.IamBinding{
				Role:         getRoleForBQEntity(a.Role),
				Member:       prefix + a.Entity,
				Resource:     entity.Id,
				ResourceType: "dataset",
			})
		}
	}

	return resultBindings, nil
}

func (c *Repository) getTableBindings(ctx context.Context, entity *org.GcpOrgEntity, entityIdParts []string) ([]types.IamBinding, error) {
	t := c.client.Dataset(entityIdParts[1]).Table(entityIdParts[2])

	policy, err := t.IAM().Policy(ctx)
	if err != nil {
		return nil, fmt.Errorf("policy of table %q: %w", entity.Id, err)
	}

	var bindings []types.IamBinding

	for _, role := range policy.Roles() {
		for _, m := range policy.Members(role) {
			bindings = append(bindings, types.IamBinding{
				Role:         string(role),
				Member:       m,
				Resource:     entity.Id,
				ResourceType: "table",
			})
		}
	}

	return bindings, nil
}

func (c *Repository) description(doType string) string {
	return fmt.Sprintf("BigQuery project %s %s", c.projectId, doType)
}
