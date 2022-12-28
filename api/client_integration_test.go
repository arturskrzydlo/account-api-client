//go:build integration
// +build integration

package api

import (
	"context"
	"testing"

	"github.com/arturskrzydlo/account-api-client/api/internal/models"
	"github.com/stretchr/testify/suite"
)

type accountApiClientSuite struct {
	suite.Suite

	accountApiClient *Client
}

func TestAccountApiClient(t *testing.T) {
	suite.Run(t, &accountApiClientSuite{})
}

func (s *accountApiClientSuite) SetupSuite() {
	s.accountApiClient = NewClient("test-api-key")
}

func (s *accountApiClientSuite) TestCreateAccount() {
	s.Run("should successfully create single account", func() {
		s.Assert().NotNil(s.accountApiClient.CreateAccount(context.Background(), &models.AccountData{}))
	})
}

func (s *accountApiClientSuite) TestFetchAccount() {
	s.Run("should successfully fetch single account", func() {
		s.Assert().NotNil(s.accountApiClient.FetchAccount(context.Background(), "account-id"))
	})
}

func (s *accountApiClientSuite) TestDeleteAccount() {
	s.Run("should successfully delete single account", func() {
		s.Assert().NotNil(s.accountApiClient.FetchAccount(context.Background(), "account-id"))
	})
}
