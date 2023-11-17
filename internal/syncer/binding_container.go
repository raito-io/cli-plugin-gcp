package syncer

import (
	importer "github.com/raito-io/cli/base/access_provider/sync_to_target"
	"github.com/raito-io/golang-set/set"

	"github.com/raito-io/cli-plugin-gcp/internal/iam"
)

type BindingsForDataObject struct {
	bindingsToAdd    set.Set[iam.IamBinding]
	bindingsToDelete set.Set[iam.IamBinding]
	accessProviders  map[iam.IamBinding][]*importer.AccessProvider
}

func NewBindingsForDataObject() *BindingsForDataObject {
	return &BindingsForDataObject{
		bindingsToAdd:    set.NewSet[iam.IamBinding](),
		bindingsToDelete: set.NewSet[iam.IamBinding](),
		accessProviders:  make(map[iam.IamBinding][]*importer.AccessProvider),
	}
}

func (b *BindingsForDataObject) BindingToAdd(binding iam.IamBinding, ap *importer.AccessProvider) {
	b.bindingsToAdd.Add(binding)

	if b.bindingsToDelete.Contains(binding) {
		b.bindingsToDelete.Remove(binding)
		b.accessProviders[binding] = nil
	}

	b.accessProviders[binding] = append(b.accessProviders[binding], ap)
}

func (b *BindingsForDataObject) BindingToDelete(binding iam.IamBinding, ap *importer.AccessProvider) {
	if b.bindingsToAdd.Contains(binding) {
		return
	}

	b.bindingsToDelete.Add(binding)
	b.accessProviders[binding] = append(b.accessProviders[binding], ap)
}

func (b *BindingsForDataObject) GetAllAccessProviders() []*importer.AccessProvider {
	result := make([]*importer.AccessProvider, 0)
	for _, ap := range b.accessProviders {
		result = append(result, ap...)
	}

	return result
}

type BindingContainer struct {
	bindings map[iam.DataObjectReference]*BindingsForDataObject
}

func NewBindingContainer() *BindingContainer {
	return &BindingContainer{
		bindings: make(map[iam.DataObjectReference]*BindingsForDataObject),
	}
}

func (c *BindingContainer) BindingToAdd(doRef iam.DataObjectReference, binding iam.IamBinding, ap *importer.AccessProvider) {
	if _, found := c.bindings[doRef]; !found {
		c.bindings[doRef] = NewBindingsForDataObject()
	}

	c.bindings[doRef].BindingToAdd(binding, ap)
}

func (c *BindingContainer) BindingToDelete(doRef iam.DataObjectReference, binding iam.IamBinding, ap *importer.AccessProvider) {
	if _, found := c.bindings[doRef]; !found {
		c.bindings[doRef] = NewBindingsForDataObject()
	}

	c.bindings[doRef].BindingToDelete(binding, ap)
}
