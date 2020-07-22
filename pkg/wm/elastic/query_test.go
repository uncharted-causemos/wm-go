package elastic

import (
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

func TestBuildFilter(t *testing.T) {
	tests := []struct {
		description string
		input       *wm.Filter
		expect      string
	}{
		{
			"Build OR filter with multiple values",
			&wm.Filter{
				Field:        wm.FieldDatacubeModel,
				StringValues: []string{"m1", "m2"},
				Operand:      wm.OperandOr,
			},
			`{"bool":{"should":[{"term":{"model":"m1"}},{"term":{"model":"m2"}}]}}`,
		},
		{
			"Build And filter with multiple values",
			&wm.Filter{
				Field:        wm.FieldDatacubeCategory,
				StringValues: []string{"Economics", "Agriculture"},
				Operand:      wm.OperandAnd,
			},
			`{"bool":{"must":[{"term":{"category":"Economics"}},{"term":{"category":"Agriculture"}}]}}`,
		},
		{
			"Build a range filter",
			&wm.Filter{
				Field: wm.FieldDatacubePeriod,
				Range: wm.Range{Minimum: 1167609600000, Maximum: 1525132800000, IsClosed: false},
			},
			`{"bool":{"must":[{"range":{"period":{"gte":1167609600000,"lt":1525132800000}}}]}}`,
		},
		{
			"Build a closed range filter",
			&wm.Filter{
				Field: wm.FieldDatacubePeriod,
				Range: wm.Range{Minimum: 0, Maximum: 3, IsClosed: true},
			},
			`{"bool":{"must":[{"range":{"period":{"gte":0,"lte":3}}}]}}`,
		},
	}
	for _, test := range tests {
		f, _ := buildFilter(test.input)
		result, _ := json.Marshal(f)
		if string(result) != test.expect {
			t.Errorf("%s\nbuildFilter returned: \n%s\ninstead of:\n%s\n for input %v", test.description, result, test.expect, spew.Sdump(test.input))
		}
	}
}

func TestBuildNestedFilter(t *testing.T) {
	type inputParam struct {
		search  string
		filters []*wm.Filter
	}
	tests := []struct {
		description string
		input       inputParam
		expect      string
		isErr       bool
	}{
		{
			"Build filter with single nested field",
			inputParam{
				"concepts",
				[]*wm.Filter{
					{
						Field:        wm.FieldDatacubeConceptName,
						StringValues: []string{"c1", "c2"},
						Operand:      wm.OperandOr,
					},
				},
			},
			`{"nested":{"path":"concepts","query":{"bool":{"filter":[{"bool":{"should":[{"term":{"concepts.name":"c1"}},{"term":{"concepts.name":"c2"}}]}}]}}}}`,
			false,
		},
		{
			"Build filter with multiple nested sibling fields",
			inputParam{
				"concepts",
				[]*wm.Filter{
					{
						Field:        wm.FieldDatacubeConceptName,
						StringValues: []string{"c1", "c2"},
						Operand:      wm.OperandOr,
					},
					{
						Field: wm.FieldDatacubeConceptScore,
						Range: wm.Range{Minimum: 0.1, Maximum: 0.3, IsClosed: false},
					},
				},
			},
			`{"nested":{"path":"concepts","query":{"bool":{"filter":[{"bool":{"should":[{"term":{"concepts.name":"c1"}},{"term":{"concepts.name":"c2"}}]}},{"bool":{"must":[{"range":{"concepts.score":{"gte":0.1,"lt":0.3}}}]}}]}}}}`,
			false,
		},
	}
	for _, test := range tests {
		f, err := buildNestedFilter(test.input.search, test.input.filters)
		if err != nil {
			if !test.isErr {
				t.Errorf("nbuildNestedFilter returned err:\n%v\nfor:\n%+v", err, spew.Sdump(test.input))
			}
		}
		result, _ := json.Marshal(f)
		if string(result) != test.expect {
			t.Errorf("%s\nbuildNestedFilter returned: \n%s\ninstead of:\n%s\n for input %v", test.description, result, test.expect, spew.Sdump(test.input))
		}
	}
}

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		description string
		input       queryOptions
		expect      string
	}{
		{
			"Build an empty query",
			queryOptions{},
			`{"bool":{"match_all":{}}}`,
		},
		{
			"Build text match queries for a search term",
			queryOptions{
				search: searchOptions{text: "testSearchTerm", fields: []string{"f1", "f2", "f3"}},
			},
			`{"bool":{"minimum_should_match":1,"should":[{"match":{"f1":"testSearchTerm"}},{"match":{"f2":"testSearchTerm"}},{"match":{"f3":"testSearchTerm"}}]}}`,
		},
		{
			"Build a query with filters",
			queryOptions{
				filters: []*wm.Filter{
					{
						Field:        wm.FieldDatacubeID,
						StringValues: []string{"id1", "id2"},
						Operand:      wm.OperandOr,
					},
					{
						Field: wm.FieldDatacubePeriod,
						Range: wm.Range{Minimum: 0.1, Maximum: 0.3, IsClosed: true},
					},
				},
			},
			`{"bool":{"filter":[{"bool":{"should":[{"term":{"id":"id1"}},{"term":{"id":"id2"}}]}},{"bool":{"must":[{"range":{"period":{"gte":0.1,"lte":0.3}}}]}}]}}`,
		},
		{
			"Build a query with filters including nested fields",
			queryOptions{
				filters: []*wm.Filter{
					{
						Field:        wm.FieldDatacubeID,
						StringValues: []string{"id1", "id2"},
						Operand:      wm.OperandOr,
					},
					{
						Field: wm.FieldDatacubePeriod,
						Range: wm.Range{Minimum: 0.1, Maximum: 0.3, IsClosed: true},
					},
					{
						Field:        wm.FieldDatacubeConceptName,
						StringValues: []string{"c1", "c2"},
						Operand:      wm.OperandOr,
					},
					{
						Field: wm.FieldDatacubeConceptScore,
						Range: wm.Range{Minimum: 0.1, Maximum: 0.3, IsClosed: false},
					},
				},
			},
			`{"bool":{"filter":[{"bool":{"should":[{"term":{"id":"id1"}},{"term":{"id":"id2"}}]}},{"bool":{"must":[{"range":{"period":{"gte":0.1,"lte":0.3}}}]}},{"nested":{"path":"concepts","query":{"bool":{"filter":[{"bool":{"should":[{"term":{"concepts.name":"c1"}},{"term":{"concepts.name":"c2"}}]}},{"bool":{"must":[{"range":{"concepts.score":{"gte":0.1,"lt":0.3}}}]}}]}}}}]}}`,
		},
		{
			"Build a query with search and filters",
			queryOptions{
				search: searchOptions{text: "testSearchTerm", fields: []string{"f1", "f2", "f3"}},
				filters: []*wm.Filter{
					{
						Field:        wm.FieldDatacubeID,
						StringValues: []string{"id1", "id2"},
						Operand:      wm.OperandOr,
					},
					{
						Field: wm.FieldDatacubePeriod,
						Range: wm.Range{Minimum: 0.1, Maximum: 0.3, IsClosed: true},
					},
				},
			},
			`{"bool":{"filter":[{"bool":{"should":[{"term":{"id":"id1"}},{"term":{"id":"id2"}}]}},{"bool":{"must":[{"range":{"period":{"gte":0.1,"lte":0.3}}}]}}],"minimum_should_match":1,"should":[{"match":{"f1":"testSearchTerm"}},{"match":{"f2":"testSearchTerm"}},{"match":{"f3":"testSearchTerm"}}]}}`,
		},
	}
	for _, test := range tests {
		f, _ := buildQuery(test.input)
		result, _ := json.Marshal(f)
		if string(result) != test.expect {
			t.Errorf("%s\nbuildFilter returned: \n%s\ninstead of:\n%s\n for input %v", test.description, result, test.expect, spew.Sdump(test.input))
		}
	}
}
