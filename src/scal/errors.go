package scal

import (
	"net/http"
)

var (
	BadRequest          = http.StatusBadRequest          // FIXME: REMOVE
	Forbidden           = http.StatusForbidden           // FIXME: REMOVE
	InternalServerError = http.StatusInternalServerError // FIXME: REMOVE
)
