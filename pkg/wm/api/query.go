package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

func getFacetNames(r *http.Request) ([]string, error) {
	raw := r.URL.Query().Get("facets")
	if raw == "" {
		return nil, errors.New("The facets list is missing from the query")
	}

	var facetNames []string
	if err := json.Unmarshal([]byte(raw), &facetNames); err != nil {
		return nil, err
	}

	return facetNames, nil
}

func getFilters(r *http.Request) ([]*wm.Filter, error) {
	raw := r.URL.Query().Get("filters")
	if raw == "" {
		return nil, nil
	}

	filters, err := parseFilters([]byte(raw))
	if err != nil {
		return nil, err
	}

	return filters, nil
}
