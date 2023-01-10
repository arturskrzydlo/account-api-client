package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/arturskrzydlo/account-api-client/api/internal/models"
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
			httpClientOptions := make([]ClientOption, 0)
			if tc.httpClient != nil {
				httpClientOptions = append(httpClientOptions, WithCustomHttpClient(tc.httpClient))
			}
			// when
			accountClient, err := NewAccountsClient(tc.apiURL, httpClientOptions...)
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

func (s *accountAPIClientSuite) TestRetryPolicies() {
	s.Run("client should retry request to an api according to retry policy defined in client config", func() {
		// given
		numCalls := 0
		testServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			numCalls++
		}))

		maxRetries := 3
		accountsClient, err := NewAccountsClient(testServ.URL, WithRetriesOnDefaultRetryPolicy(maxRetries))
		s.Assert().NoError(err)

		// when
		_, err = accountsClient.FetchAccount(context.Background(), "account-id")

		// then
		var reqErr *RequestError
		s.Assert().True(errors.As(err, &reqErr))
		s.Assert().Equal(http.StatusInternalServerError, reqErr.statusCode)
		s.Assert().Equal(maxRetries+1, numCalls)
	})

	s.Run("client should retry request to an api according to retry policy and back to valid response after second retry", func() {
		// given
		maxRetries := 2
		numCalls := 0
		testServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// there should be three calls - initial one, one with first retry,
			if numCalls < maxRetries-1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response, _ := json.Marshal(models.AccountResponse{Data: &models.AccountDataResponse{
					Attributes:     nil,
					ID:             "some-id",
					OrganisationID: "org-id",
				}})
				_, err := w.Write(response)
				s.Require().NoError(err)
			}
			numCalls++
		}))
		accountsClient, err := NewAccountsClient(testServ.URL, WithRetriesOnDefaultRetryPolicy(maxRetries))
		s.Assert().NoError(err)

		// when
		account, err := accountsClient.FetchAccount(context.Background(), "account-id")

		// then
		s.Require().NoError(err)
		s.Assert().NotNil(account)
		s.Assert().Equal(maxRetries, numCalls)
	})

	s.Run("test default retry policy", func() {
		// given
		defaultRetryPolicy := DefaultRetryPolicy{
			MaxRetries: 2,
		}

		testCases := map[string]struct {
			err   error
			res   *http.Response
			retry bool
		}{
			"should not retry when error and response are empty": {
				err:   nil,
				res:   nil,
				retry: false,
			},
			"should not retry when response has status code less than 500": {
				err:   nil,
				res:   &http.Response{StatusCode: 499},
				retry: false,
			},
			"should retry when response has status code greater than or equal to 500": {
				err:   nil,
				res:   &http.Response{StatusCode: http.StatusInternalServerError},
				retry: true,
			},
			"should not retry when err on request is of different type than *url.Error": {
				err:   errors.New("some error"),
				res:   nil,
				retry: false,
			},
			"should retry when err on request is of type *url.Error": {
				err:   &url.Error{},
				res:   nil,
				retry: true,
			},
		}

		for name, tc := range testCases {
			s.Run(name, func() {
				// when
				shouldRetry := defaultRetryPolicy.ShouldRetry(tc.err, tc.res)

				// then
				s.Assert().Equal(tc.retry, shouldRetry)
			})
		}
	})
}

func (s *accountAPIClientSuite) TestBackoffStrategies() {
	// these test is a bit brittle - it can be changed to use clock library https://github.com/benbjohnson/clock and time-consuming
	// to not make test last to long and dependent on the machine performance
	// it would require to add clock var and use it across client
	s.Run("client should apply backoff strategy to retry", func() {
		// given
		numCalls := 0
		testServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			numCalls++
		}))

		delay := time.Millisecond * 1
		maxRetries := 3
		multiplier := 10
		accountsClient, err := NewAccountsClient(testServ.URL, WithRetriesOnDefaultRetryPolicy(maxRetries), WithExponentialBackoffStrategy(delay, multiplier))
		s.Assert().NoError(err)

		// when
		startTime := time.Now()
		_, err = accountsClient.FetchAccount(context.Background(), "account-id")
		endTime := time.Now()

		// then
		s.Assert().True(endTime.Sub(startTime) > delay+(delay*time.Duration(multiplier))+delay*time.Duration(multiplier)*time.Duration(multiplier))
	})

	s.Run("should increase exponentially delay between retries", func() {
		// given
		backoff := ExponentialBackoffStrategy{
			initialDelay: time.Millisecond * 10,
			multiplier:   10,
		}

		// when
		var delay time.Duration
		for i := 0; i < 3; i++ {
			delay = backoff.delay()
		}

		// then
		s.Assert().Equal(time.Second, delay)
	})

	s.Run("should return linear delay between retries", func() {
		// given
		backoff := LinearBackoffStrategy{
			delayTime: time.Millisecond * 100,
		}

		// when
		delay := backoff.delay()

		// then
		s.Assert().Equal(backoff.delayTime, delay)
	})
}
