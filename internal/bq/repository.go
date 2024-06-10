package bigquery

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/iam"
	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/golang-set/set"
	bigquery2 "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/iterator"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	iam2 "github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

const (
	userPrefix           = "user:"
	serviceAccountPrefix = "serviceAccount:"
	groupPrefix          = "group:"
	specialGroupPrefix   = "special_group:"
)

var bqPolicyCache = make(map[string][]iam2.IamBinding)
var bqDataObjectCache = make(map[string][]*org.GcpOrgEntity)

//go:generate go run github.com/vektra/mockery/v2 --name=ProjectClient --with-expecter --inpackage
type ProjectClient interface {
	GetIamPolicy(ctx context.Context, projectId string) ([]iam2.IamBinding, error)
	UpdateBinding(ctx context.Context, dataObject *iam2.DataObjectReference, bindingsToAdd []iam2.IamBinding, bindingsToDelete []iam2.IamBinding) error
}

//go:generate go run github.com/vektra/mockery/v2 --name=BigQueryRowAccessPoliciesService --with-expecter --inpackage
type BigQueryRowAccessPoliciesService interface {
	List(projectId string, datasetId string, tableId string) *bigquery2.RowAccessPoliciesListCall
	GetIamPolicy(resource string, getiampolicyrequest *bigquery2.GetIamPolicyRequest) *bigquery2.RowAccessPoliciesGetIamPolicyCall
}

type RepositoryOptions struct {
	EnableCache bool
}

type Repository struct {
	projectClient   ProjectClient
	client          *bigquery.Client
	rowAccessClient BigQueryRowAccessPoliciesService
	projectId       string
	listHidden      bool

	options *RepositoryOptions
}

func NewRepository(projectClient ProjectClient, client *bigquery.Client, rowAccessClient BigQueryRowAccessPoliciesService, configMap *config.ConfigMap, options *RepositoryOptions) *Repository {
	return &Repository{
		projectClient:   projectClient,
		client:          client,
		rowAccessClient: rowAccessClient,
		projectId:       configMap.GetString(common.GcpProjectId),
		listHidden:      configMap.GetBool(common.BqIncludeHiddenDatasets),

		options: options,
	}
}

func (c *Repository) Project() *org.GcpOrgEntity {
	return &org.GcpOrgEntity{
		EntryName:   c.projectId,
		Id:          c.projectId,
		Name:        c.projectId,
		FullName:    c.projectId,
		Type:        data_source.Datasource,
		Description: "",
		Location:    "",
		PolicyTags:  nil,
		Parent:      nil,
	}
}

func (c *Repository) ListDataSets(ctx context.Context, parent *org.GcpOrgEntity, fn func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery.Dataset) error) error {
	if err, done := c.loadDataObjectsFromCache(ctx, parent, func(ctx context.Context, item *org.GcpOrgEntity) error {
		return fn(ctx, item, nil)
	}); done {
		return err
	}

	dsIterator := c.client.Datasets(ctx)
	dsIterator.ListHidden = c.listHidden

	var dataObjects []*org.GcpOrgEntity

	for {
		ds, err := dsIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return fmt.Errorf("dataset iterator: %w", err)
		}

		meta, err := ds.Metadata(ctx)
		if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching metadata for dataset %q: %s", ds.DatasetID, err))

			continue
		} else if err != nil {
			return fmt.Errorf("getting metadata for dataset %s: %w", ds.DatasetID, err)
		}

		id := fmt.Sprintf("%s.%s", parent.Id, ds.DatasetID)

		entity := org.GcpOrgEntity{
			Type:        data_source.Dataset,
			Name:        ds.DatasetID,
			Id:          id,
			FullName:    id,
			Description: meta.Description,
			Parent:      parent,
			Location:    meta.Location,
			Tags:        meta.Labels,
		}

		err = fn(ctx, &entity, ds)
		if err != nil {
			return err
		}

		dataObjects = append(dataObjects, &entity)
	}

	if c.options.EnableCache {
		bqDataObjectCache[parent.FullName] = dataObjects
	}

	return nil
}

func (c *Repository) ListTables(ctx context.Context, ds *bigquery.Dataset, parent *org.GcpOrgEntity, fn func(ctx context.Context, entity *org.GcpOrgEntity, tab *bigquery.Table) error) error {
	if err, done := c.loadDataObjectsFromCache(ctx, parent, func(ctx context.Context, item *org.GcpOrgEntity) error {
		return fn(ctx, item, nil)
	}); done {
		return err
	}

	if ds == nil {
		ds = c.client.Dataset(parent.Name)

		_, err := ds.Metadata(ctx)
		if err != nil {
			return fmt.Errorf("getting metadata for dataset %s: %w", parent.Id, err)
		}
	}

	tIterator := ds.Tables(ctx)

	var dataObjects []*org.GcpOrgEntity

	for {
		tab, err := tIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching table in dataset %q: %s. Ignore further tables of dataset", ds.DatasetID, err.Error()))

			break // As the iterator will not move on, we break here to avoid an infinite loop
		} else if err != nil {
			return fmt.Errorf("table iterator: %w", err)
		}

		entityType := data_source.Table

		meta, err := tab.Metadata(ctx)

		if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching metadata for table %q: %s", tab.TableID, err.Error()))

			continue
		} else if err != nil {
			return fmt.Errorf("table metadata: %w", err)
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
			Description: meta.Description,
			Parent:      parent,
			Location:    meta.Location,
			Tags:        meta.Labels,
		}

		err = fn(ctx, &entity, tab)
		if err != nil {
			return err
		}

		dataObjects = append(dataObjects, &entity)
	}

	if c.options.EnableCache {
		bqDataObjectCache[parent.FullName] = dataObjects
	}

	return nil
}

func (c *Repository) ListColumns(ctx context.Context, tab *bigquery.Table, parent *org.GcpOrgEntity, fn func(ctx context.Context, entity *org.GcpOrgEntity) error) error {
	if err, done := c.loadDataObjectsFromCache(ctx, parent, fn); done {
		return err
	}

	if tab == nil {
		ds := c.client.Dataset(parent.Parent.Name)

		_, err := ds.Metadata(ctx)
		if err != nil {
			return fmt.Errorf("getting metadata for dataset %s: %w", parent.Id, err)
		}

		tab = ds.Table(parent.Name)
	}

	tMeta, err := tab.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("table metadata: %w", err)
	}

	dataObjects := make([]*org.GcpOrgEntity, 0, len(tMeta.Schema))

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
			Description: col.Description,
			Location:    tMeta.Location,
			PolicyTags:  policyTags,
			DataType:    ptr.String(string(col.Type)),
		}

		err = fn(ctx, &entity)
		if err != nil {
			return err
		}

		dataObjects = append(dataObjects, &entity)
	}

	if c.options.EnableCache {
		bqDataObjectCache[parent.FullName] = dataObjects
	}

	return nil
}

func (c *Repository) ListViews(ctx context.Context, ds *bigquery.Dataset, parent *org.GcpOrgEntity, fn func(ctx context.Context, entity *org.GcpOrgEntity) error) error {
	if err, done := c.loadDataObjectsFromCache(ctx, parent, func(ctx context.Context, item *org.GcpOrgEntity) error {
		if item.Type == data_source.View {
			return fn(ctx, item)
		}

		return nil
	}); done {
		return err
	}

	if ds == nil {
		ds = c.client.Dataset(parent.Id)
	}

	tIterator := ds.Tables(ctx)

	for {
		tab, err := tIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching table in dataset %q: %s. Ignore furher of dataset", ds.DatasetID, err.Error()))

			break // As the iterator will not move on, we break here to avoid an infinite loop
		} else if err != nil {
			return fmt.Errorf("table iterator: %w", err)
		}

		meta, err := tab.Metadata(ctx)
		if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching metadata for table %q: %s", tab.TableID, err.Error()))

			continue
		} else if err != nil {
			return fmt.Errorf("table metadata: %w", err)
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
			Description: meta.Description,
			Parent:      parent,
			Location:    meta.Location,
			Tags:        meta.Labels,
		}

		err = fn(ctx, &entity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Repository) GetBindings(ctx context.Context, entity *org.GcpOrgEntity) ([]iam2.IamBinding, error) {
	if bindings, found := bqPolicyCache[entity.Id]; c.options.EnableCache && found {
		common.Logger.Debug(fmt.Sprintf("Found cached bindings for entity %s", entity.Id))

		return bindings, nil
	}

	common.Logger.Info(fmt.Sprintf("Fetching BigQuery IAM Policy for %s (%s)", entity.Id, entity.Type))

	entityIdParts := strings.Split(entity.Id, ".")

	var bindings []iam2.IamBinding
	var err error

	switch entity.Type {
	case "project", data_source.Datasource:
		bindings, err = c.projectClient.GetIamPolicy(ctx, c.projectId)
	case data_source.Dataset:
		bindings, err = c.getDataSetBindings(ctx, entity, entityIdParts)
	case data_source.Table, data_source.View:
		bindings, err = c.getTableBindings(ctx, entity, entityIdParts)
	case data_source.Column:
		// Do nothing
	default:
		return nil, fmt.Errorf("unsupported entity type %s", entity.Type)
	}

	if common.IsGoogle400Error(err) {
		common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching IAM Policy for %s: %s", entity.FullName, err.Error()))

		return bindings, nil
	} else if err != nil {
		return nil, err
	}

	if c.options.EnableCache {
		bqPolicyCache[entity.Id] = bindings
	}

	return bindings, nil
}

func (c *Repository) UpdateBindings(ctx context.Context, dataObject *iam2.DataObjectReference, addBindings []iam2.IamBinding, removeBindings []iam2.IamBinding) error {
	entityIdParts := strings.Split(dataObject.FullName, ".")

	if len(entityIdParts) == 1 {
		err := c.projectClient.UpdateBinding(ctx, dataObject, addBindings, removeBindings)
		if err != nil {
			return fmt.Errorf("update project bindings for %q: %w", dataObject.FullName, err)
		}
	} else if len(entityIdParts) == 2 {
		err := c.updateDatasetBindings(ctx, entityIdParts[1], addBindings, removeBindings)
		if err != nil {
			return fmt.Errorf("update dataset bindings for %q: %w", dataObject.FullName, err)
		}
	} else if len(entityIdParts) == 3 {
		err := c.updateTableBindings(ctx, entityIdParts[1], entityIdParts[2], addBindings, removeBindings)
		if err != nil {
			return fmt.Errorf("update table bindings for %q: %w", dataObject.FullName, err)
		}
	} else {
		return fmt.Errorf("unknown entity type for %s (%s)", dataObject.FullName, dataObject.ObjectType)
	}

	return nil
}

func (c *Repository) GetDataUsage(ctx context.Context, windowStart *time.Time, usageFirstUsed *time.Time, usageLastUsed *time.Time, fn func(ctx context.Context, entity *BQInformationSchemaEntity) error) error {
	regions := set.NewSet[string]()

	dsIterator := c.client.Datasets(ctx)

	for {
		ds, err := dsIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching dataset: %s", err.Error()))

			continue
		} else if err != nil {
			return fmt.Errorf("dataset iterator: %w", err)
		}

		md, err := ds.Metadata(ctx)
		if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while fetching metadata for dataset %q: %s", ds.DatasetID, err.Error()))

			continue
		} else if err != nil {
			common.Logger.Error(fmt.Sprintf("Error getting metadata for dataset %s: %s", ds.DatasetID, err))

			continue
		}

		if md.Location != "" {
			regions.Add(md.Location)
		}
	}

	allViews, err := c.getAllViews(ctx)
	if err != nil {
		return fmt.Errorf("get all views: %w", err)
	}

	for r := range regions {
		common.Logger.Info(fmt.Sprintf("querying INFORMATION_SCHEMA in BigQuery region %s", r))

		err = c.getDataUsage(ctx, strings.ToLower(r), windowStart, usageFirstUsed, usageLastUsed, allViews, fn)

		if common.IsGoogle400Error(err) {
			common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while querying INFORMATION_SCHEMA in BigQuery region %s: %s", r, err.Error()))

			continue
		} else if err != nil {
			return fmt.Errorf("get data usage: %w", err)
		}
	}

	return nil
}

func (c *Repository) ListFilters(ctx context.Context, table *org.GcpOrgEntity, fn func(ctx context.Context, rap *bigquery2.RowAccessPolicy, users []string, groups []string, internalizable bool) error) error {
	if table.Type != data_source.Table {
		return errors.New("only tables can be filtered")
	}

	err := c.rowAccessClient.List(table.Parent.Parent.Name, table.Parent.Name, table.Name).Pages(ctx, func(response *bigquery2.ListRowAccessPoliciesResponse) error {
		for _, rap := range response.RowAccessPolicies {
			policy, err := c.rowAccessClient.GetIamPolicy(fmt.Sprintf("projects/%s/datasets/%s/tables/%s/rowAccessPolicies/%s", rap.RowAccessPolicyReference.ProjectId, rap.RowAccessPolicyReference.DatasetId, rap.RowAccessPolicyReference.TableId, rap.RowAccessPolicyReference.PolicyId), &bigquery2.GetIamPolicyRequest{}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("get row level iam policy %+v: %w", *rap.RowAccessPolicyReference, err)
			}

			var users []string
			var groups []string
			internalizable := true

			for _, binding := range policy.Bindings {
				if binding.Role == roles.RolesBigQueryFilteredDataViewer.Name {
					for _, member := range binding.Members {
						entity, memberId, err2 := parseMember(member)
						if err2 != nil {
							common.Logger.Warn(fmt.Sprintf("Unknown member type for row level filter %s: %s", rap.RowAccessPolicyReference.PolicyId, member))
							internalizable = false
						}

						switch entity { //nolint:exhaustive
						case bigquery.UserEmailEntity:
							users = append(users, memberId)
						case bigquery.GroupEmailEntity:
							groups = append(groups, memberId)
						}
					}
				}
			}

			err = fn(ctx, rap, users, groups, internalizable)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if common.IsGoogle400Error(err) {
		common.Logger.Warn(fmt.Sprintf("Encountered 4xx error while querying filters on table %s: %s", table.FullName, err.Error()))

		return nil
	} else if err != nil {
		return fmt.Errorf("list row access policies: %w", err)
	}

	return nil
}

func (c *Repository) CreateOrUpdateFilter(ctx context.Context, filter *BQFilter) error {
	granteeList := make([]string, 0, len(filter.Users)+len(filter.Groups))

	for _, u := range filter.Users {
		granteeList = append(granteeList, fmt.Sprintf("'user:%s'", u))
	}

	for _, g := range filter.Groups {
		granteeList = append(granteeList, fmt.Sprintf("'group:%s'", g))
	}

	var grantStatement string

	if len(granteeList) > 0 {
		grantStatement = "GRANT TO (" + strings.Join(granteeList, ", ") + ")"
	}

	queryStr := fmt.Sprintf("CREATE OR REPLACE ROW ACCESS POLICY `%s` ON `%s`.`%s`.`%s` %s FILTER USING (%s);",
		filter.FilterName, filter.Table.Project, filter.Table.Dataset, filter.Table.Table, grantStatement, filter.FilterExpression)
	query := c.client.Query(queryStr)
	common.Logger.Debug(fmt.Sprintf("Executing query: %s", queryStr))

	j, err := query.Run(ctx)
	if err != nil {
		return fmt.Errorf("create row access policy: %w", err)
	}

	status, err := j.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for creation of row access policy: %w", err)
	}

	if status.Err() != nil {
		return fmt.Errorf("create row access policy job: %w", status.Err())
	}

	return nil
}

func (c *Repository) DeleteFilter(ctx context.Context, table *BQReferencedTable, filterName string) error {
	page, err := c.rowAccessClient.List(table.Project, table.Dataset, table.Table).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("list row access policies: %w", err)
	}

	var queryStr string

	if len(page.RowAccessPolicies) > 1 {
		queryStr = fmt.Sprintf("DROP ROW ACCESS POLICY IF EXISTS `%s` ON `%s`.`%s`.`%s`;", filterName, table.Project, table.Dataset, table.Table)
	} else if len(page.RowAccessPolicies) == 1 && page.RowAccessPolicies[0].RowAccessPolicyReference.PolicyId == filterName {
		queryStr = fmt.Sprintf("DROP ALL ROW ACCESS POLICIES ON `%s`.`%s`.`%s`;", table.Project, table.Dataset, table.Table)
	} else {
		return nil
	}

	query := c.client.Query(queryStr)
	common.Logger.Debug(fmt.Sprintf("Executing query: %s", queryStr))

	j, err := query.Run(ctx)
	if err != nil {
		return fmt.Errorf("delete row access policy: %w", err)
	}

	status, err := j.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for deletion of row access policy: %w", err)
	}

	if status.Err() != nil {
		return fmt.Errorf("delete row access policy job: %w", status.Err())
	}

	return nil
}

func (c *Repository) getAllViews(ctx context.Context) ([]org.GcpOrgEntity, error) {
	allViews := make([]org.GcpOrgEntity, 0)

	err := c.ListDataSets(ctx, c.Project(), func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery.Dataset) error {
		return c.ListViews(ctx, dataset, entity, func(ctx context.Context, entity *org.GcpOrgEntity) error {
			allViews = append(allViews, *entity)

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return allViews, nil
}

func (c *Repository) getDataUsage(ctx context.Context, region string, windowStart *time.Time, usageFirstUsed *time.Time, usageLastUsed *time.Time, allViews []org.GcpOrgEntity, fn func(ctx context.Context, entity *BQInformationSchemaEntity) error) error {
	if usageFirstUsed != nil && usageLastUsed != nil {
		common.Logger.Info(fmt.Sprintf("Using start date %s, excluding [%s, %s]", windowStart.Format(time.RFC3339), usageFirstUsed.Format(time.RFC3339), usageLastUsed.Format(time.RFC3339)))
	} else {
		common.Logger.Info(fmt.Sprintf("Using start date %s", windowStart.Format(time.RFC3339)))
	}

	timeQueryFragment := fmt.Sprintf(`end_time >= %d`, windowStart.Unix())
	if usageFirstUsed != nil && usageLastUsed != nil {
		timeQueryFragment = fmt.Sprintf(`((end_time >= %[1]d AND end_time < %[2]d) OR end_time > %[3]d)`, windowStart.Unix(), usageFirstUsed.Unix(), usageLastUsed.Unix())
	}

	common.Logger.Info(fmt.Sprintf("time fragment query: %s", timeQueryFragment))

	query := c.client.Query(fmt.Sprintf(`
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
				AND statement_type in ("SELECT", "INSERT", "UPDATE", "DELETE", "MERGE", "TRUNCATE_TABLE")
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
		return fmt.Errorf("execute query: %w", err)
	}

	common.Logger.Debug("BigQuery Query finished, processing results")

	maxTime := int64(0)
	minTime := int64(math.MaxInt64)
	minNotCachedTime := int64(math.MaxInt64)

	i := 0

	for {
		var row BQInformationSchemaEntity
		err := rows.Next(&row)

		if i%100 == 0 {
			common.Logger.Debug(fmt.Sprintf("processing record %d", i))
		}

		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			common.Logger.Warn(fmt.Sprintf("Failed to load usage row: %s", err.Error()))

			continue
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

		for viewIdx := range allViews {
			if strings.Contains(row.Query, allViews[viewIdx].FullName) {
				row.Tables = append(row.Tables, BQInformationSchemaReferencedTable{
					Project: bigquery.NullString{StringVal: c.projectId, Valid: true},
					Dataset: bigquery.NullString{StringVal: strings.Split(allViews[viewIdx].Parent.Id, ".")[1], Valid: true},
					Table:   bigquery.NullString{StringVal: allViews[viewIdx].Name, Valid: true},
				})

				common.Logger.Debug(fmt.Sprintf("Query %q contains view %q, adding a reference to it for usage", row.Query, allViews[viewIdx].FullName))

				break
			}
		}

		err = fn(ctx, &row)
		if err != nil {
			return err
		}

		i += 1
	}

	common.Logger.Debug(fmt.Sprintf("Retrieved %d records in %.2f s; minimum timestamp: %d (cache min timestamp: %d), maximum: %d ", i, time.Since(start).Seconds(), minNotCachedTime, minTime, maxTime))

	return nil
}

func (c *Repository) getDataSetBindings(ctx context.Context, entity *org.GcpOrgEntity, entityIdParts []string) ([]iam2.IamBinding, error) {
	ds := c.client.Dataset(entityIdParts[1])

	dsMeta, err := ds.Metadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("metadata of dataset %q: %w", entityIdParts[1], err)
	}

	var resultBindings []iam2.IamBinding

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

			resultBindings = append(resultBindings, iam2.IamBinding{
				Role:         getRoleForBQEntity(a.Role),
				Member:       prefix + a.Entity,
				Resource:     entity.Id,
				ResourceType: "dataset",
			})
		}
	}

	return resultBindings, nil
}

func (c *Repository) updateDatasetBindings(ctx context.Context, dataset string, bindingsToAdd []iam2.IamBinding, bindingsToRemove []iam2.IamBinding) error {
	ds := c.client.Dataset(dataset)

	dsMeta, err := ds.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("metadata of dataset %q: %w", dataset, err)
	}

	update, err := mergeBindings(dsMeta.Access, bindingsToAdd, bindingsToRemove)
	if err != nil {
		return err
	}

	_, err = ds.Update(ctx, *update, dsMeta.ETag)
	if err != nil {
		return fmt.Errorf("update dataset %q: %w", dataset, err)
	}

	return nil
}

func mergeBindings(existingAccess []*bigquery.AccessEntry, bindingsToAdd []iam2.IamBinding, bindingsToRemove []iam2.IamBinding) (*bigquery.DatasetMetadataToUpdate, error) {
	update := bigquery.DatasetMetadataToUpdate{
		Access: []*bigquery.AccessEntry{},
	}

	bindingsToRemoveMap := make(map[string]set.Set[string]) //Role -> Members
	for i := range bindingsToRemove {
		if _, found := bindingsToRemoveMap[bindingsToRemove[i].Role]; !found {
			bindingsToRemoveMap[bindingsToRemove[i].Role] = set.NewSet[string]()
		}

		bindingsToRemoveMap[bindingsToRemove[i].Role].Add(bindingsToRemove[i].Member)
	}

	// Remove old bindings
	for _, a := range existingAccess {
		memberId := fmt.Sprintf("%s:%s", entityToString(a.EntityType), a.Entity)
		role := getRoleForBQEntity(a.Role)

		if membersEntities, found := bindingsToRemoveMap[role]; found {
			if !membersEntities.Contains(memberId) {
				update.Access = append(update.Access, a)
			}
		} else {
			update.Access = append(update.Access, a)
		}
	}

	// Add new bindings
	for i := range bindingsToAdd {
		memberEntityType, memberEntityId, err2 := parseMember(bindingsToAdd[i].Member)
		if err2 != nil {
			return nil, fmt.Errorf("parse member %q: %w", bindingsToAdd[i].Member, err2)
		}

		update.Access = append(update.Access, &bigquery.AccessEntry{
			Role:       getBQEntityForRole(bindingsToAdd[i].Role),
			EntityType: memberEntityType,
			Entity:     memberEntityId,
		})
	}

	return &update, nil
}

func (c *Repository) getTableBindings(ctx context.Context, entity *org.GcpOrgEntity, entityIdParts []string) ([]iam2.IamBinding, error) {
	t := c.client.Dataset(entityIdParts[1]).Table(entityIdParts[2])

	policy, err := t.IAM().Policy(ctx)
	if err != nil {
		return nil, fmt.Errorf("policy of table %q: %w", entity.Id, err)
	}

	var bindings []iam2.IamBinding

	for _, role := range policy.Roles() {
		for _, m := range policy.Members(role) {
			bindings = append(bindings, iam2.IamBinding{
				Role:         string(role),
				Member:       m,
				Resource:     entity.Id,
				ResourceType: entity.Type,
			})
		}
	}

	return bindings, nil
}

func (c *Repository) updateTableBindings(ctx context.Context, dataset, table string, bindingsToAdd []iam2.IamBinding, bindingsToRemove []iam2.IamBinding) error {
	t := c.client.Dataset(dataset).Table(table)

	policy, err := t.IAM().Policy(ctx)
	if err != nil {
		return fmt.Errorf("policy of table '%s.%s': %w", dataset, table, err)
	}

	for i := range bindingsToRemove {
		b := &bindingsToRemove[i]
		policy.Remove(b.Member, iam.RoleName(b.Role))
	}

	for i := range bindingsToAdd {
		b := &bindingsToAdd[i]
		policy.Add(b.Member, iam.RoleName(b.Role))
	}

	err = t.IAM().SetPolicy(ctx, policy)
	if err != nil {
		return fmt.Errorf("set policy of '%s.%s': %w", dataset, table, err)
	}

	return nil
}

func (c *Repository) loadDataObjectsFromCache(ctx context.Context, parent *org.GcpOrgEntity, fn func(ctx context.Context, item *org.GcpOrgEntity) error) (error, bool) {
	if !c.options.EnableCache {
		return nil, false
	}

	if result, found := bqDataObjectCache[parent.FullName]; found {
		for _, entity := range result {
			err := fn(ctx, entity)
			if err != nil {
				return err, true
			}
		}

		return nil, true
	}

	return nil, false
}

func getBQEntityForRole(t string) bigquery.AccessRole {
	switch t {
	case "roles/bigquery.dataOwner":
		return bigquery.OwnerRole
	case "roles/bigquery.dataEditor":
		return bigquery.WriterRole
	case "roles/bigquery.dataViewer":
		return bigquery.ReaderRole
	}

	return bigquery.AccessRole(t)
}

func parseMember(m string) (bigquery.EntityType, string, error) {
	parts := strings.Split(m, ":")
	if len(parts) != 2 {
		return bigquery.UserEmailEntity, "", fmt.Errorf("invalid member format: %s", m)
	}

	switch parts[0] {
	case "user":
		return bigquery.UserEmailEntity, parts[1], nil
	case "group":
		return bigquery.GroupEmailEntity, parts[1], nil
	case "serviceAccount":
		return bigquery.UserEmailEntity, parts[1], nil
	}

	return bigquery.UserEmailEntity, "", fmt.Errorf("unknown member type: %s", m)
}

func entityToString(entity bigquery.EntityType) string {
	switch entity {
	case bigquery.DomainEntity:
		return "domain"
	case bigquery.GroupEmailEntity:
		return "group"
	case bigquery.UserEmailEntity:
		return "user"
	case bigquery.SpecialGroupEntity:
		return "special_group"
	default:
		return "other"
	}
}
