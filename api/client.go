package api

import (
	"context"
	"net/http"
	"time"

	"github.com/arturskrzydlo/account-api-client/pkg"
	"go.uber.org/zap"
)

const (
	baseURLV1 = "http://localhost:8080/v1"
)

type AccountClient interface {
	CreateAccount(ctx context.Context, accountData *pkg.AccountData) error
	FetchAccount(ctx context.Context, accountId string) (account *pkg.AccountData, err error)
	DeleteAccount(ctx context.Context, accountId string) error
}

type Client struct {
	baseURL    string
	apiKey     string
	logger     *zap.Logger
	httpClient *http.Client
}

func (c *Client) CreateAccount(ctx context.Context, accountData *pkg.AccountData) error {
	//TODO implement me
	panic("implement me")
}

func (c *Client) FetchAccount(ctx context.Context, accountId string) (account *pkg.AccountData, err error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) DeleteAccount(ctx context.Context, accountId string) error {
	//TODO implement me
	panic("implement me")
}

func NewClient(apiKey string) *Client {
	logger, _ := zap.NewProduction()

	return &Client{apiKey: apiKey,
		baseURL:    baseURLV1,
		httpClient: &http.Client{Timeout: time.Second * 20},
		logger:     logger,
	}
}
