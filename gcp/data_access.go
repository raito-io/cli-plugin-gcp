package gcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/smithy-go/ptr"
	"github.com/raito-io/cli-plugin-gcp/gcp/common"
	"github.com/raito-io/cli-plugin-gcp/gcp/iam"
	"github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/wrappers"

	exporter "github.com/raito-io/cli/base/access_provider/sync_from_target"
	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/cli/base/util/config"
)

type AccessSyncer struct {
	iamServiceProvider   func(configMap *config.ConfigMap) iam.IAMService
	raitoManagedBindings []iam.IamBinding
	getDSMetadata        func(ctx context.Context) (*data_source.MetaData, error)
}

func NewDataAccessSyncer() *AccessSyncer {
	return &AccessSyncer{
		iamServiceProvider: newIamServiceProvider,
		getDSMetadata:      GetDataSourceMetaData,
	}
}

func (a *AccessSyncer) WithDataSourceMetadataFetcher(getDSMetadata func(ctx context.Context) (*data_source.MetaData, error)) *AccessSyncer {
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
	accessProviderMap := make(map[string]*exporter.AccessProvider)

	for _, binding := range bindings {
		if strings.EqualFold(binding.ResourceType, iam.Organization.String()) {
			binding.Resource = GetOrgDataObjectName(configMap)
		}

		managed, err2 := a.isRaitoManagedBinding(ctx, binding)

		if err2 != nil {
			return nil, err2
		}

		if configMap.GetBoolWithDefault(common.ExcludeNonAplicablePermissions, true) && !managed {
			common.Logger.Warn(fmt.Sprintf("Skipping role %s for %s on %s %s as it is not an applicable permission for this datasource and %s is false", binding.Role, binding.Member, binding.Resource, binding.ResourceType, common.ExcludeNonAplicablePermissions))
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

		apName := fmt.Sprintf("%s_%s_%s", binding.ResourceType, binding.Resource, strings.Replace(binding.Role, "/", "_", -1))

		if _, f := accessProviderMap[apName]; !f {
			accessProviderMap[apName] = &exporter.AccessProvider{
				ExternalId:        apName,
				Name:              apName,
				NamingHint:        apName,
				NotInternalizable: !managed,
				WhoLocked:         ptr.Bool(false),
				WhatLocked:        ptr.Bool(true),
				WhatLockedReason:  ptr.String("This is a single resource AP"),
				Action:            exporter.Grant,
				NameLocked:        ptr.Bool(false),
				DeleteLocked:      ptr.Bool(false),
				ActualName:        apName,
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

		if strings.HasPrefix(binding.Member, "user:") || strings.HasPrefix(binding.Member, "serviceAccount:") {
			accessProviderMap[apName].Who.Users = append(accessProviderMap[apName].Who.Users, strings.Split(binding.Member, ":")[1])
		} else if strings.HasPrefix(binding.Member, "group:") {
			accessProviderMap[apName].Who.Groups = append(accessProviderMap[apName].Who.Groups, strings.Split(binding.Member, ":")[1])
		} else if strings.HasPrefix(binding.Member, "special_group:") && configMap.GetStringWithDefault(common.GcpProjectId, "") != "" && strings.Contains(binding.Member, "project") {
			// this is a special IAM construct that creates a removable link between ownership on a service resource and ownership on org level
			// e.g. owners on a GCP project are owners on BQ datasets. This binding is removable but can not be (re-)created by a user.
			accessProviderMap[apName].Who.AccessProviders = append(accessProviderMap[apName].Who.AccessProviders, fmt.Sprintf("datasource_%s_%s", configMap.GetString(common.GcpProjectId), strings.Replace(binding.Role, "/", "_", -1)))
		}
	}

	aps := make([]*exporter.AccessProvider, 0)
	for _, ap := range accessProviderMap {
		aps = append(aps, ap)
	}

	return aps, nil
}

func (a *AccessSyncer) SyncAccessProviderToTarget(ctx context.Context, accessProviders *importer.AccessProviderImport, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler, configMap *config.ConfigMap) error {
	bindingsToAdd, bindingsToDelete := ConvertAccessProviderToBindings(accessProviders)

	for _, ap := range accessProviders.AccessProviders {
		// record feedback
		err := accessProviderFeedbackHandler.AddAccessProviderFeedback(ap.Id, importer.AccessSyncFeedbackInformation{AccessId: ap.Id, ActualName: ap.Id})

		if err != nil {
			return err
		}
	}

	iamService := a.iamServiceProvider(configMap)

	for _, b := range bindingsToDelete {
		common.Logger.Info(fmt.Sprintf("Revoking binding %+v", b))

		err := iamService.RemoveIamBinding(ctx, configMap, b)

		if err != nil {
			return err
		}
	}

	for _, b := range bindingsToAdd {
		common.Logger.Info(fmt.Sprintf("Granting binding %+v", b))

		err := iamService.AddIamBinding(ctx, configMap, b)

		if err != nil {
			return err
		}
	}

	// these bindings will be ignored during SyncAccessProvidersFromTarget
	a.raitoManagedBindings = append(a.raitoManagedBindings, bindingsToAdd...)

	return nil
}

func (a *AccessSyncer) SyncAccessAsCodeToTarget(ctx context.Context, accessProviders *importer.AccessProviderImport, prefix string, configMap *config.ConfigMap) error {
	return fmt.Errorf("access as code is not yet supported by this plugin")
}

func (a *AccessSyncer) isRaitoManagedBinding(ctx context.Context, binding iam.IamBinding) (bool, error) {
	meta, err := a.getDSMetadata(ctx)

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

func ConvertAccessProviderToBindings(accessProviders *importer.AccessProviderImport) ([]iam.IamBinding, []iam.IamBinding) {
	bindingsToAdd := make([]iam.IamBinding, 0)
	bindingsToDelete := make([]iam.IamBinding, 0)

	for _, ap := range accessProviders.AccessProviders {
		members := []string{}

		for _, m := range append(ap.Who.Users, ap.Who.UsersInherited...) {
			if strings.Contains(m, "gserviceaccount.com") {
				members = append(members, "serviceAccount:"+m)
			} else {
				members = append(members, "user:"+m)
			}
		}

		for _, m := range ap.Who.Groups {
			members = append(members, "group:"+m)
		}

		delete_members := []string{}

		if ap.DeletedWho != nil {
			for _, m := range append(ap.DeletedWho.Users, ap.DeletedWho.UsersInherited...) {
				if strings.Contains(m, "gserviceaccount.com") {
					delete_members = append(delete_members, "serviceAccount:"+m)
				} else {
					delete_members = append(delete_members, "user:"+m)
				}
			}

			for _, m := range ap.DeletedWho.Groups {
				delete_members = append(delete_members, "group:"+m)
			}
		}

		// process the WhatItems
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
						bindingsToDelete = append(bindingsToDelete, binding)
					} else {
						bindingsToAdd = append(bindingsToAdd, binding)
					}
				}

				// for deleted members remove bindings
				for _, m := range delete_members {
					binding := iam.IamBinding{
						Member:       m,
						Role:         p,
						Resource:     w.DataObject.FullName,
						ResourceType: w.DataObject.Type,
					}

					bindingsToDelete = append(bindingsToDelete, binding)
				}
			}
		}

		// process the Deled WhatItems
		if ap.DeleteWhat != nil {
			for _, w := range ap.DeleteWhat {
				for _, p := range w.Permissions {
					// for ALL members delete the bindings
					for _, m := range append(members, delete_members...) {
						binding := iam.IamBinding{
							Member:       m,
							Role:         p,
							Resource:     w.DataObject.FullName,
							ResourceType: w.DataObject.Type,
						}

						bindingsToDelete = append(bindingsToDelete, binding)
					}
				}
			}
		}
	}

	return bindingsToAdd, bindingsToDelete
}
