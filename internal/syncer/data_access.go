package syncer

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/raito-io/golang-set/set"

	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli/base/access_provider"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/wrappers"

	exporter "github.com/raito-io/cli/base/access_provider/sync_from_target"
	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

//go:generate go run github.com/vektra/mockery/v2 --name=BindingRepository --with-expecter --inpackage
type BindingRepository interface {
	Bindings(ctx context.Context, fn func(ctx context.Context, dataObject *org.GcpOrgEntity, bindings []iam.IamBinding) error) error
	UpdateBindings(ctx context.Context, dataObject *iam.DataObjectReference, addBindings []iam.IamBinding, removeBindings []iam.IamBinding) error

	DataSourceType() string
}

//go:generate go run github.com/vektra/mockery/v2 --name=ProjectRepo --with-expecter --inpackage
type ProjectRepo interface {
	GetProjectOwner(ctx context.Context, projectId string) (owner []string, editor []string, viewer []string, err error)
}

//go:generate go run github.com/vektra/mockery/v2 --name=MaskingService --with-expecter --inpackage
type MaskingService interface {
	ImportMasks(ctx context.Context, accessProviderHandler wrappers.AccessProviderHandler, locations set.Set[string], maskingTags map[string][]string, raitoMasks set.Set[string]) error
	ExportMasks(ctx context.Context, accessProvider *importer.AccessProvider, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler) ([]string, error)
	MaskedBinding(ctx context.Context, members []string) ([]iam.IamBinding, error)
}

type AccessSyncer struct {
	bindingRepo    BindingRepository
	projectRepo    ProjectRepo
	maskingService MaskingService
	metadata       *data_source.MetaData

	maskingSupport  bool
	addMaskedReader bool

	// cache
	raitoManagedBindings set.Set[iam.IamBinding]
	raitoMasks           set.Set[string]
}

func NewDataAccessSyncer(bindingRepo BindingRepository, projectRepo ProjectRepo, maskingService MaskingService, metadata *data_source.MetaData, configmap *config.ConfigMap) *AccessSyncer {
	maskingSupport := false

	for _, feature := range metadata.SupportedFeatures {
		if feature == data_source.ColumnMasking {
			maskingSupport = true
			break
		}
	}

	return &AccessSyncer{
		bindingRepo:          bindingRepo,
		projectRepo:          projectRepo,
		maskingService:       maskingService,
		metadata:             metadata,
		maskingSupport:       maskingSupport,
		addMaskedReader:      configmap.GetBoolWithDefault(common.GcpMaskedReader, false) || configmap.GetBoolWithDefault(common.BqCatalogEnabled, false),
		raitoManagedBindings: set.NewSet[iam.IamBinding](),
		raitoMasks:           set.NewSet[string](),
	}
}

func (a *AccessSyncer) SyncAccessProvidersFromTarget(ctx context.Context, accessProviderHandler wrappers.AccessProviderHandler, configMap *config.ConfigMap) error {
	var allBindings []iam.IamBinding
	locations := set.NewSet[string]()
	maskingTags := make(map[string][]string)

	err := a.bindingRepo.Bindings(ctx, func(ctx context.Context, dataObject *org.GcpOrgEntity, bindings []iam.IamBinding) error {
		if len(bindings) == 0 {
			return nil
		}

		allBindings = append(allBindings, bindings...)

		if a.maskingSupport && dataObject.Type == data_source.Column && len(dataObject.PolicyTags) > 0 {
			locations.Add(dataObject.Location)

			for tagIdx := range dataObject.PolicyTags {
				if !a.raitoMasks.Contains(dataObject.PolicyTags[tagIdx]) {
					maskingTags[dataObject.PolicyTags[tagIdx]] = append(maskingTags[dataObject.PolicyTags[tagIdx]], dataObject.FullName)
				}
			}
		}

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

	if a.maskingSupport {
		err = a.maskingService.ImportMasks(ctx, accessProviderHandler, locations, maskingTags, a.raitoMasks)
		if err != nil {
			return fmt.Errorf("import masks: %w", err)
		}
	}

	return nil
}

func (a *AccessSyncer) SyncAccessProviderToTarget(ctx context.Context, accessProviders *importer.AccessProviderImport, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler, _ *config.ConfigMap) error {
	common.Logger.Info(fmt.Sprintf("Start converting %d access providers to bindings", len(accessProviders.AccessProviders)))

	grants := make([]*importer.AccessProvider, 0, len(accessProviders.AccessProviders))

	// Handle masks
	for _, ap := range accessProviders.AccessProviders {
		if ap.Action == importer.Grant {
			grants = append(grants, ap)
		} else if ap.Action == importer.Mask {
			raitoMask, err := a.maskingService.ExportMasks(ctx, ap, accessProviderFeedbackHandler)
			if err != nil {
				return fmt.Errorf("export masks: %w", err)
			}

			if raitoMask != nil {
				a.raitoMasks.Add(raitoMask...)
			}
		} else {
			err := accessProviderFeedbackHandler.AddAccessProviderFeedback(importer.AccessProviderSyncFeedback{
				AccessProvider: ap.Id,
				ActualName:     ap.Id,
				Errors:         []string{fmt.Sprintf("unsupported action: %d", ap.Action)},
			})
			if err != nil {
				return fmt.Errorf("add access provider feedback: %w", err)
			}
		}
	}

	// Handle grants
	bindings := a.convertAccessProviderToBindings(ctx, grants)

	common.Logger.Info("Done converting access providers to bindings.")

	apFeedback := make(map[string]*importer.AccessProviderSyncFeedback)

	for _, ap := range grants {
		apFeedback[ap.Id] = &importer.AccessProviderSyncFeedback{AccessProvider: ap.Id, ActualName: ap.Id, Type: ptr.String(access_provider.AclSet)}
	}

	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}

	for do := range bindings.bindings {
		wg.Add(1)

		go func(do iam.DataObjectReference) {
			defer wg.Done()

			common.Logger.Debug(fmt.Sprintf("Update bindings for %s %q", do.ObjectType, do.FullName))

			bindingsToAdd := bindings.bindings[do].bindingsToAdd.Slice()
			bindingsToDelete := bindings.bindings[do].bindingsToDelete.Slice()

			err := a.bindingRepo.UpdateBindings(ctx, &do, bindingsToAdd, bindingsToDelete)

			// LOCKED part!
			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				a.handleErrors(fmt.Errorf("update bindings of %s %q: %w", do.ObjectType, do.FullName, err), apFeedback, bindings.bindings[do].GetAllAccessProviders())
			}

			a.raitoManagedBindings.Add(bindingsToAdd...)
			a.raitoManagedBindings.Add(bindingsToDelete...) // Add also bindings to delete as if an AP failed to delete we do not want those bindings to be importer as external AP
		}(do)
	}

	wg.Wait()

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
		managed := a.isRaitoManagedBinding(binding)

		if configMap.GetBoolWithDefault(common.ExcludeNonAplicablePermissions, true) && !managed {
			common.Logger.Info(fmt.Sprintf("Skipping role %s for %s on %s %s as it is not an applicable permission for this datasource and %s is false", binding.Role, binding.Member, binding.Resource, binding.ResourceType, common.ExcludeNonAplicablePermissions))
			continue
		}

		if a.raitoManagedBindings.Contains(binding) {
			common.Logger.Debug(fmt.Sprintf("Skipping role %s for %s on %s %s as it is managed by raito", binding.Role, binding.Member, binding.Resource, binding.ResourceType))
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
		projectOwnerIds, projectEditorIds, projectViewerIDs, err := a.projectRepo.GetProjectOwner(ctx, gcpProject)

		if err != nil {
			return nil, nil, nil, fmt.Errorf("get project %q owner: %w", gcpProject, err)
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

func (a *AccessSyncer) isRaitoManagedBinding(binding iam.IamBinding) bool {
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

func (a *AccessSyncer) convertAccessProviderToBindings(ctx context.Context, accessProviders []*importer.AccessProvider) *BindingContainer {
	bindings := NewBindingContainer()

	for _, ap := range accessProviders {
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
			objectType := w.DataObject.Type
			if objectType == data_source.Datasource {
				objectType = a.bindingRepo.DataSourceType()
			}

			objectReference := iam.DataObjectReference{
				FullName:   w.DataObject.FullName,
				ObjectType: objectType,
			}

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
						bindings.BindingToDelete(objectReference, binding, ap)
					} else {
						bindings.BindingToAdd(objectReference, binding, ap)
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

					bindings.BindingToDelete(objectReference, binding, ap)
				}
			}
		}

		if a.addMaskedReader && !ap.Delete {
			additionalMaskBindings, err := a.maskingService.MaskedBinding(ctx, members)
			if err != nil {
				common.Logger.Error(fmt.Sprintf("error while masking binding: %s", err.Error()))
			}

			for _, b := range additionalMaskBindings {
				dataObjectReference := iam.DataObjectReference{
					FullName:   b.Resource,
					ObjectType: b.ResourceType,
				}

				bindings.BindingToAdd(dataObjectReference, b, ap)
			}
		}

		// process the Deleted WhatItems
		if ap.DeleteWhat != nil {
			for _, w := range ap.DeleteWhat {
				dataObjectReference := iam.DataObjectReference{
					FullName:   w.DataObject.FullName,
					ObjectType: w.DataObject.Type,
				}

				for _, p := range w.Permissions {
					// for ALL members delete the bindings
					for _, m := range append(members, deleteMembers...) {
						binding := iam.IamBinding{
							Member:       m,
							Role:         p,
							Resource:     w.DataObject.FullName,
							ResourceType: w.DataObject.Type,
						}

						bindings.BindingToDelete(dataObjectReference, binding, ap)
					}
				}
			}
		}
	}

	return bindings
}

func generateNamingHint(name string) string {
	const maxLength = 128

	if len(name) <= maxLength {
		return name
	}

	return name[len(name)-maxLength:]
}
