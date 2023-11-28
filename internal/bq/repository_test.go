package bigquery

import (
	"cloud.google.com/go/bigquery"

	"testing"

	"github.com/raito-io/cli-plugin-gcp/internal/iam"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
