package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arturskrzydlo/account-api-client/api/internal/models"
	"go.uber.org/zap"
)

const (
	baseURLV1 = "http://localhost:8080/v1"
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
		c.logger.Error("failed to create a new account", zap.Error(err))
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	if res.StatusCode >= 400 {
		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		var errResponseBody ErrResponseBody
		err = json.Unmarshal(responseBody, &errResponseBody)
		return NewRequestErr(400, errors.New(errResponseBody.ErrorMessage))
	}

	return nil
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
		c.logger.Error("failed to fetch an account", zap.Error(err), zap.String("accountID", accountID))
	}

	if res.StatusCode == http.StatusOK {
		body, err := io.ReadAll(res.Body)
		defer func() {
			if errClose := res.Body.Close(); errClose != nil {
				c.logger.Warn("failed to close response body", zap.Error(errClose))
			}
		}()
		if err != nil {
			c.logger.Error("failed to read pca response body: ", zap.Error(err))
		}

		var account models.AccountResponse
		if err := json.Unmarshal(body, &account); err != nil {
			c.logger.Error("failed to read pca response body: ", zap.Error(err))
		}
		return &account, nil
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, NewRequestErr(http.StatusNotFound, errors.New(fmt.Sprintf("can't find account with id %s", accountID)))
	}
	return nil, nil
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
		c.logger.Error("failed to delete an account", zap.Error(err), zap.String("accountID", accountID))
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	return nil
}

func NewAccountsClient(baseURL string) *Client {
	logger, _ := zap.NewProduction()
	if baseURL == "" {
		baseURL = baseURLV1
	}
	return &Client{
		baseURL:    baseURLV1,
		httpClient: &http.Client{Timeout: time.Second * 5},
		logger:     logger,
	}
}
