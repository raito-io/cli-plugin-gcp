package bigquery

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery/datapolicies/apiv1/datapoliciespb"
	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/access_provider/sync_from_target"
	"github.com/raito-io/golang-set/set"

	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/gcp/gcp"
	"github.com/raito-io/cli-plugin-gcp/gcp/iam"

	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/util/config"
)

type AccessSyncer struct {
	datacatalogRepo    *dataCatalogIamRepository
	iamServiceProvider func(configMap *config.ConfigMap) iam.IAMService
	gcpAccessSyncer    *gcp.AccessSyncer

	raitoMasks set.Set[string]
}

func NewDataAccessSyncer() *AccessSyncer {
	datacatalogRepo := &dataCatalogIamRepository{
		bigQueryRepo: &BigQueryRepository{},
	}

	iamServiceProvider := func(configMap *config.ConfigMap) iam.IAMService {
		return newIamServiceProvider(configMap).WithServiceIamRepo([]string{"dataset", "table"}, &bigQueryIamRepository{}, GetResourceIds).WithBindingHook(func(ap *importer.AccessProvider, members, deletedMembers []string, what importer.WhatItem) ([]iam.IamBinding, []iam.IamBinding) {
			if configMap.GetBoolWithDefault(BqCatalogEnabled, false) && !ap.Delete {
				var bindingsToAdd []iam.IamBinding

				for _, member := range members {
					bindingsToAdd = append(bindingsToAdd, iam.IamBinding{
						Member:       member,
						Role:         "roles/bigquerydatapolicy.maskedReader",
						Resource:     strings.SplitN(what.DataObject.FullName, ".", 2)[0],
						ResourceType: "project",
					})
				}

				return bindingsToAdd, nil
			}

			return nil, nil
		})
	}

	return &AccessSyncer{
		datacatalogRepo:    datacatalogRepo,
		iamServiceProvider: iamServiceProvider,
		gcpAccessSyncer:    gcp.NewDataAccessSyncer().WithIAMServiceProvider(iamServiceProvider).WithDataSourceMetadataFetcher(GetAlteredDataSourceMetaData),
		raitoMasks:         set.NewSet[string](),
	}
}

// GetAlteredDataSourceMetaData provides an altered version of the data source meta data to also add the 'project' type.
// This is needed to correctly look up the applicable permissions for the mapped data object type (datasource = project).
func GetAlteredDataSourceMetaData(ctx context.Context, config *config.ConfigMap) (*ds.MetaData, error) {
	md, err := GetDataSourceMetaData(ctx, config)

	var permissions []*ds.DataObjectTypePermission

	for _, dot := range md.DataObjectTypes {
		if dot.Name == ds.Datasource {
			permissions = dot.Permissions
		}
	}

	md.DataObjectTypes = append(md.DataObjectTypes, &ds.DataObjectType{
		Name:        project,
		Type:        project,
		Permissions: permissions,
	})

	return md, err
}

func (a *AccessSyncer) SyncAccessProvidersFromTarget(ctx context.Context, accessProviderHandler wrappers.AccessProviderHandler, configMap *config.ConfigMap) error {
	bindings, err := a.iamServiceProvider(configMap).GetIAMPolicyBindings(ctx, configMap)
	if err != nil || len(bindings) == 0 {
		return err
	}

	aps, err := a.gcpAccessSyncer.ConvertBindingsToAccessProviders(ctx, configMap, bindings)

	if err != nil || len(aps) == 0 {
		return err
	}

	for i := range aps {
		ap := aps[i]

		for wi := range ap.What {
			what := ap.What[wi]
			if what.DataObject.Type == project {
				ap.What[wi] = sync_from_target.WhatItem{
					DataObject: &ds.DataObjectReference{
						FullName: what.DataObject.FullName,
						Type:     "datasource",
					},
					Permissions: what.Permissions,
				}
			}
		}

		err = accessProviderHandler.AddAccessProviders(ap)

		if err != nil {
			return err
		}
	}

	if configMap.GetBoolWithDefault(BqCatalogEnabled, false) {
		logger.Info("Import masks.")
		logger.Debug(fmt.Sprintf("%d masks created by raito. Those will be ingored during import: %v", len(a.raitoMasks), a.raitoMasks.Slice()))

		err = a.importMasks(ctx, accessProviderHandler, configMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *AccessSyncer) SyncAccessProviderToTarget(ctx context.Context, accessProviders *importer.AccessProviderImport, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler, configMap *config.ConfigMap) error {
	grants := make([]*importer.AccessProvider, 0, len(accessProviders.AccessProviders))

	for i, ap := range accessProviders.AccessProviders {
		for j, w := range ap.What {
			if w.DataObject.Type == ds.Datasource {
				accessProviders.AccessProviders[i].What[j].DataObject.Type = project
			}
		}

		for j, w := range ap.DeleteWhat {
			if w.DataObject.Type == ds.Datasource {
				accessProviders.AccessProviders[i].DeleteWhat[j].DataObject.Type = project
			}
		}

		if ap.Action == importer.Grant {
			grants = append(grants, accessProviders.AccessProviders[i])
		} else if ap.Action == importer.Mask {
			if !configMap.GetBoolWithDefault(BqCatalogEnabled, false) {
				err := accessProviderFeedbackHandler.AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
					AccessProvider: ap.Id,
					Errors:         []string{"BigQuery catalog is not enabled"},
				})

				if err != nil {
					return err
				}
			}

			var actualName, externalId, errors []string
			var maskType *string
			var err error

			if ap.Delete {
				actualName, maskType, externalId, err = a.deleteMask(ctx, ap, configMap)
			} else {
				actualName, maskType, externalId, err = a.exportMasks(ctx, ap, configMap)
			}

			if err != nil {
				errors = append(errors, err.Error())
			}

			err = accessProviderFeedbackHandler.AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
				AccessProvider: ap.Id,
				ActualName:     strings.Join(actualName, ","),
				Type:           maskType,
				ExternalId:     ptr.String(strings.Join(externalId, ",")),
				Errors:         errors,
			})

			if err != nil {
				return err
			}

			a.raitoMasks.Add(externalId...)
		}
	}

	return a.gcpAccessSyncer.SyncAccessProviderToTarget(ctx, &importer.AccessProviderImport{
		LastCalculated:  accessProviders.LastCalculated,
		AccessProviders: grants,
	}, accessProviderFeedbackHandler, configMap)
}

func (a *AccessSyncer) SyncAccessAsCodeToTarget(ctx context.Context, accessProviders *importer.AccessProviderImport, prefix string, configMap *config.ConfigMap) error {
	return a.gcpAccessSyncer.SyncAccessAsCodeToTarget(ctx, accessProviders, prefix, configMap)
}

func (a *AccessSyncer) importMasks(ctx context.Context, accessProviderHandler wrappers.AccessProviderHandler, configMap *config.ConfigMap) error {
	locations := set.NewSet[string]()
	maskingTags := make(map[string][]string)

	dataSets, err := a.datacatalogRepo.bigQueryRepo.GetDataSets(ctx, configMap)
	if err != nil {
		return err
	}

	for dsIdx := range dataSets {
		tablesAndColumns, err2 := a.datacatalogRepo.bigQueryRepo.GetTables(ctx, configMap, dataSets[dsIdx])
		if err2 != nil {
			return err2
		}

		for idx := range tablesAndColumns {
			if tablesAndColumns[idx].Type != "column" {
				continue
			}

			if len(tablesAndColumns[idx].PolicyTags) > 0 {
				locations.Add(tablesAndColumns[idx].Location)

				for tagIdx := range tablesAndColumns[idx].PolicyTags {
					if !a.raitoMasks.Contains(tablesAndColumns[idx].PolicyTags[tagIdx]) {
						maskingTags[tablesAndColumns[idx].PolicyTags[tagIdx]] = append(maskingTags[tablesAndColumns[idx].PolicyTags[tagIdx]], tablesAndColumns[idx].FullName)
					}
				}
			}
		}
	}

	logger.Debug(fmt.Sprintf("Need to check masks for %d locations: %+v", len(locations), locations.Slice()))
	logger.Debug(fmt.Sprintf("Policy tags found for %d columns: %+v", len(maskingTags), maskingTags))

	masks, err := a.datacatalogRepo.ListDataPolicies(ctx, configMap)
	if err != nil {
		return err
	}

	for maskTag, columns := range maskingTags {
		mask, found := masks[maskTag]
		if !found {
			logger.Warn(fmt.Sprintf("Data policy for tag %s not found", maskTag))

			continue
		} else if a.raitoMasks.Contains(mask.DataPolicy.FullName) {
			logger.Debug(fmt.Sprintf("Ingore raito created mask %q", maskTag))

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

		members, err := a.datacatalogRepo.GetFineGrainedReaderMembers(ctx, configMap, mask.PolicyTag.FullName)
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
			return err
		}
	}

	return nil
}

func (a *AccessSyncer) deleteMask(ctx context.Context, ap *importer.AccessProvider, configMap *config.ConfigMap) (actualNames []string, apType *string, externalIds []string, err error) {
	if ap.ExternalId == nil || *ap.ExternalId == "" {
		logger.Warn(fmt.Sprintf("No external ID found for mask %s. Mask probably already deleted.", ap.Name))

		return nil, nil, nil, nil
	}

	externalIds = strings.Split(*ap.ExternalId, ",")

	if ap.ActualName != nil {
		actualNames = strings.Split(*ap.ActualName, ",")
	}

	logger.Info(fmt.Sprintf("Deleting mask %s with %d policy tags", ap.Name, len(externalIds)))

	for _, externalId := range externalIds {
		logger.Debug(fmt.Sprintf("Delete data policy %s for tag %s", externalId, ap.Name))

		err = a.datacatalogRepo.DeletePolicyAndTag(ctx, configMap, externalId)
		if err != nil {
			return actualNames, ap.Type, externalIds, err
		}
	}

	return actualNames, ap.Type, externalIds, nil
}

func (a *AccessSyncer) exportMasks(ctx context.Context, accessProvider *importer.AccessProvider, configMap *config.ConfigMap) (actualName []string, apType *string, externalIds []string, err error) {
	defer func() {
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to export mask %q: %s", accessProvider.Name, err.Error()))
		}
	}()

	return a.exportRaitoMask(ctx, accessProvider, configMap)
}

func (a *AccessSyncer) exportRaitoMask(ctx context.Context, accessProvider *importer.AccessProvider, configMap *config.ConfigMap) (actualName []string, apType *string, externalId []string, err error) {
	logger.Info(fmt.Sprintf("Update mask %s", accessProvider.Name))

	// List all locations required for mask
	dataPolicyLocations, doLocations, deletedDoLocations, err := a.exportRaitoMaskListAllDoLocations(ctx, accessProvider, configMap)
	if err != nil {
		return nil, nil, nil, err
	}

	// First remove old data policies
	err = a.exportRaitoMaskRemoveOldPolicies(ctx, accessProvider, configMap, deletedDoLocations, doLocations, dataPolicyLocations)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create or update all data policies
	apType, dataPolicyMap, err := a.exportRaitoMaskCreateAndUpdateDataPolicies(ctx, accessProvider, configMap, doLocations, dataPolicyLocations)
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
			logger.Debug(fmt.Sprintf("Update who for policy tag %q", maskingPolicy.PolicyTag.FullName))

			err = a.datacatalogRepo.UpdateAccess(ctx, configMap, &maskingPolicy, &accessProvider.Who, accessProvider.DeletedWho)
			if err != nil {
				return actualName, apType, externalId, err
			}

			// Update What of policy tag
			logger.Debug(fmt.Sprintf("Update what for policy tag %q", maskingPolicy.PolicyTag.FullName))

			err = a.datacatalogRepo.UpdateWhatOfDataPolicy(ctx, configMap, &maskingPolicy, dataObjectsToAdd, dataObjectsToRemove)
			if err != nil {
				return actualName, apType, externalId, err
			}
		} else {
			err = a.datacatalogRepo.DeletePolicyAndTag(ctx, configMap, dataPolicyMap[location].DataPolicy.FullName)
			if err != nil {
				return actualName, apType, externalId, err
			}
		}
	}

	return actualName, apType, externalId, nil
}

func (a *AccessSyncer) exportRaitoMaskCreateAndUpdateDataPolicies(ctx context.Context, accessProvider *importer.AccessProvider, configMap *config.ConfigMap, doLocations map[string][]string, dataPolicyLocations map[string]string) (*string, map[string]BQMaskingInformation, error) {
	logger.Debug(fmt.Sprintf("Create or update data policies for mask %s", accessProvider.Name))

	maskingTypeInt, found := datapoliciespb.DataMaskingPolicy_PredefinedExpression_value[GetValueIfExists(accessProvider.Type, "ALWAYS_NULL")]
	if !found {
		maskingTypeInt = int32(datapoliciespb.DataMaskingPolicy_ALWAYS_NULL)
	}

	maskingType := datapoliciespb.DataMaskingPolicy_PredefinedExpression(maskingTypeInt)
	apType := ptr.String(maskingType.String())

	dataPolicyMap := make(map[string]BQMaskingInformation)

	for doLocation := range doLocations {
		if dataPolicyId, found := dataPolicyLocations[doLocation]; found {
			// Get MaskingInformation for existing policy
			maskingInformation, err := a.datacatalogRepo.UpdatePolicyTag(ctx, configMap, doLocation, maskingType, accessProvider, dataPolicyId)
			if err != nil {
				return nil, nil, fmt.Errorf("update mask %q: %w", dataPolicyId, err)
			}

			dataPolicyMap[doLocation] = *maskingInformation
		} else {
			// Create new data policy
			logger.Info(fmt.Sprintf("Create new data policy and policy tag %s in location %s", accessProvider.Name, doLocation))

			maskingInformation, err := a.datacatalogRepo.CreatePolicyTagWithDataPolicy(ctx, configMap, doLocation, maskingType, accessProvider)
			if err != nil {
				return nil, nil, fmt.Errorf("data policy creation: %w", err)
			}

			dataPolicyMap[doLocation] = *maskingInformation
		}
	}

	return apType, dataPolicyMap, nil
}

func (a *AccessSyncer) exportRaitoMaskRemoveOldPolicies(ctx context.Context, accessProvider *importer.AccessProvider, configMap *config.ConfigMap, deletedDoLocations map[string][]string, doLocations map[string][]string, dataPolicyLocations map[string]string) error {
	logger.Debug(fmt.Sprintf("Rolmove old policies for mask %s", accessProvider.Name))

	for doLocation := range deletedDoLocations {
		logger.Debug(fmt.Sprintf("check if we should delete policies in location %s", doLocation))

		if _, doFound := doLocations[doLocation]; !doFound {
			logger.Debug("location not used by active data objects")

			if dataPolicyId, policyFound := dataPolicyLocations[doLocation]; policyFound {
				logger.Info(fmt.Sprintf("Delete data policy and policy tag %s in location %s", accessProvider.Name, doLocation))

				err := a.datacatalogRepo.DeletePolicyAndTag(ctx, configMap, dataPolicyId)
				if err != nil {
					return fmt.Errorf("delete data policy %q: %w", dataPolicyId, err)
				}
			} else {
				logger.Debug("No policy found. Assuming policy does not exist")
			}
		} else {
			logger.Debug("location used by active data objects")
		}
	}

	return nil
}

func (a *AccessSyncer) exportRaitoMaskListAllDoLocations(ctx context.Context, accessProvider *importer.AccessProvider, configMap *config.ConfigMap) (map[string]string, map[string][]string, map[string][]string, error) {
	var dataPolicies []string
	if accessProvider.ExternalId != nil {
		dataPolicies = strings.Split(*accessProvider.ExternalId, ",")
	}

	dataPolicyLocations := map[string]string{}

	for _, dataPolicy := range dataPolicies {
		dpNameSplit := strings.Split(dataPolicy, "/")
		dataPolicyLocations[dpNameSplit[3]] = dataPolicy
	}

	dataObjectLocationMap, deletedDataObjectLocationMap, err := a.datacatalogRepo.GetLocationsForDataObjects(ctx, configMap, accessProvider)
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
