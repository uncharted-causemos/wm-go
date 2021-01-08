package api

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

func TestParseFilters(t *testing.T) {
	for _, test := range []struct {
		raw     string
		isErr   bool
		wantLen int
	}{
		{
			`{"clauses":[]}`,
			false,
			0,
		},
		{
			`{"clauses":[
				{"field":"category","operand":"or","isNot":false,"values":["a","b","c"]}
			]}`,
			false,
			1,
		},
		{
			`{"clauses":[
				{"field":"category","operand":"or","isNot":false,"values":["a","b","c"]},
				{"field":"type","operand":"or","isNot":false,"values":["d","e"]}
			]}`,
			false,
			2,
		},
	} {
		got, err := parseFilters([]byte(test.raw), wm.ContextDatacube)
		if err != nil {
			if !test.isErr {
				t.Errorf("parseFilters returned err:\n%v\nfor:\n%v", err, spew.Sdump(test))
			}
		} else if len(got) != test.wantLen {
			t.Errorf("parseFilters returned %d filters instead of %d for:\n%s\ngot:%v", len(got), test.wantLen, test.raw, spew.Sdump(got))
		}
	}
}

func TestParseFilter(t *testing.T) {
	for _, test := range []struct {
		raw   string
		isErr bool
		want  *wm.Filter
	}{
		{
			`{"field":"category","operand":"or","isNot":false,"values":["a","b","c"]}`,
			false,
			&wm.Filter{
				Field:        wm.FieldDatacubeCategory,
				Operand:      wm.OperandOr,
				IsNot:        false,
				StringValues: []string{"a", "b", "c"},
			},
		},
		{
			`{"field":"period","operand":"and","isNot":false,"values":[[2010,2020]]}`,
			false,
			&wm.Filter{
				Field:   wm.FieldDatacubePeriod,
				Operand: wm.OperandAnd,
				IsNot:   false,
				Range:   wm.Range{Minimum: 2010, Maximum: 2020, IsClosed: false},
			},
		},
	} {
		got, err := parseFilter([]byte(test.raw), wm.ContextDatacube)
		if err != nil {
			if !test.isErr {
				t.Errorf("parseFilter returned err:\n%v\nfor:\n%v", err, spew.Sdump(test))
			}
		} else if !reflect.DeepEqual(got, test.want) {
			t.Errorf("parseFilter returned:\n%v\ninstead of:\n%v\nfor:\n%s", spew.Sdump(got), spew.Sdump(test.want), test.raw)
		}
	}
}

func TestParseValues(t *testing.T) {
	for _, test := range []struct {
		field       wm.Field
		raw         string
		isErr       bool
		wantStrVals []string
		wantIntVals []int
		wantRange   wm.Range
	}{
		{
			wm.FieldDatacubeCountry,
			`[]`,
			false,
			[]string{},
			nil,
			wm.Range{},
		},
		{
			wm.FieldDatacubeCountry,
			`["Ethiopia"]`,
			false,
			[]string{"Ethiopia"},
			nil,
			wm.Range{},
		},
		{
			wm.FieldDatacubePeriod,
			`[[2010,2020]]`,
			false,
			nil,
			nil,
			wm.Range{Minimum: 2010, Maximum: 2020, IsClosed: false},
		},
		{
			wm.FieldDatacubePeriod,
			`"broken"`,
			true,
			nil,
			nil,
			wm.Range{},
		},
	} {
		strVals, intVals, rng, err := parseValues(test.field, []byte(test.raw))
		if err != nil {
			if !test.isErr {
				t.Errorf("parseValues returned err:\n%v\nfor:\n%+v", err, test)
			}
		} else if !reflect.DeepEqual(strVals, test.wantStrVals) ||
			!reflect.DeepEqual(intVals, test.wantIntVals) ||
			rng != test.wantRange {
			t.Errorf("parseValues returned:\n%v\n%v\n%v\ninstead of:\n%v\n%v\n%v\nfor:\n%v %s", strVals, intVals, rng, test.wantStrVals, test.wantIntVals, test.wantRange, test.field, test.raw)
		}
	}
}

func TestParseStringValues(t *testing.T) {
	for _, test := range []struct {
		raw   string
		isErr bool
		want  []string
	}{
		{
			`[]`,
			false,
			[]string{},
		},
		{
			`["one"]`,
			false,
			[]string{"one"},
		},
		{
			`["a", "b", "c"]`,
			false,
			[]string{"a", "b", "c"},
		},
		{
			`"broken"`,
			true,
			nil,
		},
		{
			`[1,2,3]`,
			true,
			nil,
		},
	} {
		got, err := parseStringValues([]byte(test.raw))
		if err != nil {
			if !test.isErr {
				t.Errorf("parseStringValues returned err:\n%v\nfor:\n%+v", err, test)
			}
		} else if !reflect.DeepEqual(got, test.want) {
			t.Errorf("parseStringValues returned:\n%v\ninstead of:\n%v\nfor:\n%s", got, test.want, test.raw)
		}
	}
}

func TestParseIntValues(t *testing.T) {
	for _, test := range []struct {
		raw   string
		isErr bool
		want  []int
	}{
		{
			"[]",
			false,
			nil,
		},
		{
			"[1,2,3]",
			false,
			[]int{1, 2, 3},
		},
		{
			`"broken"`,
			true,
			nil,
		},
		{
			`["a", "b", "c"]`,
			true,
			nil,
		},
	} {
		got, err := parseIntValues([]byte(test.raw))
		if err != nil {
			if !test.isErr {
				t.Errorf("parseIntValues returned err:\n%v\nfor:\n%+v", err, test)
			}
		} else if !reflect.DeepEqual(got, test.want) {
			t.Errorf("parseIntValues returned:\n%v\ninstead of:\n%v\nfor:\n%s", got, test.want, test.raw)
		}
	}
}

func TestParseRange(t *testing.T) {
	for _, test := range []struct {
		raw   string
		isErr bool
		want  wm.Range
	}{
		{
			"[[3.141,2.718]]",
			false,
			wm.Range{Minimum: 3.141, Maximum: 2.718, IsClosed: false},
		},
		{
			"[[3.141]]",
			true,
			wm.Range{},
		},
		{
			"[[3.141,2.718,1.618]]",
			true,
			wm.Range{},
		},
	} {
		got, err := parseRange([]byte(test.raw))
		if err != nil {
			if !test.isErr {
				t.Errorf("parseRange returned err:\n%v\nfor:\n%+v", err, test)
			}
		} else if got != test.want {
			t.Errorf("parseRange returned:\n%v\ninstead of:\n%v\nfor:\n%s", got, test.want, test.raw)
		}
	}
}
