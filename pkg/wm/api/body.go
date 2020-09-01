package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// decodeJSONBody decodes json request body and stores it in the value pointed by dst
func decodeJSONBody(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&dst)
	if err != nil {
		var invalidUnmarshalError *json.InvalidUnmarshalError
		// catch and handle specific errors here
		switch {
		case errors.Is(err, io.EOF):
			msg := fmt.Sprintf("Request body must not be empty")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}
		case errors.As(err, &invalidUnmarshalError):
			return err
		default:
			msg := fmt.Sprintf("Request body is invalid")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}
		}
	}
	return nil
}
