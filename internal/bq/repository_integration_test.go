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
			Id:          "raito-integration-test.RAITO_TESTING",
			Name:        "RAITO_TESTING",
			FullName:    "raito-integration-test.RAITO_TESTING",
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

	dataset := client.Dataset("RAITO_TESTING")
	parent := &org.GcpOrgEntity{
		Id:          "raito-integration-test.RAITO_TESTING",
		Name:        "RAITO_TESTING",
		FullName:    "raito-integration-test.RAITO_TESTING",
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

	assert.Len(t, tables, 72)

	assert.Contains(t, tables, &org.GcpOrgEntity{
		Id:          "raito-integration-test.RAITO_TESTING.HumanResources_Department",
		Name:        "HumanResources_Department",
		FullName:    "raito-integration-test.RAITO_TESTING.HumanResources_Department",
		Type:        "table",
		Location:    "EU",
		Description: "Human resource department table",
		Parent:      parent,
		Tags:        map[string]string{"label1": "value1"},
	})

	assert.Contains(t, tables, &org.GcpOrgEntity{
		Id:          "raito-integration-test.RAITO_TESTING.HumanResources_Employee",
		Name:        "HumanResources_Employee",
		FullName:    "raito-integration-test.RAITO_TESTING.HumanResources_Employee",
		Type:        "table",
		Location:    "EU",
		Description: "",
		Parent:      parent,
	})
}

func TestRepository_ListColumns(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	repository, client, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	dataset := client.Dataset("RAITO_TESTING")
	table := dataset.Table("HumanResources_Department")
	parent := &org.GcpOrgEntity{
		Id:          "raito-integration-test.RAITO_TESTING.HumanResources_Department",
		Name:        "HumanResources_Department",
		FullName:    "raito-integration-test.RAITO_TESTING.HumanResources_Department",
		Type:        "table",
		Location:    "EU",
		Description: "Human resource department table",
		Tags:        map[string]string{"label1": "value1"},
		Parent: &org.GcpOrgEntity{
			Id:          "raito-integration-test.RAITO_TESTING",
			Name:        "public_dataset",
			FullName:    "raito-integration-test.RAITO_TESTING",
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
			Id:          "raito-integration-test.RAITO_TESTING.HumanResources_Department.DepartmentID",
			Name:        "DepartmentID",
			FullName:    "raito-integration-test.RAITO_TESTING.HumanResources_Department.DepartmentID",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("INTEGER"),
		}, {
			Id:          "raito-integration-test.RAITO_TESTING.HumanResources_Department.Name",
			Name:        "Name",
			FullName:    "raito-integration-test.RAITO_TESTING.HumanResources_Department.Name",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("STRING"),
		},
		{
			Id:          "raito-integration-test.RAITO_TESTING.HumanResources_Department.GroupName",
			Name:        "GroupName",
			FullName:    "raito-integration-test.RAITO_TESTING.HumanResources_Department.GroupName",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("STRING"),
		},
		{
			Id:          "raito-integration-test.RAITO_TESTING.HumanResources_Department.ModifiedDate",
			Name:        "ModifiedDate",
			FullName:    "raito-integration-test.RAITO_TESTING.HumanResources_Department.ModifiedDate",
			Type:        "column",
			Location:    "EU",
			Description: "",
			Parent:      parent,
			DataType:    ptr.String("TIMESTAMP"),
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

	dataset := client.Dataset("RAITO_TESTING")
	parent := &org.GcpOrgEntity{
		Id:          "raito-integration-test.RAITO_TESTING",
		Name:        "RAITO_TESTING",
		FullName:    "raito-integration-test.RAITO_TESTING",
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
		Id:          "raito-integration-test.RAITO_TESTING.Sales_Customer_Limited",
		Name:        "Sales_Customer_Limited",
		FullName:    "raito-integration-test.RAITO_TESTING.Sales_Customer_Limited",
		Type:        "view",
		Location:    "EU",
		Description: "",
		Parent:      parent,
		Tags:        map[string]string{"max_store_id": "100"},
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
					Member:       "user:b_stewart@raito.dev",
					Role:         "roles/bigquery.jobUser",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
				{
					Member:       "user:c_harris@raito.dev",
					Role:         "roles/bigquery.jobUser",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
				{
					Member:       "user:d_hayden@raito.dev",
					Role:         "roles/bigquery.jobUser",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
				{
					Member:       "user:m_carissa@raito.dev",
					Role:         "roles/bigquery.jobUser",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
				{
					Member:       "user:n_nguyen@raito.dev",
					Role:         "roles/bigquery.jobUser",
					Resource:     "raito-integration-test",
					ResourceType: "project",
				},
				{
					Member:       "user:n_nguyen@raito.dev",
					Role:         "roles/editor",
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
					Id:          "raito-integration-test.RAITO_TESTING",
					Name:        "RAITO_TESTING",
					FullName:    "raito-integration-test.RAITO_TESTING",
					Type:        data_source.Dataset,
					Location:    "EU",
					Description: "",
				},
			},
			wantBindings: []iam.IamBinding{
				{
					Member:       "special_group:projectWriters",
					Role:         "roles/bigquery.dataEditor",
					Resource:     "raito-integration-test.RAITO_TESTING",
					ResourceType: "dataset",
				},
				{
					Member:       "special_group:projectOwners",
					Role:         "roles/bigquery.dataOwner",
					Resource:     "raito-integration-test.RAITO_TESTING",
					ResourceType: "dataset",
				},
				{
					Member:       "user:d_hayden@raito.dev",
					Role:         "roles/bigquery.dataViewer",
					Resource:     "raito-integration-test.RAITO_TESTING",
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
					Id:          "raito-integration-test.RAITO_TESTING.Sales_CountryRegionCurrency",
					Name:        "Sales_CountryRegionCurrency",
					FullName:    "raito-integration-test.RAITO_TESTING.Sales_CountryRegionCurrency",
					Type:        data_source.Table,
					Location:    "eu",
					Description: "",
				},
			},
			wantBindings: []iam.IamBinding{
				{
					Member:       "user:m_carissa@raito.dev",
					Role:         "roles/bigquery.dataViewer",
					Resource:     "raito-integration-test.RAITO_TESTING.Sales_CountryRegionCurrency",
					ResourceType: "table",
				},
				{
					Member:       "user:d_hayden@raito.dev",
					Role:         "roles/bigquery.dataViewer",
					Resource:     "raito-integration-test.RAITO_TESTING.Sales_CountryRegionCurrency",
					ResourceType: "table",
				},
				{
					Member:       "group:sales@raito.dev",
					Role:         "roles/bigquery.dataViewer",
					Resource:     "raito-integration-test.RAITO_TESTING.Sales_CountryRegionCurrency",
					ResourceType: "table",
				},
				{
					Member:       "group:dev@raito.dev",
					Role:         "roles/bigquery.dataViewer",
					Resource:     "raito-integration-test.RAITO_TESTING.Sales_CountryRegionCurrency",
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
			Id:          "raito-integration-test.RAITO_TESTING.Sales_CountryRegionCurrency.CurrencyCode",
			Name:        "CurrencyCode",
			FullName:    "raito-integration-test.RAITO_TESTING.Sales_CountryRegionCurrency.CurrencyCode",
			Type:        data_source.Column,
			Location:    "eu",
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
					Id:          "raito-integration-test.RAITO_TESTING",
					Name:        "RAITO_TESTING",
					FullName:    "raito-integration-test.RAITO_TESTING",
					Type:        "dataset",
					Location:    "eu",
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
					Id:          "raito-integration-test.RAITO_TESTING",
					Name:        "RAITO_TESTING",
					FullName:    "raito-integration-test.RAITO_TESTING",
					Type:        "dataset",
					Location:    "EU",
					Description: "",
					Parent:      repository.Project(),
				},
				bindings: []iam.IamBinding{
					{
						Member:       "user:m_carissa@raito.dev",
						Role:         roles.RolesBigQueryMetadataViewer.Name,
						ResourceType: "dataset",
						Resource:     "raito-integration-test.RAITO_TESTING",
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
					Id:       "raito-integration-test.RAITO_TESTING.Sales_SalesTerritoryHistory",
					Name:     "Sales_SalesTerritoryHistory",
					FullName: "raito-integration-test.RAITO_TESTING.Sales_SalesTerritoryHistory",
					Type:     "table",
					Location: "EU",
					Parent: &org.GcpOrgEntity{
						Id:          "raito-integration-test.RAITO_TESTING",
						Name:        "RAITO_TESTING",
						FullName:    "raito-integration-test.RAITO_TESTING",
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
						Resource:     "raito-integration-test.RAITO_TESTING.Sales_SalesTerritoryHistory",
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
		Id:          "raito-integration-test.RAITO_TESTING.Person_Address",
		Name:        "Person_Address",
		FullName:    "raito-integration-test.RAITO_TESTING.Person_Address",
		Type:        "table",
		Location:    "EU",
		Description: "",
		Parent: &org.GcpOrgEntity{
			Id:          "raito-integration-test.RAITO_TESTING",
			Name:        "RAITO_TESTING",
			FullName:    "raito-integration-test.RAITO_TESTING",
			Type:        "dataset",
			Location:    "EU",
			Description: "",
			Parent:      repository.Project(),
		},
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
			externaId:      "raito-integration-test.RAITO_TESTING.Person_Address.person_address_group",
			policy:         "StateProvinceID = 0",
			groups:         []string{"dev@raito.dev"},
			internalizable: true,
		},
		{
			externaId:      "raito-integration-test.RAITO_TESTING.Person_Address.person_address_user",
			policy:         "true",
			users:          []string{"n_nguyen@raito.dev"},
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
		Dataset: "RAITO_TESTING",
		Table:   "Person_BusinessEntityAddress",
	}

	filterName := "BusinessEntityID_invalid"

	t.Run("Create filter", func(t *testing.T) {
		err = repository.CreateOrUpdateFilter(ctx, &BQFilter{
			Table:            table,
			Users:            []string{"m_carissa@raito.dev"},
			Groups:           []string{"dev@raito.dev"},
			FilterExpression: "BusinessEntityID < 0",
			FilterName:       filterName,
		})

		require.NoError(t, err)
	})

	t.Run("List created filter", func(t *testing.T) {
		foundFilter := false

		err := repository.ListFilters(ctx, &org.GcpOrgEntity{
			Id:       "raito-integration-test.RAITO_TESTING.Person_BusinessEntityAddress",
			Name:     "Person_BusinessEntityAddress",
			FullName: "raito-integration-test.RAITO_TESTING.Person_BusinessEntityAddress",
			Type:     "table",
			Location: "eu",
			Parent: &org.GcpOrgEntity{
				Id:          "raito-integration-test.RAITO_TESTING",
				Name:        "RAITO_TESTING",
				FullName:    "raito-integration-test.RAITO_TESTING",
				Type:        "dataset",
				Location:    "eu",
				Description: "",
				Parent:      repository.Project(),
			},
		}, func(ctx context.Context, rap *bigquery.RowAccessPolicy, users []string, groups []string, internalizable bool) error {
			if rap.RowAccessPolicyReference.PolicyId == filterName {
				foundFilter = true

				assert.ElementsMatch(t, []string{"m_carissa@raito.dev"}, users)
				assert.ElementsMatch(t, []string{"dev@raito.dev"}, groups)
				assert.Equal(t, "BusinessEntityID < 0", rap.FilterPredicate)
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
			Id:       "raito-integration-test.RAITO_TESTING.Person_BusinessEntityAddress",
			Name:     "Person_BusinessEntityAddress",
			FullName: "raito-integration-test.RAITO_TESTING.Person_BusinessEntityAddress",
			Type:     "table",
			Location: "eu",
			Parent: &org.GcpOrgEntity{
				Id:          "raito-integration-test.RAITO_TESTING",
				Name:        "RAITO_TESTING",
				FullName:    "raito-integration-test.RAITO_TESTING",
				Type:        "dataset",
				Location:    "eu",
				Description: "",
				Parent:      repository.Project(),
			},
		}, func(ctx context.Context, rap *bigquery.RowAccessPolicy, users []string, groups []string, internalizable bool) error {
			if rap.RowAccessPolicyReference.PolicyId == filterName {
				require.Fail(t, "Filter still exists")
			}

			return nil
		})

		require.NoError(t, err)
	})
}

func TestRepository_CreateAndDeleteFilter_noGrantees(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository, _, _, cleanup, err := createRepository(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	table := BQReferencedTable{
		Project: "raito-integration-test",
		Dataset: "RAITO_TESTING",
		Table:   "Person_Person",
	}

	filterName := "Person_Person_BE_1"

	t.Run("Create filter", func(t *testing.T) {
		err = repository.CreateOrUpdateFilter(ctx, &BQFilter{
			Table:            table,
			FilterExpression: "Demographics = \"BEL\"",
			FilterName:       filterName,
		})

		require.NoError(t, err)
	})

	t.Run("List created filter", func(t *testing.T) {
		foundFilter := false

		err := repository.ListFilters(ctx, &org.GcpOrgEntity{
			Id:       "raito-integration-test.RAITO_TESTING.Person_Person",
			Name:     "Person_Person",
			FullName: "raito-integration-test.RAITO_TESTING.Person_Person",
			Type:     "table",
			Location: "EU",
			Parent: &org.GcpOrgEntity{
				Id:          "raito-integration-test.RAITO_TESTING",
				Name:        "RAITO_TESTING",
				FullName:    "raito-integration-test.RAITO_TESTING",
				Type:        "dataset",
				Location:    "EU",
				Description: "",
				Parent:      repository.Project(),
			},
		}, func(ctx context.Context, rap *bigquery.RowAccessPolicy, users []string, groups []string, internalizable bool) error {
			if rap.RowAccessPolicyReference.PolicyId == filterName {
				foundFilter = true

				assert.Empty(t, users)
				assert.Empty(t, groups)
				assert.Equal(t, "Demographics = \"BEL\"", rap.FilterPredicate)
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
			Id:       "raito-integration-test.RAITO_TESTING.Person_Person",
			Name:     "Person_Person",
			FullName: "raito-integration-test.RAITO_TESTING.Person_Person",
			Type:     "table",
			Location: "eu",
			Parent: &org.GcpOrgEntity{
				Id:          "raito-integration-test.RAITO_TESTING",
				Name:        "RAITO_TESTING",
				FullName:    "raito-integration-test.RAITO_TESTING",
				Type:        "dataset",
				Location:    "eu",
				Description: "",
				Parent:      repository.Project(),
			},
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
