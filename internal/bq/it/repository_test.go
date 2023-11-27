//go:build integration

package it

import (
	"context"
	"testing"

	bigquery2 "cloud.google.com/go/bigquery"
	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bigquery "github.com/raito-io/cli-plugin-gcp/internal/bq"
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
			Description: "BigQuery project raito-integration-test dataset",
			Parent:      project,
		},
		{
			Id:          "raito-integration-test.public_dataset",
			Name:        "public_dataset",
			FullName:    "raito-integration-test.public_dataset",
			Type:        "dataset",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test dataset",
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
		Location:    "europe-west1",
		Description: "BigQuery project raito-integration-test dataset",
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
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test table",
			Parent:      parent,
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
			Name:        "covid_19_geographic_distribution_worldwide",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide",
			Type:        "table",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test table",
			Parent:      parent,
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_belgium",
			Name:        "covid_19_belgium",
			FullName:    "raito-integration-test.public_dataset.covid_19_belgium",
			Type:        "view",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test view",
			Parent:      parent,
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
		Location:    "europe-west1",
		Description: "BigQuery project raito-integration-test table",
		Parent: &org.GcpOrgEntity{
			Id:          "raito-integration-test.public_dataset",
			Name:        "public_dataset",
			FullName:    "raito-integration-test.public_dataset",
			Type:        "dataset",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test dataset",
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

	assert.ElementsMatch(t, []*org.GcpOrgEntity{
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.date",
			Name:        "date",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.date",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("DATE"),
		}, {
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.day",
			Name:        "day",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.day",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.month",
			Name:        "month",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.month",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.year",
			Name:        "year",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.year",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_confirmed_cases",
			Name:        "daily_confirmed_cases",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_confirmed_cases",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_deaths",
			Name:        "daily_deaths",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.daily_deaths",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.confirmed_cases",
			Name:        "confirmed_cases",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.confirmed_cases",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.deaths",
			Name:        "deaths",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.deaths",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.countries_and_territories",
			Name:        "countries_and_territories",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.countries_and_territories",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("STRING"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.geo_id",
			Name:        "geo_id",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.geo_id",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("STRING"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.country_territory_code",
			Name:        "country_territory_code",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.country_territory_code",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
			Parent:      parent,
			DataType:    ptr.String("STRING"),
		},
		{
			Id:          "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.pop_data_2019",
			Name:        "pop_data_2019",
			FullName:    "raito-integration-test.public_dataset.covid_19_geographic_distribution_worldwide.pop_data_2019",
			Type:        "column",
			Location:    "europe-west1",
			Description: "BigQuery project raito-integration-test column",
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
		Location:    "europe-west1",
		Description: "BigQuery project raito-integration-test dataset",
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
		Location:    "europe-west1",
		Description: "BigQuery project raito-integration-test view",
		Parent:      parent,
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
					Location:    "europe-west1",
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
					Description: "BigQuery project raito-integration-test table",
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
			t.Parallel()

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
			Description: "BigQuery project raito-integration-test table",
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
		ctx            context.Context
		dataObject     *iam.DataObjectReference
		addBindings    []iam.IamBinding
		removeBindings []iam.IamBinding
	}
	tests := []struct {
		name      string
		args      args
		wantError require.ErrorAssertionFunc
	}{
		{
			name: "No bindings to update",
			args: args{
				ctx:            ctx,
				dataObject:     &iam.DataObjectReference{FullName: "raito-integration-test.public_dataset", ObjectType: "dataset"},
				addBindings:    []iam.IamBinding{},
				removeBindings: []iam.IamBinding{},
			},
			wantError: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repository.UpdateBindings(tt.args.ctx, tt.args.dataObject, tt.args.addBindings, tt.args.removeBindings)

			tt.wantError(t, err)

			if err != nil {
				return
			}

			// TODO
		})
	}
}

func createRepository(ctx context.Context, t *testing.T) (*bigquery.Repository, *bigquery2.Client, *config.ConfigMap, func(), error) {
	t.Helper()

	configMap := it.IntegrationTestConfigMap()
	testServices, cleanup, err := InitializeBigqueryRepository(ctx, configMap)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return testServices.Repository, testServices.Client, configMap, cleanup, err
}
