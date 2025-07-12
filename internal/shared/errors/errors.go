package errors

import (
	"net/http"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Cause   error  `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

func NewAppErrorWithCause(code, message string, status int, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Cause:   cause,
	}
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func BadRequest(code, message string) *AppError {
	return NewAppError(code, message, http.StatusBadRequest)
}

func BadRequestWithCause(code, message string, cause error) *AppError {
	return NewAppErrorWithCause(code, message, http.StatusBadRequest, cause)
}

func Unauthorized(code, message string) *AppError {
	return NewAppError(code, message, http.StatusUnauthorized)
}

func UnauthorizedWithCause(code, message string, cause error) *AppError {
	return NewAppErrorWithCause(code, message, http.StatusUnauthorized, cause)
}

func Forbidden(code, message string) *AppError {
	return NewAppError(code, message, http.StatusForbidden)
}

func ForbiddenWithCause(code, message string, cause error) *AppError {
	return NewAppErrorWithCause(code, message, http.StatusForbidden, cause)
}

func NotFound(code, message string) *AppError {
	return NewAppError(code, message, http.StatusNotFound)
}

func NotFoundWithCause(code, message string, cause error) *AppError {
	return NewAppErrorWithCause(code, message, http.StatusNotFound, cause)
}

func Conflict(code, message string) *AppError {
	return NewAppError(code, message, http.StatusConflict)
}

func ConflictWithCause(code, message string, cause error) *AppError {
	return NewAppErrorWithCause(code, message, http.StatusConflict, cause)
}

func InternalServer(code, message string) *AppError {
	return NewAppError(code, message, http.StatusInternalServerError)
}

func InternalServerWithCause(code, message string, cause error) *AppError {
	return NewAppErrorWithCause(code, message, http.StatusInternalServerError, cause)
}
