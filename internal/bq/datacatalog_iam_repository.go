package bigquery

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"cloud.google.com/go/bigquery"
	datapolicies "cloud.google.com/go/bigquery/datapolicies/apiv1"
	"cloud.google.com/go/bigquery/datapolicies/apiv1/datapoliciespb"
	datacatalog "cloud.google.com/go/datacatalog/apiv1"
	"cloud.google.com/go/datacatalog/apiv1/datacatalogpb"
	"cloud.google.com/go/iam/apiv1/iampb"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/util/config"
	"github.com/raito-io/golang-set/set"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

const (
	fineGrainedReaderRole = "roles/datacatalog.categoryFineGrainedReader"
	taxonomy_prefix       = "raito_taxonomy_"

	idAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type dataCatalogIamRepository struct {
	bigQueryRepo *BigQueryRepository

	// Cache
	dataPolicies map[string]BQMaskingInformation
	datasetCache map[string]BQEntity
}

func (r *dataCatalogIamRepository) UpdateAccess(ctx context.Context, configMap *config.ConfigMap, maskingInformation *BQMaskingInformation, who *sync_to_target.WhoItem, deletedWho *sync_to_target.WhoItem) error {
	client, err := r.createPolicyTagClient(ctx, configMap)
	if err != nil {
		return err
	}

	defer client.Close()

	policy, err := client.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: maskingInformation.PolicyTag.FullName})
	if err != nil {
		return fmt.Errorf("failed to get iam policy of policy tag %q: %w", maskingInformation.PolicyTag.FullName, err)
	}

	updatedPolicy := &iampb.Policy{
		Version:      policy.Version,
		AuditConfigs: policy.AuditConfigs,
		Etag:         policy.Etag,
	}

	membersToDelete := set.NewSet[string]()

	if deletedWho != nil {
		membersToDelete.Add(parseWhoToMembers(deletedWho)...)
	}

	membersToAdd := parseWhoToMembers(who)

	updatedFineGrainedAccess := false

	for _, binding := range policy.Bindings {
		if binding.Role != fineGrainedReaderRole {
			updatedPolicy.Bindings = append(updatedPolicy.Bindings, binding)
		} else {
			newBinding := &iampb.Binding{
				Role: fineGrainedReaderRole,
			}

			for _, bindingMembers := range binding.Members {
				if !membersToDelete.Contains(bindingMembers) {
					newBinding.Members = append(newBinding.Members, bindingMembers)
				}
			}

			newBinding.Members = append(newBinding.Members, membersToAdd...)

			updatedPolicy.Bindings = append(updatedPolicy.Bindings, newBinding)
			updatedFineGrainedAccess = true
		}
	}

	if !updatedFineGrainedAccess {
		newBinding := &iampb.Binding{
			Role: fineGrainedReaderRole,
		}

		newBinding.Members = append(newBinding.Members, membersToAdd...)

		updatedPolicy.Bindings = append(updatedPolicy.Bindings, newBinding)
	}

	_, err = client.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{Policy: updatedPolicy, Resource: maskingInformation.PolicyTag.FullName})
	if err != nil {
		return fmt.Errorf("set fine grained reader role on %q: %w", maskingInformation.PolicyTag.FullName, err)
	}

	return nil
}

func (r *dataCatalogIamRepository) UpdatePolicyTag(ctx context.Context, configMap *config.ConfigMap, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *sync_to_target.AccessProvider, dataPolicyId string) (*BQMaskingInformation, error) {
	maskInfo, err := r.GetMaskingInformationForDataPolicy(ctx, configMap, dataPolicyId)
	if err != nil {
		return nil, err
	}

	if maskInfo == nil {
		return r.CreatePolicyTagWithDataPolicy(ctx, configMap, location, maskingType, ap)
	}

	tagClient, err := r.createPolicyTagClient(ctx, configMap)
	if err != nil {
		return nil, err
	}

	defer tagClient.Close()

	var displayName string

	if strings.HasPrefix(maskInfo.PolicyTag.Name, ap.Name+"_") {
		displayName = maskInfo.PolicyTag.Name
	} else {
		displayName = createTagDisplayname(ap)
	}

	_, err = tagClient.UpdatePolicyTag(ctx, &datacatalogpb.UpdatePolicyTagRequest{
		PolicyTag: &datacatalogpb.PolicyTag{
			Name:            maskInfo.PolicyTag.FullName,
			DisplayName:     displayName,
			Description:     ap.Description,
			ParentPolicyTag: maskInfo.PolicyTag.ParentTag,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("update policy tag %q: %w", maskInfo.PolicyTag.FullName, err)
	}

	maskInfo.PolicyTag.Name = displayName
	maskInfo.PolicyTag.Description = ap.Description

	return maskInfo, nil
}

func (r *dataCatalogIamRepository) ListDataPolicies(ctx context.Context, configMap *config.ConfigMap) (map[string]BQMaskingInformation, error) {
	if len(r.dataPolicies) == 0 {
		gcpProject := configMap.GetString(common.GcpProjectId)

		client, err := r.createDataPolicyClient(ctx, configMap)
		if err != nil {
			return nil, err
		}

		defer client.Close()

		locations := set.NewSet[string]()

		dataSets, err := r.bigQueryRepo.GetDataSets(ctx, configMap)
		if err != nil {
			return nil, err
		}

		for _, dataSet := range dataSets {
			locations.Add(dataSet.Location)
		}

		r.dataPolicies = make(map[string]BQMaskingInformation)
		for location := range locations {
			err = r.listDataPoliciesForLocation(ctx, gcpProject, strings.ToLower(location), configMap, r.dataPolicies)
			if err != nil {
				return nil, err
			}
		}
	}

	return r.dataPolicies, nil
}

func (r *dataCatalogIamRepository) listDataPoliciesForLocation(ctx context.Context, gcpProject string, location string, configMap *config.ConfigMap, result map[string]BQMaskingInformation) error {
	logger.Info(fmt.Sprintf("Listing policy tags for project %s in location %s", gcpProject, location))

	client, err2 := r.createDataPolicyClient(ctx, configMap)
	if err2 != nil {
		return fmt.Errorf("unabled to create data policy client: %w", err2)
	}

	defer client.Close()

	it := client.ListDataPolicies(ctx, &datapoliciespb.ListDataPoliciesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", gcpProject, location),
	})

	for {
		policy, err := it.Next()

		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return fmt.Errorf("list data policy iterator: %w", err)
		}

		if policy.DataPolicyType == datapoliciespb.DataPolicy_DATA_MASKING_POLICY {
			maskingInformation, err3 := r.createBqMaskingInformation(ctx, policy, configMap)
			if err3 != nil {
				return fmt.Errorf("create bq masking information: %w", err3)
			}

			if maskingInformation == nil {
				logger.Warn(fmt.Sprintf("Data policy %q is not associated with a policy tag. This data policy will be ignored.", policy.GetPolicyTag()))

				continue
			}

			keyRegex := regexp.MustCompile(`projects/\d*/`)
			key := keyRegex.ReplaceAllString(policy.GetPolicyTag(), fmt.Sprintf("projects/%s/", gcpProject))

			result[key] = *maskingInformation
		}
	}

	return nil
}

func (r *dataCatalogIamRepository) GetMaskingInformationForDataPolicy(ctx context.Context, configMap *config.ConfigMap, dataPolicyId string) (*BQMaskingInformation, error) {
	client, err := r.createDataPolicyClient(ctx, configMap)
	if err != nil {
		return nil, err
	}

	defer client.Close()

	dataPolicy, err := client.GetDataPolicy(ctx, &datapoliciespb.GetDataPolicyRequest{
		Name: dataPolicyId,
	})
	if err != nil {
		return nil, fmt.Errorf("get data policy %q: %w", dataPolicyId, err)
	}

	return r.createBqMaskingInformation(ctx, dataPolicy, configMap)
}

func (r *dataCatalogIamRepository) createBqMaskingInformation(ctx context.Context, policy *datapoliciespb.DataPolicy, configMap *config.ConfigMap) (*BQMaskingInformation, error) {
	maskType := policy.GetDataMaskingPolicy().GetPredefinedExpression()

	policyTagClient, err := r.createPolicyTagClient(ctx, configMap)
	if err != nil {
		return nil, err
	}

	policyTag, err := policyTagClient.GetPolicyTag(ctx, &datacatalogpb.GetPolicyTagRequest{Name: policy.GetPolicyTag()})

	var e *googleapi.Error
	if ok := errors.As(err, &e); ok && e.Code == 404 {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("get policy tag %q: %w", policy.GetPolicyTag(), err)
	}

	maskingInformation := &BQMaskingInformation{
		DataPolicy: BQDataPolicy{
			FullName:   policy.Name,
			PolicyType: maskType,
		},
		PolicyTag: BQPolicyTag{
			FullName:    policyTag.Name,
			Description: policyTag.Description,
			Name:        policyTag.DisplayName,
			ParentTag:   policyTag.ParentPolicyTag,
		},
	}

	return maskingInformation, nil
}

func (r *dataCatalogIamRepository) createPolicyTagClient(ctx context.Context, configMap *config.ConfigMap) (*datacatalog.PolicyTagManagerClient, error) {
	config, err := getConfig(configMap, admin.CloudPlatformScope)

	if err != nil {
		return nil, err
	}

	client, err := datacatalog.NewPolicyTagManagerRESTClient(ctx, option.WithHTTPClient(config.Client(ctx)))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (r *dataCatalogIamRepository) createDataPolicyClient(ctx context.Context, configMap *config.ConfigMap) (*datapolicies.DataPolicyClient, error) {
	config, err := getConfig(configMap, admin.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	client, err := datapolicies.NewDataPolicyRESTClient(ctx, option.WithHTTPClient(config.Client(ctx)))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (r *dataCatalogIamRepository) DeletePolicyAndTag(ctx context.Context, configMap *config.ConfigMap, policyTagId string) error {
	info, err := r.GetMaskingInformationForDataPolicy(ctx, configMap, policyTagId)

	var e *googleapi.Error
	if ok := errors.As(err, &e); ok && e.Code == 404 {
		logger.Warn(fmt.Sprintf("Cannot found data policy %q. Assuming data policy is already deleted", policyTagId))

		return nil
	} else if err != nil {
		return err
	}

	logger.Debug(fmt.Sprintf("Deleting data policy %s", info.DataPolicy.FullName))

	err = r.deleteDataPolicy(ctx, configMap, info.DataPolicy.FullName)
	if err != nil {
		return fmt.Errorf("delete data policy %q: %w", info.DataPolicy.FullName, err)
	}

	client, err := r.createPolicyTagClient(ctx, configMap)
	if err != nil {
		return err
	}

	defer client.Close()

	taxonomyId := info.PolicyTag.Taxonomy()

	logger.Debug(fmt.Sprintf("Get taxonomy: %s", taxonomyId))

	taxonomy, err := client.GetTaxonomy(ctx, &datacatalogpb.GetTaxonomyRequest{
		Name: taxonomyId,
	})

	if err != nil {
		return fmt.Errorf("get taxonomy %q: %w", taxonomyId, err)
	}

	if strings.HasPrefix(taxonomy.GetDisplayName(), taxonomy_prefix) {
		logger.Debug(fmt.Sprintf("Delete policyTag: %s", info.PolicyTag.FullName))

		err = r.deletePolicyTag(ctx, configMap, info.PolicyTag.FullName)
		if err != nil {
			return fmt.Errorf("delete policy tag %q: %w", info.PolicyTag.FullName, err)
		}

		taxonomy, err = client.GetTaxonomy(ctx, &datacatalogpb.GetTaxonomyRequest{
			Name: taxonomyId,
		})
		if err != nil {
			return fmt.Errorf("reload taxonomy %q: %w", taxonomyId, err)
		}

		if taxonomy.GetPolicyTagCount() == 0 {
			logger.Debug(fmt.Sprintf("Delete taxonomy: %s", taxonomy.GetName()))

			err = client.DeleteTaxonomy(ctx, &datacatalogpb.DeleteTaxonomyRequest{
				Name: taxonomy.GetName(),
			})

			if err != nil {
				return fmt.Errorf("delete taxonomy %q: %w", taxonomy.GetName(), err)
			}
		}
	}

	return nil
}

func (r *dataCatalogIamRepository) deletePolicyTag(ctx context.Context, configMap *config.ConfigMap, id string) error {
	client, err := r.createPolicyTagClient(ctx, configMap)
	if err != nil {
		return err
	}

	defer client.Close()

	return client.DeletePolicyTag(ctx, &datacatalogpb.DeletePolicyTagRequest{Name: id})
}

func (r *dataCatalogIamRepository) deleteDataPolicy(ctx context.Context, configMap *config.ConfigMap, id string) error {
	client, err := r.createDataPolicyClient(ctx, configMap)
	if err != nil {
		return err
	}

	defer client.Close()

	return client.DeleteDataPolicy(ctx, &datapoliciespb.DeleteDataPolicyRequest{Name: id})
}
func (r *dataCatalogIamRepository) GetFineGrainedReaderMembers(ctx context.Context, configMap *config.ConfigMap, tagId string) ([]string, error) {
	tagClient, err := r.createPolicyTagClient(ctx, configMap)
	if err != nil {
		return nil, err
	}

	defer tagClient.Close()

	logger.Debug(fmt.Sprintf("Getting iam policy for policy tag %s", tagId))

	iamPolicy, err := tagClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: tagId,
	})

	if err != nil {
		return nil, fmt.Errorf("get iam policy for policy tag %q: %w", tagId, err)
	}

	var result []string

	for _, binding := range iamPolicy.Bindings {
		logger.Debug(fmt.Sprintf("Binding for %q with role %q: %v", tagId, binding.Role, binding.Members))

		if binding.Role == fineGrainedReaderRole {
			result = binding.Members
		}
	}

	return result, nil
}

func (r *dataCatalogIamRepository) CreatePolicyTagWithDataPolicy(ctx context.Context, configMap *config.ConfigMap, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *sync_to_target.AccessProvider) (_ *BQMaskingInformation, err error) {
	policyTagClient, err := r.createPolicyTagClient(ctx, configMap)
	if err != nil {
		return nil, err
	}

	defer policyTagClient.Close()

	location = strings.ToLower(location)

	// 1. Create taxonomy if not exists
	taxonomyName := taxonomy_prefix + location
	parent := fmt.Sprintf("projects/%s/locations/%s", configMap.GetString(common.GcpProjectId), location)

	taxIt := policyTagClient.ListTaxonomies(ctx, &datacatalogpb.ListTaxonomiesRequest{
		Parent: parent,
	})

	var taxonomy *datacatalogpb.Taxonomy

	for {
		tmpTax, err2 := taxIt.Next()

		if errors.Is(err2, iterator.Done) {
			break
		} else if err2 != nil {
			logger.Error(fmt.Sprintf("failed to list taxonomies with parent %q: %s", parent, err2.Error()))

			return nil, fmt.Errorf("list taxonomies: %w", err2)
		}

		if tmpTax.DisplayName == taxonomyName {
			if taxonomy != nil {
				return nil, fmt.Errorf("taxonomy %s already found before", taxonomyName)
			}

			taxonomy = tmpTax
		}
	}

	if taxonomy == nil {
		taxonomy, err = policyTagClient.CreateTaxonomy(ctx, &datacatalogpb.CreateTaxonomyRequest{
			Taxonomy: &datacatalogpb.Taxonomy{
				DisplayName:          taxonomyName,
				Description:          fmt.Sprintf("Raito managed taxonomy for location %s", location),
				ActivatedPolicyTypes: []datacatalogpb.Taxonomy_PolicyType{datacatalogpb.Taxonomy_FINE_GRAINED_ACCESS_CONTROL},
			},
			Parent: parent,
		})

		if err != nil {
			return nil, fmt.Errorf("create taxonomy %q: %w", taxonomyName, err)
		}
	}

	// 2. Create policy tag
	displayName := createTagDisplayname(ap)

	policyTag, err := policyTagClient.CreatePolicyTag(ctx, &datacatalogpb.CreatePolicyTagRequest{
		Parent: taxonomy.Name,
		PolicyTag: &datacatalogpb.PolicyTag{
			DisplayName: displayName,
			Description: ap.Description,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("create policy tag %q in taxonomy %q: %w", displayName, taxonomy.Name, err)
	}

	defer func() {
		if err != nil {
			policyTagClient.DeletePolicyTag(ctx, &datacatalogpb.DeletePolicyTagRequest{Name: policyTag.Name}) //nolint:errcheck
		}
	}()

	// 3. Create data policy
	dataPolicyId := gonanoid.MustGenerate(idAlphabet, 24) // Must be unique in the project and location

	dataPolicyClient, err := r.createDataPolicyClient(ctx, configMap)
	if err != nil {
		return nil, err
	}

	dataPolicy, err := dataPolicyClient.CreateDataPolicy(ctx, &datapoliciespb.CreateDataPolicyRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", configMap.GetString(common.GcpProjectId), location),
		DataPolicy: &datapoliciespb.DataPolicy{
			DataPolicyId:   dataPolicyId,
			DataPolicyType: datapoliciespb.DataPolicy_DATA_MASKING_POLICY,
			MatchingLabel: &datapoliciespb.DataPolicy_PolicyTag{
				PolicyTag: policyTag.Name,
			},
			Policy: &datapoliciespb.DataPolicy_DataMaskingPolicy{
				DataMaskingPolicy: &datapoliciespb.DataMaskingPolicy{
					MaskingExpression: &datapoliciespb.DataMaskingPolicy_PredefinedExpression_{
						PredefinedExpression: maskingType,
					},
				},
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("create data policy %q in policy tag %q: %w", dataPolicyId, policyTag.Name, err)
	}

	return r.createBqMaskingInformation(ctx, dataPolicy, configMap)
}

func createTagDisplayname(ap *sync_to_target.AccessProvider) string {
	displayName := ap.Name + "_" + gonanoid.MustGenerate(idAlphabet, 8) // Must be unique in taxonomy
	return displayName
}

func (r *dataCatalogIamRepository) GetLocationsForDataObjects(ctx context.Context, configMap *config.ConfigMap, ap *sync_to_target.AccessProvider) (map[string]string, map[string]string, error) {
	datasets, err := r.getDataSets(ctx, configMap)
	if err != nil {
		return nil, nil, err
	}

	dos := make(map[string]string)
	deletedDos := make(map[string]string)

	for _, whatItem := range ap.What {
		doNameSplit := strings.SplitN(whatItem.DataObject.FullName, ".", 3)
		if dsInfo, found := datasets[strings.Join(doNameSplit[0:2], ".")]; found {
			dos[whatItem.DataObject.FullName] = strings.ToLower(dsInfo.Location)
		} else {
			return nil, nil, fmt.Errorf("data object %s not found", whatItem.DataObject.FullName)
		}
	}

	for _, whatItem := range ap.DeleteWhat {
		doNameSplit := strings.SplitN(whatItem.DataObject.FullName, ".", 3)
		if dsInfo, found := datasets[strings.Join(doNameSplit[0:2], ".")]; found {
			deletedDos[whatItem.DataObject.FullName] = strings.ToLower(dsInfo.Location)
		} else {
			return nil, nil, fmt.Errorf("deleted data object %s not found", whatItem.DataObject.FullName)
		}
	}

	return dos, deletedDos, nil
}

type tableMaskUpdate struct {
	ColumnsToAddMask    set.Set[string]
	ColumnsToRemoveMask set.Set[string]
}

func (r *dataCatalogIamRepository) UpdateWhatOfDataPolicy(ctx context.Context, configMap *config.ConfigMap, policy *BQMaskingInformation, dataObjects []string, deletedDataObjects []string) error {
	columnsToUpdatePerTable := make(map[string]tableMaskUpdate)

	parseColumnsToUpdatePerTable := func(dos []string, toRemove bool) {
		for _, do := range dos {
			doNameSplit := strings.Split(do, ".")
			tableName := strings.Join(doNameSplit[0:len(doNameSplit)-1], ".")
			columnName := doNameSplit[len(doNameSplit)-1]

			if maskUpdates, found := columnsToUpdatePerTable[tableName]; found {
				if toRemove {
					maskUpdates.ColumnsToRemoveMask.Add(columnName)
					columnsToUpdatePerTable[tableName] = maskUpdates
				} else {
					maskUpdates.ColumnsToAddMask.Add(columnName)
					columnsToUpdatePerTable[tableName] = maskUpdates
				}
			} else {
				if toRemove {
					columnsToUpdatePerTable[tableName] = tableMaskUpdate{
						ColumnsToRemoveMask: set.NewSet(columnName),
					}
				} else {
					columnsToUpdatePerTable[tableName] = tableMaskUpdate{
						ColumnsToAddMask: set.NewSet(columnName),
					}
				}
			}
		}
	}

	parseColumnsToUpdatePerTable(dataObjects, false)
	parseColumnsToUpdatePerTable(deletedDataObjects, true)

	client, err := ConnectToBigQuery(configMap, ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	for table, maskUpdates := range columnsToUpdatePerTable {
		nameSplit := strings.Split(table, ".")
		ds := client.Dataset(nameSplit[1])
		bqTable := ds.Table(nameSplit[2])

		metadata, err := bqTable.Metadata(ctx)
		if err != nil {
			return fmt.Errorf("loading metadata for %q: %w", nameSplit[1:3], err)
		}

		var schemaUpdate []*bigquery.FieldSchema

		for _, column := range metadata.Schema {
			if maskUpdates.ColumnsToAddMask.Contains(column.Name) {
				if column.PolicyTags == nil {
					column.PolicyTags = &bigquery.PolicyTagList{Names: []string{policy.PolicyTag.FullName}}
				}
			} else if maskUpdates.ColumnsToRemoveMask.Contains(column.Name) {
				if column.PolicyTags != nil {
					column.PolicyTags = &bigquery.PolicyTagList{Names: []string{}}
				}
			}

			schemaUpdate = append(schemaUpdate, column)
		}

		_, err = bqTable.Update(ctx, bigquery.TableMetadataToUpdate{
			Schema: schemaUpdate,
		}, metadata.ETag)

		if err != nil {
			return fmt.Errorf("update schema of table %q: %w", table, err)
		}
	}

	return nil
}

func (r *dataCatalogIamRepository) getDataSets(ctx context.Context, configMap *config.ConfigMap) (map[string]BQEntity, error) {
	if len(r.datasetCache) == 0 {
		dataSets, err := r.bigQueryRepo.GetDataSets(ctx, configMap)
		if err != nil {
			return nil, err
		}

		r.datasetCache = make(map[string]BQEntity)

		for _, dataSet := range dataSets {
			r.datasetCache[dataSet.FullName] = dataSet
		}
	}

	return r.datasetCache, nil
}

func parseWhoToMembers(who *sync_to_target.WhoItem) []string {
	members := make([]string, 0, len(who.Users)+len(who.Groups))

	for _, m := range who.Users {
		if strings.Contains(m, "gserviceaccount.com") {
			members = append(members, "serviceAccount:"+m)
		} else {
			members = append(members, "user:"+m)
		}
	}

	for _, m := range who.Groups {
		members = append(members, "group:"+m)
	}

	return members
}
