package api

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

func TestGetFacetNames(t *testing.T) {
	for _, test := range []struct {
		r     *http.Request
		isErr bool
		want  []string
	}{
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `facets=["topic","score"]&filters={"clauses":[]}`,
				},
			},
			false,
			[]string{"topic", "score"},
		},
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `facets="broken"`,
				},
			},
			true,
			nil,
		},
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `facets=[1,2,3]`,
				},
			},
			true,
			nil,
		},
	} {
		got, err := getFacetNames(test.r)
		if err != nil {
			if !test.isErr {
				t.Errorf("getFacetNames returned err:\n%v\nfor:\n%v", err, spew.Sdump(test))
			}
		} else if !reflect.DeepEqual(got, test.want) {
			t.Errorf("getFacetNames returned:\n%v\ninstead of:\n%v\nfor:\n%v", got, test.want, spew.Sdump(test.r))
		}
	}
}

func TestGetFilters(t *testing.T) {
	for _, test := range []struct {
		r       *http.Request
		isErr   bool
		wantLen int
	}{
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `facets=["topic","score"]`,
				},
			},
			false,
			0,
		},
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `facets=["topic","score"]&filters={"clauses":[]}`,
				},
			},
			false,
			0,
		},
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `facets=["topic","score"]&filters={"clauses":[
						{"field":"category","operand":"or","isNot":false,"values":["a","b","c"]}
					]}`,
				},
			},
			false,
			1,
		},
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `facets=["topic","score"]&filters={"clauses":[
						{"field":"category","operand":"or","isNot":false,"values":["a","b","c"]},
						{"field":"type","operand":"or","isNot":false,"values":["d","e"]}
					]}`,
				},
			},
			false,
			2,
		},
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `filters="broken"`,
				},
			},
			true,
			0,
		},
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `filters={"clauses":"broken"}`,
				},
			},
			true,
			0,
		},
	} {
		got, err := getFilters(test.r, wm.ContextDatacube)
		if err != nil {
			if !test.isErr {
				t.Errorf("getFilters returned err:\n%v\nfor:\n%v", err, spew.Sdump(test))
			}
		} else if len(got) != test.wantLen {
			t.Errorf("getFilters returned %d filters instead of %d for:\n%v\ngot:%v", len(got), test.wantLen, spew.Sdump(test.r), spew.Sdump(got))
		}
	}
}

func TestGetTileDataSpecs(t *testing.T) {
	for _, test := range []struct {
		r     *http.Request
		isErr bool
		want  wm.GridTileOutputSpecs
	}{
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `specs=[
						{"model":"population", "runId":"rid", "feature":"f1", "date": "2020-01", "valueProp": "v1"},
					  {"model":"population2", "runId":"rid2", "feature":"f2", "date":"2020-02", "valueProp": "v2"}
					]`,
				},
			},
			false,
			wm.GridTileOutputSpecs{
				wm.GridTileOutputSpec{Model: "population", RunID: "rid", Feature: "f1", Date: "2020-01", ValueProp: "v1"},
				wm.GridTileOutputSpec{Model: "population2", RunID: "rid2", Feature: "f2", Date: "2020-02", ValueProp: "v2"},
			},
		},
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `specs="broken"`,
				},
			},
			true,
			nil,
		},
		{
			&http.Request{
				URL: &url.URL{
					RawQuery: `specs=[]`,
				},
			},
			false,
			wm.GridTileOutputSpecs{},
		},
	} {
		got, err := getGridTileOutputSpecs(test.r)
		if err != nil {
			if !test.isErr {
				t.Errorf("getTileRequestSpecs returned err:\n%v\nfor:\n%v", err, spew.Sdump(test))
			}
		} else if !reflect.DeepEqual(got, test.want) {
			t.Errorf("getTileRequestSpecs returned:\n%v\ninstead of:\n%v\nfor:\n%s", spew.Sdump(got), spew.Sdump(test.want), spew.Sdump(test.r.URL))
		}
	}
}
