package gcp

import (
	"context"
	"errors"
	"reflect"
	"testing"

	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/wrappers/mocks"
)

func TestNoMasking_ExportMasks(t *testing.T) {
	type fields struct {
		organisationId string
	}
	type args struct {
		ctx                               context.Context
		accessProvider                    *importer.AccessProvider
		accessProviderFeedbackHandleSetup func(handler *mocks.AccessProviderFeedbackHandler)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "export masks result in not support",
			fields: fields{
				organisationId: "test-org",
			},
			args: args{
				ctx: context.Background(),
				accessProvider: &importer.AccessProvider{
					Name: "test-access-provider",
					Id:   "test-access-provider-id",
				},
				accessProviderFeedbackHandleSetup: func(handler *mocks.AccessProviderFeedbackHandler) {
					handler.EXPECT().AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
						AccessProvider: "test-access-provider-id",
						ActualName:     "test-access-provider",
						Errors:         []string{"masking is not supported in data source"},
					}).Return(nil)
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "export masks result in not support - error in handler",
			fields: fields{
				organisationId: "test-org",
			},
			args: args{
				ctx: context.Background(),
				accessProvider: &importer.AccessProvider{
					Name: "test-access-provider",
					Id:   "test-access-provider-id",
				},
				accessProviderFeedbackHandleSetup: func(handler *mocks.AccessProviderFeedbackHandler) {
					handler.EXPECT().AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
						AccessProvider: "test-access-provider-id",
						ActualName:     "test-access-provider",
						Errors:         []string{"masking is not supported in data source"},
					}).Return(errors.New("test error"))
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NoMasking{
				organisationId: tt.fields.organisationId,
			}

			handlerMock := mocks.NewAccessProviderFeedbackHandler(t)
			tt.args.accessProviderFeedbackHandleSetup(handlerMock)

			got, err := n.ExportMasks(tt.args.ctx, tt.args.accessProvider, handlerMock)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportMasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExportMasks() got = %v, want %v", got, tt.want)
			}
		})
	}
}
