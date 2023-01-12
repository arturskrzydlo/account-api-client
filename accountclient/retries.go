package accountclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
)

type retrier struct {
	retryPolicy RetryPolicy
	backoff     BackOffStrategy
}

func (r retrier) retry(request *http.Request, fn func(request *http.Request) (*http.Response, error)) (*http.Response, error) {
	var originalBody []byte
	var err error

	maxRetries := r.retryPolicy.NumberOfRetries()
	retriesCount := 0

	// need to copy body between retries because body is closed on each
	// request call automatically
	if request != nil && request.Body != http.NoBody {
		originalBody, err = copyBody(request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed fo copy request body: %w", err)
		}
		resetBody(request, originalBody)
	}

	res, err := fn(request)

	for {
		if !r.retryPolicy.ShouldRetry(err, res) {
			break
		}
		if retriesCount == maxRetries {
			return res, err
		}

		time.Sleep(r.backoff.delay(retriesCount))
		resetBody(request, originalBody)
		res, err = fn(request)
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

func resetBody(request *http.Request, originalBody []byte) {
	request.Body = io.NopCloser(bytes.NewBuffer(originalBody))
}

func copyBody(src io.ReadCloser) ([]byte, error) {
	body, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	err = src.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close request body: %w", err)
	}
	return body, nil
}
