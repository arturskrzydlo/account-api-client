package api

import (
	"encoding/json"
	"errors"
	"fmt"
)

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

func (c *client) reqErrFromResponse(responseBody []byte, statusCode int) error {
	var errResponseBody ErrResponseBody
	err := json.Unmarshal(responseBody, &errResponseBody)
	if err != nil {
		// in case when error message is not in defined format try to get whole response as a string
		// I've noticed that there are differences in api and returned error message format i.e between 400 and 403.
		// Also when there is no body like for 404 we should be able to still return
		// requestErr but with empty error message
		return NewRequestErr(statusCode, errors.New(string(responseBody)))
	}
	return NewRequestErr(statusCode, errors.New(errResponseBody.ErrorMessage))
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.statusCode, r.errMsg)
}
