package elastic

import "gitlab.uncharted.software/WM/wm-go/pkg/wm"

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

func buildFilter(filter *wm.Filter) (map[string]interface{}, error) {
	fieldName := fieldNames[filter.Field]
	clause := operandClause[filter.Operand]
	// Build terms
	var terms []interface{}
	for _, value := range filter.StringValues {
		terms = append(terms, map[string]interface{}{
			"term": map[string]interface{}{fieldName: value},
		})
	}
	f := map[string]interface{}{
		"bool": map[string]interface{}{
			clause: terms,
		},
	}

	// Build Ranges

	// Build Integers
	return f, nil
}

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

func buildSearch(term string, fields []string) ([]interface{}, error) {
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
	filterContext, err := buildFilters(options.filters)
	if err != nil {
		return nil, err
	}
	searchContext, err := buildSearch(options.search.text, options.search.fields)
	if err != nil {
		return nil, err
	}
	boolClause := make(map[string]interface{})
	if filterContext != nil {
		boolClause["filter"] = filterContext
	}
	if options.search.text != "" {
		boolClause["should"] = searchContext
		boolClause["minimum_should_match"] = 1
	}
	esQuery := map[string]interface{}{
		"bool": boolClause,
	}
	return esQuery, nil
}
