package bigquery

import (
	"context"
	"errors"

	"cloud.google.com/go/bigquery"
	"github.com/raito-io/cli/base/data_source"
	"github.com/stretchr/testify/mock"

	"testing"

	"github.com/raito-io/cli-plugin-gcp/internal/iam"
	"github.com/raito-io/cli-plugin-gcp/internal/org"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_GetBindings_Project(t *testing.T) {
	type fields struct {
		projectId string
		setup     func(client *MockProjectClient)
	}
	type args struct {
		ctx    context.Context
		entity *org.GcpOrgEntity
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []iam.IamBinding
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "get bindings for project",
			fields: fields{
				projectId: "projectId",
				setup: func(client *MockProjectClient) {
					client.EXPECT().GetIamPolicy(mock.Anything, "projectId").Return([]iam.IamBinding{
						{
							Role:         "roles/owner",
							Member:       "user:ruben@raito.io",
							Resource:     "projectId",
							ResourceType: "project",
						},
					}, nil)
				},
			},
			args: args{
				ctx:    context.Background(),
				entity: &org.GcpOrgEntity{Id: "projectId", Type: "project"},
			},
			want: []iam.IamBinding{{
				Role:         "roles/owner",
				Member:       "user:ruben@raito.io",
				Resource:     "projectId",
				ResourceType: "project",
			}},
			wantErr: require.NoError,
		},
		{
			name: "get bindings for datasource",
			fields: fields{
				projectId: "projectId",
				setup: func(client *MockProjectClient) {
					client.EXPECT().GetIamPolicy(mock.Anything, "projectId").Return([]iam.IamBinding{
						{
							Role:         "roles/owner",
							Member:       "user:ruben@raito.io",
							Resource:     "projectId",
							ResourceType: "project",
						},
					}, nil)
				},
			},
			args: args{
				ctx:    context.Background(),
				entity: &org.GcpOrgEntity{Id: "projectId", Type: data_source.Datasource},
			},
			want: []iam.IamBinding{{
				Role:         "roles/owner",
				Member:       "user:ruben@raito.io",
				Resource:     "projectId",
				ResourceType: "project",
			}},
			wantErr: require.NoError,
		},
		{
			name: "get bindings for datasource result in error",
			fields: fields{
				projectId: "projectId",
				setup: func(client *MockProjectClient) {
					client.EXPECT().GetIamPolicy(mock.Anything, "projectId").Return(nil, errors.New("boom"))
				},
			},
			args: args{
				ctx:    context.Background(),
				entity: &org.GcpOrgEntity{Id: "projectId", Type: data_source.Datasource},
			},
			wantErr: require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			projectRepo := NewMockProjectClient(t)

			tt.fields.setup(projectRepo)

			repo := Repository{
				projectClient: projectRepo,
				projectId:     tt.fields.projectId,

				options: &RepositoryOptions{EnableCache: false},
			}

			// When
			result, err := repo.GetBindings(tt.args.ctx, tt.args.entity)

			// Then
			tt.wantErr(t, err)
			if err != nil {
				return
			}

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRepository_UpdateBindings_Project(t *testing.T) {
	dataObject := iam.DataObjectReference{
		ObjectType: "project",
		FullName:   "projectId",
	}

	addBindings := []iam.IamBinding{
		{
			Role:         "roles/owner",
			Member:       "user:ruben@raito.io",
			Resource:     "projectId",
			ResourceType: "project",
		},
	}

	removeBindings := []iam.IamBinding{
		{
			Role:         "roles/editor",
			Member:       "user:michael@raito.io",
			Resource:     "projectId",
			ResourceType: "project",
		},
	}

	type fields struct {
		projectId string
		setup     func(client *MockProjectClient)
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "successful update project bindings",
			fields: fields{
				projectId: "projectId",
				setup: func(client *MockProjectClient) {
					client.EXPECT().UpdateBinding(mock.Anything, &dataObject, addBindings, removeBindings).Return(nil)
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: require.NoError,
		},
		{
			name: "failed update project bindings",
			fields: fields{
				projectId: "projectId",
				setup: func(client *MockProjectClient) {
					client.EXPECT().UpdateBinding(mock.Anything, &dataObject, addBindings, removeBindings).Return(errors.New("failed"))
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			projectRepo := NewMockProjectClient(t)

			tt.fields.setup(projectRepo)

			repo := Repository{
				projectClient: projectRepo,
				projectId:     tt.fields.projectId,

				options: &RepositoryOptions{EnableCache: false},
			}

			// When
			err := repo.UpdateBindings(tt.args.ctx, &dataObject, addBindings, removeBindings)

			// Then
			tt.wantErr(t, err)
		})
	}
}

func TestAccessMerge(t *testing.T) {
	type TestData struct {
		Name     string
		Existing []*bigquery.AccessEntry
		ToAdd    []iam.IamBinding
		ToRemove []iam.IamBinding
		Expected []string
	}

	tests := []TestData{
		{
			Name: "Remove group and add user",
			Existing: []*bigquery.AccessEntry{
				{
					Role:       bigquery.OwnerRole,
					EntityType: bigquery.UserEmailEntity,
					Entity:     "user@raito.io",
				},
				{
					Role:       bigquery.WriterRole,
					EntityType: bigquery.GroupEmailEntity,
					Entity:     "group@raito.io",
				},
			},
			ToAdd: []iam.IamBinding{
				{
					Role:   getRoleForBQEntity(bigquery.ReaderRole),
					Member: "user:user2@raito.io",
				},
			},
			ToRemove: []iam.IamBinding{
				{
					Role:   getRoleForBQEntity(bigquery.WriterRole),
					Member: "group:group@raito.io",
				},
			},
			Expected: []string{
				"user@raito.io|OWNER",
				"user2@raito.io|READER",
			},
		},
		{
			Name: "Remove all",
			Existing: []*bigquery.AccessEntry{
				{
					Role:       bigquery.OwnerRole,
					EntityType: bigquery.UserEmailEntity,
					Entity:     "user@raito.io",
				},
				{
					Role:       bigquery.WriterRole,
					EntityType: bigquery.GroupEmailEntity,
					Entity:     "group@raito.io",
				},
			},
			ToAdd: []iam.IamBinding{},
			ToRemove: []iam.IamBinding{
				{
					Role:   getRoleForBQEntity(bigquery.WriterRole),
					Member: "group:group@raito.io",
				},
				{
					Role:   getRoleForBQEntity(bigquery.OwnerRole),
					Member: "user:user@raito.io",
				},
			},
			Expected: []string{},
		},
		{
			Name: "Remove with other role",
			Existing: []*bigquery.AccessEntry{
				{
					Role:       bigquery.OwnerRole,
					EntityType: bigquery.UserEmailEntity,
					Entity:     "user@raito.io",
				},
				{
					Role:       bigquery.WriterRole,
					EntityType: bigquery.GroupEmailEntity,
					Entity:     "group@raito.io",
				},
			},
			ToAdd: []iam.IamBinding{},
			ToRemove: []iam.IamBinding{
				{
					Role:   getRoleForBQEntity(bigquery.ReaderRole),
					Member: "group:group@raito.io",
				},
			},
			Expected: []string{
				"group@raito.io|WRITER",
				"user@raito.io|OWNER",
			},
		},
		{
			Name: "Same is added and removed",
			Existing: []*bigquery.AccessEntry{
				{
					Role:       bigquery.OwnerRole,
					EntityType: bigquery.UserEmailEntity,
					Entity:     "user@raito.io",
				},
				{
					Role:       bigquery.WriterRole,
					EntityType: bigquery.GroupEmailEntity,
					Entity:     "group@raito.io",
				},
			},
			ToAdd: []iam.IamBinding{
				{
					Role:   getRoleForBQEntity(bigquery.OwnerRole),
					Member: "user:user@raito.io",
				},
			},
			ToRemove: []iam.IamBinding{
				{
					Role:   getRoleForBQEntity(bigquery.OwnerRole),
					Member: "user:user@raito.io",
				},
			},
			Expected: []string{
				"group@raito.io|WRITER",
				"user@raito.io|OWNER",
			},
		},
		{
			Name: "Added and removed with different role",
			Existing: []*bigquery.AccessEntry{
				{
					Role:       bigquery.OwnerRole,
					EntityType: bigquery.UserEmailEntity,
					Entity:     "user@raito.io",
				},
				{
					Role:       bigquery.WriterRole,
					EntityType: bigquery.GroupEmailEntity,
					Entity:     "group@raito.io",
				},
			},
			ToAdd: []iam.IamBinding{
				{
					Role:   getRoleForBQEntity(bigquery.WriterRole),
					Member: "user:user@raito.io",
				},
			},
			ToRemove: []iam.IamBinding{
				{
					Role:   getRoleForBQEntity(bigquery.OwnerRole),
					Member: "user:user@raito.io",
				},
			},
			Expected: []string{
				"group@raito.io|WRITER",
				"user@raito.io|WRITER",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			update, err := mergeBindings(test.Existing, test.ToAdd, test.ToRemove)
			require.NoError(t, err)
			assert.Equal(t, len(test.Expected), len(update.Access))
			entities := make([]string, 0, len(update.Access))
			for _, a := range update.Access {
				entities = append(entities, a.Entity+"|"+string(a.Role))
			}
			assert.ElementsMatch(t, test.Expected, entities)
		})
	}
}
