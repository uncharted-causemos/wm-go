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
				Field:   wm.FieldDatacubePeriod,
				Range:   [2]float64{1.2, 3.2},
				Operand: wm.OperandOr,
			},
			`{"bool":{"must":[{"term":{"category":"Economics"}},{"term":{"category":"Agriculture"}}]}}`,
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
