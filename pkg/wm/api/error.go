package api

import (
	"net/http"
)

// malformedRequest represent an error caused by malformed/invalid request syntax
type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func (a *api) errorResponse(w http.ResponseWriter, err error, status int) {
	a.logger.Errorw("API error",
		"err", err,
		"status", status,
	)
	http.Error(w, http.StatusText(status), status)
}
