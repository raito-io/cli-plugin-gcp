package gcp

import (
	"context"
	"errors"
	"fmt"

	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/golang-set/set"
)

type NoFiltering struct {
}

func NewNoFiltering() *NoFiltering {
	return &NoFiltering{}
}

func (f *NoFiltering) ImportFilters(_ context.Context, _ *data_source.DataSourceSyncConfig, _ wrappers.AccessProviderHandler, _ set.Set[string]) error {
	return errors.New("filtering is not supported for GCP")
}

func (f *NoFiltering) ExportFilter(_ context.Context, accessProvider *importer.AccessProvider, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler) (*string, error) {
	err := accessProviderFeedbackHandler.AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
		AccessProvider: accessProvider.Id,
		ActualName:     accessProvider.Name,
		Errors:         []string{"filtering is not supported in data source"},
	})
	if err != nil {
		return nil, fmt.Errorf("add access provider feedback: %w", err)
	}

	return nil, nil
}
