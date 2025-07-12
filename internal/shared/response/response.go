package response

import (
	"encoding/json"
	"net/http"

	"go-clean-template/internal/shared/errors"
)

// SuccessResponse represents a successful API response with metadata
type SuccessResponse struct {
	Meta *Meta `json:"meta,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error *ErrorInfo `json:"error"`
}

// ErrorInfo represents error information in response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta represents metadata for responses (pagination, etc.)
type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Success creates a successful response
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// SuccessWithMeta creates a successful response with metadata
func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	// For responses with metadata, we need to create a wrapper
	response := struct {
		Data interface{} `json:"data"`
		Meta *Meta       `json:"meta"`
	}{
		Data: data,
		Meta: meta,
	}
	JSON(w, http.StatusOK, response)
}

// Error creates an error response
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorResponse{
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorFromAppError creates an error response from an AppError
func ErrorFromAppError(w http.ResponseWriter, err *errors.AppError) {
	JSON(w, err.Status, ErrorResponse{
		Error: &ErrorInfo{
			Code:    err.Code,
			Message: err.Message,
		},
	})
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data) // Ignore encoding errors as headers are already written
}
