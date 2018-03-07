package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
)

// -------------------------------------------------------------------
// • DataError
// -------------------------------------------------------------------

// DataError defines the properties for a basic data error structure.
type DataError struct {
	Data interface{} `json:"data"`
}

// Error returns the data error message.
func (e *DataError) Error() string {
	return fmt.Sprintf("Invalid data: %v", e.Data)
}

// MarshalJSON handles json serialization
func (e *DataError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Data)
}

// NewDataError initilizes and returns a new DataError.
func NewDataError(data interface{}) *DataError {
	return &DataError{Data: data}
}

// -------------------------------------------------------------------
// • ApiError
// -------------------------------------------------------------------

// ApiError defines the properties for a basic api error response.
type ApiError struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Error returns the error message.
func (e *ApiError) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code.
func (e *ApiError) StatusCode() int {
	return e.Status
}

// NewApiError creates and returns new normalized ApiError instance.
func NewApiError(status int, message string, data interface{}) *ApiError {
	message = Sentenize(message)

	switch v := data.(type) {
	case map[string]string:
		// do nothing...
	case validation.Errors:
		// do nothing...
	case *DataError:
		data = v.Data
	case error:
		message += " " + Sentenize(v.Error())
		data = nil
	}

	return &ApiError{status, strings.TrimSpace(message), data}
}

// NewNotFoundError creates and returns 404 ApiError.
func NewNotFoundError(message string) *ApiError {
	if message == "" {
		message = "Oops, the requested resource wasn't found."
	}

	return NewApiError(http.StatusNotFound, message, nil)
}

// NewBadRequestError creates and returns 400 ApiError.
func NewBadRequestError(message string, data interface{}) *ApiError {
	if message == "" {
		message = "Oops, something went wrong while proceeding your request."
	}

	return NewApiError(http.StatusBadRequest, message, data)
}
