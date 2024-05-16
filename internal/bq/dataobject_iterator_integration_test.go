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
		objectsPerType map[string]int
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
					DataObjectExcludes: []string{"raito-integration-test.MASTER_DATA"},
					DataObjectParent:   "",
				},
				skipColumns: false,
			},
			wants: wants{
				objectsPerType: map[string]int{
					"datasource": 1,
					"dataset":    1,
					"table":      71,
					"view":       1,
					"column":     493,
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
					DataObjectExcludes: []string{"raito-integration-test.MASTER_DATA"},
					DataObjectParent:   "",
				},
				skipColumns: true,
			},
			wants: wants{
				objectsPerType: map[string]int{
					"datasource": 1,
					"dataset":    1,
					"table":      71,
					"view":       1,
					"column":     0,
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
					DataObjectParent:   "raito-integration-test.RAITO_TESTING",
				},
				skipColumns: false,
			},
			wants: wants{
				objectsPerType: map[string]int{
					"datasource": 0,
					"dataset":    0,
					"table":      71,
					"view":       1,
					"column":     493,
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
					DataObjectExcludes: []string{"RAITO_TESTING.dbo_AWBuildVersion", "RAITO_TESTING.dbo_DatabaseLog", "RAITO_TESTING.dbo_ErrorLog"},
				},
				skipColumns: false,
			},
			wants: wants{
				objectsPerType: map[string]int{
					"datasource": 0,
					"dataset":    1,
					"table":      68,
					"view":       1,
					"column":     472,
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

			for objectType, count := range tt.wants.objectsPerType {
				assert.Equalf(t, objectCounts[objectType], count, "Found %d objects of type %q. Expected at least %d", objectCounts[objectType], objectType, count)
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
