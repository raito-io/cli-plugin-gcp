package bigquery

import (
	"context"
	"testing"

	"cloud.google.com/go/bigquery/datapolicies/apiv1/datapoliciespb"
	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/access_provider/sync_from_target"
	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/raito-io/golang-set/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

func TestBqMaskingService_ImportMasks(t *testing.T) {
	type fields struct {
		setup          func(repository *mockMaskingDataCatalogRepository)
		projectId      string
		maskingEnabled bool
	}
	type args struct {
		ctx         context.Context
		locations   set.Set[string]
		maskingTags map[string][]string
		raitoMasks  set.Set[string]
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantFeedback []sync_from_target.AccessProvider
		wantErr      require.ErrorAssertionFunc
	}{
		{
			name: "No masks to import",
			fields: fields{
				setup: func(repository *mockMaskingDataCatalogRepository) {
					repository.EXPECT().ListDataPolicies(mock.Anything).Return(nil, nil)
				},
				projectId:      "test-project",
				maskingEnabled: true,
			},
			args: args{
				ctx:         context.Background(),
				locations:   set.NewSet("europe-west1"),
				raitoMasks:  set.NewSet("existing-mask"),
				maskingTags: nil,
			},
			wantFeedback: []sync_from_target.AccessProvider{},
			wantErr:      require.NoError,
		},
		{
			name: "No masks if not supported",
			fields: fields{
				setup: func(repository *mockMaskingDataCatalogRepository) {
				},
				projectId:      "test-project",
				maskingEnabled: false,
			},
			args: args{
				ctx:         context.Background(),
				locations:   set.NewSet("europe-west1"),
				raitoMasks:  set.NewSet("existing-mask"),
				maskingTags: nil,
			},
			wantFeedback: []sync_from_target.AccessProvider{},
			wantErr:      require.NoError,
		},
		{
			name: "Get mask",
			fields: fields{
				setup: func(repository *mockMaskingDataCatalogRepository) {
					repository.EXPECT().ListDataPolicies(mock.Anything).Return(map[string]BQMaskingInformation{
						"maskTag1": {
							DataPolicy: BQDataPolicy{
								FullName:   "DataPolicy1",
								PolicyType: datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS,
							},
							PolicyTag: BQPolicyTag{
								FullName: "maskTag1",
								Name:     "maskNameTag1",
							},
						},
					}, nil)

					repository.EXPECT().GetFineGrainedReaderMembers(mock.Anything, "maskTag1").Return([]string{"user:user1@raito.io", "group:sales@raito.io"}, nil)
				},
				projectId:      "test-project",
				maskingEnabled: true,
			},
			args: args{
				ctx:         context.Background(),
				locations:   set.NewSet("europe-west1"),
				raitoMasks:  set.NewSet("existing-mask"),
				maskingTags: map[string][]string{"maskTag1": {"column1", "column3"}},
			},
			wantErr: require.NoError,
			wantFeedback: []sync_from_target.AccessProvider{
				{
					ExternalId: "DataPolicy1",
					Name:       "maskNameTag1",
					Type:       ptr.String(datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS.String()),
					Action:     sync_from_target.Mask,
					Who: &sync_from_target.WhoItem{
						Users:  []string{"user1@raito.io"},
						Groups: []string{"sales@raito.io"},
					},
					ActualName: "maskNameTag1",
					What: []sync_from_target.WhatItem{
						{
							DataObject: &data_source.DataObjectReference{
								FullName: "column1",
								Type:     "column",
							},
							Permissions: []string{},
						},
						{

							DataObject: &data_source.DataObjectReference{
								FullName: "column3",
								Type:     "column",
							},
							Permissions: []string{},
						},
					},
				},
			},
		},
		{
			name: "Ignore unknown policy tag",
			fields: fields{
				setup: func(repository *mockMaskingDataCatalogRepository) {
					repository.EXPECT().ListDataPolicies(mock.Anything).Return(map[string]BQMaskingInformation{}, nil)
				},
				projectId:      "test-project",
				maskingEnabled: true,
			},
			args: args{
				ctx:         context.Background(),
				locations:   set.NewSet("europe-west1"),
				raitoMasks:  set.NewSet("existing-mask"),
				maskingTags: map[string][]string{"maskTag1": {"column1", "column3"}},
			},
			wantErr:      require.NoError,
			wantFeedback: []sync_from_target.AccessProvider{},
		},
		{
			name: "Ignore raito mask",
			fields: fields{
				setup: func(repository *mockMaskingDataCatalogRepository) {
					repository.EXPECT().ListDataPolicies(mock.Anything).Return(map[string]BQMaskingInformation{"maskTag1": {
						DataPolicy: BQDataPolicy{
							FullName:   "existing-mask",
							PolicyType: datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS,
						},
						PolicyTag: BQPolicyTag{
							FullName: "maskTag1",
							Name:     "maskNameTag1",
						},
					}}, nil)
				},
				projectId:      "test-project",
				maskingEnabled: true,
			},
			args: args{
				ctx:         context.Background(),
				locations:   set.NewSet("europe-west1"),
				raitoMasks:  set.NewSet("existing-mask"),
				maskingTags: map[string][]string{"maskTag1": {"column1", "column3"}},
			},
			wantErr:      require.NoError,
			wantFeedback: []sync_from_target.AccessProvider{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maskingService, repo := createMaskingService(t, tt.fields.projectId, tt.fields.maskingEnabled)
			tt.fields.setup(repo)

			apHandler := mocks.NewSimpleAccessProviderHandler(t, 1)

			err := maskingService.ImportMasks(tt.args.ctx, apHandler, tt.args.locations, tt.args.maskingTags, tt.args.raitoMasks)

			tt.wantErr(t, err)
			if err != nil {
				return
			}

			assert.Equal(t, tt.wantFeedback, apHandler.AccessProviders)
		})
	}
}

func TestBqMaskingService_ExportMasks(t *testing.T) {
	newMask := importer.AccessProvider{
		Id:     "MaskId1",
		Type:   ptr.String(datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS.String()),
		Action: importer.Mask,
		Who: importer.WhoItem{
			Users:  []string{"user1@raito.io"},
			Groups: []string{"sales@raito.io"},
		},
		What: []importer.WhatItem{
			{
				DataObject: &data_source.DataObjectReference{
					FullName: "column1",
					Type:     "column",
				},
			},
			{
				DataObject: &data_source.DataObjectReference{
					FullName: "column2",
					Type:     "column",
				},
			},
		},
		DeleteWhat: []importer.WhatItem{
			{
				DataObject: &data_source.DataObjectReference{
					FullName: "column3",
					Type:     "column",
				},
			},
		},
	}

	deleteMask := importer.AccessProvider{
		Id:         "MaskId1",
		ExternalId: ptr.String("DataPolicy1,DataPolicy2"),
		Type:       ptr.String(datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS.String()),
		Action:     importer.Mask,
		Who: importer.WhoItem{
			Users:  []string{"user1@raito.io"},
			Groups: []string{"sales@raito.io"},
		},
		ActualName: ptr.String("maskNameTag1,maskNameTag2"),
		What: []importer.WhatItem{
			{
				DataObject: &data_source.DataObjectReference{
					FullName: "column1",
					Type:     "column",
				},
			},
			{
				DataObject: &data_source.DataObjectReference{
					FullName: "column2",
					Type:     "column",
				},
			},
		},
		Delete: true,
	}

	maskInfo := BQMaskingInformation{
		DataPolicy: BQDataPolicy{
			FullName:   "DataPolicy1",
			PolicyType: datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS,
		},
		PolicyTag: BQPolicyTag{
			FullName: "maskTag1",
			Name:     "maskNameTag1",
		},
	}

	maskInfo2 := BQMaskingInformation{
		DataPolicy: BQDataPolicy{
			FullName:   "DataPolicy2",
			PolicyType: datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS,
		},
		PolicyTag: BQPolicyTag{
			FullName: "maskTag2",
			Name:     "maskNameTag2",
		},
	}

	type fields struct {
		setup          func(repository *mockMaskingDataCatalogRepository)
		projectId      string
		maskingEnabled bool
	}
	type args struct {
		ctx            context.Context
		accessProvider *importer.AccessProvider
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		want         []string
		wantFeedback []importer.AccessProviderSyncFeedback
		wantErr      require.ErrorAssertionFunc
	}{
		{
			name: "Error feedback if masking is not enabled",
			fields: fields{
				setup: func(repository *mockMaskingDataCatalogRepository) {
				},
				projectId:      "test-project",
				maskingEnabled: false,
			},
			args: args{
				ctx:            context.Background(),
				accessProvider: &newMask,
			},
			wantErr: require.NoError,
			wantFeedback: []importer.AccessProviderSyncFeedback{
				{
					AccessProvider: newMask.Id,
					ActualName:     "",
					Errors:         []string{"BigQuery catalog is not enabled"},
				},
			},
			want: []string{},
		},
		{
			name: "Create new mask",
			fields: fields{
				setup: func(repository *mockMaskingDataCatalogRepository) {
					repository.EXPECT().GetLocationsForDataObjects(mock.Anything, &newMask).Return(map[string]string{"column1": "europe-west1", "column2": "europe-west2"}, map[string]string{"column3": "europe-west1"}, nil)
					repository.EXPECT().CreatePolicyTagWithDataPolicy(mock.Anything, "europe-west1", datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS, &newMask).Return(&maskInfo, nil)
					repository.EXPECT().CreatePolicyTagWithDataPolicy(mock.Anything, "europe-west2", datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS, &newMask).Return(&maskInfo2, nil)
					repository.EXPECT().UpdateAccess(mock.Anything, &maskInfo, &newMask.Who, newMask.DeletedWho).Return(nil)
					repository.EXPECT().UpdateAccess(mock.Anything, &maskInfo2, &newMask.Who, newMask.DeletedWho).Return(nil)
					repository.EXPECT().UpdateWhatOfDataPolicy(mock.Anything, &maskInfo2, []string{"column2"}, []string(nil)).Return(nil)

					repository.EXPECT().UpdateWhatOfDataPolicy(mock.Anything, &maskInfo, []string{"column1"}, []string{"column3"}).Return(nil)
				},
				projectId:      "test-project",
				maskingEnabled: true,
			},
			args: args{
				ctx:            context.Background(),
				accessProvider: &newMask,
			},
			wantErr: require.NoError,
			wantFeedback: []importer.AccessProviderSyncFeedback{
				{
					AccessProvider: newMask.Id,
					ActualName:     "maskNameTag1,maskNameTag2",
					ExternalId:     ptr.String("DataPolicy1,DataPolicy2"),
					Type:           ptr.String(datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS.String()),
				},
			},
			want: []string{
				"DataPolicy1",
				"DataPolicy2",
			},
		},
		{
			name: "Delete mask",
			fields: fields{
				setup: func(repository *mockMaskingDataCatalogRepository) {
					repository.EXPECT().DeletePolicyAndTag(mock.Anything, "DataPolicy1").Return(nil)
					repository.EXPECT().DeletePolicyAndTag(mock.Anything, "DataPolicy2").Return(nil)
				},
				projectId:      "test-project",
				maskingEnabled: true,
			},
			args: args{
				ctx:            context.Background(),
				accessProvider: &deleteMask,
			},
			wantErr: require.NoError,
			wantFeedback: []importer.AccessProviderSyncFeedback{
				{
					AccessProvider: deleteMask.Id,
					ActualName:     "maskNameTag1,maskNameTag2",
					ExternalId:     ptr.String("DataPolicy1,DataPolicy2"),
					Type:           ptr.String(datapoliciespb.DataMaskingPolicy_FIRST_FOUR_CHARACTERS.String()),
				},
			},
			want: []string{
				"DataPolicy1",
				"DataPolicy2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maskingService, repo := createMaskingService(t, tt.fields.projectId, tt.fields.maskingEnabled)
			tt.fields.setup(repo)

			feedbackHandler := mocks.NewSimpleAccessProviderFeedbackHandler(t)

			result, err := maskingService.ExportMasks(tt.args.ctx, tt.args.accessProvider, feedbackHandler)

			tt.wantErr(t, err)
			if err != nil {
				return
			}

			assert.ElementsMatch(t, result, tt.want)
			assert.ElementsMatch(t, feedbackHandler.AccessProviderFeedback, tt.wantFeedback)
		})
	}
}

func createMaskingService(t *testing.T, projectId string, maskingEnabled bool) (*BqMaskingService, *mockMaskingDataCatalogRepository) {
	t.Helper()
	repo := newMockMaskingDataCatalogRepository(t)

	service := NewBqMaskingService(repo, &config.ConfigMap{Parameters: map[string]string{common.GcpProjectId: projectId}})
	service.maskingEnabled = maskingEnabled

	return service, repo
}
