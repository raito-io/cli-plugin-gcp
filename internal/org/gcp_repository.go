package org

import (
	"context"
	"fmt"
	"strings"

	"github.com/raito-io/cli/base/util/config"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
)

type GCPRepository struct {
}

func NewGCPRepository() *GCPRepository { return &GCPRepository{} }

func (r *GCPRepository) GetProjects(ctx context.Context, configMap *config.ConfigMap) ([]GcpOrgEntity, error) {
	crm, err := common.CrmService(ctx, configMap)

	if err != nil {
		return nil, err
	}

	res, err := crm.Projects.List().Do()

	if err != nil {
		if strings.Contains(err.Error(), "403") {
			common.Logger.Warn(fmt.Sprintf("Failed to fetch the GCP projects: %s", err.Error()))
			return nil, nil
		}

		return nil, err
	}

	projects := make([]GcpOrgEntity, len(res.Projects))

	for i, p := range res.Projects {
		projects[i].Name = p.Name
		projects[i].Id = p.ProjectId
		projects[i].Type = "project"

		if p.Parent != nil {
			parent := GcpOrgEntity{
				Id:   p.Parent.Id,
				Type: p.Parent.Type,
			}

			projects[i].Parent = &parent
		}
	}

	return projects, nil
}

func (r *GCPRepository) GetFolders(ctx context.Context, configMap *config.ConfigMap) ([]GcpOrgEntity, error) {
	orgId := configMap.GetString(common.GcpOrgId)

	if orgId == "" {
		return nil, nil
	}

	return getFoldersForParent(ctx, configMap, fmt.Sprintf("organizations/%s", orgId))
}

func getFoldersForParent(ctx context.Context, configMap *config.ConfigMap, parent string) ([]GcpOrgEntity, error) {
	crm, err := common.CrmServiceV2(ctx, configMap)

	if err != nil {
		return nil, err
	}

	res, err := crm.Folders.List().Parent(parent).Do()

	if err != nil {
		if strings.Contains(err.Error(), "403") {
			common.Logger.Warn(fmt.Sprintf("Failed to fetch the GCP folders in %s: %s", parent, err.Error()))
			return nil, nil
		}

		return nil, err
	}

	folders := make([]GcpOrgEntity, len(res.Folders))

	for i, p := range res.Folders {
		folders[i].Name = p.DisplayName
		folders[i].Id = strings.Split(p.Name, "/")[1]
		folders[i].Type = "folder"

		split := strings.Split(p.Parent, "s/") // the "s" is not a typo, resource IDs are always plural e.g. folders/<id> for a folder
		if len(split) == 2 && !strings.HasPrefix(parent, "organizations/") {
			pObj := GcpOrgEntity{
				Id:   split[1],
				Type: split[0],
			}
			folders[i].Parent = &pObj
		}

		subFolders, err := getFoldersForParent(ctx, configMap, p.Name)

		if err != nil {
			return nil, err
		}

		folders = append(folders, subFolders...)
	}

	return folders, nil
}
