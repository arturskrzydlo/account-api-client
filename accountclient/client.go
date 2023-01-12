// Package accountclient allows to make operations on fake form3 account api
//
// This is package for http client which is consuming fake form3 account api. This api needs to be run locally,
// and it can be done by running this https://github.com/form3tech-oss/interview-accountapi/blob/master/docker-compose.yml
// This package was created as a result of interview coding task
package accountclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/google/uuid"

	"github.com/arturskrzydlo/account-api-client/accountclient/models"
)

const (
	// http.Client default timeout. It is default assigned to http.Client if it hasn't been configured on Client creation
	defaultTimeout     = time.Second * 10
	jsonType           = "application/json"
	hystrixCommandName = "account-client"
	// its threshold measured int percentages of errors in all requests which tells circuit breaker to open
	defaultHystrixErrorPercentageThreshold = 30
)

// Client which performs rest api operations
//
// All the operations are wrapped by circuit breaker to avoid flooding api server with invalid requests
// Circuit breaker reacts both on 4xx error code like 5xx error codes.
// Depending on configuration in ClientConfig requests might be also retries. By default, retries are switched off
type Client struct {
	baseURL    string
	httpClient *http.Client
	retrier    retrier
}

// NewAccountClient creates Client - we have to pass baseURL which has no default value as fake account api has no permanent address
// baseURL must be a valid URL otherwise creation of a new account will finish with error
// ClientOption are optional parameters which modify ClientConfig. Those parameters are evaluated and can modify ClientConfig
func NewAccountClient(baseURL string, options ...ClientOption) (*Client, error) {
	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url provided: %w", err)
	}

	// default client config
	cfg := ClientConfig{
		HTTPClient:      &http.Client{Timeout: defaultTimeout},
		RetryPolicy:     DefaultRetryPolicy{maxRetries: 0},
		BackoffStrategy: NoBackoffStrategy{},
	}

	for _, option := range options {
		option(&cfg)
	}

	hystrix.ConfigureCommand(hystrixCommandName, hystrix.CommandConfig{
		ErrorPercentThreshold: defaultHystrixErrorPercentageThreshold,
		Timeout:               int(cfg.HTTPClient.Timeout.Milliseconds()),
	})

	return &Client{
		baseURL:    baseURL,
		httpClient: cfg.HTTPClient,
		retrier: retrier{
			retryPolicy: cfg.RetryPolicy,
			backoff:     cfg.BackoffStrategy,
		},
	}, nil
}

// ClientOption is function which can modify ClientConfig
type ClientOption func(config *ClientConfig)

// ClientConfig contains currently all modifiable parameters
type ClientConfig struct {
	// HTTPClient param can be used to create custom http.Client
	HTTPClient *http.Client
	// RetryPolicy allows to set custom RetryPolicy
	RetryPolicy RetryPolicy
	// BackoffStrategy allows to defined strategy to make delays between next retries
	BackoffStrategy BackoffStrategy
}

// WithRetriesOnDefaultRetryPolicy is a predefined DefaultRetryPolicy to use in NewAccountClient
func WithRetriesOnDefaultRetryPolicy(maxRetries int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RetryPolicy = DefaultRetryPolicy{maxRetries: maxRetries}
	}
}

// WithCustomHTTPClient is a predefined option to create custom http.Client to use in NewAccountClient
func WithCustomHTTPClient(httpClient *http.Client) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.HTTPClient = httpClient
	}
}

// WithCustomRetryPolicy is a predefined option allowing to create own custom RetryPolicy
func WithCustomRetryPolicy(retryPolicy RetryPolicy) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RetryPolicy = retryPolicy
	}
}

// WithCustomBackoffStrategy is a predefined option allowing to create own custom BackoffStrategy
func WithCustomBackoffStrategy(backoff BackoffStrategy) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.BackoffStrategy = backoff
	}
}

// WithExponentialBackoffStrategy is a predefined ExponentialBackoffStrategy option to be added on NewAccountClient creation
func WithExponentialBackoffStrategy(initialDelay time.Duration, multiplier int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.BackoffStrategy = &ExponentialBackoffStrategy{
			initialDelay: initialDelay,
			multiplier:   multiplier,
		}
	}
}

// WithLinearBackoffStrategy is a predefined LinearBackoffStrategy option to be added on NewAccountClient creation
func WithLinearBackoffStrategy(delay time.Duration) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.BackoffStrategy = LinearBackoffStrategy{delayTime: delay}
	}
}

// CreateAccount creates account based on models.CreateAccountRequest data
// If there will be 4xx or 500x error it can be in a form of RequestError, but currently not all 4xx errors are in the same format
// In that case error msg will remain empty and only status code will be available
// Other errors are returned as simple errors
func (c *Client) CreateAccount(ctx context.Context, accountData *models.CreateAccountRequest) (*models.AccountResponse, error) {
	reqBody, err := json.Marshal(accountData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize account body: %w", err)
	}
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/organisation/accounts", c.baseURL),
		bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create a request to create a new account: %w", err)
	}

	var accountResponse models.AccountResponse
	err = c.sendRequest(ctx, request, &accountResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to send create account request: %w", err)
	}

	return &accountResponse, nil
}

// FetchAccount fetches account in form of models.AccountResponse. To fetch account existing accountID should be provided
// If there will be 4xx or 500x error it can be in a form of RequestError, but currently not all 4xx errors are in the same format
// In that case error msg will remain empty and only status code will be available
// Other errors are returned as simple errors
func (c *Client) FetchAccount(ctx context.Context, accountID uuid.UUID) (account *models.AccountResponse, err error) {
	request, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/organisation/accounts/%s", c.baseURL, accountID.String()), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create fetche account request: %w", err)
	}

	var accountResponse models.AccountResponse
	err = c.sendRequest(ctx, request, &accountResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to send fetch account request: %w", err)
	}
	return &accountResponse, nil
}

// DeleteAccount delete existing account based on accountID. Also version of account must be provided. Version field is updated with each update
// and can be obtained from models.AccountResponse
// If there will be 4xx or 500x error it can be in a form of RequestError, but currently not all 4xx errors are in the same format
// In that case error msg will remain empty and only status code will be available
// Other errors are returned as simple errors
func (c *Client) DeleteAccount(ctx context.Context, accountID uuid.UUID, version *int64) error {
	request, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/organisation/accounts/%s?version=%d", c.baseURL, accountID, *version),
		http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to delete account request: %w", err)
	}

	err = c.sendRequest(ctx, request, nil)
	if err != nil {
		return fmt.Errorf("failed to send delete account request: %w", err)
	}

	return nil
}

func (c *Client) sendRequest(ctx context.Context, request *http.Request, result interface{}) error {
	request = request.WithContext(ctx)
	setContentType(request)

	var resBody []byte
	err := hystrix.Do(hystrixCommandName, func() error {
		body, err := c.sendRequestWithRetries(request)
		resBody = body
		return err
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to send request to an api: %w", err)
	}

	if result != nil && resBody != nil {
		if err = json.Unmarshal(resBody, result); err != nil {
			return fmt.Errorf("failed to unmarshall response body: %w", err)
		}
	}

	return nil
}

func (c *Client) sendRequestWithRetries(request *http.Request) ([]byte, error) {
	res, err := c.retrier.retry(request, func(req *http.Request) (*http.Response, error) {
		response, resErr := c.httpClient.Do(request)
		if resErr != nil {
			return nil, fmt.Errorf("failed to make request to an api : %w", resErr)
		}

		return response, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send request : %w", err)
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			logger := log.New(os.Stderr, "", 0)
			logger.Printf("failed to close response body: %s", err.Error())
		}
	}()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return nil, c.reqErrFromResponse(resBody, res.StatusCode)
	}

	return resBody, nil
}

func setContentType(req *http.Request) string {
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", jsonType)
		contentType = jsonType
	}
	return contentType
}
