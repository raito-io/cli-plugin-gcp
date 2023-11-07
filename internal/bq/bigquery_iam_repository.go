package bigquery

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	gcp_iam "cloud.google.com/go/iam"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

const (
	userPrefix           = "user:"
	serviceAccountPrefix = "serviceAccount:"
	groupPrefix          = "group:"
	specialGroupPrefix   = "special_group:"
)

var bqPolicyCache = make(map[string]iam.IAMPolicyContainer)
var resourceIds []string = nil

type bigQueryIamRepository struct {
}

func GetResourceIds(ctx context.Context, configMap *config.ConfigMap) ([]string, error) {
	if resourceIds != nil {
		return resourceIds, nil
	}

	repo := BigQueryRepository{}
	ids := []string{}

	datasets, err := repo.GetDataSets(ctx, configMap)

	if err != nil {
		return ids, err
	}

	for _, d := range datasets {
		ids = append(ids, d.ID)

		tables, err2 := repo.GetTables(ctx, configMap, BQEntity{ID: d.ID})

		if err2 != nil {
			return ids, err2
		}

		for _, t := range tables {
			if t.Type != "table" {
				continue
			}

			ids = append(ids, fmt.Sprintf("%s.%s", d.ID, t.ID))
		}
	}

	resourceIds = ids

	return ids, err
}

func (r *bigQueryIamRepository) getUserEntities(ctx context.Context, configMap *config.ConfigMap, id string, sa bool) ([]iam.UserEntity, error) {
	users := []iam.UserEntity{}
	iamPolicy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return users, err
	}

	prefix := userPrefix
	if sa {
		prefix = serviceAccountPrefix
	}

	for _, b := range iamPolicy.Service {
		if !strings.HasPrefix(b.Member, prefix) {
			continue
		}

		users = append(users, iam.UserEntity{
			Email:      strings.Replace(b.Member, prefix, "", 1),
			Name:       strings.Replace(b.Member, prefix, "", 1),
			ExternalId: b.Member,
		})
	}

	return users, nil
}

func (r *bigQueryIamRepository) GetUsers(ctx context.Context, configMap *config.ConfigMap, id string) ([]iam.UserEntity, error) {
	return r.getUserEntities(ctx, configMap, id, false)
}

func (r *bigQueryIamRepository) GetServiceAccounts(ctx context.Context, configMap *config.ConfigMap, id string) ([]iam.UserEntity, error) {
	return r.getUserEntities(ctx, configMap, id, true)
}

func (r *bigQueryIamRepository) GetGroups(ctx context.Context, configMap *config.ConfigMap, id string) ([]iam.GroupEntity, error) {
	groups := []iam.GroupEntity{}
	iamPolicy, err := r.GetIamPolicy(ctx, configMap, id)

	if err != nil {
		return groups, err
	}

	prefix := groupPrefix

	for _, b := range iamPolicy.Service {
		if !strings.HasPrefix(b.Member, prefix) {
			continue
		}

		groups = append(groups, iam.GroupEntity{
			Email:      strings.Replace(b.Member, prefix, "", 1),
			ExternalId: b.Member,
		})
	}

	return groups, nil
}

func (r *bigQueryIamRepository) GetIamPolicy(ctx context.Context, configMap *config.ConfigMap, id string) (iam.IAMPolicyContainer, error) {
	if policy, f := bqPolicyCache[id]; f {
		return policy, nil
	}

	common.Logger.Info(fmt.Sprintf("Fetching BigQuery IAM Policy for %s", id))
	parts := strings.Split(id, ".")

	policy := iam.IAMPolicyContainer{}
	var err error = nil

	if len(parts) == 1 { // dataset
		policy, err = r.getDataSetIamPolicy(ctx, configMap, id)
	} else if len(parts) == 2 { // table
		policy, err = r.getTableIamPolicy(ctx, configMap, id)
	}

	if err != nil && strings.Contains(err.Error(), "404") {
		common.Logger.Warn(fmt.Sprintf("Encountered error while fetching IAM Policy for %s: %s", id, err.Error()))
		return policy, nil
	}

	if err == nil {
		bqPolicyCache[id] = policy
	}

	return policy, err
}

func (r *bigQueryIamRepository) AddBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	return r.updateIamPolicy(ctx, configMap, member, role, id, false)
}
func (r *bigQueryIamRepository) RemoveBinding(ctx context.Context, configMap *config.ConfigMap, id, member, role string) error {
	return r.updateIamPolicy(ctx, configMap, member, role, id, true)
}

func (r *bigQueryIamRepository) getDataSetIamPolicy(ctx context.Context, configMap *config.ConfigMap, id string) (iam.IAMPolicyContainer, error) {
	conn, err := ConnectToBigQuery(configMap, ctx)

	if err != nil {
		return iam.IAMPolicyContainer{Service: nil}, err
	}
	defer conn.Close()

	ds := conn.Dataset(id)

	if ds == nil {
		return iam.IAMPolicyContainer{Service: nil}, err
	}

	dsMeta, err := ds.Metadata(ctx)

	if err != nil {
		return iam.IAMPolicyContainer{Service: nil}, err
	}

	bindings := []iam.IamBinding{}

	for _, a := range dsMeta.Access {
		if a.EntityType == bigquery.UserEmailEntity || a.EntityType == bigquery.GroupEmailEntity || a.EntityType == bigquery.SpecialGroupEntity {
			prefix := userPrefix

			if a.EntityType == bigquery.GroupEmailEntity {
				prefix = groupPrefix
			} else if a.EntityType == bigquery.SpecialGroupEntity {
				prefix = specialGroupPrefix
			} else if strings.Contains(a.Entity, "gserviceaccount") {
				prefix = serviceAccountPrefix
			}

			bindings = append(bindings, iam.IamBinding{
				Role:         getRoleForBQEntity(a.Role),
				Member:       prefix + a.Entity,
				Resource:     fmt.Sprintf("%s.%s", configMap.GetString(common.GcpProjectId), ds.DatasetID),
				ResourceType: "dataset",
			})
		}
	}

	return iam.IAMPolicyContainer{Service: bindings}, nil
}

func (r *bigQueryIamRepository) getTableIamPolicy(ctx context.Context, configMap *config.ConfigMap, id string) (iam.IAMPolicyContainer, error) {
	parts := strings.Split(id, ".")

	if len(parts) != 2 {
		return iam.IAMPolicyContainer{Service: nil}, fmt.Errorf("invalid table id: %s", id)
	}

	conn, err := ConnectToBigQuery(configMap, ctx)

	if err != nil {
		return iam.IAMPolicyContainer{Service: nil}, err
	}
	defer conn.Close()

	t := conn.Dataset(parts[0]).Table(parts[1])

	policy, err := t.IAM().Policy(ctx)

	if err != nil {
		return iam.IAMPolicyContainer{Service: nil}, err
	}

	bindings := []iam.IamBinding{}

	for _, role := range policy.Roles() {
		for _, m := range policy.Members(role) {
			bindings = append(bindings, iam.IamBinding{
				Role:         string(role),
				Member:       m,
				Resource:     fmt.Sprintf("%s.%s", configMap.GetString(common.GcpProjectId), id),
				ResourceType: "table",
			})
		}
	}

	return iam.IAMPolicyContainer{Service: bindings}, nil
}

func (r *bigQueryIamRepository) updateBQDatasetAccess(ctx context.Context, configMap *config.ConfigMap, member, role, datasetID string, revoke bool) error {
	conn, err := ConnectToBigQuery(configMap, ctx)

	if err != nil {
		return fmt.Errorf("error while updating dataset access bindings: %w", err)
	}

	defer conn.Close()

	ds := conn.Dataset(datasetID)
	meta, err := ds.Metadata(ctx)

	if err != nil {
		return fmt.Errorf("error while fetching data set metadata: %w", err)
	}

	update := bigquery.DatasetMetadataToUpdate{}
	entityType, entity, err := parseMember(member)

	if err != nil {
		return err
	}

	if !revoke {
		update.Access = meta.Access

		update.Access = append(update.Access, &bigquery.AccessEntry{
			Role:       getBQEntityForRole(role),
			EntityType: entityType,
			Entity:     entity},
		)
	} else {
		update.Access = []*bigquery.AccessEntry{}
		for _, a := range meta.Access {
			if a.Entity != entity && a.EntityType != entityType && a.Role != getBQEntityForRole(role) {
				update.Access = append(update.Access, a)
			}
		}
	}

	_, err = ds.Update(ctx, update, meta.ETag)

	if err != nil {
		return fmt.Errorf("error while updating access bindings for data set %s: %s", datasetID, err.Error())
	}

	return nil
}

func (r *bigQueryIamRepository) updateIamPolicy(ctx context.Context, configMap *config.ConfigMap, member, role, resource string, revoke bool) error {
	action := "Adding"
	if revoke {
		action = "Removing"
	}

	resource = strings.Replace(resource, configMap.GetString(common.GcpProjectId)+".", "", 1)

	common.Logger.Info(fmt.Sprintf("%s binding %s on %s for %s", action, role, member, resource))

	parts := strings.Split(resource, ".")

	if len(parts) == 1 { // dataset
		err := r.updateBQDatasetAccess(ctx, configMap, member, role, parts[0], revoke)
		if err != nil {
			return err
		}
	} else if len(parts) == 2 { // table
		err := r.updateBQTableAccess(ctx, configMap, member, role, parts[0], parts[1], revoke)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid resource: %s", resource)
	}

	return nil
}

func (r *bigQueryIamRepository) updateBQTableAccess(ctx context.Context, configMap *config.ConfigMap, member, role, datasetID, tableID string, revoke bool) error {
	conn, err := ConnectToBigQuery(configMap, ctx)

	if err != nil {
		return fmt.Errorf("error while updating table access bindings: %w", err)
	}

	defer conn.Close()

	table := conn.Dataset(datasetID).Table(tableID)
	policy, err := table.IAM().Policy(ctx)

	if err != nil {
		return fmt.Errorf("error while fetching table policy: %w", err)
	}

	iamRole := gcp_iam.RoleName(role)

	if revoke {
		policy.Remove(member, iamRole)
	} else {
		policy.Add(member, iamRole)
	}

	err = table.IAM().SetPolicy(ctx, policy)

	if err != nil {
		return fmt.Errorf("error while updating access bindings for table %s: %w", tableID, err)
	}

	return nil
}

func parseMember(m string) (bigquery.EntityType, string, error) {
	parts := strings.Split(m, ":")
	if len(parts) != 2 {
		return bigquery.UserEmailEntity, "", fmt.Errorf("invalid member format: %s", m)
	}

	switch parts[0] {
	case "user":
		return bigquery.UserEmailEntity, parts[1], nil
	case "group":
		return bigquery.GroupEmailEntity, parts[1], nil
	case "serviceAccount":
		return bigquery.UserEmailEntity, parts[1], nil
	}

	return bigquery.UserEmailEntity, "", fmt.Errorf("unknown member type: %s", m)
}
