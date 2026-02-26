package errors

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

var (
	ErrNotFound     = ErrorResponse{Code: "not_found", Message: "The requested resource was not found"}
	ErrUnauthorized = ErrorResponse{Code: "unauthorized", Message: "Authentication is required"}
	ErrBadGateway   = ErrorResponse{Code: "bad_gateway", Message: "Upstream service is unavailable"}
	ErrInternal     = ErrorResponse{Code: "internal_error", Message: "An unexpected error occurred"}
)

func WriteJSON(w http.ResponseWriter, status int, resp ErrorResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
