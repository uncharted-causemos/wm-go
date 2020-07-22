package elastic

import (
	"errors"
	"strings"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type queryOptions struct {
	filters []*wm.Filter
	search  searchOptions
}
type searchOptions struct {
	text   string
	fields []string
}

// fieldNames ia a filterable field type to document field name mapping
var fieldNames = map[wm.Field]string{
	// Datacube Fields
	wm.FieldDatacubeID:           "id",
	wm.FieldDatacubeType:         "type",
	wm.FieldDatacubeModel:        "model",
	wm.FieldDatacubeCategory:     "category",
	wm.FieldDatacubeLabel:        "label",
	wm.FieldDatacubeMaintainer:   "maintainer",
	wm.FieldDatacubeSource:       "source",
	wm.FieldDatacubeOutputName:   "output_name",
	wm.FieldDatacubeParameters:   "parameters",
	wm.FieldDatacubeConceptName:  "concepts.name",
	wm.FieldDatacubeConceptScore: "concepts.score",
	wm.FieldDatacubeCountry:      "country",
	wm.FieldDatacubeAdmin1:       "admin1",
	wm.FieldDatacubeAdmin2:       "admin2",
	wm.FieldDatacubePeriod:       "period",
}

var operandClause = map[wm.Operand]string{
	wm.OperandAnd: "must",
	wm.OperandOr:  "should",
}

// buildFilter builds ES bool filter query satisfying given filter
func buildFilter(filter *wm.Filter) (map[string]interface{}, error) {
	clause := operandClause[filter.Operand]
	fieldName, ok := fieldNames[filter.Field]
	if !ok {
		return nil, errors.New("buildFilter: Unrecognized field")
	}
	var queries []interface{}
	if filter.IsNot == true {
		return nil, errors.New("buildFilter: Not yet Implemented")
	}
	if filter.StringValues != nil {
		// Build terms
		for _, value := range filter.StringValues {
			queries = append(queries, map[string]interface{}{
				"term": map[string]interface{}{fieldName: value},
			})
		}
	} else if filter.IntValues != nil {
		return nil, errors.New("buildFilter: Not Yet Implemented")
	} else {
		// Build range
		lt := "lt"
		if filter.Range.IsClosed {
			lt = "lte"
		}
		queries = []interface{}{
			map[string]interface{}{
				"range": map[string]interface{}{fieldName: map[string]interface{}{"gte": filter.Range.Minimum, lt: filter.Range.Maximum, "relation": "within"}},
			},
		}
	}

	f := map[string]interface{}{
		"bool": map[string]interface{}{
			clause: queries,
		},
	}
	return f, nil
}

// buildNestedFilter builds ES nested filter query with given filters.
// Provided filters must have fields with same parent field
func buildNestedFilter(filters []*wm.Filter) (map[string]interface{}, error) {
	var path string
	// Note: currently, only supported nested fields are `concepts.*`, when more nested fields are added in the future,
	// It should also validate if the fields of input filters all have same parent.
	for _, filter := range filters {
		fields := strings.Split(fieldNames[filter.Field], ".")
		if len(fields) != 2 {
			// only support nested field with 1 level deep
			return nil, errors.New("buildNestedFilter: Invalid nested field")
		}
		path = fields[0]
	}
	fs, err := buildFilters(filters)
	if err != nil {
		return nil, err
	}
	nested := map[string]interface{}{
		"nested": map[string]interface{}{
			"path": path,
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"filter": fs,
				},
			},
		},
	}
	return nested, nil
}

// buildFilters builds ES filter quires with given filters
func buildFilters(filters []*wm.Filter) ([]interface{}, error) {
	var fs []interface{}
	for _, filter := range filters {
		f, err := buildFilter(filter)
		if err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}
	return fs, nil
}

// buildFilterContext builds ES filter used in the filter context (root filter query)
func buildFilterContext(filters []*wm.Filter) ([]interface{}, error) {
	return buildFilters(filters)
}

// buildSearchQueries builds ES text search queries on given fields with a provided search term
func buildSearchQueries(term string, fields []string) ([]interface{}, error) {
	var matches []interface{}
	for _, field := range fields {
		match := map[string]interface{}{
			"match": map[string]interface{}{field: term},
		}
		matches = append(matches, match)
	}
	return matches, nil
}

func buildQuery(options queryOptions) (map[string]interface{}, error) {
	filterContext, err := buildFilterContext(options.filters)
	if err != nil {
		return nil, err
	}
	boolClause := make(map[string]interface{})
	if filterContext != nil {
		boolClause["filter"] = filterContext
	}
	if options.search.text != "" {
		searchContext, err := buildSearchQueries(options.search.text, options.search.fields)
		if err != nil {
			return nil, err
		}
		boolClause["should"] = searchContext
		boolClause["minimum_should_match"] = 1
	}
	esQuery := map[string]interface{}{
		"bool": boolClause,
	}
	return esQuery, nil
}
