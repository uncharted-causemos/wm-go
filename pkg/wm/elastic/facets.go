package elastic

import "gitlab.uncharted.software/WM/wm-go/pkg/wm"

// GetFacets returns the facets.
func (es *ES) GetFacets(facetNames []string, filters []*wm.Filter) (*wm.Facets, error) {
	return &wm.Facets{}, nil
}
