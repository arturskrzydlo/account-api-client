package api

import (
	"errors"
	"math"
	"net/http"
	"net/url"
	"time"
)

type RetryPolicy interface {
	ShouldRetry(err error, response *http.Response) bool
	NumberOfRetries() int
}

type BackOffStrategy interface {
	delay() time.Duration
}

type DefaultRetryPolicy struct {
	MaxRetries int
}

type NoBackoffStrategy struct{}

func (n NoBackoffStrategy) delay() time.Duration {
	return 0
}

type LinearBackoffStrategy struct {
	delayTime time.Duration
}

func (l LinearBackoffStrategy) delay() time.Duration {
	return l.delayTime
}

func (mrp DefaultRetryPolicy) NumberOfRetries() int {
	return mrp.MaxRetries
}

type ExponentialBackoffStrategy struct {
	initialDelay time.Duration
	retryCount   int
	multiplier   int
}

func (e *ExponentialBackoffStrategy) delay() time.Duration {
	multiplier := math.Pow(float64(e.multiplier), float64(e.retryCount))
	e.retryCount++
	return e.initialDelay * time.Duration(multiplier)
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

func retry(retryPolicy RetryPolicy, backoff BackOffStrategy, fn func() (*http.Response, error)) (*http.Response, error) {
	retriesLeft := retryPolicy.NumberOfRetries()
	res, err := fn()
	for {
		if !retryPolicy.ShouldRetry(err, res) {
			break
		}
		if retriesLeft == 0 {
			return res, err
		}

		time.Sleep(backoff.delay())
		res, err = fn()
		retriesLeft--
	}
	return res, err
}
