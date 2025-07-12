package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with context
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Cause   error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// New creates a new AppError with the given status, code, and message
func New(status int, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// NewWithCause creates a new AppError with a cause
func NewWithCause(status int, code, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Cause:   cause,
	}
}

// Common error constructors for convenience
func NewBadRequest(message string) *AppError {
	return New(http.StatusBadRequest, "BAD_REQUEST", message)
}

func NewUnauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func NewForbidden(message string) *AppError {
	return New(http.StatusForbidden, "FORBIDDEN", message)
}

func NewNotFound(message string) *AppError {
	return New(http.StatusNotFound, "NOT_FOUND", message)
}

func NewInternalError(message string, cause error) *AppError {
	return NewWithCause(http.StatusInternalServerError, "INTERNAL_ERROR", message, cause)
}
