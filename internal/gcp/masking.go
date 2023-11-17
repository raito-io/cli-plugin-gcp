package gcp

import (
	"context"
	"errors"
	"fmt"

	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

type NoMasking struct {
	organisationId string
}

func NewNoMasking(configmap *config.ConfigMap) *NoMasking {
	return &NoMasking{
		organisationId: fmt.Sprintf("gcp-org-%s", configmap.GetString(common.GcpOrgId)),
	}
}

func (n *NoMasking) ImportMasks(_ context.Context, _ wrappers.AccessProviderHandler, _ set.Set[string], _ map[string][]string, _ set.Set[string]) error {
	return errors.New("masking is not supported for GCP")
}

func (n *NoMasking) ExportMasks(ctx context.Context, accessProvider *importer.AccessProvider, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler) ([]string, error) {
	err := accessProviderFeedbackHandler.AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
		AccessProvider: accessProvider.Id,
		ActualName:     accessProvider.Name,
		Errors:         []string{"masking is not supported in data source"},
	})
	if err != nil {
		return nil, fmt.Errorf("add access provider feedback: %w", err)
	}

	return nil, nil
}

func (n *NoMasking) MaskedBinding(_ context.Context, members []string) ([]iam.IamBinding, error) {
	bindings := make([]iam.IamBinding, 0, len(members))
	for _, member := range members {
		bindings = append(bindings, iam.IamBinding{
			Member:       member,
			Role:         "roles/bigquerydatapolicy.maskedReader",
			Resource:     n.organisationId,
			ResourceType: "organization",
		})
	}

	return bindings, nil
}
