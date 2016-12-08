package scal

import (
	"net/http"
)

var (
	BadRequest          = http.StatusBadRequest
	Forbidden           = http.StatusForbidden
	InternalServerError = http.StatusInternalServerError
)
