package api

import (
	"errors"
	"math"
	"net/http"
	"net/url"
	"time"
)

type Retrier struct {
	retryPolicy RetryPolicy
	backoff     BackOffStrategy
}

func (r Retrier) retry(fn func() (*http.Response, error)) (*http.Response, error) {
	maxRetries := r.retryPolicy.NumberOfRetries()
	retriesCount := 0
	res, err := fn()
	for {
		if !r.retryPolicy.ShouldRetry(err, res) {
			break
		}
		if retriesCount == maxRetries {
			return res, err
		}

		time.Sleep(r.backoff.delay(retriesCount))
		res, err = fn()
		retriesCount++
	}
	return res, err
}

type RetryPolicy interface {
	ShouldRetry(err error, response *http.Response) bool
	NumberOfRetries() int
}

type BackOffStrategy interface {
	delay(retryCount int) time.Duration
}

type DefaultRetryPolicy struct {
	MaxRetries int
}

type NoBackoffStrategy struct{}

func (n NoBackoffStrategy) delay(_ int) time.Duration {
	return 0
}

type LinearBackoffStrategy struct {
	delayTime time.Duration
}

func (l LinearBackoffStrategy) delay(_ int) time.Duration {
	return l.delayTime
}

func (mrp DefaultRetryPolicy) NumberOfRetries() int {
	return mrp.MaxRetries
}

type ExponentialBackoffStrategy struct {
	initialDelay time.Duration
	multiplier   int
}

func (e ExponentialBackoffStrategy) delay(retryCount int) time.Duration {
	multiplier := math.Pow(float64(e.multiplier), float64(retryCount))
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
