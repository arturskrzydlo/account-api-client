package api

import (
	"net/http"
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
	validAPIURL := "http://some-api.com"
	testCases := map[string]struct {
		apiURL      string
		httpClient  *http.Client
		expectedErr bool
	}{
		"should create successfully account client": {
			apiURL:      validAPIURL,
			httpClient:  nil,
			expectedErr: false,
		},
		"should not create client and return error when baseURL param is invalid URL": {
			apiURL:      "invalidURL",
			httpClient:  nil,
			expectedErr: true,
		},
		"should create account client with custom http client": {
			apiURL: validAPIURL,
			httpClient: &http.Client{
				Timeout: 999,
			},
			expectedErr: false,
		},
		"should create account client with default http client": {
			apiURL:      validAPIURL,
			httpClient:  nil,
			expectedErr: false,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			// when
			accountClient, err := NewAccountsClient(tc.apiURL, tc.httpClient)
			// then
			if !tc.expectedErr {
				s.NoError(err)
				s.Assert().Equal(tc.apiURL, accountClient.baseURL)
				if tc.httpClient != nil {
					s.Assert().Equal(tc.httpClient, accountClient.httpClient)
				} else {
					s.Assert().Equal(&http.Client{Timeout: defaultTimeout}, accountClient.httpClient)
				}
			}
		})
	}
}
