package syncer

import (
	"context"
	"testing"
	"time"

	bigquery2 "cloud.google.com/go/bigquery"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/data_usage"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	bigquery "github.com/raito-io/cli-plugin-gcp/internal/bq"
)

func TestDataUsageSyncer_SyncDataUsage(t *testing.T) {
	startTime := time.Now().Add(-1 * time.Minute).Unix()
	endTime := time.Now().Unix()

	type fields struct {
		mockSetup   func(dataUsageRepoMock *MockDataUsageRepository, idGen *MockIdGen, statementHandler *mocks.DataUsageStatementHandler)
		usageWindow int
	}
	type args struct {
		ctx          context.Context
		configParams *config.ConfigMap
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		expect  []data_usage.Statement
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Single statements",
			fields: fields{
				mockSetup: func(dataUsageRepoMock *MockDataUsageRepository, idGen *MockIdGen, statementHandler *mocks.DataUsageStatementHandler) {
					idGen.EXPECT().New().Return("id123")

					statementHandler.EXPECT().GetImportFileSize().Return(uint64(1024))

					dataUsageRepoMock.EXPECT().GetDataUsage(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, _ *time.Time, _ *time.Time, _ *time.Time, f func(context.Context, *bigquery.BQInformationSchemaEntity) error) error {
						err := f(ctx, &bigquery.BQInformationSchemaEntity{
							Tables: []bigquery.BQInformationSchemaReferencedTable{
								{
									Project: bigquery2.NullString{StringVal: "project1", Valid: true},
									Table:   bigquery2.NullString{StringVal: "table1", Valid: true},
									Dataset: bigquery2.NullString{StringVal: "dataset1", Valid: true},
								},
							},
							Query:         "SELECT * FROM `project1`.`dataset1`.`table1`",
							User:          "ruben@raito.io",
							StartTime:     startTime,
							EndTime:       endTime,
							CachedQuery:   false,
							StatementType: "SELECT",
						})

						return err
					})
				},
				usageWindow: 90,
			},
			args: args{
				ctx:          context.Background(),
				configParams: &config.ConfigMap{},
			},
			expect: []data_usage.Statement{
				{
					ExternalId: "id123",
					User:       "ruben@raito.io",
					StartTime:  startTime,
					EndTime:    endTime,
					AccessedDataObjects: []data_usage.UsageDataObjectItem{
						{
							DataObject: data_usage.UsageDataObjectReference{
								FullName: "project1.dataset1.table1",
								Type:     data_source.Table,
							},
							GlobalPermission: data_usage.Read,
						},
					},
					Success: true,
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statementHandler := mocks.NewSimpleDataUsageStatementHandler(t)
			s, repo, idGen := createDataUsageSyncer(t)
			tt.fields.mockSetup(repo, idGen, statementHandler.DataUsageStatementHandler)
			s.usageWindow = tt.fields.usageWindow

			tt.wantErr(t, s.SyncDataUsage(tt.args.ctx, statementHandler, tt.args.configParams))

			assert.ElementsMatch(t, tt.expect, statementHandler.Statements)
		})
	}
}

func createDataUsageSyncer(t *testing.T) (*DataUsageSyncer, *MockDataUsageRepository, *MockIdGen) {
	repo := NewMockDataUsageRepository(t)
	idGen := NewMockIdGen(t)

	s := NewDataUsageSyncer(repo, idGen, &config.ConfigMap{})

	return s, repo, idGen
}
