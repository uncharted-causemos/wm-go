package wm

// KnowledgeBase defines the methods that a database implementation needs to
// satisfy.
type KnowledgeBase interface {
	GetFacets(facetNames []string, filters []*Filter) (*Facets, error)
}
