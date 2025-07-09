package bigquery

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/bexpression"
	"github.com/raito-io/bexpression/base"
	"github.com/raito-io/bexpression/datacomparison"
	"github.com/raito-io/cli/base/access_provider/sync_from_target"
	"github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/access_provider/types"
	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/raito-io/golang-set/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/bigquery/v2"

	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func Test_createFilterExpression(t *testing.T) {
	type args struct {
		expr *bexpression.DataComparisonExpression
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "simple expression",
			args: args{
				expr: &bexpression.DataComparisonExpression{
					Literal: ptr.Bool(true),
				},
			},
			want:    "TRUE",
			wantErr: assert.NoError,
		},
		{
			name: "simple comparison expression - data object",
			args: args{
				expr: &bexpression.DataComparisonExpression{
					Comparison: &datacomparison.DataComparison{
						Operator: datacomparison.ComparisonOperatorGreaterThan,
						LeftOperand: datacomparison.Operand{
							Reference: &datacomparison.Reference{
								EntityType: datacomparison.EntityTypeDataObject,
								EntityID:   `{"fullName":"bq-demodata.MASTER_DATA.Sales_CreditCard.ExpYear","id":"JJGSpyjrssv94KPk9dNuI","type":"column"}`,
							},
						},
						RightOperand: datacomparison.Operand{
							Literal: &datacomparison.Literal{
								Int: ptr.Int(2020),
							},
						},
					},
				},
			},
			want:    "ExpYear > 2020",
			wantErr: assert.NoError,
		},
		{
			name: "simple comparison expression - column by reference",
			args: args{
				expr: &bexpression.DataComparisonExpression{
					Comparison: &datacomparison.DataComparison{
						Operator: datacomparison.ComparisonOperatorGreaterThan,
						LeftOperand: datacomparison.Operand{
							Reference: &datacomparison.Reference{
								EntityType: datacomparison.EntityTypeColumnReferenceByName,
								EntityID:   `ExpYear`,
							},
						},
						RightOperand: datacomparison.Operand{
							Literal: &datacomparison.Literal{
								Int: ptr.Int(2020),
							},
						},
					},
				},
			},
			want:    "ExpYear > 2020",
			wantErr: assert.NoError,
		},
		{
			name: "aggregation expression",
			args: args{
				expr: &bexpression.DataComparisonExpression{
					Aggregator: &bexpression.DataComparisonAggregator{
						Operator: base.AggregatorOperatorAnd,
						Operands: []bexpression.DataComparisonExpression{
							{
								Literal: ptr.Bool(true),
							},
							{
								Aggregator: &bexpression.DataComparisonAggregator{
									Operator: base.AggregatorOperatorOr,
									Operands: []bexpression.DataComparisonExpression{
										{
											Comparison: &datacomparison.DataComparison{
												Operator: datacomparison.ComparisonOperatorEqual,
												LeftOperand: datacomparison.Operand{
													Reference: &datacomparison.Reference{
														EntityType: datacomparison.EntityTypeColumnReferenceByName,
														EntityID:   `State`,
													},
												},
												RightOperand: datacomparison.Operand{
													Literal: &datacomparison.Literal{
														Str: ptr.String("CA"),
													},
												},
											},
										},
										{
											UnaryExpression: &bexpression.DataComparisonUnaryExpression{
												Operator: base.UnaryOperatorNot,
												Operand: bexpression.DataComparisonExpression{
													Comparison: &datacomparison.DataComparison{
														Operator: datacomparison.ComparisonOperatorEqual,
														LeftOperand: datacomparison.Operand{
															Reference: &datacomparison.Reference{
																EntityType: datacomparison.EntityTypeColumnReferenceByName,
																EntityID:   `Currency`,
															},
														},
														RightOperand: datacomparison.Operand{
															Literal: &datacomparison.Literal{
																Str: ptr.String("EURO"),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want:    `TRUE AND ((State = "CA") OR (NOT (Currency = "EURO")))`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createFilterExpression(context.Background(), tt.args.expr)
			if !tt.wantErr(t, err, fmt.Sprintf("createFilterExpression(%v)", tt.args.expr)) {
				return
			}
			assert.Equalf(t, tt.want, got, "createFilterExpression(%v)", tt.args.expr)
		})
	}
}

func TestBqFilteringService_ImportFilters(t *testing.T) {
	type fields struct {
		setup func(repositoryMock *mockFilteringRepository, doIteratorMock *mockFilteringDataObjectIterator)
	}
	type args struct {
		ctx          context.Context
		config       *ds.DataSourceSyncConfig
		raitoFilters set.Set[string]
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []sync_from_target.AccessProvider
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "no filters",
			fields: fields{
				setup: func(repositoryMock *mockFilteringRepository, doIteratorMock *mockFilteringDataObjectIterator) {
					doIteratorMock.EXPECT().Sync(mock.Anything, mock.Anything, true, mock.Anything).RunAndReturn(func(ctx context.Context, config *ds.DataSourceSyncConfig, b bool, f func(context.Context, *org.GcpOrgEntity) error) error {
						err := f(ctx, &org.GcpOrgEntity{
							Name: "dataset1",
							Id:   "dataset1",
							Type: ds.Dataset,
						})
						if err != nil {
							return err
						}

						err = f(ctx, &org.GcpOrgEntity{
							Name: "table1",
							Id:   "table1",
							Type: ds.Table,
						})
						if err != nil {
							return err
						}

						return f(ctx, &org.GcpOrgEntity{
							Name: "table2",
							Id:   "table2",
							Type: ds.Table,
						})
					}).Once()

					repositoryMock.EXPECT().ListFilters(mock.Anything, &org.GcpOrgEntity{Name: "table1", Id: "table1", Type: ds.Table}, mock.Anything).Return(nil).Once()
					repositoryMock.EXPECT().ListFilters(mock.Anything, &org.GcpOrgEntity{Name: "table2", Id: "table2", Type: ds.Table}, mock.Anything).Return(nil).Once()
				},
			},
			args: args{
				ctx:          context.Background(),
				config:       &ds.DataSourceSyncConfig{},
				raitoFilters: set.NewSet[string]("raitoFilter1"),
			},
			want:    []sync_from_target.AccessProvider{},
			wantErr: require.NoError,
		},
		{
			name: "ignore Raito managed filters",
			fields: fields{
				setup: func(repositoryMock *mockFilteringRepository, doIteratorMock *mockFilteringDataObjectIterator) {
					doIteratorMock.EXPECT().Sync(mock.Anything, mock.Anything, true, mock.Anything).RunAndReturn(func(ctx context.Context, config *ds.DataSourceSyncConfig, b bool, f func(context.Context, *org.GcpOrgEntity) error) error {
						err := f(ctx, &org.GcpOrgEntity{
							Name: "dataset1",
							Id:   "dataset1",
							Type: ds.Dataset,
						})
						if err != nil {
							return err
						}

						err = f(ctx, &org.GcpOrgEntity{
							Name: "table1",
							Id:   "table1",
							Type: ds.Table,
						})
						if err != nil {
							return err
						}

						return f(ctx, &org.GcpOrgEntity{
							Name: "table2",
							Id:   "table2",
							Type: ds.Table,
						})
					}).Once()

					repositoryMock.EXPECT().ListFilters(mock.Anything, &org.GcpOrgEntity{Name: "table1", Id: "table1", Type: ds.Table}, mock.Anything).RunAndReturn(func(ctx context.Context, entity *org.GcpOrgEntity, f func(context.Context, *bigquery.RowAccessPolicy, []string, []string, bool) error) error {
						err := f(ctx, &bigquery.RowAccessPolicy{
							FilterPredicate: "column1 = \"value1\"",
							RowAccessPolicyReference: &bigquery.RowAccessPolicyReference{
								PolicyId:  "policyId1",
								TableId:   "table1",
								DatasetId: "dataset1",
								ProjectId: "projectId1",
							},
							ForceSendFields: nil,
							NullFields:      nil,
						}, []string{"ruben@raito.io"}, []string{"sales@raito.io"}, true)

						if err != nil {
							return err
						}

						return f(ctx, &bigquery.RowAccessPolicy{
							FilterPredicate: "column2 = \"value2\"",
							RowAccessPolicyReference: &bigquery.RowAccessPolicyReference{
								PolicyId:  "raitoFilter1",
								TableId:   "table1",
								DatasetId: "dataset1",
								ProjectId: "projectId1",
							},
							ForceSendFields: nil,
							NullFields:      nil,
						}, []string{"ruben@raito.io"}, []string{"sales@raito.io"}, true)
					}).Once()
					repositoryMock.EXPECT().ListFilters(mock.Anything, &org.GcpOrgEntity{Name: "table2", Id: "table2", Type: ds.Table}, mock.Anything).Return(nil).Once()
				},
			},
			args: args{
				ctx:          context.Background(),
				config:       &ds.DataSourceSyncConfig{},
				raitoFilters: set.NewSet[string]("projectId1.dataset1.table1.raitoFilter1"),
			},
			want: []sync_from_target.AccessProvider{
				{
					ExternalId: "projectId1.dataset1.table1.policyId1",
					Name:       "policyId1",
					NamingHint: "policyId1",
					Action:     types.Filtered,
					Policy:     "column1 = \"value1\"",
					Who: &sync_from_target.WhoItem{
						Users:  []string{"ruben@raito.io"},
						Groups: []string{"sales@raito.io"},
					},
					NotInternalizable: false,
					ActualName:        "policyId1",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &ds.DataObjectReference{
								FullName: "projectId1.dataset1.table1",
								Type:     ds.Table,
							},
						},
					},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "non internalizable filter",
			fields: fields{
				setup: func(repositoryMock *mockFilteringRepository, doIteratorMock *mockFilteringDataObjectIterator) {
					doIteratorMock.EXPECT().Sync(mock.Anything, mock.Anything, true, mock.Anything).RunAndReturn(func(ctx context.Context, config *ds.DataSourceSyncConfig, b bool, f func(context.Context, *org.GcpOrgEntity) error) error {
						err := f(ctx, &org.GcpOrgEntity{
							Name: "dataset1",
							Id:   "dataset1",
							Type: ds.Dataset,
						})
						if err != nil {
							return err
						}

						return f(ctx, &org.GcpOrgEntity{
							Name: "table1",
							Id:   "table1",
							Type: ds.Table,
						})
					}).Once()

					repositoryMock.EXPECT().ListFilters(mock.Anything, &org.GcpOrgEntity{Name: "table1", Id: "table1", Type: ds.Table}, mock.Anything).RunAndReturn(func(ctx context.Context, entity *org.GcpOrgEntity, f func(context.Context, *bigquery.RowAccessPolicy, []string, []string, bool) error) error {
						return f(ctx, &bigquery.RowAccessPolicy{
							FilterPredicate: "column1 = \"value2\"",
							RowAccessPolicyReference: &bigquery.RowAccessPolicyReference{
								PolicyId:  "filter1",
								TableId:   "table1",
								DatasetId: "dataset1",
								ProjectId: "projectId1",
							},
							ForceSendFields: nil,
							NullFields:      nil,
						}, []string{"ruben@raito.io"}, []string{"sales@raito.io"}, false)
					}).Once()
				},
			},
			args: args{
				ctx:          context.Background(),
				config:       &ds.DataSourceSyncConfig{},
				raitoFilters: set.NewSet[string]("projectId1.dataset1.table1.raitoFilter1"),
			},
			want: []sync_from_target.AccessProvider{
				{
					ExternalId: "projectId1.dataset1.table1.filter1",
					Name:       "filter1",
					NamingHint: "filter1",
					Action:     types.Filtered,
					Policy:     "column1 = \"value2\"",
					Who: &sync_from_target.WhoItem{
						Users:  []string{"ruben@raito.io"},
						Groups: []string{"sales@raito.io"},
					},
					NotInternalizable: true,
					ActualName:        "filter1",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &ds.DataObjectReference{
								FullName: "projectId1.dataset1.table1",
								Type:     ds.Table,
							},
						},
					},
				},
			},
			wantErr: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, mockRepo, mockDoIterator := createFilteringService(t)
			tt.fields.setup(mockRepo, mockDoIterator)

			apHandler := mocks.NewSimpleAccessProviderHandler(t, 1)

			err := s.ImportFilters(tt.args.ctx, tt.args.config, apHandler, tt.args.raitoFilters)
			tt.wantErr(t, err)

			if err != nil {
				return
			}

			assert.ElementsMatch(t, tt.want, apHandler.AccessProviders)
		})
	}
}

func TestBqFilteringService_ExportFilter(t *testing.T) {
	type fields struct {
		setup func(repositoryMock *mockFilteringRepository, doIteratorMock *mockFilteringDataObjectIterator)
	}
	type args struct {
		ctx            context.Context
		accessProvider *sync_to_target.AccessProvider
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		want         *string
		wantFeedback *sync_to_target.AccessProviderSyncFeedback
		wantErr      require.ErrorAssertionFunc
	}{
		{
			name: "New filter with policy rule",
			fields: fields{
				setup: func(repositoryMock *mockFilteringRepository, _ *mockFilteringDataObjectIterator) {
					repositoryMock.EXPECT().CreateOrUpdateFilter(mock.Anything, &BQFilter{
						FilterName: "filter1",
						Table: BQReferencedTable{
							Project: "project1",
							Dataset: "dataset1",
							Table:   "table1",
						},
						Users:            []string{"ruben@raito.io"},
						Groups:           []string{"sales@raito.io"},
						FilterExpression: "column1 = \"value1\"",
					}).Return(nil).Once()
				},
			},
			args: args{
				ctx: context.Background(),
				accessProvider: &sync_to_target.AccessProvider{
					Id:          "apId1",
					Name:        "filter1-name",
					Description: "filter1 description",
					NamingHint:  "filter1",
					Action:      types.Filtered,
					Who: sync_to_target.WhoItem{
						Users:  []string{"ruben@raito.io"},
						Groups: []string{"sales@raito.io"},
					},
					Delete:         false,
					PolicyRule:     ptr.String("column1 = \"value1\""),
					FilterCriteria: nil,
					What: []sync_to_target.WhatItem{
						{
							DataObject: &ds.DataObjectReference{
								FullName: "project1.dataset1.table1",
								Type:     ds.Table,
							},
						},
					},
					DeleteWhat: nil,
				},
			},
			want: ptr.String("project1.dataset1.table1.filter1"),
			wantFeedback: &sync_to_target.AccessProviderSyncFeedback{
				AccessProvider: "apId1",
				ActualName:     "filter1",
				ExternalId:     ptr.String("project1.dataset1.table1.filter1"),
				State: &sync_to_target.AccessProviderFeedbackState{
					Who: sync_to_target.AccessProviderWhoFeedbackState{
						Users:  []string{"ruben@raito.io"},
						Groups: []string{"sales@raito.io"},
					},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Update existing filter with filter criteria",
			fields: fields{
				setup: func(repositoryMock *mockFilteringRepository, _ *mockFilteringDataObjectIterator) {
					repositoryMock.EXPECT().CreateOrUpdateFilter(mock.Anything, &BQFilter{
						FilterName: "filter3",
						Table: BQReferencedTable{
							Project: "project1",
							Dataset: "dataset1",
							Table:   "table1",
						},
						Users:            []string{"ruben@raito.io"},
						Groups:           []string{"sales@raito.io"},
						FilterExpression: "column1 = \"value2\"",
					}).Return(nil).Once()
				},
			},
			args: args{
				ctx: context.Background(),
				accessProvider: &sync_to_target.AccessProvider{
					Id:          "apId1",
					Name:        "filter1-name",
					Description: "filter1 description",
					NamingHint:  "filter1",
					ExternalId:  ptr.String("project1.dataset1.table1.filter3"),
					Action:      types.Filtered,
					Who: sync_to_target.WhoItem{
						Users:  []string{"ruben@raito.io"},
						Groups: []string{"sales@raito.io"},
					},
					Delete:     false,
					PolicyRule: nil,
					FilterCriteria: &bexpression.DataComparisonExpression{
						Comparison: &datacomparison.DataComparison{
							Operator: datacomparison.ComparisonOperatorEqual,
							LeftOperand: datacomparison.Operand{
								Reference: &datacomparison.Reference{
									EntityType: datacomparison.EntityTypeDataObject,
									EntityID:   `{"fullName": "project1.dataset1.table1.column1", "type": "column"}`,
								},
							},
							RightOperand: datacomparison.Operand{
								Literal: &datacomparison.Literal{
									Str: ptr.String("value2"),
								},
							},
						},
					},
					What: []sync_to_target.WhatItem{
						{
							DataObject: &ds.DataObjectReference{
								FullName: "project1.dataset1.table1",
								Type:     ds.Table,
							},
						},
					},
					DeleteWhat: nil,
				},
			},
			want: ptr.String("project1.dataset1.table1.filter3"),
			wantFeedback: &sync_to_target.AccessProviderSyncFeedback{
				AccessProvider: "apId1",
				ActualName:     "filter3",
				ExternalId:     ptr.String("project1.dataset1.table1.filter3"),
				State: &sync_to_target.AccessProviderFeedbackState{
					Who: sync_to_target.AccessProviderWhoFeedbackState{
						Users:  []string{"ruben@raito.io"},
						Groups: []string{"sales@raito.io"},
					},
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Delete existing filter",
			fields: fields{
				setup: func(repositoryMock *mockFilteringRepository, _ *mockFilteringDataObjectIterator) {
					repositoryMock.EXPECT().DeleteFilter(mock.Anything, &BQReferencedTable{Project: "project1", Dataset: "dataset1", Table: "table1"}, "filter2").Return(nil).Once()
				},
			},
			args: args{
				ctx: context.Background(),
				accessProvider: &sync_to_target.AccessProvider{
					Id:          "apId1",
					Name:        "filter1-name",
					Description: "filter1 description",
					NamingHint:  "filter1",
					ExternalId:  ptr.String("project1.dataset1.table1.filter2"),
					Action:      types.Filtered,
					Who: sync_to_target.WhoItem{
						Users:  []string{"ruben@raito.io"},
						Groups: []string{"sales@raito.io"},
					},
					Delete:     true,
					PolicyRule: nil,
					FilterCriteria: &bexpression.DataComparisonExpression{
						Comparison: &datacomparison.DataComparison{
							Operator: datacomparison.ComparisonOperatorEqual,
							LeftOperand: datacomparison.Operand{
								Reference: &datacomparison.Reference{
									EntityType: datacomparison.EntityTypeDataObject,
									EntityID:   `{"fullName": "project1.dataset1.table1.column1", "type": "column"}`,
								},
							},
							RightOperand: datacomparison.Operand{
								Literal: &datacomparison.Literal{
									Str: ptr.String("value2"),
								},
							},
						},
					},
					What: []sync_to_target.WhatItem{
						{
							DataObject: &ds.DataObjectReference{
								FullName: "project1.dataset1.table1",
								Type:     ds.Table,
							},
						},
					},
					DeleteWhat: nil,
				},
			},
			want: ptr.String("project1.dataset1.table1.filter2"),
			wantFeedback: &sync_to_target.AccessProviderSyncFeedback{
				AccessProvider: "apId1",
				ActualName:     "filter2",
				ExternalId:     ptr.String("project1.dataset1.table1.filter2"),
				State: &sync_to_target.AccessProviderFeedbackState{
					Who: sync_to_target.AccessProviderWhoFeedbackState{
						Users:  []string{"ruben@raito.io"},
						Groups: []string{"sales@raito.io"},
					},
				},
			},
			wantErr: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, mockRepo, mockDoIterator := createFilteringService(t)
			tt.fields.setup(mockRepo, mockDoIterator)

			apFeedbackHandler := mocks.NewSimpleAccessProviderFeedbackHandler(t)

			got, err := s.ExportFilter(tt.args.ctx, tt.args.accessProvider, apFeedbackHandler)
			tt.wantErr(t, err)

			if err != nil {
				return
			}

			assert.Equal(t, tt.want, got)

			if tt.wantFeedback == nil {
				assert.Empty(t, apFeedbackHandler.AccessProviderFeedback)
			} else {
				assert.Len(t, apFeedbackHandler.AccessProviderFeedback, 1)
				assert.Contains(t, apFeedbackHandler.AccessProviderFeedback, *tt.wantFeedback)
			}
		})
	}
}

func createFilteringService(t *testing.T) (*BqFilteringService, *mockFilteringRepository, *mockFilteringDataObjectIterator) {
	t.Helper()

	repoMock := newMockFilteringRepository(t)
	doIteratorMock := newMockFilteringDataObjectIterator(t)

	return NewBqFilteringService(repoMock, doIteratorMock), repoMock, doIteratorMock
}
