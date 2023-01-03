package api

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type accountApiClientSuite struct {
	suite.Suite

	accountApiClient *Client
}

func TestAccountApiClientUnit(t *testing.T) {
	suite.Run(t, &accountApiClientSuite{})
}

func (s *accountApiClientSuite) TestAccountClientCreation() {
	s.Run("should create successfully account client", func() {
		// given
		validApiURL := "http://some-api.com"
		// when
		client, err := NewAccountsClient(validApiURL)
		// then
		s.NoError(err)
		s.Assert().Equal(validApiURL, client.baseURL)
	})
	s.Run("should not create client and return error when baseURL param is invalid URL", func() {
		// given
		validApiURL := "invalidURL"
		// when
		_, err := NewAccountsClient(validApiURL)
		// then
		s.NotNil(err)
	})
}
