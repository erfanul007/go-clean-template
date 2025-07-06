package response

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
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
	response := Response{
		Success: true,
		Data:    data,
	}
	JSON(w, http.StatusOK, response)
}

// SuccessWithMeta creates a successful response with metadata
func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	response := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
	JSON(w, http.StatusOK, response)
}

// Error creates an error response
func Error(w http.ResponseWriter, status int, code, message string) {
	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	JSON(w, status, response)
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log the error but don't return it since headers are already written
		// In a real application, you might want to use a proper logger here
		_ = err
	}
}
