package api

import (
	"context"
	"net/http"
	"time"

	"github.com/arturskrzydlo/account-api-client/api/internal/models"
	"go.uber.org/zap"
)

const (
	baseURLV1 = "http://localhost:8080/v1"
)

type AccountClient interface {
	CreateAccount(ctx context.Context, accountData *models.AccountData) error
	FetchAccount(ctx context.Context, accountID string) (account *models.AccountData, err error)
	DeleteAccount(ctx context.Context, accountID string) error
}

type Client struct {
	baseURL    string
	apiKey     string
	logger     *zap.Logger
	httpClient *http.Client
}

func (c *Client) CreateAccount(ctx context.Context, accountData *models.AccountData) error {
	// TODO implement me
	panic("implement me")
}

func (c *Client) FetchAccount(ctx context.Context, accountId string) (account *models.AccountData, err error) {
	// TODO implement me
	panic("implement me")
}

func (c *Client) DeleteAccount(ctx context.Context, accountId string) error {
	// TODO implement me
	panic("implement me")
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
