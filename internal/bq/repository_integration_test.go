//go:build integration

package bigquery

import (
	"context"
	"fmt"
	"testing"
	"time"

	bigquery2 "cloud.google.com/go/bigquery"
	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/bigquery/v2"

	"github.com/raito-io/cli-plugin-gcp/internal/common/roles"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/it"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func TestRepository_ListDataSets(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	repository, _, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	var datasets []*org.GcpOrgEntity
	project := repository.Project()

	// When
	err = repository.ListDataSets(ctx, project, func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery2.Dataset) error {
		assert.Equal(t, entity.Name, dataset.DatasetID)
		assert.Equal(t, project.Id, dataset.ProjectID)

		datasets = append(datasets, entity)

		return nil
	})

	// Then
	require.NoError(t, err)

	assert.ElementsMatch(t, []*org.GcpOrgEntity{
		{
			Id:          "raito-integration-test.private_dataset",
			Name:        "private_dataset",
			FullName:    "raito-integration-test.private_dataset",
			Type:        "dataset",
			Location:    "EU",
			Description: "",
			Parent:      project,
		},
		{
			Id:          "raito-integration-test.public_dataset",
			Name:        "public_dataset",
			FullName:    "raito-integration-test.public_dataset",
			Type:        "dataset",
			Location:    "EU",
			Description: "",
			Parent:      project,
		},
	}, datasets)
}

func TestRepository_ListTables(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	repository, client, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	dataset := client.Dataset("public_dataset")
	parent := &org.GcpOrgEntity{
		Id:          "raito-integration-test.public_dataset",
		Name:        "public_dataset",
		FullName:    "raito-integration-test.public_dataset",
		Type:        "dataset",
		Location:    "EU",
		Description: "",
		Parent:      repository.Project(),
	}

	var tables []*org.GcpOrgEntity

	// When
	err = repository.ListTables(ctx, dataset, parent, func(ctx context.Context, entity *org.GcpOrgEntity, tab *bigquery2.Table) error {
		assert.Equal(t, entity.Name, tab.TableID)
		assert.Equal(t, parent.Name, tab.DatasetID)

		tables = append(tables, entity)

		return nil
	})

	// Then
	require.NoError(t, err)

	assert.ElementsMatch(t, []*org.GcpOrgEntity{
		{
			Id:          "raito-integration-test.public_dataset.covid19_open_data",
			Name:        "covid19_open_data",
			FullName:    "raito-integration-test.public_dataset.covid19_open_data",
			Type:        "table",
			Location:    "EU",
			Description: "This dataset contains country-level datasets of daily time-series data related to COVID-19 globally. You can find the list of sources available here: https://github.com/open-covid-19/data",
			Parent:      parent,
			Tags:        map[string]string{"freebqcovid": ""},
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
			Name:        "covid_19_geographic_distribution_worldwide",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
			Type:        "table",
			Location:    "EU",
			Description: "",
			Parent:      parent,
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_belgium",
			Name:        "covid_19_belgium",
			FullName:    "raito-integration-test.public_dataset.covid_19_belgium",
			Type:        "view",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			Tags:        map[string]string{"country": "belgium"},
		},
	}, tables)
}

func TestRepository_ListColumns(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	repository, client, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	dataset := client.Dataset("public_dataset")
	table := dataset.Table("covid_19_geographic_distribution_worldwide")
	parent := &org.GcpOrgEntity{
		Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
		Name:        "covid_19_geographic_distribution_worldwide",
		FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
		Type:        "table",
		Location:    "EU",
		Description: "",
		Parent: &org.GcpOrgEntity{
			Id:          "raito-integration-test.public_dataset",
			Name:        "public_dataset",
			FullName:    "raito-integration-test.public_dataset",
			Type:        "dataset",
			Location:    "EU",
			Description: "",
			Parent:      repository.Project(),
		},
	}

	var columns []*org.GcpOrgEntity

	// When
	err = repository.ListColumns(ctx, table, parent, func(ctx context.Context, entity *org.GcpOrgEntity) error {
		columns = append(columns, entity)

		return nil
	})

	// Then
	require.NoError(t, err)

	// Ignore policyTags
	for i := range columns {
		columns[i].PolicyTags = nil
	}

	assert.ElementsMatch(t, []*org.GcpOrgEntity{
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.date",
			Name:        "date",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.date",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("DATE"),
		}, {
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.day",
			Name:        "day",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.day",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.month",
			Name:        "month",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.month",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.year",
			Name:        "year",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.year",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_confirmed_cases",
			Name:        "daily_confirmed_cases",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_confirmed_cases",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_deaths",
			Name:        "daily_deaths",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_deaths",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.confirmed_cases",
			Name:        "confirmed_cases",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.confirmed_cases",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.deaths",
			Name:        "deaths",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.deaths",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.countries_and_territories",
			Name:        "countries_and_territories",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.countries_and_territories",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("STRING"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.geo_id",
			Name:        "geo_id",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.geo_id",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("STRING"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.country_territory_code",
			Name:        "country_territory_code",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.country_territory_code",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("STRING"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.pop_data_2019",
			Name:        "pop_data_2019",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.pop_data_2019",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
	}, columns)
}

func TestRepository_ListViews(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()
	repository, client, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	dataset := client.Dataset("public_dataset")
	parent := &org.GcpOrgEntity{
		Id:          "raito-integration-test.public_dataset",
		Name:        "public_dataset",
		FullName:    "raito-integration-test.public_dataset",
		Type:        "dataset",
		Location:    "EU",
		Description: "",
		Parent:      repository.Project(),
	}

	var views []*org.GcpOrgEntity

	// When
	err = repository.ListViews(ctx, dataset, parent, func(ctx context.Context, entity *org.GcpOrgEntity) error {
		views = append(views, entity)
		return nil
	})

	// Then
	require.NoError(t, err)

	assert.ElementsMatch(t, []*org.GcpOrgEntity{{
		Id:          "raito-integration-test.public_dataset.covid_19_belgium",
		Name:        "covid_19_belgium",
		FullName:    "raito-integration-test.public_dataset.covid_19_belgium",
		Type:        "view",
		Location:    "EU",
		Description: "",
		Parent:      parent,
		Tags:        map[string]string{"country": "belgium"},
	}}, views)
}

func TestRepository_GetBindings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository, _, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	type args struct {
		ctx    context.Context
		entity *org.GcpOrgEntity
	}
	tests := []struct {
		name         string
		args         args
		wantBindings []iam.IamBinding
		wantErr      require.ErrorAssertionFunc
	}{
		{
			name: "datasource bindings",
			args: args{
				ctx: ctx,
				entity: &org.GcpOrgEntity{
					Id:          "raito-integration-test",
					Name:        "raito-integration-test",
					FullName:    "raito-integration-test",
					Type:        data_source.Datasource,
					Location:    "europe-west1",
					Description: "BigQuery project raito-integration-test",
				},
			},
			wantBindings: []iam.IamBinding{
				{
					Member:       "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
					Role:         "organizations/905493414429/roles/RaitoGcpRole",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
				{
					Member:       "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
					Role:         "organizations/905493414429/roles/RaitoGcpRoleMasking",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
				{
					Member:       "serviceAccount:service-account-for-raito-cli@raito-integration-test.iam.gserviceaccount.com",
					Role:         "roles/bigquery.admin",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
				{
					Member:       "serviceAccount:service-204677507107@gcp-sa-bigquerydatatransfer.iam.gserviceaccount.com",
					Role:         "roles/bigquerydatatransfer.serviceAgent",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				}, {
					Member:       "user:dieter@raito.dev",
					Role:         "roles/owner",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
				{
					Member:       "user:ruben@raito.dev",
					Role:         "roles/owner",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "dataset bindings",
			args: args{
				ctx: ctx,
				entity: &org.GcpOrgEntity{
					Id:          "raito-integration-test.public_dataset",
					Name:        "public_dataset",
					FullName:    "raito-integration-test.public_dataset",
					Type:        data_source.Dataset,
					Location:    "EU",
					Description: "BigQuery project raito-integration-test",
				},
			},
			wantBindings: []iam.IamBinding{
				{
					Member:       "special_group:projectWriters",
					Role:         "roles/bigquery.dataEditor",
					Resource:     "raito-integration-test.public_dataset",
					ResourceType: "dataset",
				},
				{
					Member:       "special_group:projectOwners",
					Role:         "roles/bigquery.dataOwner",
					Resource:     "raito-integration-test.public_dataset",
					ResourceType: "dataset",
				},
				{
					Member:       "special_group:projectReaders",
					Role:         "roles/bigquery.dataViewer",
					Resource:     "raito-integration-test.public_dataset",
					ResourceType: "dataset",
				},
				{
					Member:       "user:d_hayden@raito.dev",
					Role:         "roles/bigquery.dataOwner",
					Resource:     "raito-integration-test.public_dataset",
					ResourceType: "dataset",
				},
				{
					Member:       "user:ruben@raito.dev",
					Role:         "roles/bigquery.dataOwner",
					Resource:     "raito-integration-test.public_dataset",
					ResourceType: "dataset",
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "table bindings",
			args: args{
				ctx: ctx,
				entity: &org.GcpOrgEntity{
					Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
					Name:        "covid_19_geographic_distribution_worldwide",
					FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
					Type:        data_source.Table,
					Location:    "europe-west1",
					Description: "",
				},
			},
			wantBindings: []iam.IamBinding{
				{
					Member:       "user:m_carissa@raito.dev",
					Role:         "roles/bigquery.dataViewer",
					Resource:     "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
					ResourceType: "table",
				},
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repository.GetBindings(tt.args.ctx, tt.args.entity)

			tt.wantErr(t, err)

			if err != nil {
				return
			}

			for _, binding := range tt.wantBindings {
				assert.Contains(t, result, binding)
			}
		})
	}

	t.Run("column bindings should be empty", func(t *testing.T) {
		t.Parallel()

		result, err := repository.GetBindings(ctx, &org.GcpOrgEntity{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.deaths",
			Name:        "deaths",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.deaths",
			Type:        data_source.Column,
			Location:    "europe-west1",
			Description: "",
		})

		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestRepository_UpdateBindings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository, _, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	type args struct {
		ctx        context.Context
		dataObject *org.GcpOrgEntity
		bindings   []iam.IamBinding
	}
	tests := []struct {
		name      string
		args      args
		wantError require.ErrorAssertionFunc
	}{
		{
			name: "No bindings to update",
			args: args{
				ctx: ctx,
				dataObject: &org.GcpOrgEntity{
					Id:          "raito-integration-test.public_dataset",
					Name:        "public_dataset",
					FullName:    "raito-integration-test.public_dataset",
					Type:        "dataset",
					Location:    "europe-west1",
					Description: "",
					Parent:      repository.Project(),
				},
				bindings: []iam.IamBinding{},
			},
			wantError: require.NoError,
		},
		// Bindings for project are repository tests in project gcp package
		{
			name: "Update dataset bindings",
			args: args{
				ctx: ctx,
				dataObject: &org.GcpOrgEntity{
					Id:          "raito-integration-test.private_dataset",
					Name:        "private_dataset",
					FullName:    "raito-integration-test.private_dataset",
					Type:        "dataset",
					Location:    "EU",
					Description: "",
					Parent:      repository.Project(),
				},
				bindings: []iam.IamBinding{
					{
						Member:       "user:m_carissa@raito.dev",
						Role:         roles.RolesBigQueryDataViewer.Name,
						ResourceType: "dataset",
						Resource:     "raito-integration-test.private_dataset",
					},
				},
			},
			wantError: require.NoError,
		},
		{
			name: "Update table bindings",
			args: args{
				ctx: ctx,
				dataObject: &org.GcpOrgEntity{
					Id:       "raito-integration-test.private_dataset.private_table",
					Name:     "private_table",
					FullName: "raito-integration-test.private_dataset.private_table",
					Type:     "table",
					Location: "EU",
					Parent: &org.GcpOrgEntity{
						Id:          "raito-integration-test.private_dataset",
						Name:        "private_dataset",
						FullName:    "raito-integration-test.private_dataset",
						Type:        "dataset",
						Location:    "EU",
						Description: "",
						Parent:      repository.Project(),
					},
				},
				bindings: []iam.IamBinding{
					{
						Member:       "user:m_carissa@raito.dev",
						Role:         roles.RolesBigQueryEditor.Name,
						ResourceType: "table",
						Resource:     "raito-integration-test.private_dataset.private_table",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalBindings, err := repository.GetBindings(tt.args.ctx, tt.args.dataObject)
			require.NoError(t, err)

			dataobject := iam.DataObjectReference{FullName: tt.args.dataObject.FullName, ObjectType: tt.args.dataObject.Type}

			t.Run("Add bindings", func(t *testing.T) {
				// When
				err = repository.UpdateBindings(ctx, &dataobject, tt.args.bindings, nil)

				// Then
				require.NoError(t, err)

				updatedBindings, err := repository.GetBindings(tt.args.ctx, tt.args.dataObject)
				require.NoError(t, err)

				assert.GreaterOrEqual(t, len(updatedBindings), len(originalBindings))

				for _, binding := range tt.args.bindings {
					assert.Contains(t, updatedBindings, binding)
				}

				for _, binding := range originalBindings {
					assert.Contains(t, updatedBindings, binding)
				}

				originalBindings = updatedBindings
			})

			t.Run("Remove bindings", func(t *testing.T) {
				// When
				err = repository.UpdateBindings(ctx, &dataobject, nil, tt.args.bindings)

				//Then
				require.NoError(t, err)

				updatedBindings, err := repository.GetBindings(tt.args.ctx, tt.args.dataObject)
				require.NoError(t, err)

				assert.Equal(t, len(updatedBindings), len(originalBindings)-len(tt.args.bindings))

				for _, binding := range tt.args.bindings {
					assert.NotContains(t, updatedBindings, binding)
				}
			})
		})
	}
}

func TestRepository_GetDataUsage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository, _, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	var dataUsage []*BQInformationSchemaEntity

	err = repository.GetDataUsage(ctx, ptr.Time(time.Now().Add(-14*24*time.Hour)), nil, nil, func(ctx context.Context, entity *BQInformationSchemaEntity) error {
		dataUsage = append(dataUsage, entity)

		return nil
	})

	require.NoError(t, err)
	assert.NotEmpty(t, dataUsage)
}

func TestRepository_ListFilters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository, _, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	type filter struct {
		externaId      string
		policy         string
		users          []string
		groups         []string
		internalizable bool
	}

	filters := []filter{}

	err = repository.ListFilters(ctx, &org.GcpOrgEntity{
		Id:          "raito-integration-test.public_dataset.covid19_open_data",
		Name:        "covid19_open_data",
		FullName:    "raito-integration-test.public_dataset.covid19_open_data",
		Type:        "table",
		Location:    "EU",
		Description: "This dataset contains country-level datasets of daily time-series data related to COVID-19 globally. You can find the list of sources available here: https://github.com/open-covid-19/data",
		Parent: &org.GcpOrgEntity{
			Id:          "raito-integration-test.public_dataset",
			Name:        "public_dataset",
			FullName:    "raito-integration-test.public_dataset",
			Type:        "dataset",
			Location:    "EU",
			Description: "",
			Parent:      repository.Project(),
		},
		Tags: map[string]string{"freebqcovid": ""},
	}, func(ctx context.Context, rap *bigquery.RowAccessPolicy, users []string, groups []string, internalizable bool) error {
		filters = append(filters, filter{
			externaId:      fmt.Sprintf("%s.%s.%s.%s", rap.RowAccessPolicyReference.ProjectId, rap.RowAccessPolicyReference.DatasetId, rap.RowAccessPolicyReference.TableId, rap.RowAccessPolicyReference.PolicyId),
			policy:         rap.FilterPredicate,
			users:          users,
			groups:         groups,
			internalizable: internalizable,
		})

		return nil
	})

	require.NoError(t, err)

	assert.ElementsMatch(t, filters, []filter{
		{
			externaId:      "raito-integration-test.public_dataset.covid19_open_data.covid_all",
			policy:         "true",
			users:          []string{"ruben@raito.dev"},
			internalizable: true,
		},
		{
			externaId:      "raito-integration-test.public_dataset.covid19_open_data.covid_us",
			policy:         "country_code = \"US\"",
			users:          []string{"d_hayden@raito.dev"},
			groups:         []string{"dev@raito.dev"},
			internalizable: true,
		},
	})
}

func TestRepository_CreateAndDeleteFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository, _, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	table := BQReferencedTable{
		Project: "raito-integration-test",
		Dataset: "public_dataset",
		Table:   "covid_19_geographic_distribution_worldwide",
	}

	filterName := "covid_bel"

	t.Run("Create filter", func(t *testing.T) {
		err = repository.CreateOrUpdateFilter(ctx, &BQFilter{
			Table:            table,
			Users:            []string{"m_carissa@raito.dev"},
			Groups:           []string{"dev@raito.dev"},
			FilterExpression: "country_territory_code = \"BEL\"",
			FilterName:       filterName,
		})

		require.NoError(t, err)
	})

	t.Run("List created filter", func(t *testing.T) {
		foundFilter := false

		err := repository.ListFilters(ctx, &org.GcpOrgEntity{
			Id:       "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
			Name:     "covid_19_geographic_distribution_worldwide",
			FullName: "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
			Type:     "table",
			Location: "europe-west1",
			Parent: &org.GcpOrgEntity{
				Id:          "raito-integration-test.public_dataset",
				Name:        "public_dataset",
				FullName:    "raito-integration-test.public_dataset",
				Type:        "dataset",
				Location:    "europe-west1",
				Description: "",
				Parent:      repository.Project(),
			},
			Tags: map[string]string{"freebqcovid": ""},
		}, func(ctx context.Context, rap *bigquery.RowAccessPolicy, users []string, groups []string, internalizable bool) error {
			if rap.RowAccessPolicyReference.PolicyId == filterName {
				foundFilter = true

				assert.ElementsMatch(t, []string{"m_carissa@raito.dev"}, users)
				assert.ElementsMatch(t, []string{"dev@raito.dev"}, groups)
				assert.Equal(t, "country_territory_code = \"BEL\"", rap.FilterPredicate)
			}

			return nil
		})

		require.NoError(t, err)
		assert.True(t, foundFilter)
	})

	t.Run("Delete filter", func(t *testing.T) {
		err = repository.DeleteFilter(ctx, &table, filterName)

		require.NoError(t, err)
	})

	t.Run("Check if filter is deleted", func(t *testing.T) {
		err := repository.ListFilters(ctx, &org.GcpOrgEntity{
			Id:       "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
			Name:     "covid_19_geographic_distribution_worldwide",
			FullName: "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
			Type:     "table",
			Location: "europe-west1",
			Parent: &org.GcpOrgEntity{
				Id:          "raito-integration-test.public_dataset",
				Name:        "public_dataset",
				FullName:    "raito-integration-test.public_dataset",
				Type:        "dataset",
				Location:    "europe-west1",
				Description: "",
				Parent:      repository.Project(),
			},
			Tags: map[string]string{"freebqcovid": ""},
		}, func(ctx context.Context, rap *bigquery.RowAccessPolicy, users []string, groups []string, internalizable bool) error {
			if rap.RowAccessPolicyReference.PolicyId == filterName {
				require.Fail(t, "Filter still exists")
			}

			return nil
		})

		require.NoError(t, err)
	})
}

func createRepository(ctx context.Context, t *testing.T) (*Repository, *bigquery2.Client, *config.ConfigMap, func(), error) {
	t.Helper()

	configMap := it.IntegrationTestConfigMap()
	testServices, cleanup, err := InitializeBigqueryRepository(ctx, configMap)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return testServices.Repository, testServices.Client, configMap, cleanup, err
}
