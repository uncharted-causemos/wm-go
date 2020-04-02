package api

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
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
						{"field":"cause","operand":"or","isNot":false,"values":["a","b","c"]}
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
						{"field":"cause","operand":"or","isNot":false,"values":["a","b","c"]},
						{"field":"effect","operand":"or","isNot":false,"values":["d","e"]}
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
		got, err := getFilters(test.r)
		if err != nil {
			if !test.isErr {
				t.Errorf("getFilters returned err:\n%v\nfor:\n%v", err, spew.Sdump(test))
			}
		} else if len(got) != test.wantLen {
			t.Errorf("getFilters returned %d filters instead of %d for:\n%v\ngot:%v", len(got), test.wantLen, spew.Sdump(test.r), spew.Sdump(got))
		}
	}
}
