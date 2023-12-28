//go:build integration

package bigquery

import (
	"context"
	"testing"

	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/it"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func TestDataObjectIterator_Sync(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	iterator, configMap, cleanup, err := createIterator(ctx, t)
	require.NoError(t, err)

	defer cleanup()

	type args struct {
		ctx         context.Context
		config      *ds.DataSourceSyncConfig
		skipColumns bool
	}
	type wants struct {
		minimalObjectsPerType map[string]int
		maximalObjectsPerType map[string]int
	}
	tests := []struct {
		name    string
		args    args
		wants   wants
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "List all data objects",
			args: args{
				ctx: ctx,
				config: &ds.DataSourceSyncConfig{
					ConfigMap:          configMap,
					DataSourceId:       "datasourceId1",
					DataObjectExcludes: nil,
					DataObjectParent:   "",
				},
				skipColumns: false,
			},
			wants: wants{
				minimalObjectsPerType: map[string]int{
					"datasource": 1,
					"dataset":    2,
					"table":      3,
					"view":       1,
					"column":     700,
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Skip columns",
			args: args{
				ctx: ctx,
				config: &ds.DataSourceSyncConfig{
					ConfigMap:          configMap,
					DataSourceId:       "datasourceId1",
					DataObjectExcludes: nil,
					DataObjectParent:   "",
				},
				skipColumns: true,
			},
			wants: wants{
				minimalObjectsPerType: map[string]int{
					"datasource": 1,
					"dataset":    2,
					"table":      3,
					"view":       1,
				},
				maximalObjectsPerType: map[string]int{
					"column": 0,
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Load single dataset",
			args: args{
				ctx: ctx,
				config: &ds.DataSourceSyncConfig{
					ConfigMap:          configMap,
					DataSourceId:       "datasourceId1",
					DataObjectExcludes: nil,
					DataObjectParent:   "raito-integration-test.public_dataset",
				},
				skipColumns: false,
			},
			wants: wants{
				minimalObjectsPerType: map[string]int{
					"table":  2,
					"view":   1,
					"column": 700,
				},
				maximalObjectsPerType: map[string]int{
					"datasource": 0,
					"dataset":    0,
					"table":      2,
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Exclude some objects",
			args: args{
				ctx: ctx,
				config: &ds.DataSourceSyncConfig{
					ConfigMap:          configMap,
					DataSourceId:       "datasourceId1",
					DataObjectParent:   "raito-integration-test",
					DataObjectExcludes: []string{"private_dataset", "public_dataset.covid_19_geographic_distribution_worldwide"},
				},
				skipColumns: false,
			},
			wants: wants{
				minimalObjectsPerType: map[string]int{
					"dataset": 1,
					"table":   1,
					"view":    1,
					"column":  700,
				},
				maximalObjectsPerType: map[string]int{
					"datasource": 0,
					"dataset":    1,
					"table":      1,
				},
			},
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objectCounts := make(map[string]int)

			err = iterator.Sync(tt.args.ctx, tt.args.config, tt.args.skipColumns, func(ctx context.Context, object *org.GcpOrgEntity) error {
				if _, found := objectCounts[object.Type]; found {
					objectCounts[object.Type]++
				} else {
					objectCounts[object.Type] = 1
				}

				return nil
			})

			tt.wantErr(t, err)

			if err != nil {
				return
			}

			for objectType, count := range tt.wants.minimalObjectsPerType {
				assert.GreaterOrEqualf(t, objectCounts[objectType], count, "Found %d objects of type %q. Expected at least %d", objectCounts[objectType], objectType, count)
			}

			for objectType, count := range tt.wants.maximalObjectsPerType {
				assert.LessOrEqualf(t, objectCounts[objectType], count, "Found %d objects of type %q. Expected at most %d", objectCounts[objectType], objectType, count)
			}
		})
	}
}

func createIterator(ctx context.Context, t *testing.T) (*DataObjectIterator, *config.ConfigMap, func(), error) {
	t.Helper()

	configMap := it.IntegrationTestConfigMap()

	iterator, cleanup, err := InitializeDataObjectIterator(ctx, configMap)
	if err != nil {
		return nil, nil, nil, err
	}

	return iterator, configMap, cleanup, nil
}
