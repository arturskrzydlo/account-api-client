package api

import (
	"errors"
	"net/http"
	"net/url"
)

type RetryPolicy interface {
	ShouldRetry(err error, response *http.Response) bool
	NumberOfRetries() int
}

type DefaultRetryPolicy struct {
	MaxRetries int
}

func (mrp DefaultRetryPolicy) NumberOfRetries() int {
	return mrp.MaxRetries
}

func (mrp DefaultRetryPolicy) ShouldRetry(err error, response *http.Response) bool {
	if response == nil && err == nil {
		return false
	}

	errFromHTTPClient := false
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			errFromHTTPClient = true
		}
	}

	serverSideStatusCode := false
	if response != nil && response.StatusCode >= 500 {
		serverSideStatusCode = true
	}

	return errFromHTTPClient || serverSideStatusCode
}

func retry(retryPolicy RetryPolicy, fn func() (*http.Response, error)) (*http.Response, error) {
	retriesLeft := retryPolicy.NumberOfRetries()
	res, err := fn()
	for {
		if !retryPolicy.ShouldRetry(err, res) {
			break
		}
		if retriesLeft == 0 {
			return res, err
		}
		res, err = fn()
		retriesLeft--
	}
	return res, err
}
