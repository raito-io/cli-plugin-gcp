package bigquery

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery/datapolicies/apiv1/datapoliciespb"
	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/access_provider/sync_from_target"
	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

//go:generate go run github.com/vektra/mockery/v2 --name=MaskingDataCatalogRepository --with-expecter --inpackage
type MaskingDataCatalogRepository interface {
	ListDataPolicies(ctx context.Context) (map[string]BQMaskingInformation, error)
	GetFineGrainedReaderMembers(ctx context.Context, tagId string) ([]string, error)
	DeletePolicyAndTag(ctx context.Context, policyTagId string) error
	UpdateAccess(ctx context.Context, maskingInformation *BQMaskingInformation, who *importer.WhoItem, deletedWho *importer.WhoItem) error
	UpdateWhatOfDataPolicy(ctx context.Context, policy *BQMaskingInformation, dataObjects []string, deletedDataObjects []string) error
	UpdatePolicyTag(ctx context.Context, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *importer.AccessProvider, dataPolicyId string) (*BQMaskingInformation, error)
	CreatePolicyTagWithDataPolicy(ctx context.Context, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *importer.AccessProvider) (_ *BQMaskingInformation, err error)
	GetLocationsForDataObjects(ctx context.Context, ap *importer.AccessProvider) (map[string]string, map[string]string, error)
}

type BqMaskingService struct {
	datacatalogRepo MaskingDataCatalogRepository
	projectId       string
	maskingEnabled  bool
}

func NewBqMaskingService(dataCatalogRepository MaskingDataCatalogRepository, configMap *config.ConfigMap) *BqMaskingService {
	return &BqMaskingService{
		datacatalogRepo: dataCatalogRepository,
		projectId:       configMap.GetString(common.GcpProjectId),
		maskingEnabled:  configMap.GetBoolWithDefault(common.BqCatalogEnabled, false),
	}
}

func (m *BqMaskingService) ImportMasks(ctx context.Context, accessProviderHandler wrappers.AccessProviderHandler, locations set.Set[string], maskingTags map[string][]string, raitoMasks set.Set[string]) error {
	if !m.maskingEnabled {
		return nil
	}

	common.Logger.Debug(fmt.Sprintf("Need to check masks for %d locations: %+v", len(locations), locations.Slice()))
	common.Logger.Debug(fmt.Sprintf("Policy tags found for %d columns: %+v", len(maskingTags), maskingTags))

	masks, err := m.datacatalogRepo.ListDataPolicies(ctx)
	if err != nil {
		return err
	}

	for maskTag, columns := range maskingTags {
		mask, found := masks[maskTag]
		if !found {
			common.Logger.Warn(fmt.Sprintf("Data policy for tag %s not found", maskTag))

			continue
		} else if raitoMasks.Contains(mask.DataPolicy.FullName) {
			common.Logger.Debug(fmt.Sprintf("Ingore raito created mask %q", maskTag))

			continue
		}

		whatItems := make([]sync_from_target.WhatItem, 0, len(columns))

		for _, column := range columns {
			whatItems = append(whatItems, sync_from_target.WhatItem{
				DataObject: &ds.DataObjectReference{
					FullName: column,
					Type:     "column",
				},
				Permissions: []string{},
			})
		}

		whoItem := sync_from_target.WhoItem{}

		members, err := m.datacatalogRepo.GetFineGrainedReaderMembers(ctx, mask.PolicyTag.FullName)
		if err != nil {
			return err
		}

		for _, member := range members {
			if strings.HasPrefix(member, "user:") {
				whoItem.Users = append(whoItem.Users, member[5:])
			} else if strings.HasPrefix(member, "serviceAccount:") {
				whoItem.Users = append(whoItem.Users, member[10:])
			} else if strings.HasPrefix(member, "group:") {
				whoItem.Groups = append(whoItem.Groups, member[6:])
			}
		}

		err = accessProviderHandler.AddAccessProviders(
			&sync_from_target.AccessProvider{
				Name:       mask.PolicyTag.Name,
				Type:       ptr.String(mask.DataPolicy.PolicyType.String()),
				What:       whatItems,
				Action:     sync_from_target.Mask,
				ExternalId: mask.DataPolicy.FullName,
				Who:        &whoItem,
				ActualName: mask.PolicyTag.Name,
			},
		)

		if err != nil {
			return fmt.Errorf("add mask to ap handler: %w", err)
		}
	}

	return nil
}

func (m *BqMaskingService) ExportMasks(ctx context.Context, accessProvider *importer.AccessProvider, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler) ([]string, error) {
	if !m.maskingEnabled {
		err := accessProviderFeedbackHandler.AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
			AccessProvider: accessProvider.Id,
			Errors:         []string{"BigQuery catalog is not enabled"},
		})
		if err != nil {
			return nil, fmt.Errorf("add ap feedback to handler: %w", err)
		}
	}

	var actualName, externalId, errors []string
	var maskType *string
	var err error

	if accessProvider.Delete {
		actualName, maskType, externalId, err = m.deleteMask(ctx, accessProvider)
	} else {
		actualName, maskType, externalId, err = m.exportMasks(ctx, accessProvider)
	}

	if err != nil {
		errors = append(errors, err.Error())
	}

	err = accessProviderFeedbackHandler.AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
		AccessProvider: accessProvider.Id,
		ActualName:     strings.Join(actualName, ","),
		Type:           maskType,
		ExternalId:     ptr.String(strings.Join(externalId, ",")),
		Errors:         errors,
	})

	if err != nil {
		return nil, fmt.Errorf("add ap feedback to handler: %w", err)
	}

	return externalId, nil
}

func (m *BqMaskingService) MaskedBinding(_ context.Context, members []string) ([]iam.IamBinding, error) {
	bindings := make([]iam.IamBinding, 0, len(members))

	for _, member := range members {
		bindings = append(bindings, iam.IamBinding{
			Member:       member,
			Role:         "roles/bigquerydatapolicy.maskedReader",
			Resource:     m.projectId,
			ResourceType: "project",
		})
	}

	return bindings, nil
}

func (m *BqMaskingService) deleteMask(ctx context.Context, ap *importer.AccessProvider) (actualNames []string, apType *string, externalIds []string, err error) {
	if ap.ExternalId == nil || *ap.ExternalId == "" {
		common.Logger.Warn(fmt.Sprintf("No external ID found for mask %s. Mask probably already deleted.", ap.Name))

		return nil, nil, nil, nil
	}

	externalIds = strings.Split(*ap.ExternalId, ",")

	if ap.ActualName != nil {
		actualNames = strings.Split(*ap.ActualName, ",")
	}

	common.Logger.Info(fmt.Sprintf("Deleting mask %s with %d policy tags", ap.Name, len(externalIds)))

	for _, externalId := range externalIds {
		common.Logger.Debug(fmt.Sprintf("Delete data policy %s for tag %s", externalId, ap.Name))

		err = m.datacatalogRepo.DeletePolicyAndTag(ctx, externalId)
		if err != nil {
			return actualNames, ap.Type, externalIds, err
		}
	}

	return actualNames, ap.Type, externalIds, nil
}

func (m *BqMaskingService) exportMasks(ctx context.Context, accessProvider *importer.AccessProvider) (actualName []string, apType *string, externalId []string, err error) {
	common.Logger.Info(fmt.Sprintf("Update mask %s", accessProvider.Name))

	defer func() {
		if err != nil {
			common.Logger.Error(fmt.Sprintf("Failed to export mask %q: %s", accessProvider.Name, err.Error()))
		}
	}()

	// List all locations required for mask
	dataPolicyLocations, doLocations, deletedDoLocations, err := m.exportRaitoMaskListAllDoLocations(ctx, accessProvider)
	if err != nil {
		return nil, nil, nil, err
	}

	// First remove old data policies
	err = m.exportRaitoMaskRemoveOldPolicies(ctx, accessProvider, deletedDoLocations, doLocations, dataPolicyLocations)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create or update all data policies
	apType, dataPolicyMap, err := m.exportRaitoMaskCreateAndUpdateDataPolicies(ctx, accessProvider, doLocations, dataPolicyLocations)
	if err != nil {
		return nil, nil, nil, err
	}

	externalId = make([]string, 0, len(dataPolicyMap))
	actualName = make([]string, 0, len(dataPolicyMap))

	for _, maskingInformation := range dataPolicyMap {
		externalId = append(externalId, maskingInformation.DataPolicy.FullName)
		actualName = append(actualName, maskingInformation.PolicyTag.Name)
	}

	for location := range dataPolicyMap {
		dataObjectsToAdd := doLocations[location]
		dataObjectsToRemove := deletedDoLocations[location]

		if len(dataObjectsToAdd) > 0 {
			maskingPolicy := dataPolicyMap[location]

			// Update WHO of policy tag and data policy
			common.Logger.Debug(fmt.Sprintf("Update who for policy tag %q", maskingPolicy.PolicyTag.FullName))

			err = m.datacatalogRepo.UpdateAccess(ctx, &maskingPolicy, &accessProvider.Who, accessProvider.DeletedWho)
			if err != nil {
				return actualName, apType, externalId, err
			}

			// Update What of policy tag
			common.Logger.Debug(fmt.Sprintf("Update what for policy tag %q", maskingPolicy.PolicyTag.FullName))

			err = m.datacatalogRepo.UpdateWhatOfDataPolicy(ctx, &maskingPolicy, dataObjectsToAdd, dataObjectsToRemove)
			if err != nil {
				return actualName, apType, externalId, err
			}
		} else {
			err = m.datacatalogRepo.DeletePolicyAndTag(ctx, dataPolicyMap[location].DataPolicy.FullName)
			if err != nil {
				return actualName, apType, externalId, err
			}
		}
	}

	return actualName, apType, externalId, nil
}

func (m *BqMaskingService) exportRaitoMaskCreateAndUpdateDataPolicies(ctx context.Context, accessProvider *importer.AccessProvider, doLocations map[string][]string, dataPolicyLocations map[string]string) (*string, map[string]BQMaskingInformation, error) {
	common.Logger.Debug(fmt.Sprintf("Create or update data policies for mask %s", accessProvider.Name))

	maskingTypeInt := int32(datapoliciespb.DataMaskingPolicy_ALWAYS_NULL)

	if accessProvider.Type != nil {
		if accessProviderMaskType, found := datapoliciespb.DataMaskingPolicy_PredefinedExpression_value[*accessProvider.Type]; found {
			maskingTypeInt = accessProviderMaskType
		}
	}

	maskingType := datapoliciespb.DataMaskingPolicy_PredefinedExpression(maskingTypeInt)
	apType := ptr.String(maskingType.String())

	dataPolicyMap := make(map[string]BQMaskingInformation)

	for doLocation := range doLocations {
		if dataPolicyId, found := dataPolicyLocations[doLocation]; found {
			// Get MaskingInformation for existing policy
			maskingInformation, err := m.datacatalogRepo.UpdatePolicyTag(ctx, doLocation, maskingType, accessProvider, dataPolicyId)
			if err != nil {
				return nil, nil, fmt.Errorf("update mask %q: %w", dataPolicyId, err)
			}

			dataPolicyMap[doLocation] = *maskingInformation
		} else {
			// Create new data policy
			common.Logger.Info(fmt.Sprintf("Create new data policy and policy tag %s in location %s", accessProvider.Name, doLocation))

			maskingInformation, err := m.datacatalogRepo.CreatePolicyTagWithDataPolicy(ctx, doLocation, maskingType, accessProvider)
			if err != nil {
				return nil, nil, fmt.Errorf("data policy creation: %w", err)
			}

			dataPolicyMap[doLocation] = *maskingInformation
		}
	}

	return apType, dataPolicyMap, nil
}

func (m *BqMaskingService) exportRaitoMaskRemoveOldPolicies(ctx context.Context, accessProvider *importer.AccessProvider, deletedDoLocations map[string][]string, doLocations map[string][]string, dataPolicyLocations map[string]string) error {
	common.Logger.Debug(fmt.Sprintf("Rolmove old policies for mask %s", accessProvider.Name))

	for doLocation := range deletedDoLocations {
		common.Logger.Debug(fmt.Sprintf("check if we should delete policies in location %s", doLocation))

		if _, doFound := doLocations[doLocation]; !doFound {
			common.Logger.Debug("location not used by active data objects")

			if dataPolicyId, policyFound := dataPolicyLocations[doLocation]; policyFound {
				common.Logger.Info(fmt.Sprintf("Delete data policy and policy tag %s in location %s", accessProvider.Name, doLocation))

				err := m.datacatalogRepo.DeletePolicyAndTag(ctx, dataPolicyId)
				if err != nil {
					return fmt.Errorf("delete data policy %q: %w", dataPolicyId, err)
				}
			} else {
				common.Logger.Debug("No policy found. Assuming policy does not exist")
			}
		} else {
			common.Logger.Debug("location used by active data objects")
		}
	}

	return nil
}

func (m *BqMaskingService) exportRaitoMaskListAllDoLocations(ctx context.Context, accessProvider *importer.AccessProvider) (map[string]string, map[string][]string, map[string][]string, error) {
	var dataPolicies []string
	if accessProvider.ExternalId != nil && *accessProvider.ExternalId != "" {
		dataPolicies = strings.Split(*accessProvider.ExternalId, ",")
	}

	common.Logger.Debug(fmt.Sprintf("List all data policies for mask %s: %+v (%d)", accessProvider.Name, dataPolicies, len(dataPolicies)))

	dataPolicyLocations := map[string]string{}

	for _, dataPolicy := range dataPolicies {
		dpNameSplit := strings.Split(dataPolicy, "/")
		dataPolicyLocations[dpNameSplit[3]] = dataPolicy
	}

	dataObjectLocationMap, deletedDataObjectLocationMap, err := m.datacatalogRepo.GetLocationsForDataObjects(ctx, accessProvider)
	if err != nil {
		return nil, nil, nil, err
	}

	doLocations := map[string][]string{}
	deletedDoLocations := map[string][]string{}

	for doObject, dataObjectLocation := range dataObjectLocationMap {
		doLocations[dataObjectLocation] = append(doLocations[dataObjectLocation], doObject)
	}

	for doObject, doObjectLocation := range deletedDataObjectLocationMap {
		deletedDoLocations[doObjectLocation] = append(deletedDoLocations[doObjectLocation], doObject)
	}

	return dataPolicyLocations, doLocations, deletedDoLocations, nil
}
