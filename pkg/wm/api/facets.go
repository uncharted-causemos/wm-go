package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type facetsResponse struct {
	*wm.Facets
}

// Render allows Project to satisfy the render.Renderer interface.
func (fr *facetsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getFacets(w http.ResponseWriter, r *http.Request) {
	facetNames, err := getFacetNames(r)
	if err != nil {
		a.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	filters, err := getFilters(r, wm.ContextDatacube)
	if err != nil {
		a.errorResponse(w, err, http.StatusBadRequest)
		return
	}
	fmt.Println(filters, facetNames)
	// NYI
	// fs, err := a.kb.GetFacets(facetNames, filters)
	// if err != nil {
	// 	a.errorResponse(w, err, http.StatusInternalServerError)
	// 	return
	// }

	render.Render(w, r, &facetsResponse{})
}
