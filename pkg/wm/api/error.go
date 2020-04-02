package api

import (
	"net/http"
)

func (a *api) errorResponse(w http.ResponseWriter, err error, status int) {
	a.logger.Errorw("API error",
		"err", err,
		"status", status,
	)
	http.Error(w, http.StatusText(status), status)
}
