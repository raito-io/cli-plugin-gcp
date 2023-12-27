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
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

const (
	fineGrainedReaderRole = "roles/datacatalog.categoryFineGrainedReader"
	taxonomy_prefix       = "raito_taxonomy_"

	idAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

//go:generate go run github.com/vektra/mockery/v2 --name=dataCatalogBqRepository --with-expecter --inpackage
type dataCatalogBqRepository interface {
	ListDataSets(ctx context.Context, parent *org.GcpOrgEntity, fn func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery.Dataset) error) error
	Project() *org.GcpOrgEntity
}

type DataCatalogRepository struct {
	bigQueryRepo     dataCatalogBqRepository
	policyTagClient  *datacatalog.PolicyTagManagerClient
	dataPolicyClient *datapolicies.DataPolicyClient
	bigQueryClient   *bigquery.Client

	projectId string

	// Cache
	dataPolicies map[string]BQMaskingInformation
	datasetCache map[string]org.GcpOrgEntity
}

func NewDataCatalogRepository(repository dataCatalogBqRepository, tagClient *datacatalog.PolicyTagManagerClient, dataPolicyClient *datapolicies.DataPolicyClient, bqClient *bigquery.Client, configMap *config.ConfigMap) *DataCatalogRepository {
	return &DataCatalogRepository{
		bigQueryRepo:     repository,
		policyTagClient:  tagClient,
		dataPolicyClient: dataPolicyClient,
		bigQueryClient:   bqClient,

		projectId: configMap.GetString(common.GcpProjectId),

		dataPolicies: make(map[string]BQMaskingInformation),
		datasetCache: make(map[string]org.GcpOrgEntity),
	}
}

func (r *DataCatalogRepository) UpdateAccess(ctx context.Context, maskingInformation *BQMaskingInformation, who *sync_to_target.WhoItem, deletedWho *sync_to_target.WhoItem) error {
	policy, err := r.policyTagClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: maskingInformation.PolicyTag.FullName})
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

	_, err = r.policyTagClient.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{Policy: updatedPolicy, Resource: maskingInformation.PolicyTag.FullName})
	if err != nil {
		return fmt.Errorf("set fine grained reader role on %q: %w", maskingInformation.PolicyTag.FullName, err)
	}

	return nil
}

func (r *DataCatalogRepository) UpdatePolicyTag(ctx context.Context, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *sync_to_target.AccessProvider, dataPolicyId string) (*BQMaskingInformation, error) {
	maskInfo, err := r.GetMaskingInformationForDataPolicy(ctx, dataPolicyId)
	if err != nil {
		return nil, err
	}

	if maskInfo == nil {
		return r.CreatePolicyTagWithDataPolicy(ctx, location, maskingType, ap)
	}

	var displayName string

	if strings.HasPrefix(maskInfo.PolicyTag.Name, ap.Name+"_") {
		displayName = maskInfo.PolicyTag.Name
	} else {
		displayName = createTagDisplayname(ap)
	}

	_, err = r.policyTagClient.UpdatePolicyTag(ctx, &datacatalogpb.UpdatePolicyTagRequest{
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

func (r *DataCatalogRepository) ListDataPolicies(ctx context.Context) (map[string]BQMaskingInformation, error) {
	if len(r.dataPolicies) == 0 {
		locations := set.NewSet[string]()

		err := r.bigQueryRepo.ListDataSets(ctx, r.bigQueryRepo.Project(), func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery.Dataset) error {
			locations.Add(entity.Location)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("list dataset: %w", err)
		}

		r.dataPolicies = make(map[string]BQMaskingInformation)
		for location := range locations {
			err = r.listDataPoliciesForLocation(ctx, strings.ToLower(location), r.dataPolicies)
			if err != nil {
				return nil, err
			}
		}
	}

	return r.dataPolicies, nil
}

func (r *DataCatalogRepository) listDataPoliciesForLocation(ctx context.Context, location string, result map[string]BQMaskingInformation) error {
	common.Logger.Info(fmt.Sprintf("Listing policy tags for project %s in location %s", r.projectId, location))

	it := r.dataPolicyClient.ListDataPolicies(ctx, &datapoliciespb.ListDataPoliciesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", r.projectId, location),
	})

	for {
		policy, err := it.Next()

		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return fmt.Errorf("list data policy iterator: %w", err)
		}

		if policy.DataPolicyType == datapoliciespb.DataPolicy_DATA_MASKING_POLICY {
			maskingInformation, err3 := r.createBqMaskingInformation(ctx, policy)
			if err3 != nil {
				return fmt.Errorf("create bq masking information: %w", err3)
			}

			if maskingInformation == nil {
				common.Logger.Warn(fmt.Sprintf("Data policy %q is not associated with a policy tag. This data policy will be ignored.", policy.GetPolicyTag()))

				continue
			}

			keyRegex := regexp.MustCompile(`projects/\d*/`)
			key := keyRegex.ReplaceAllString(policy.GetPolicyTag(), fmt.Sprintf("projects/%s/", r.projectId))

			result[key] = *maskingInformation
		}
	}

	return nil
}

func (r *DataCatalogRepository) GetMaskingInformationForDataPolicy(ctx context.Context, dataPolicyId string) (*BQMaskingInformation, error) {
	dataPolicy, err := r.dataPolicyClient.GetDataPolicy(ctx, &datapoliciespb.GetDataPolicyRequest{
		Name: dataPolicyId,
	})
	if err != nil {
		return nil, fmt.Errorf("get data policy %q: %w", dataPolicyId, err)
	}

	return r.createBqMaskingInformation(ctx, dataPolicy)
}

func (r *DataCatalogRepository) createBqMaskingInformation(ctx context.Context, policy *datapoliciespb.DataPolicy) (*BQMaskingInformation, error) {
	maskType := policy.GetDataMaskingPolicy().GetPredefinedExpression()

	policyTag, err := r.policyTagClient.GetPolicyTag(ctx, &datacatalogpb.GetPolicyTagRequest{Name: policy.GetPolicyTag()})

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

func (r *DataCatalogRepository) DeletePolicyAndTag(ctx context.Context, policyTagId string) error {
	info, err := r.GetMaskingInformationForDataPolicy(ctx, policyTagId)

	var e *googleapi.Error
	if ok := errors.As(err, &e); ok && e.Code == 404 {
		common.Logger.Warn(fmt.Sprintf("Cannot found data policy %q. Assuming data policy is already deleted", policyTagId))

		return nil
	} else if err != nil {
		return err
	}

	common.Logger.Debug(fmt.Sprintf("Deleting data policy %s", info.DataPolicy.FullName))

	err = r.deleteDataPolicy(ctx, info.DataPolicy.FullName)
	if err != nil {
		return fmt.Errorf("delete data policy %q: %w", info.DataPolicy.FullName, err)
	}

	taxonomyId := info.PolicyTag.Taxonomy()

	common.Logger.Debug(fmt.Sprintf("Get taxonomy: %s", taxonomyId))

	taxonomy, err := r.policyTagClient.GetTaxonomy(ctx, &datacatalogpb.GetTaxonomyRequest{
		Name: taxonomyId,
	})

	if err != nil {
		return fmt.Errorf("get taxonomy %q: %w", taxonomyId, err)
	}

	if strings.HasPrefix(taxonomy.GetDisplayName(), taxonomy_prefix) {
		common.Logger.Debug(fmt.Sprintf("Delete policyTag: %s", info.PolicyTag.FullName))

		err = r.deletePolicyTag(ctx, info.PolicyTag.FullName)
		if err != nil {
			return fmt.Errorf("delete policy tag %q: %w", info.PolicyTag.FullName, err)
		}

		taxonomy, err = r.policyTagClient.GetTaxonomy(ctx, &datacatalogpb.GetTaxonomyRequest{
			Name: taxonomyId,
		})
		if err != nil {
			return fmt.Errorf("reload taxonomy %q: %w", taxonomyId, err)
		}

		if taxonomy.GetPolicyTagCount() == 0 {
			common.Logger.Debug(fmt.Sprintf("Delete taxonomy: %s", taxonomy.GetName()))

			err = r.policyTagClient.DeleteTaxonomy(ctx, &datacatalogpb.DeleteTaxonomyRequest{
				Name: taxonomy.GetName(),
			})

			if err != nil {
				return fmt.Errorf("delete taxonomy %q: %w", taxonomy.GetName(), err)
			}
		}
	}

	return nil
}

func (r *DataCatalogRepository) deletePolicyTag(ctx context.Context, id string) error {
	err := r.policyTagClient.DeletePolicyTag(ctx, &datacatalogpb.DeletePolicyTagRequest{Name: id})
	if err != nil {
		return fmt.Errorf("delete policy tag %q: %w", id, err)
	}

	return nil
}

func (r *DataCatalogRepository) deleteDataPolicy(ctx context.Context, id string) error {
	err := r.dataPolicyClient.DeleteDataPolicy(ctx, &datapoliciespb.DeleteDataPolicyRequest{Name: id})
	if err != nil {
		return fmt.Errorf("delete data policy %q: %w", id, err)
	}

	return nil
}
func (r *DataCatalogRepository) GetFineGrainedReaderMembers(ctx context.Context, tagId string) ([]string, error) {
	common.Logger.Debug(fmt.Sprintf("Getting iam policy for policy tag %s", tagId))

	iamPolicy, err := r.policyTagClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: tagId,
	})

	if err != nil {
		return nil, fmt.Errorf("get iam policy for policy tag %q: %w", tagId, err)
	}

	var result []string

	for _, binding := range iamPolicy.Bindings {
		common.Logger.Debug(fmt.Sprintf("Binding for %q with role %q: %v", tagId, binding.Role, binding.Members))

		if binding.Role == fineGrainedReaderRole {
			result = binding.Members
		}
	}

	return result, nil
}

func (r *DataCatalogRepository) CreatePolicyTagWithDataPolicy(ctx context.Context, location string, maskingType datapoliciespb.DataMaskingPolicy_PredefinedExpression, ap *sync_to_target.AccessProvider) (_ *BQMaskingInformation, err error) {
	location = strings.ToLower(location)

	// 1. Create taxonomy if not exists
	taxonomyName := taxonomy_prefix + location
	parent := fmt.Sprintf("projects/%s/locations/%s", r.projectId, location)

	taxIt := r.policyTagClient.ListTaxonomies(ctx, &datacatalogpb.ListTaxonomiesRequest{
		Parent: parent,
	})

	var taxonomy *datacatalogpb.Taxonomy

	for {
		tmpTax, err2 := taxIt.Next()

		if errors.Is(err2, iterator.Done) {
			break
		} else if err2 != nil {
			common.Logger.Error(fmt.Sprintf("failed to list taxonomies with parent %q: %s", parent, err2.Error()))

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
		taxonomy, err = r.policyTagClient.CreateTaxonomy(ctx, &datacatalogpb.CreateTaxonomyRequest{
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

	policyTag, err := r.policyTagClient.CreatePolicyTag(ctx, &datacatalogpb.CreatePolicyTagRequest{
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
			r.policyTagClient.DeletePolicyTag(ctx, &datacatalogpb.DeletePolicyTagRequest{Name: policyTag.Name}) //nolint:errcheck
		}
	}()

	// 3. Create data policy
	dataPolicyId := gonanoid.MustGenerate(idAlphabet, 24) // Must be unique in the project and location

	dataPolicy, err := r.dataPolicyClient.CreateDataPolicy(ctx, &datapoliciespb.CreateDataPolicyRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", r.projectId, location),
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

	return r.createBqMaskingInformation(ctx, dataPolicy)
}

func createTagDisplayname(ap *sync_to_target.AccessProvider) string {
	displayName := validSqlName(ap.NamingHint) + "_" + gonanoid.MustGenerate(idAlphabet, 8) // Must be unique in taxonomy
	return displayName
}

func (r *DataCatalogRepository) GetLocationsForDataObjects(ctx context.Context, ap *sync_to_target.AccessProvider) (map[string]string, map[string]string, error) {
	datasets, err := r.getDataSets(ctx)
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

func (r *DataCatalogRepository) UpdateWhatOfDataPolicy(ctx context.Context, policy *BQMaskingInformation, dataObjects []string, deletedDataObjects []string) error {
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

	for table, maskUpdates := range columnsToUpdatePerTable {
		nameSplit := strings.Split(table, ".")
		ds := r.bigQueryClient.Dataset(nameSplit[1])
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

func (r *DataCatalogRepository) getDataSets(ctx context.Context) (map[string]org.GcpOrgEntity, error) {
	if len(r.datasetCache) == 0 {
		r.datasetCache = make(map[string]org.GcpOrgEntity)

		err := r.bigQueryRepo.ListDataSets(ctx, r.bigQueryRepo.Project(), func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery.Dataset) error {
			r.datasetCache[entity.FullName] = *entity

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("list datasets: %w", err)
		}
	}

	return r.datasetCache, nil
}

func parseWhoToMembers(who *sync_to_target.WhoItem) []string {
	if who == nil {
		return nil
	}

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
