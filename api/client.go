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
	defaultTimeout = time.Second * 5
)

type AccountClient interface {
	CreateAccount(ctx context.Context, accountData *models.CreateAccountRequest) error
	FetchAccount(ctx context.Context, accountID string) (account *models.AccountResponse, err error)
	DeleteAccount(ctx context.Context, accountID string, version int64) error
}

type client struct {
	baseURL    string
	logger     *zap.Logger
	httpClient *http.Client
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
		return fmt.Errorf("failed to create a new account: %w", err)
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

func (c *client) FetchAccount(ctx context.Context, accountID string) (account *models.AccountResponse, err error) {
	request, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/organisation/accounts/%s", c.baseURL, accountID), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to get accounts: %w", err)
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
		body, respErr := io.ReadAll(res.Body)
		if respErr != nil {
			return nil, fmt.Errorf("failed to read pca response body: %w", respErr)
		}

		var accountResponse models.AccountResponse
		if respErr = json.Unmarshal(body, &accountResponse); respErr != nil {
			return nil, fmt.Errorf("failed to read pca response body: %w", respErr)
		}
		return &accountResponse, nil
	}

	return nil, c.reqErrFromResponse(res)
}

func (c *client) DeleteAccount(ctx context.Context, accountID string, version int64) error {
	request, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/organisation/accounts/%s?version=%d", c.baseURL, accountID, version),
		http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to delete an account: %w", err)
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

func NewAccountsClient(baseURL string) (*client, error) {
	logger, _ := zap.NewProduction()
	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url provided: %w", err)
	}
	return &client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		logger:     logger,
	}, nil
}
