package api

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type accountAPIClientSuite struct {
	suite.Suite
}

func TestAccountApiClientUnit(t *testing.T) {
	suite.Run(t, &accountAPIClientSuite{})
}

func (s *accountAPIClientSuite) TestAccountClientCreation() {
	s.Run("should create successfully account client", func() {
		// given
		validAPIURL := "http://some-api.com"
		// when
		client, err := NewAccountsClient(validAPIURL)
		// then
		s.NoError(err)
		s.Assert().Equal(validAPIURL, client.baseURL)
	})
	s.Run("should not create client and return error when baseURL param is invalid URL", func() {
		// given
		validAPIURL := "invalidURL"
		// when
		_, err := NewAccountsClient(validAPIURL)
		// then
		s.NotNil(err)
	})
}
