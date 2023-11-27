package syncer

import (
	"testing"

	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/golang-set/set"
	"github.com/stretchr/testify/assert"

	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

func TestBindingsForDataObject_BindingToAdd(t *testing.T) {
	t.Run("Add New Binding", func(t *testing.T) {
		// Given
		binding := iam.IamBinding{Role: "role", Resource: "resource", ResourceType: "type", Member: "member"}
		ap := importer.AccessProvider{Id: "id"}

		b := NewBindingsForDataObject()

		// When
		b.BindingToAdd(binding, &ap)

		// Then
		assert.Equal(t, set.NewSet(binding), b.bindingsToAdd)
		assert.Empty(t, b.bindingsToDelete)
		assert.Equal(t, map[iam.IamBinding][]*importer.AccessProvider{binding: {&ap}}, b.accessProviders)
	})

	t.Run("Add Binding Removes remove Binding", func(t *testing.T) {
		// Given
		binding := iam.IamBinding{Role: "role", Resource: "resource", ResourceType: "type", Member: "member"}
		ap1 := importer.AccessProvider{Id: "id1"}
		ap2 := importer.AccessProvider{Id: "id2"}

		b := NewBindingsForDataObject()
		b.BindingToDelete(binding, &ap1)

		// When
		b.BindingToAdd(binding, &ap2)

		// Then
		assert.Equal(t, set.NewSet(binding), b.bindingsToAdd)
		assert.Empty(t, b.bindingsToDelete)
		assert.Equal(t, map[iam.IamBinding][]*importer.AccessProvider{binding: {&ap2}}, b.accessProviders)
	})
}

func TestBindingsForDataObject_BindingToDelete(t *testing.T) {
	t.Run("Add Binding to Remove", func(t *testing.T) {
		// Given
		binding := iam.IamBinding{Role: "role", Resource: "resource", ResourceType: "type", Member: "member"}
		ap := importer.AccessProvider{Id: "id"}

		b := NewBindingsForDataObject()

		// When
		b.BindingToDelete(binding, &ap)

		// Then
		assert.Empty(t, b.bindingsToAdd)
		assert.Equal(t, set.NewSet(binding), b.bindingsToDelete)
		assert.Equal(t, map[iam.IamBinding][]*importer.AccessProvider{binding: {&ap}}, b.accessProviders)
	})

	t.Run("Ignore binding to remove if already in add", func(t *testing.T) {
		// Given
		binding := iam.IamBinding{Role: "role", Resource: "resource", ResourceType: "type", Member: "member"}
		ap1 := importer.AccessProvider{Id: "id1"}
		ap2 := importer.AccessProvider{Id: "id2"}

		b := NewBindingsForDataObject()
		b.BindingToAdd(binding, &ap1)

		// When
		b.BindingToDelete(binding, &ap2)

		// Then
		assert.Empty(t, b.bindingsToDelete)
		assert.Equal(t, set.NewSet(binding), b.bindingsToAdd)
		assert.Equal(t, map[iam.IamBinding][]*importer.AccessProvider{binding: {&ap1}}, b.accessProviders)
	})
}

func TestBindingsForDataObject_GetAllAccessProviders(t *testing.T) {
	type fields struct {
		accessProviders map[iam.IamBinding][]*importer.AccessProvider
	}
	tests := []struct {
		name   string
		fields fields
		want   []*importer.AccessProvider
	}{
		{
			name: "Empty",
			fields: fields{
				accessProviders: make(map[iam.IamBinding][]*importer.AccessProvider),
			},
			want: []*importer.AccessProvider{},
		},
		{
			name: "One Binding",
			fields: fields{
				accessProviders: map[iam.IamBinding][]*importer.AccessProvider{
					{Role: "role", Resource: "resource", ResourceType: "type", Member: "member"}: {
						&importer.AccessProvider{Id: "id1"},
					},
				},
			},
			want: []*importer.AccessProvider{
				{Id: "id1"},
			},
		},
		{
			name: "Multiple binding",
			fields: fields{
				accessProviders: map[iam.IamBinding][]*importer.AccessProvider{
					{Role: "role", Resource: "resource", ResourceType: "type", Member: "member"}: {
						&importer.AccessProvider{Id: "id1"},
					},
					{Role: "role2", Resource: "resource", ResourceType: "type", Member: "member"}: {
						&importer.AccessProvider{Id: "id2"},
					},
				},
			},
			want: []*importer.AccessProvider{
				{Id: "id1"},
				{Id: "id2"},
			},
		},
		{
			name: "Multiple aps per binding",
			fields: fields{
				accessProviders: map[iam.IamBinding][]*importer.AccessProvider{
					{Role: "role", Resource: "resource", ResourceType: "type", Member: "member"}: {
						&importer.AccessProvider{Id: "id1"},
						&importer.AccessProvider{Id: "id3"},
						&importer.AccessProvider{Id: "id4"},
					},
					{Role: "role2", Resource: "resource", ResourceType: "type", Member: "member"}: {
						&importer.AccessProvider{Id: "id2"},
					},
				},
			},
			want: []*importer.AccessProvider{
				{Id: "id1"},
				{Id: "id3"},
				{Id: "id4"},
				{Id: "id2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BindingsForDataObject{
				accessProviders: tt.fields.accessProviders,
			}
			assert.ElementsMatchf(t, tt.want, b.GetAllAccessProviders(), "GetAllAccessProviders()")
		})
	}
}
