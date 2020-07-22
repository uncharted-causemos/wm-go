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
			`{"bool":{"must":[{"range":{"period":{"gte":1167609600000,"lt":1525132800000,"relation":"within"}}}]}}`,
		},
		{
			"Build a closed range filter",
			&wm.Filter{
				Field: wm.FieldDatacubePeriod,
				Range: wm.Range{Minimum: 0, Maximum: 3, IsClosed: true},
			},
			`{"bool":{"must":[{"range":{"period":{"gte":0,"lte":3,"relation":"within"}}}]}}`,
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
	tests := []struct {
		description string
		input       []*wm.Filter
		expect      string
		isErr       bool
	}{
		{
			"Build filter with single nested field",
			[]*wm.Filter{
				{
					Field:        wm.FieldDatacubeConceptName,
					StringValues: []string{"c1", "c2"},
					Operand:      wm.OperandOr,
				},
			},
			`{"nested":{"path":"concepts","query":{"bool":{"filter":[{"bool":{"should":[{"term":{"concepts.name":"c1"}},{"term":{"concepts.name":"c2"}}]}}]}}}}`,
			false,
		},
		{
			"Build filter with multiple nested sibling fields",
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
			`{"nested":{"path":"concepts","query":{"bool":{"filter":[{"bool":{"should":[{"term":{"concepts.name":"c1"}},{"term":{"concepts.name":"c2"}}]}},{"bool":{"must":[{"range":{"concepts.score":{"gte":0.1,"lt":0.3,"relation":"within"}}}]}}]}}}}`,
			false,
		},
		{
			"Throw errorc with invalid nested fields",
			[]*wm.Filter{
				{
					Field:        wm.FieldDatacubeAdmin1,
					StringValues: []string{"c1", "c2"},
					Operand:      wm.OperandOr,
				},
				{
					Field: wm.FieldDatacubePeriod,
					Range: wm.Range{Minimum: 0.1, Maximum: 0.3, IsClosed: false},
				},
			},
			"null",
			true,
		},
	}
	for _, test := range tests {
		f, err := buildNestedFilter(test.input)
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
