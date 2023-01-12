package accountclient

import (
	"encoding/json"
	"errors"
	"fmt"
)

type errResponseBody struct {
	ErrorMessage string `json:"error_message"`
}

type RequestError struct {
	StatusCode int
	ErrMsg     string
}

func newRequestErr(statusCode int, err error) *RequestError {
	return &RequestError{
		StatusCode: statusCode,
		ErrMsg:     err.Error(),
	}
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("status %d: error: %v", r.StatusCode, r.ErrMsg)
}

func (c *Client) reqErrFromResponse(responseBody []byte, statusCode int) error {
	var errResBody errResponseBody
	err := json.Unmarshal(responseBody, &errResBody)
	if err != nil {
		// in case when error message is not in defined format try to get whole response as a string
		// I've noticed that there are differences in api and returned error message format i.e. between 400 and 403.
		// Also, when there is no response body, like for 404 we should be able to still return
		// requestErr but with empty error message
		return newRequestErr(statusCode, errors.New(string(responseBody)))
	}
	return newRequestErr(statusCode, errors.New(errResBody.ErrorMessage))
}
