package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-clean-template/internal/shared/errors"
)

type SuccessResponse struct {
	Meta *Meta `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Error *ErrorInfo `json:"error"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func Success(w http.ResponseWriter, data interface{}) {
	sendJSON(w, http.StatusOK, data)
}

func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	response := struct {
		Data interface{} `json:"data"`
		Meta *Meta       `json:"meta"`
	}{
		Data: data,
		Meta: meta,
	}
	sendJSON(w, http.StatusOK, response)
}

func Error(w http.ResponseWriter, status int, code, message string) {
	sendJSON(w, status, ErrorResponse{
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

func ErrorFromAppError(w http.ResponseWriter, err *errors.AppError) {
	sendJSON(w, err.Status, ErrorResponse{
		Error: &ErrorInfo{
			Code:    err.Code,
			Message: err.Message,
		},
	})
}

func GetErrorChain(err *errors.AppError) string {
	if err.Cause == nil {
		return err.Message
	}
	return fmt.Sprintf("%s: %v", err.Message, err.Cause)
}

func GetFullErrorChain(err *errors.AppError) []string {
	var chain []string
	chain = append(chain, err.Message)

	current := err.Cause
	for current != nil {
		chain = append(chain, current.Error())
		if unwrapper, ok := current.(interface{ Unwrap() error }); ok {
			current = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return chain
}

func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
