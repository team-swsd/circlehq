package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/team-swsd/circlehq/internal/model"
)

func writeHealthResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	msg := "OK"
	status := http.StatusOK
	json.NewEncoder(w).Encode(HealthStatusResponse{
		Message: &msg,
		Status:  &status,
	})
}

func writeErrorResponse(w http.ResponseWriter, err error) {
	var resError *model.ResError
	if !errors.As(err, &resError) {
		resError = &model.ErrUnexpected
	}

	response := ErrorResponse{
		StatusCode: &resError.StatusCode,
		Detail:     &resError.Detail,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resError.StatusCode)
	json.NewEncoder(w).Encode(response)
}

func writeStatusOnlyResponse(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}
