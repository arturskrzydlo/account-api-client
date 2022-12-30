package api

import "fmt"

type ErrResponseBody struct {
	ErrorMessage string `json:"error_message"`
}

type RequestError struct {
	statusCode int
	errMsg     string
}

func NewRequestErr(statusCode int, err error) *RequestError {
	return &RequestError{
		statusCode: statusCode,
		errMsg:     err.Error(),
	}
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.statusCode, r.errMsg)
}
