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
	backoff     BackoffStrategy
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

		time.Sleep(r.backoff.Delay(retriesCount))
		resetBody(request, originalBody)
		res, err = fn(request)
		retriesCount++
	}
	return res, err
}

// RetryPolicy allows to create custom policy for errors on which library will try to retry request
type RetryPolicy interface {
	// ShouldRetry based on error and http.Response decides if request should be retried
	ShouldRetry(err error, response *http.Response) bool
	// NumberOfRetries return how many times request should be retried
	NumberOfRetries() int
}

// BackoffStrategy allows to define strategy to make delays between next retries
type BackoffStrategy interface {
	// Delay returns how long should last a pause between next retries in time.Duration format.
	// Output might be dependent on current retryCount
	Delay(retryCount int) time.Duration
}

// DefaultRetryPolicy is simple policy which will retry when there is an error coming from http.Client (*url.Error) or response status code
// is server side status code (5xx)
type DefaultRetryPolicy struct {
	// maxRetries how many retries should be applied in retry process
	maxRetries int
}

// NoBackoffStrategy is marker of no delay (no backoff) between retries
type NoBackoffStrategy struct{}

func (n NoBackoffStrategy) Delay(_ int) time.Duration {
	return 0
}

// LinearBackoffStrategy is backoff in which delays between retries is constant
type LinearBackoffStrategy struct {
	// delayTime time between next retries
	delayTime time.Duration
}

func (l LinearBackoffStrategy) Delay(_ int) time.Duration {
	return l.delayTime
}

func (mrp DefaultRetryPolicy) NumberOfRetries() int {
	return mrp.maxRetries
}

// ExponentialBackoffStrategy is backoff in which delays growing exponentially between next retries
type ExponentialBackoffStrategy struct {
	initialDelay time.Duration
	multiplier   int
}

func (e ExponentialBackoffStrategy) Delay(retryCount int) time.Duration {
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
