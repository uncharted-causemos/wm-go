package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ClientError is an error whose details to be shared with client.
type ClientError interface {
	Error() string
	// ResponseBody returns response body.
	ResponseBody() ([]byte, error)
	// ResponseHeaders returns http status code and headers.
	ResponseHeaders() (int, map[string]string)
}

// HTTPError represent an http error
type HTTPError struct {
	Cause  error  `json:"-"`
	Detail string `json:"detail"`
	Status int    `json:"-"`
}

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Detail
	}
	return e.Detail + " : " + e.Cause.Error()
}

// ResponseBody returns JSON response body.
func (e *HTTPError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing response body: %v", err)
	}
	return body, nil
}

// ResponseHeaders returns http status code and headers.
func (e *HTTPError) ResponseHeaders() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

// NewHTTPError creates new HTTPError
func NewHTTPError(err error, status int, detail string) error {
	return &HTTPError{
		Cause:  err,
		Detail: detail,
		Status: status,
	}
}

func (a *api) errorResponse(w http.ResponseWriter, err error, status int) {
	// default error code and message
	code := status
	errMsg := http.StatusText(code)

	// If error is HTTPError
	var httpError *HTTPError
	if errors.As(err, &httpError) {
		code = httpError.Status
		body, err := httpError.ResponseBody()
		if err != nil {
			code = http.StatusInternalServerError
		}
		errMsg = string(body)
	}

	a.logger.Errorw("API error",
		"err", err,
		"status", code,
	)
	http.Error(w, errMsg, code)
}
