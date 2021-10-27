package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

var datacubeFields = map[string]wm.Field{
	"id":             wm.FieldDatacubeID,
	"type":           wm.FieldDatacubeType,
	"model":          wm.FieldDatacubeModel,
	"model_id":       wm.FieldDatacubeModelID,
	"category":       wm.FieldDatacubeCategory,
	"label":          wm.FieldDatacubeLabel,
	"maintainer":     wm.FieldDatacubeMaintainer,
	"source":         wm.FieldDatacubeSource,
	"output_name":    wm.FieldDatacubeOutputName,
	"output_units":   wm.FieldDatacubeOutputUnits,
	"parameters":     wm.FieldDatacubeParameters,
	"concepts.name":  wm.FieldDatacubeConceptName,
	"concepts.score": wm.FieldDatacubeConceptScore,
	"country":        wm.FieldDatacubeCountry,
	"admin1":         wm.FieldDatacubeAdmin1,
	"admin2":         wm.FieldDatacubeAdmin2,
	"period":         wm.FieldDatacubePeriod,
	"variable":       wm.FieldDatacubeVariable,
	"_search":        wm.FieldDatacubeSearch,
}

var indicatorFields = map[string]wm.Field{
	"variable":   wm.FieldIndicatorVariable,
	"dataset":    wm.FieldIndicatorDataset,
	"value_unit": wm.FieldIndicatorUnit,
}

var operands = map[string]wm.Operand{
	"and": wm.OperandAnd,
	"or":  wm.OperandOr,
}

// parseFilters extracts a slice of filters from a byte slice.
func parseFilters(raw []byte, context wm.FilterContext) ([]*wm.Filter, error) {
	op := "api.parseFilters"
	var fs []*wm.Filter

	if err := parseArray(raw, func(val []byte) error {
		f, err := parseFilter(val, context)
		if err != nil {
			return err
		}

		fs = append(fs, f)
		return nil
	}, "clauses"); err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	return fs, nil
}

// parseFilter extracts a single filter from a byte slice.
func parseFilter(raw []byte, context wm.FilterContext) (*wm.Filter, error) {
	fieldStr, err := jsonparser.GetString(raw, "field")
	if err != nil {
		return nil, err
	}
	var field wm.Field
	var ok bool

	switch context {
	case wm.ContextDatacube:
		field, ok = datacubeFields[fieldStr]
	case wm.ContextIndicator:
		field, ok = indicatorFields[fieldStr]
	default:
		return nil, fmt.Errorf("Unrecognized filter context")
	}
	if !ok {
		return nil, fmt.Errorf("Unrecognized field: %s", fieldStr)
	}

	operandStr, err := jsonparser.GetString(raw, "operand")
	if err != nil {
		return nil, err
	}

	operand, ok := operands[operandStr]
	if !ok {
		return nil, fmt.Errorf("Unrecognized operand: %s", operandStr)
	}

	isNot, err := jsonparser.GetBoolean(raw, "isNot")
	if err != nil {
		return nil, err
	}

	values, _, _, err := jsonparser.Get(raw, "values")
	if err != nil {
		return nil, err
	}

	strVals, intVals, rng, err := parseValues(field, values)
	if err != nil {
		return nil, err
	}

	return &wm.Filter{
		Field:        field,
		Operand:      operand,
		IsNot:        isNot,
		IntValues:    intVals,
		StringValues: strVals,
		Range:        rng,
	}, nil
}

// parseValues extracts the values field into an appropriately typed value.
func parseValues(field wm.Field, raw []byte) ([]string, []int, wm.Range, error) {
	var strVals []string
	var intVals []int
	var rng wm.Range
	var err error

	switch field {
	case
		wm.FieldDatacubeID,
		wm.FieldDatacubeType,
		wm.FieldDatacubeModel,
		wm.FieldDatacubeModelID,
		wm.FieldDatacubeCategory,
		wm.FieldDatacubeLabel,
		wm.FieldDatacubeMaintainer,
		wm.FieldDatacubeSource,
		wm.FieldDatacubeOutputName,
		wm.FieldDatacubeOutputUnits,
		wm.FieldDatacubeParameters,
		wm.FieldDatacubeConceptName,
		wm.FieldDatacubeCountry,
		wm.FieldDatacubeAdmin1,
		wm.FieldDatacubeAdmin2,
		wm.FieldDatacubeVariable,
		wm.FieldDatacubeSearch,
		wm.FieldIndicatorVariable,
		wm.FieldIndicatorDataset,
		wm.FieldIndicatorUnit:
		strVals, err = parseStringValues(raw)
	case
		wm.FieldDatacubePeriod,
		wm.FieldDatacubeConceptScore:
		rng, err = parseRange(raw)
	default:
		err = errors.New("parseValues failed: Unhandled values")
	}
	if err != nil {
		return nil, nil, wm.Range{}, err
	}

	return strVals, intVals, rng, nil
}

// parseArray wraps the error handling of parsing a JSON array.
func parseArray(raw []byte, cb func([]byte) error, keys ...string) error {
	var errs []string

	jsonparser.ArrayEach(raw, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			errs = append(errs, err.Error())
			return
		}

		if err := cb(value); err != nil {
			errs = append(errs, err.Error())
			return
		}
	}, keys...)

	if len(errs) > 0 {
		return fmt.Errorf("parseArray failed: %s", strings.Join(errs, ","))
	}

	return nil
}

// parseStringValues extracts the contents of the values field as a string slice.
func parseStringValues(raw []byte) ([]string, error) {
	var strVals []string
	if err := json.Unmarshal(raw, &strVals); err != nil {
		return nil, err
	}

	return strVals, nil
}

// parseIntValues extracts the contents of the values field as an int slice.
func parseIntValues(raw []byte) ([]int, error) {
	op := "api.parseIntValues"
	var intVals []int

	if err := parseArray(raw, func(val []byte) error {
		intVal, err := strconv.Atoi(string(val))
		if err != nil {
			return err
		}
		intVals = append(intVals, intVal)
		return nil
	}); err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	return intVals, nil
}

// parseRange extracts the contents of the values field as a 2-element float
// array.
func parseRange(raw []byte) (wm.Range, error) {
	op := "api.parseRange"
	var rng wm.Range

	var fs []float64
	if err := parseArray(raw, func(val []byte) error {
		return parseArray(val, func(subVal []byte) error {
			f, err := strconv.ParseFloat(string(subVal), 64)
			if err != nil {
				return err
			}
			fs = append(fs, f)
			return nil
		})
	}); err != nil {
		return rng, &wm.Error{Op: op, Err: err}
	}

	if len(fs) != 2 {
		return rng, &wm.Error{Op: op, Err: fmt.Errorf("Too many values (%d) for range filter", len(fs))}
	}

	rng.Minimum = fs[0]
	rng.Maximum = fs[1]

	return rng, nil
}
