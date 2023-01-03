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

type AccountClient interface {
	CreateAccount(ctx context.Context, accountData *models.CreateAccountRequest) error
	FetchAccount(ctx context.Context, accountID string) (account *models.AccountResponse, err error)
	DeleteAccount(ctx context.Context, accountID string, version int64) error
}

type Client struct {
	baseURL    string
	logger     *zap.Logger
	httpClient *http.Client
}

type ResponseBody struct {
	ErrorMessage string `json:"error_message"`
}

func (c *Client) CreateAccount(ctx context.Context, accountData *models.CreateAccountRequest) error {
	reqBody, err := json.Marshal(accountData)
	if err != nil {
		c.logger.Error("failed to serialize account body", zap.Error(err))
	}
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/organisation/accounts", c.baseURL),
		bytes.NewBuffer(reqBody))
	if err != nil {
		c.logger.Error("failed to create a new account", zap.Error(err))
	}

	request = request.WithContext(ctx)
	request.Header.Set("content-type", "application/json")
	request.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to make request to an api : %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	return c.reqErrFromResponse(res)
}

func (c *Client) FetchAccount(ctx context.Context, accountID string) (account *models.AccountResponse, err error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/organisation/accounts/%s", c.baseURL, accountID), http.NoBody)
	if err != nil {
		c.logger.Error("failed to get accounts", zap.Error(err))
	}

	request = request.WithContext(ctx)
	request.Header.Set("content-type", "application/json")
	request.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to an api : %w", err)
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			c.logger.Warn("failed to close response body", zap.Error(errClose))
		}
	}()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			c.logger.Error("failed to read pca response body: ", zap.Error(err))
		}

		var account models.AccountResponse
		if err := json.Unmarshal(body, &account); err != nil {
			c.logger.Error("failed to read pca response body: ", zap.Error(err))
		}
		return &account, nil
	}

	return nil, c.reqErrFromResponse(res)
}

func (c *Client) DeleteAccount(ctx context.Context, accountID string, version int64) error {
	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/organisation/accounts/%s?version=%d", c.baseURL, accountID, version),
		http.NoBody)
	if err != nil {
		c.logger.Error("failed to delete a new account", zap.Error(err))
	}

	request = request.WithContext(ctx)
	request.Header.Set("content-type", "application/json")
	request.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to make request to an api : %w", err)
	}
	// TODO: handle close errors
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	return c.reqErrFromResponse(res)
}

func NewAccountsClient(baseURL string) (*Client, error) {
	logger, _ := zap.NewProduction()
	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url provided: %w", err)
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: time.Second * 5},
		logger:     logger,
	}, nil
}
