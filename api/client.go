package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/arturskrzydlo/account-api-client/api/internal/models"
	"go.uber.org/zap"
)

const (
	defaultTimeout = time.Second * 10
	jsonType       = "application/json"
)

type AccountClient interface {
	CreateAccount(ctx context.Context, accountData *models.CreateAccountRequest) error
	FetchAccount(ctx context.Context, accountID string) (account *models.AccountResponse, err error)
	DeleteAccount(ctx context.Context, accountID string, version int64) error
}

type client struct {
	baseURL     string
	logger      *zap.Logger
	httpClient  *http.Client
	retryPolicy RetryPolicy
}

type ResponseBody struct {
	ErrorMessage string `json:"error_message"`
}

func (c *client) CreateAccount(ctx context.Context, accountData *models.CreateAccountRequest) error {
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

func (c *client) FetchAccount(ctx context.Context, accountID string) (account *models.AccountResponse, err error) {
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

func (c *client) DeleteAccount(ctx context.Context, accountID string, version int64) error {
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

func (c *client) sendRequest(ctx context.Context, request *http.Request, result interface{}) error {
	request = request.WithContext(ctx)
	setContentType(request)

	res, err := retry(c.retryPolicy, func() (*http.Response, error) {
		res, err := c.httpClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf("failed to make request to an api : %w", err)
		}

		return res, nil
	})
	if err != nil {
		return fmt.Errorf("failed to send request : %w", err)
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			c.logger.Warn("failed to close response body", zap.Error(errClose))
		}
	}()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return c.reqErrFromResponse(resBody, res.StatusCode)
	}

	if result != nil {
		if err = json.Unmarshal(resBody, result); err != nil {
			return fmt.Errorf("failed to unmarshall response body: %w", err)
		}
	}

	return nil
}

func setContentType(req *http.Request) string {
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", jsonType)
		contentType = jsonType
	}
	return contentType
}

func NewAccountsClient(baseURL string, httpClient *http.Client, retryPolicy RetryPolicy) (*client, error) {
	logger, _ := zap.NewProduction()
	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url provided: %w", err)
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	if retryPolicy == nil {
		retryPolicy = DefaultRetryPolicy{}
	}

	return &client{
		baseURL:     baseURL,
		httpClient:  httpClient,
		logger:      logger,
		retryPolicy: retryPolicy,
	}, nil
}
