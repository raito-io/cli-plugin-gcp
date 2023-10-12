package gcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/raito-io/golang-set/set"

	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/access_provider"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/wrappers"

	"github.com/raito-io/cli-plugin-gcp/gcp/common"
	"github.com/raito-io/cli-plugin-gcp/gcp/iam"

	exporter "github.com/raito-io/cli/base/access_provider/sync_from_target"
	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/util/config"
)

type AccessSyncer struct {
	iamServiceProvider   func(configMap *config.ConfigMap) iam.IAMService
	raitoManagedBindings []iam.IamBinding
	getDSMetadata        func(ctx context.Context, configMap *config.ConfigMap) (*data_source.MetaData, error)
}

func NewDataAccessSyncer() *AccessSyncer {
	return &AccessSyncer{
		iamServiceProvider: newIamServiceProvider,
		getDSMetadata:      GetDataSourceMetaData,
	}
}

func (a *AccessSyncer) WithDataSourceMetadataFetcher(getDSMetadata func(ctx context.Context, configMap *config.ConfigMap) (*data_source.MetaData, error)) *AccessSyncer {
	a.getDSMetadata = getDSMetadata
	return a
}

func (a *AccessSyncer) WithIAMServiceProvider(provider func(configMap *config.ConfigMap) iam.IAMService) *AccessSyncer {
	a.iamServiceProvider = provider
	return a
}

func (a *AccessSyncer) SyncAccessProvidersFromTarget(ctx context.Context, accessProviderHandler wrappers.AccessProviderHandler, configMap *config.ConfigMap) error {
	bindings, err := a.iamServiceProvider(configMap).GetIAMPolicyBindings(ctx, configMap)

	if err != nil || len(bindings) == 0 {
		return err
	}

	aps, err := a.ConvertBindingsToAccessProviders(ctx, configMap, bindings)

	if err != nil || len(aps) == 0 {
		return err
	}

	for _, ap := range aps {
		err = accessProviderHandler.AddAccessProviders(ap)

		if err != nil {
			return err
		}
	}

	return nil
}

func (a *AccessSyncer) ConvertBindingsToAccessProviders(ctx context.Context, configMap *config.ConfigMap, bindings []iam.IamBinding) ([]*exporter.AccessProvider, error) {
	rolesToGroupByIdentity := set.NewSet[string]()

	toGroupConfig := configMap.GetString(common.GcpRolesToGroupByIdentity)
	if toGroupConfig != "" {
		for _, role := range strings.Split(toGroupConfig, ",") {
			rolesToGroupByIdentity.Add(role)
		}
	}

	accessProviderMap := make(map[string]*exporter.AccessProvider)
	specialGroupAccessProviderMap := make(map[string]*exporter.AccessProvider)
	groupedByIdentityAccesProviderMap := make(map[string]*exporter.AccessProvider)

	projectOwnersWho, projectEditorWho, projectReaderWho, err := a.projectRolesWhoItem(ctx, configMap)
	if err != nil {
		return nil, err
	}

	for _, binding := range bindings {
		if strings.EqualFold(binding.ResourceType, iam.Organization.String()) {
			binding.Resource = GetOrgDataObjectName(configMap)
		}

		managed, err2 := a.isRaitoManagedBinding(ctx, configMap, binding)

		if err2 != nil {
			return nil, err2
		}

		if configMap.GetBoolWithDefault(common.ExcludeNonAplicablePermissions, true) && !managed {
			common.Logger.Info(fmt.Sprintf("Skipping role %s for %s on %s %s as it is not an applicable permission for this datasource and %s is false", binding.Role, binding.Member, binding.Resource, binding.ResourceType, common.ExcludeNonAplicablePermissions))
			continue
		}

		ignore := false

		for _, ignoredBinding := range a.raitoManagedBindings {
			if ignoredBinding.Equals(binding) {
				ignore = true
				break
			}
		}

		if ignore {
			continue
		}

		if strings.HasPrefix(binding.Member, "special_group:") {
			a.generateSpecialGroupOwnerAccessProvider(binding, specialGroupAccessProviderMap, projectOwnersWho, projectEditorWho, projectReaderWho)
		} else if rolesToGroupByIdentity.Contains(binding.Role) {
			a.generateGroupedByIdentityAcccessProvider(binding, groupedByIdentityAccesProviderMap)
		} else {
			a.generateAccessProvider(binding, accessProviderMap, managed)
		}
	}

	aps := make([]*exporter.AccessProvider, 0)
	for _, ap := range accessProviderMap {
		aps = append(aps, ap)
	}

	for _, specialGroupAp := range specialGroupAccessProviderMap {
		aps = append(aps, specialGroupAp)
	}

	for _, groupedByIdentityAp := range groupedByIdentityAccesProviderMap {
		aps = append(aps, groupedByIdentityAp)
	}

	return aps, nil
}

func (a *AccessSyncer) generateAccessProvider(binding iam.IamBinding, accessProviderMap map[string]*exporter.AccessProvider, managed bool) {
	apName := fmt.Sprintf("%s_%s_%s", binding.ResourceType, binding.Resource, strings.Replace(binding.Role, "/", "_", -1))

	if _, f := accessProviderMap[apName]; !f {
		accessProviderMap[apName] = &exporter.AccessProvider{
			ExternalId:        apName,
			Name:              apName,
			NamingHint:        generateNamingHint(apName),
			NotInternalizable: !managed,
			WhoLocked:         ptr.Bool(false),
			WhatLocked:        ptr.Bool(true),
			WhatLockedReason:  ptr.String("This Access Control was imported from GCP and can only cover 1 Data Object. If you want a GCP Access Control with multiple Data Objects, you can create a new one in Raito"),
			Action:            exporter.Grant,
			NameLocked:        ptr.Bool(false),
			DeleteLocked:      ptr.Bool(false),
			ActualName:        apName,
			Type:              ptr.String(access_provider.AclSet),
			What: []exporter.WhatItem{
				{
					DataObject: &data_source.DataObjectReference{
						FullName: binding.Resource,
						Type:     binding.ResourceType,
					},
					Permissions: []string{binding.Role},
				},
			},
			Who: &exporter.WhoItem{
				Users:           make([]string, 0),
				Groups:          make([]string, 0),
				AccessProviders: make([]string, 0),
			},
		}
	}

	a.addBindingMemberToAccessProvider(binding.Member, accessProviderMap[apName])
}

func (a *AccessSyncer) addBindingMemberToAccessProvider(bindingMember string, accessProvider *exporter.AccessProvider) {
	if strings.HasPrefix(bindingMember, "user:") || strings.HasPrefix(bindingMember, "serviceAccount:") {
		accessProvider.Who.Users = append(accessProvider.Who.Users, strings.Split(bindingMember, ":")[1])
	} else if strings.HasPrefix(bindingMember, "group:") {
		accessProvider.Who.Groups = append(accessProvider.Who.Groups, strings.Split(bindingMember, ":")[1])
	}
}

func (a *AccessSyncer) generateGroupedByIdentityAcccessProvider(binding iam.IamBinding, groupedByIdentityAccesProviderMap map[string]*exporter.AccessProvider) {
	member := binding.Member

	memberName := strings.Replace(member, ":", " ", -1)
	apName := fmt.Sprintf("Grouped permissions for %s", memberName)

	groupedByIdentityAccesProvider, ok := groupedByIdentityAccesProviderMap[apName]

	if !ok {
		groupedByIdentityAccesProvider = &exporter.AccessProvider{
			ExternalId:        apName,
			Name:              apName,
			NamingHint:        generateNamingHint(apName),
			NotInternalizable: true,
			Action:            exporter.Grant,
			ActualName:        apName,
			Type:              ptr.String(access_provider.AclSet),
			Who:               &exporter.WhoItem{},
		}

		a.addBindingMemberToAccessProvider(binding.Member, groupedByIdentityAccesProvider)
	}

	groupedByIdentityAccesProvider.What = append(groupedByIdentityAccesProvider.What, exporter.WhatItem{
		DataObject: &data_source.DataObjectReference{
			FullName: binding.Resource,
			Type:     binding.ResourceType,
		},
		Permissions: []string{binding.Role},
	})

	groupedByIdentityAccesProviderMap[apName] = groupedByIdentityAccesProvider
}

func (a *AccessSyncer) generateSpecialGroupOwnerAccessProvider(binding iam.IamBinding, specialGroupAccessProviderMap map[string]*exporter.AccessProvider, projectOwnersWho *exporter.WhoItem, projectEditorsWho *exporter.WhoItem, projectReadersWho *exporter.WhoItem) {
	mapping := map[string]struct {
		whoItem  *exporter.WhoItem
		roleName string
	}{
		"roles/bigquery.dataViewer": {
			whoItem:  projectReadersWho,
			roleName: "Viewer",
		},
		"roles/bigquery.dataEditor": {
			whoItem:  projectEditorsWho,
			roleName: "Editor",
		},
		"roles/bigquery.dataOwner": {
			whoItem:  projectOwnersWho,
			roleName: "Owner",
		},
	}

	r, ok := mapping[binding.Role]
	if !ok {
		common.Logger.Warn(fmt.Sprintf("Skipping role %s for special group binding %+v", binding.Role, binding))
		return
	}

	apName := fmt.Sprintf("Project %s Mapping", r.roleName)
	specialGroupAccessProvider, ok := specialGroupAccessProviderMap[apName]

	if !ok {
		specialGroupAccessProvider = &exporter.AccessProvider{
			ExternalId:        apName,
			Name:              apName,
			NamingHint:        generateNamingHint(apName),
			NotInternalizable: true,
			Action:            exporter.Grant,
			ActualName:        apName,
			Type:              ptr.String(access_provider.AclSet),
			Who:               r.whoItem,
		}
	}

	specialGroupAccessProvider.What = append(specialGroupAccessProvider.What, exporter.WhatItem{
		DataObject: &data_source.DataObjectReference{
			FullName: binding.Resource,
			Type:     binding.ResourceType,
		},
		Permissions: []string{binding.Role},
	})

	specialGroupAccessProviderMap[apName] = specialGroupAccessProvider
}

func (a *AccessSyncer) projectRolesWhoItem(ctx context.Context, configMap *config.ConfigMap) (*exporter.WhoItem, *exporter.WhoItem, *exporter.WhoItem, error) {
	projectOwnersWho := &exporter.WhoItem{}
	projectEditorWho := &exporter.WhoItem{}
	projectViewerWho := &exporter.WhoItem{}

	gcpProject := configMap.GetString(common.GcpProjectId)
	if gcpProject != "" {
		projectOwnerIds, projectEditorIds, projectViewerIDs, err := a.iamServiceProvider(configMap).GetProjectOwners(ctx, configMap, gcpProject)

		if err != nil {
			return nil, nil, nil, err
		}

		projectOwnersWho = a.generateProjectWhoItem(projectOwnerIds)
		projectEditorWho = a.generateProjectWhoItem(projectEditorIds)
		projectViewerWho = a.generateProjectWhoItem(projectViewerIDs)
	}

	return projectOwnersWho, projectEditorWho, projectViewerWho, nil
}

func (a *AccessSyncer) generateProjectWhoItem(projectOwnerIds []string) *exporter.WhoItem {
	result := &exporter.WhoItem{}

	for _, ownerId := range projectOwnerIds {
		ownerRaitoId := strings.Split(ownerId, ":")[1]

		if strings.HasPrefix(ownerId, "user:") || strings.HasPrefix(ownerId, "serviceAccount:") {
			result.Users = append(result.Users, ownerRaitoId)
		} else if strings.HasPrefix(ownerId, "group:") {
			result.Groups = append(result.Groups, ownerRaitoId)
		} else {
			common.Logger.Warn("Unknown owner type: " + ownerId)
		}
	}

	return result
}

func (a *AccessSyncer) SyncAccessProviderToTarget(ctx context.Context, accessProviders *importer.AccessProviderImport, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler, configMap *config.ConfigMap) error {
	common.Logger.Info(fmt.Sprintf("Start converting %d access providers to bindings", len(accessProviders.AccessProviders)))

	bindingsToAdd, bindingsToDelete := ConvertAccessProviderToBindings(accessProviders)

	common.Logger.Info(fmt.Sprintf("Done converting access providers to bindings: %d bindings to add, %d bindings to remove", len(bindingsToAdd), len(bindingsToDelete)))

	apFeedback := make(map[string]*importer.AccessProviderSyncFeedback)

	for _, ap := range accessProviders.AccessProviders {
		apFeedback[ap.Id] = &importer.AccessProviderSyncFeedback{AccessProvider: ap.Id, ActualName: ap.Id, Type: ptr.String(access_provider.AclSet)}
	}

	iamService := a.iamServiceProvider(configMap)

	for b, aps := range bindingsToDelete {
		common.Logger.Info(fmt.Sprintf("Revoking binding %+v", b))

		a.handleErrors(iamService.RemoveIamBinding(ctx, configMap, b), apFeedback, aps)
	}

	for b, aps := range bindingsToAdd {
		common.Logger.Info(fmt.Sprintf("Granting binding %+v", b))

		a.handleErrors(iamService.AddIamBinding(ctx, configMap, b), apFeedback, aps)

		a.raitoManagedBindings = append(a.raitoManagedBindings, b)
	}

	var merr error

	for _, apsf := range apFeedback {
		err := accessProviderFeedbackHandler.AddAccessProviderFeedback(*apsf)
		if err != nil {
			merr = multierror.Append(merr, err)
		}
	}

	return merr
}

func (a *AccessSyncer) handleErrors(err error, apFeedback map[string]*importer.AccessProviderSyncFeedback, aps []*importer.AccessProvider) {
	if err != nil {
		for _, ap := range aps {
			apFeedback[ap.Id].Errors = append(apFeedback[ap.Id].Errors, err.Error())
		}
	}
}

func (a *AccessSyncer) SyncAccessAsCodeToTarget(ctx context.Context, accessProviders *importer.AccessProviderImport, prefix string, configMap *config.ConfigMap) error {
	return fmt.Errorf("access as code is not yet supported by this plugin")
}

func (a *AccessSyncer) isRaitoManagedBinding(ctx context.Context, configMap *config.ConfigMap, binding iam.IamBinding) (bool, error) {
	meta, err := a.getDSMetadata(ctx, configMap)

	if err != nil {
		return false, err
	}

	for _, doType := range meta.DataObjectTypes {
		if strings.EqualFold(binding.ResourceType, doType.Type) {
			for _, perm := range doType.Permissions {
				if strings.EqualFold(binding.Role, perm.Permission) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func ConvertAccessProviderToBindings(accessProviders *importer.AccessProviderImport) (map[iam.IamBinding][]*importer.AccessProvider, map[iam.IamBinding][]*importer.AccessProvider) {
	bindingsToAdd := make(map[iam.IamBinding][]*importer.AccessProvider)
	bindingsToDelete := make(map[iam.IamBinding][]*importer.AccessProvider)

	for _, ap := range accessProviders.AccessProviders {
		// Process the Who items
		members := []string{}

		for _, m := range ap.Who.Users {
			if strings.Contains(m, "gserviceaccount.com") {
				members = append(members, "serviceAccount:"+m)
			} else {
				members = append(members, "user:"+m)
			}
		}

		for _, m := range ap.Who.Groups {
			members = append(members, "group:"+m)
		}

		deleteMembers := []string{}

		if ap.DeletedWho != nil {
			for _, m := range ap.DeletedWho.Users {
				if strings.Contains(m, "gserviceaccount.com") {
					deleteMembers = append(deleteMembers, "serviceAccount:"+m)
				} else {
					deleteMembers = append(deleteMembers, "user:"+m)
				}
			}

			for _, m := range ap.DeletedWho.Groups {
				deleteMembers = append(deleteMembers, "group:"+m)
			}
		}

		// Process the What Items
		for _, w := range ap.What {
			for _, p := range w.Permissions {
				// for active members add bindings (except if AP gets deleted)
				for _, m := range members {
					binding := iam.IamBinding{
						Member:       m,
						Role:         p,
						Resource:     w.DataObject.FullName,
						ResourceType: w.DataObject.Type,
					}

					if ap.Delete {
						bindingsToDelete[binding] = append(bindingsToDelete[binding], ap)
					} else {
						bindingsToAdd[binding] = append(bindingsToAdd[binding], ap)
					}
				}

				// for deleted members remove bindings
				for _, m := range deleteMembers {
					binding := iam.IamBinding{
						Member:       m,
						Role:         p,
						Resource:     w.DataObject.FullName,
						ResourceType: w.DataObject.Type,
					}

					bindingsToDelete[binding] = append(bindingsToDelete[binding], ap)
				}
			}
		}

		// process the Deleted WhatItems
		if ap.DeleteWhat != nil {
			for _, w := range ap.DeleteWhat {
				for _, p := range w.Permissions {
					// for ALL members delete the bindings
					for _, m := range append(members, deleteMembers...) {
						binding := iam.IamBinding{
							Member:       m,
							Role:         p,
							Resource:     w.DataObject.FullName,
							ResourceType: w.DataObject.Type,
						}

						bindingsToDelete[binding] = append(bindingsToDelete[binding], ap)
					}
				}
			}
		}
	}

	// Go over all the ones in the add list and remove them from the delete list
	for addBinding := range bindingsToAdd {
		delete(bindingsToDelete, addBinding)
	}

	return bindingsToAdd, bindingsToDelete
}

func generateNamingHint(name string) string {
	const maxLength = 128

	if len(name) <= maxLength {
		return name
	}

	return name[len(name)-maxLength:]
}
