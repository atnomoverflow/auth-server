package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type JSONResponse struct {
	StatusCode int
	Message    any
}

func (resp *JSONResponse) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(resp.Message)
}

type ApiError struct {
	Param   string `json:"param"`
	Message string `json:"message"`
}

func ValidationErrors(err error) (*JSONResponse, error) {
	var apiErrors []ApiError
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, err
	}
	for _, err := range validationErrs {
		fieldName := err.StructField()
		var message string

		switch err.Tag() {
		case "required":
			message = fmt.Sprintf("%s is required", fieldName)
		case "email":
			message = fmt.Sprintf("%s must be a valid email address", fieldName)
		default:
			message = fmt.Sprintf("%s is invalid", fieldName)
		}

		apiError := ApiError{
			Param:   fieldName,
			Message: message,
		}

		apiErrors = append(apiErrors, apiError)
	}
	return &JSONResponse{
		StatusCode: http.StatusBadRequest,
		Message:    apiErrors,
	}, nil
}

type RawMessage struct {
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	resp := JSONResponse{
		StatusCode: statusCode,
		Message:    RawMessage{Message: message},
	}
	resp.Write(w)
}

var (
	OK = func(w http.ResponseWriter, responce any) {
		resp := JSONResponse{
			StatusCode: http.StatusOK,
			Message:    responce,
		}
		resp.Write(w)
	}
)
var (
	InternalServerError = func(w http.ResponseWriter) {
		WriteError(w, http.StatusInternalServerError, "Something went wrong, please try again later!")
	}
	ConflictError = func(w http.ResponseWriter, msg string) {
		WriteError(w, http.StatusConflict, msg)
	}
	ForbidenError = func(w http.ResponseWriter, msg string) {
		WriteError(w, http.StatusForbidden, msg)
	}
	NotFoundError = func(w http.ResponseWriter) {
		WriteError(w, http.StatusForbidden, "Not Found!")
	}
	UnauthorizedError = func(w http.ResponseWriter, message *string) {
		msg := "Unauthorized"
		if message != nil {
			msg = *message
		}
		WriteError(w, http.StatusUnauthorized, msg)
	}
)
