package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

// decodeJSONBody decodes json request body and stores it in the value pointed by dst
func decodeJSONBody(r *http.Request, dst interface{}) error {
	op := "decodeJSONBody"
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&dst)
	if err != nil {
		var invalidUnmarshalError *json.InvalidUnmarshalError
		// catch and handle specific errors here
		switch {
		case errors.Is(err, io.EOF):
			return &wm.Error{Code: wm.EINVALID, Message: "Request body must not be empty"}
		case errors.As(err, &invalidUnmarshalError):
			return &wm.Error{Op: op, Err: err}
		default:
			return &wm.Error{Code: wm.EINVALID, Message: "Request body is invalid"}
		}
	}
	return nil
}
