package model

import (
	"fmt"
	"net/http"
)

type ResError struct {
	StatusCode int    `json:"status_code,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

var (
	ErrNotFound = ResError{
		StatusCode: http.StatusNotFound,
		Detail:     "not found",
	}
	ErrMethodNotAllowed = ResError{
		StatusCode: http.StatusMethodNotAllowed,
		Detail:     "method not allowed",
	}
	ErrBadRequest = ResError{
		StatusCode: http.StatusBadRequest,
		Detail:     "bad request",
	}
	ErrDBConnection = ResError{
		StatusCode: http.StatusInternalServerError,
		Detail:     "DB Error",
	}
	ErrUnexpected = ResError{
		StatusCode: http.StatusInternalServerError,
		Detail:     "unexpected error",
	}
)

func (e *ResError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Detail)
}
