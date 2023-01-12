package accountclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"go.uber.org/zap"

	"github.com/arturskrzydlo/account-api-client/accountclient/models"
)

const (
	defaultTimeout                         = time.Second * 10
	jsonType                               = "application/json"
	hystrixCommandName                     = "account-client"
	defaultHystrixErrorPercentageThreshold = 30
)

type Client struct {
	baseURL    string
	logger     *zap.Logger
	httpClient *http.Client
	retrier    retrier
}

func NewAccountClient(baseURL string, options ...ClientOption) (*Client, error) {
	logger, _ := zap.NewProduction()
	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url provided: %w", err)
	}

	// default client config
	cfg := ClientConfig{
		HTTPClient:      &http.Client{Timeout: defaultTimeout},
		RetryPolicy:     DefaultRetryPolicy{MaxRetries: 0},
		BackoffStrategy: NoBackoffStrategy{},
	}

	for _, option := range options {
		option(&cfg)
	}

	hystrix.ConfigureCommand(hystrixCommandName, hystrix.CommandConfig{
		ErrorPercentThreshold: defaultHystrixErrorPercentageThreshold,
		Timeout:               int(defaultTimeout.Milliseconds()),
	})

	return &Client{
		baseURL:    baseURL,
		httpClient: cfg.HTTPClient,
		logger:     logger,
		retrier: retrier{
			retryPolicy: cfg.RetryPolicy,
			backoff:     cfg.BackoffStrategy,
		},
	}, nil
}

type ClientOption func(config *ClientConfig)

type ClientConfig struct {
	HTTPClient      *http.Client
	RetryPolicy     RetryPolicy
	BackoffStrategy BackOffStrategy
}

func WithRetriesOnDefaultRetryPolicy(maxRetries int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RetryPolicy = DefaultRetryPolicy{MaxRetries: maxRetries}
	}
}

func WithCustomHTTPClient(httpClient *http.Client) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.HTTPClient = httpClient
	}
}

func WithExponentialBackoffStrategy(initialDelay time.Duration, multiplier int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.BackoffStrategy = &ExponentialBackoffStrategy{
			initialDelay: initialDelay,
			multiplier:   multiplier,
		}
	}
}

func WithLinearBackoffStrategy(delay time.Duration) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.BackoffStrategy = LinearBackoffStrategy{delayTime: delay}
	}
}

func (c *Client) CreateAccount(ctx context.Context, accountData *models.CreateAccountRequest) error {
	reqBody, err := json.Marshal(accountData)
	if err != nil {
		return fmt.Errorf("failed to serialize account body: %w", err)
	}
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/organisation/accounts", c.baseURL),
		bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create a request to create a new account: %w", err)
	}

	err = c.sendRequest(ctx, request, nil)
	if err != nil {
		return fmt.Errorf("failed to send create account request: %w", err)
	}

	return nil
}

func (c *Client) FetchAccount(ctx context.Context, accountID string) (account *models.AccountResponse, err error) {
	request, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/organisation/accounts/%s", c.baseURL, accountID), http.NoBody)
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

func (c *Client) DeleteAccount(ctx context.Context, accountID string, version int64) error {
	request, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/organisation/accounts/%s?version=%d", c.baseURL, accountID, version),
		http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create delete account request: %w", err)
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
	res, err := c.retrier.retry(func() (*http.Response, error) {
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
			c.logger.Warn("failed to close response body", zap.Error(errClose))
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
