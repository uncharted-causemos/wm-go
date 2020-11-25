package elastic

import (
	"errors"

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
	wm.FieldDatacubeVariable:     "variable",
	wm.FieldDatacubeSearch:       "_search",

	// Analysis Fields
	wm.FieldAnalysisProjectID: "project_id",

	// Analysis Fields
	wm.FieldIndicatorVariable: "variable",
	wm.FieldIndicatorDataset:  "dataset",
}

const (
	conceptsPath = "concepts"
)

// Available nested fields and its path mapping
var nestedPath = map[wm.Field]string{
	wm.FieldDatacubeConceptName:  conceptsPath,
	wm.FieldDatacubeConceptScore: conceptsPath,
}

var isSearchField = map[wm.Field]bool{
	wm.FieldDatacubeSearch: true,
}

var operandClause = map[wm.Operand]string{
	wm.OperandAnd: "must",
	wm.OperandOr:  "should",
}

// buildClause builds ES bool clause query satisfying given filter
func buildClause(filter *wm.Filter) (map[string]interface{}, error) {
	clauseType := operandClause[filter.Operand]
	fieldName, ok := fieldNames[filter.Field]
	matchType := "term"
	if isSearchField[filter.Field] {
		matchType = "match"
	}
	if !ok {
		return nil, errors.New("buildClause: Unrecognized field")
	}
	var queries []interface{}
	if filter.IsNot == true {
		return nil, errors.New("buildClause: Not yet Implemented")
	}
	if filter.StringValues != nil {
		// Build terms
		for _, value := range filter.StringValues {
			queries = append(queries, map[string]interface{}{
				matchType: map[string]interface{}{fieldName: value},
			})
		}
	} else if filter.IntValues != nil {
		return nil, errors.New("buildClause: Not Yet Implemented")
	} else {
		// Build range
		lt := "lt"
		if filter.Range.IsClosed {
			lt = "lte"
		}
		queries = []interface{}{
			map[string]interface{}{
				"range": map[string]interface{}{fieldName: map[string]interface{}{"gte": filter.Range.Minimum, lt: filter.Range.Maximum}},
			},
		}
	}

	f := map[string]interface{}{
		"bool": map[string]interface{}{
			clauseType: queries,
		},
	}
	return f, nil
}

// buildNestedFilter builds ES nested filter query with given filters.
// Provided filters must have fields with same parent field
func buildNestedClause(path string, filters []*wm.Filter) (map[string]interface{}, error) {
	fs, err := buildClauses(filters)
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
func buildClauses(filters []*wm.Filter) ([]interface{}, error) {
	var fs []interface{}
	for _, filter := range filters {
		f, err := buildClause(filter)
		if err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}
	return fs, nil
}

// buildFilterContext builds ES filter used in the filter context (root filter query)
func buildQueryClauses(filters []*wm.Filter) ([]interface{}, error) {
	var results []interface{}
	nested := make(map[string][]*wm.Filter)
	var normals []*wm.Filter

	for _, filter := range filters {
		path := nestedPath[filter.Field]
		if path != "" {
			nested[path] = append(nested[path], filter)
		} else {
			normals = append(normals, filter)
		}
	}
	normalFilters, err := buildClauses(normals)
	if err != nil {
		return nil, err
	}
	results = append(results, normalFilters...)

	for p, fs := range nested {
		nestedFilter, err := buildNestedClause(p, fs)
		if err != nil {
			return nil, err
		}
		results = append(results, nestedFilter)
	}
	return results, nil
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

func buildBoolQuery(options queryOptions) (map[string]interface{}, error) {
	boolClause := make(map[string]interface{})

	// Matching document with searches will contribute to the score
	var searches []*wm.Filter
	// Scores will be ignored for documents matching filters
	var filters []*wm.Filter

	for _, filter := range options.filters {
		if isSearchField[filter.Field] {
			searches = append(searches, filter)
		} else {
			filters = append(searches, filter)
		}
	}
	queryContext, err := buildQueryClauses(searches)
	if err != nil {
		return nil, err
	}
	filterContext, err := buildQueryClauses(filters)
	if err != nil {
		return nil, err
	}
	if queryContext != nil {
		boolClause["must"] = queryContext
	}
	if filterContext != nil {
		boolClause["filter"] = filterContext
	}

	esQuery := make(map[string]interface{})
	if len(boolClause) > 0 {
		esQuery["bool"] = boolClause
	}

	return esQuery, nil
}

// TODO: this will be deprecated and replaced by buildBoolQuery. Need to update the corresponding tests as well.
func buildQuery(options queryOptions) (map[string]interface{}, error) {
	filterContext, err := buildQueryClauses(options.filters)
	if err != nil {
		return nil, err
	}
	boolClause := make(map[string]interface{})
	if options.search.text != "" {
		searchContext, err := buildSearchQueries(options.search.text, options.search.fields)
		if err != nil {
			return nil, err
		}
		boolClause["should"] = searchContext
		boolClause["minimum_should_match"] = 1
	}
	if filterContext != nil {
		boolClause["filter"] = filterContext
	}
	esQuery := make(map[string]interface{})
	if len(boolClause) > 0 {
		esQuery["bool"] = boolClause
	}
	return esQuery, nil
}
