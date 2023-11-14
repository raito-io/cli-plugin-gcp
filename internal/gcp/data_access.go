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

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/iam/types"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
	"github.com/raito-io/cli-plugin-gcp/internal/syncer"

	exporter "github.com/raito-io/cli/base/access_provider/sync_from_target"
	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/util/config"
)

//go:generate go run github.com/vektra/mockery/v2 --name=GcpBindingRepository --with-expecter --inpackage
type GcpBindingRepository interface {
	Bindings(ctx context.Context, fn func(ctx context.Context, dataObject *org.GcpOrgEntity, bindings []types.IamBinding) error) error
	AddBinding(ctx context.Context, binding types.IamBinding) error
	RemoveBinding(ctx context.Context, binding types.IamBinding) error
}

//go:generate go run github.com/vektra/mockery/v2 --name=ProjectRepo --with-expecter --inpackage
type ProjectRepo interface {
	GetProjectOwner(ctx context.Context, projectId string) (owner []string, editor []string, viewer []string, err error)
}

type AccessSyncer struct {
	bindingRepo GcpBindingRepository
	projectRepo ProjectRepo
	metadata    *data_source.MetaData

	// cache
	raitoManagedBindings []types.IamBinding
}

func NewDataAccessSyncer(bindingRepo GcpBindingRepository, projectRepo ProjectRepo, metadata *data_source.MetaData) *AccessSyncer {
	return &AccessSyncer{
		bindingRepo: bindingRepo,
		projectRepo: projectRepo,
		metadata:    metadata,
	}
}

func (a *AccessSyncer) SyncAccessProvidersFromTarget(ctx context.Context, accessProviderHandler wrappers.AccessProviderHandler, configMap *config.ConfigMap) error {
	var allBindings []types.IamBinding

	err := a.bindingRepo.Bindings(ctx, func(ctx context.Context, dataObject *org.GcpOrgEntity, bindings []types.IamBinding) error {
		if len(bindings) == 0 {
			return nil
		}

		allBindings = append(allBindings, bindings...)

		return nil
	})
	if err != nil {
		return fmt.Errorf("processing bindings: %w", err)
	}

	aps, err := a.ConvertBindingsToAccessProviders(ctx, configMap, allBindings)
	if err != nil {
		return fmt.Errorf("convert bindings to access providers: %w", err)
	}

	err = accessProviderHandler.AddAccessProviders(aps...)
	if err != nil {
		return fmt.Errorf("add access providers: %w", err)
	}

	return nil
}

func (a *AccessSyncer) SyncAccessProviderToTarget(ctx context.Context, accessProviders *importer.AccessProviderImport, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler, configMap *config.ConfigMap) error {
	common.Logger.Info(fmt.Sprintf("Start converting %d access providers to bindings", len(accessProviders.AccessProviders)))

	bindingsToAdd, bindingsToDelete := ConvertAccessProviderToBindings(accessProviders)

	common.Logger.Info(fmt.Sprintf("Done converting access providers to bindings: %d bindings to add, %d bindings to remove", len(bindingsToAdd), len(bindingsToDelete)))

	apFeedback := make(map[string]*importer.AccessProviderSyncFeedback)

	for _, ap := range accessProviders.AccessProviders {
		apFeedback[ap.Id] = &importer.AccessProviderSyncFeedback{AccessProvider: ap.Id, ActualName: ap.Id, Type: ptr.String(access_provider.AclSet)}
	}

	for b, aps := range bindingsToDelete {
		common.Logger.Info(fmt.Sprintf("Revoking binding %+v", b))

		a.handleErrors(a.bindingRepo.RemoveBinding(ctx, b), apFeedback, aps)
	}

	for b, aps := range bindingsToAdd {
		common.Logger.Info(fmt.Sprintf("Granting binding %+v", b))

		a.handleErrors(a.bindingRepo.AddBinding(ctx, b), apFeedback, aps)

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

func (a *AccessSyncer) SyncAccessAsCodeToTarget(_ context.Context, _ *importer.AccessProviderImport, _ string, _ *config.ConfigMap) error {
	return fmt.Errorf("access as code is not yet supported by this plugin")
}

func (a *AccessSyncer) ConvertBindingsToAccessProviders(ctx context.Context, configMap *config.ConfigMap, bindings []types.IamBinding) ([]*exporter.AccessProvider, error) {
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
		if strings.EqualFold(binding.ResourceType, types.Organization.String()) {
			binding.Resource = syncer.GetOrgDataObjectName(configMap)
		}

		managed := a.isRaitoManagedBinding(binding)

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

func (a *AccessSyncer) generateAccessProvider(binding types.IamBinding, accessProviderMap map[string]*exporter.AccessProvider, managed bool) {
	apName := fmt.Sprintf("%s_%s_%s", binding.ResourceType, binding.Resource, strings.Replace(binding.Role, "/", "_", -1))

	if _, f := accessProviderMap[apName]; !f {
		accessProviderMap[apName] = &exporter.AccessProvider{
			ExternalId:        apName,
			Name:              apName,
			NamingHint:        generateNamingHint(apName),
			NotInternalizable: !managed,
			WhoLocked:         ptr.Bool(false),
			WhatLocked:        ptr.Bool(false),
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

func (a *AccessSyncer) generateGroupedByIdentityAcccessProvider(binding types.IamBinding, groupedByIdentityAccesProviderMap map[string]*exporter.AccessProvider) {
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

func (a *AccessSyncer) generateSpecialGroupOwnerAccessProvider(binding types.IamBinding, specialGroupAccessProviderMap map[string]*exporter.AccessProvider, projectOwnersWho *exporter.WhoItem, projectEditorsWho *exporter.WhoItem, projectReadersWho *exporter.WhoItem) {
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
		projectOwnerIds, projectEditorIds, projectViewerIDs, err := a.projectRepo.GetProjectOwner(ctx, gcpProject)

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

func (a *AccessSyncer) handleErrors(err error, apFeedback map[string]*importer.AccessProviderSyncFeedback, aps []*importer.AccessProvider) {
	if err != nil {
		common.Logger.Error(fmt.Sprintf("error while updating bindings: %s", err.Error()))

		for _, ap := range aps {
			apFeedback[ap.Id].Errors = append(apFeedback[ap.Id].Errors, err.Error())
		}
	}
}

func (a *AccessSyncer) isRaitoManagedBinding(binding types.IamBinding) bool {
	for _, doType := range a.metadata.DataObjectTypes {
		if strings.EqualFold(binding.ResourceType, doType.Type) {
			for _, perm := range doType.Permissions {
				if strings.EqualFold(binding.Role, perm.Permission) {
					return true
				}
			}
		}
	}

	return false
}

func ConvertAccessProviderToBindings(accessProviders *importer.AccessProviderImport, hooks ...iam.AccessProviderBindingHook) (map[types.IamBinding][]*importer.AccessProvider, map[types.IamBinding][]*importer.AccessProvider) {
	bindingsToAdd := make(map[types.IamBinding][]*importer.AccessProvider)
	bindingsToDelete := make(map[types.IamBinding][]*importer.AccessProvider)

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
					binding := types.IamBinding{
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
					binding := types.IamBinding{
						Member:       m,
						Role:         p,
						Resource:     w.DataObject.FullName,
						ResourceType: w.DataObject.Type,
					}

					bindingsToDelete[binding] = append(bindingsToDelete[binding], ap)
				}
			}

			for _, hook := range hooks {
				bindingsToAddFromHook, bindingsToRemoveFromHook := hook(ap, members, deleteMembers, w)
				for _, b := range bindingsToAddFromHook {
					bindingsToAdd[b] = append(bindingsToAdd[b], ap)
				}

				for _, b := range bindingsToRemoveFromHook {
					bindingsToDelete[b] = append(bindingsToDelete[b], ap)
				}
			}
		}

		// process the Deleted WhatItems
		if ap.DeleteWhat != nil {
			for _, w := range ap.DeleteWhat {
				for _, p := range w.Permissions {
					// for ALL members delete the bindings
					for _, m := range append(members, deleteMembers...) {
						binding := types.IamBinding{
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
