package common

import (
	"encoding/json"
	"net/http"
)

// APIError represents a structured API error response.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeUnprocessableEntity = "UNPROCESSABLE_ENTITY"
	ErrCodeTooManyRequests     = "TOO_MANY_REQUESTS"
	ErrCodeInternalServer      = "INTERNAL_SERVER_ERROR"
)

// NewAPIError creates a new API error.
func NewAPIError(code, message string, details any) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// ErrBadRequest creates a bad request error.
func ErrBadRequest(message string) *APIError {
	return NewAPIError(ErrCodeBadRequest, message, nil)
}

// ErrUnauthorized creates an unauthorized error.
func ErrUnauthorized(message string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, nil)
}

// ErrForbidden creates a forbidden error.
func ErrForbidden(message string) *APIError {
	return NewAPIError(ErrCodeForbidden, message, nil)
}

// ErrNotFound creates a not found error.
func ErrNotFound(message string) *APIError {
	return NewAPIError(ErrCodeNotFound, message, nil)
}

// ErrConflict creates a conflict error.
func ErrConflict(message string) *APIError {
	return NewAPIError(ErrCodeConflict, message, nil)
}

// ErrTooManyRequests creates a rate limit error.
func ErrTooManyRequests(message string) *APIError {
	return NewAPIError(ErrCodeTooManyRequests, message, nil)
}

// ErrInternalServer creates an internal server error.
func ErrInternalServer(message string) *APIError {
	return NewAPIError(ErrCodeInternalServer, message, nil)
}

// WriteError writes an error response to the HTTP response writer.
func WriteError(w http.ResponseWriter, statusCode int, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(err)
}

// WriteJSON writes a JSON response to the HTTP response writer.
func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
