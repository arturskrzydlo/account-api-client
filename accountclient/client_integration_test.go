//go:build integration

package accountclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/arturskrzydlo/account-api-client/accountclient/models"
)

type accountApiClientIntegrationSuite struct {
	suite.Suite

	accountApiClient *Client
}

type CustomRetryPolicy struct {
	maxRetries int
}

func (c CustomRetryPolicy) ShouldRetry(err error, response *http.Response) bool {
	if response != nil {
		return response.StatusCode >= http.StatusBadRequest
	}
	return false
}

func (c CustomRetryPolicy) NumberOfRetries() int {
	return c.maxRetries
}

func TestAccountApiClient(t *testing.T) {
	suite.Run(t, &accountApiClientIntegrationSuite{})
}

func (s *accountApiClientIntegrationSuite) SetupSuite() {
	s.accountApiClient = createAccountClient()
}

func getHostname() string {
	return os.Getenv("ACCOUNT_API_HOSTNAME")
}

func createAccountClient() *Client {
	accountApiClient, err := NewAccountClient(fmt.Sprintf("http://%s:8080/v1", getHostname()),
		WithRetriesOnDefaultRetryPolicy(3),
		WithLinearBackoffStrategy(time.Millisecond*100))
	if err != nil {
		log.Fatal("failed to create account api client")
	}
	return accountApiClient
}

// client to test e2e retries and set more restrictive retry policy
// where retries are triggered also on 4xx codes
func customRetryPolicyAccountClient() *Client {
	accountApiClient, err := NewAccountClient(fmt.Sprintf("http://%s:8080/v1", getHostname()),
		WithCustomRetryPolicy(CustomRetryPolicy{3}), WithCustomHTTPClient(&http.Client{Timeout: time.Second * 60}))
	if err != nil {
		log.Fatal("failed to create account api client")
	}
	return accountApiClient
}

func (s *accountApiClientIntegrationSuite) TestCreateAccount() {
	s.Run("should successfully create single account", func() {
		// given
		account := createAccountRequest()

		// when
		accountResp, err := s.accountApiClient.CreateAccount(context.Background(), account)

		// then
		s.Assert().NoError(err)
		s.assertCreatedAccount(account.Data, accountResp.Data)
		fetchedAccount, err := s.accountApiClient.FetchAccount(context.Background(), account.Data.ID)
		s.Assert().NoError(err)
		s.assertCreatedAccount(account.Data, fetchedAccount.Data)
	})

	s.Run("should not create account and return error with error code when account creation fails", func() {
		// given
		account := createAccountRequest()
		account.Data.Attributes.Country = nil

		// when
		accountResp, err := s.accountApiClient.CreateAccount(context.Background(), account)

		// then
		s.Assert().Nil(accountResp)
		var reqErr *RequestError
		s.Assert().ErrorAs(err, &reqErr)
		s.Assert().Equal(reqErr.StatusCode, 400)
		s.Assert().NotEmpty(reqErr.ErrMsg)
		_, err = s.accountApiClient.FetchAccount(context.Background(), account.Data.ID)
		s.Assert().ErrorAs(err, &reqErr)
		s.Assert().Equal(reqErr.StatusCode, 404)
	})

	s.Run("should return error without error code when there is issue with request", func() {
		// given
		account := createAccountRequest()
		s.accountApiClient.baseURL = "http://localhost:9999/fake/url/v1"
		// when
		accountResp, err := s.accountApiClient.CreateAccount(context.Background(), account)

		// then
		s.Assert().Nil(accountResp)
		var reqErr *RequestError
		s.Assert().False(errors.As(err, &reqErr))
		s.Assert().NotNil(err)
		s.accountApiClient = createAccountClient()
	})
}

func (s *accountApiClientIntegrationSuite) TestFetchAccount() {
	s.Run("should successfully fetch single account", func() {
		// given
		account := createAccountRequest()
		_, err := s.accountApiClient.CreateAccount(context.Background(), account)
		s.Require().NoError(err)

		// when
		fetchedAccount, err := s.accountApiClient.FetchAccount(context.Background(), account.Data.ID)

		// then
		s.Assert().NoError(err)
		s.assertCreatedAccount(account.Data, fetchedAccount.Data)
	})

	s.Run("should return error when there is no account for given accountID", func() {
		// given
		accountID := uuid.New()

		// when
		fetchedAccount, err := s.accountApiClient.FetchAccount(context.Background(), accountID)

		// then
		s.Assert().Nil(fetchedAccount)
		var reqErr *RequestError
		s.Assert().ErrorAs(err, &reqErr)
		s.Assert().Equal(reqErr.StatusCode, http.StatusNotFound)
	})

	s.Run("should return error without error code when there is any issue with request", func() {
		// given
		accountID := uuid.New()
		s.accountApiClient.baseURL = "http://localhost:9999/fake/url/v1"

		// when
		fetchedAccount, err := s.accountApiClient.FetchAccount(context.Background(), accountID)

		// then
		s.Assert().Nil(fetchedAccount)
		var reqErr *RequestError
		s.Assert().False(errors.As(err, &reqErr))
		s.Assert().NotNil(err)
		s.accountApiClient = createAccountClient()
	})
}

func (s *accountApiClientIntegrationSuite) TestDeleteAccount() {
	s.Run("should successfully delete single account", func() {
		// given
		account := createAccountRequest()
		accountRes, err := s.accountApiClient.CreateAccount(context.Background(), account)
		s.Require().NoError(err)

		// when
		err = s.accountApiClient.DeleteAccount(context.Background(), account.Data.ID, accountRes.Data.Version)

		// then
		s.Require().NoError(err)
		fetchedAccount, err := s.accountApiClient.FetchAccount(context.Background(), account.Data.ID)
		var reqErr *RequestError
		s.Assert().ErrorAs(err, &reqErr)
		s.Assert().Equal(reqErr.StatusCode, http.StatusNotFound)
		s.Assert().Nil(fetchedAccount)
	})

	s.Run("should return error with status code when it was not possible to delete account", func() {
		// given
		accountID := uuid.New()
		accountVersion := int64(0)

		// when
		err := s.accountApiClient.DeleteAccount(context.Background(), accountID, &accountVersion)

		// then
		var reqErr *RequestError
		s.Assert().ErrorAs(err, &reqErr)
		s.Assert().Equal(reqErr.StatusCode, http.StatusNotFound)
		s.Assert().Empty(reqErr.ErrMsg)
	})

	s.Run("should return error without error code when there is any issue with request", func() {
		// given
		accountID := uuid.New()
		accountVersion := int64(0)
		s.accountApiClient.baseURL = "http://localhost:9999/fake/url/v1"

		// when
		err := s.accountApiClient.DeleteAccount(context.Background(), accountID, &accountVersion)

		// then
		var reqErr *RequestError
		s.Assert().False(errors.As(err, &reqErr))
		s.Assert().NotNil(err)
		s.accountApiClient = createAccountClient()
	})
}

func (s *accountApiClientIntegrationSuite) TestRetriesAreApplied() {
	// these tests actually check if after retries we are receiving request errors
	// it has been created after issue where on retries nil request were sent
	// thus url.Error was returned from function
	s.Run("should retry failed requests and return request error", func() {
		// given
		account := createAccountRequest()
		s.accountApiClient = customRetryPolicyAccountClient()
		_, err := s.accountApiClient.CreateAccount(context.Background(), account)
		s.Require().NoError(err)

		// when
		// it should be retried and finished with error because we are attempting to
		// create account with the same id
		_, err = s.accountApiClient.CreateAccount(context.Background(), account)

		// then
		s.Require().Error(err)
		var reqErr *RequestError
		s.Assert().ErrorAs(err, &reqErr)
		s.Assert().Equal(reqErr.StatusCode, http.StatusConflict)
	})

	s.Run("should retry failed requests for fetching when we should have response body", func() {
		// given
		s.accountApiClient = customRetryPolicyAccountClient()

		// when
		// it should be retried and finished with error because we are attempting to
		// create account with the same id
		accountResp, err := s.accountApiClient.FetchAccount(context.Background(), uuid.New())

		// then
		s.Require().Error(err)
		s.Assert().Nil(accountResp)
		var reqErr *RequestError
		s.Assert().ErrorAs(err, &reqErr)
		s.Assert().Equal(reqErr.StatusCode, http.StatusNotFound)
	})
}

func (s *accountApiClientIntegrationSuite) TestSampleAccountFlow() {
	s.Run("should create account, then fetch it and at the it should successfully delete it", func() {
		// given
		account := createAccountRequest()

		// when creating account
		accountResp, err := s.accountApiClient.CreateAccount(context.Background(), account)

		// then
		s.Assert().NoError(err)
		s.assertCreatedAccount(account.Data, accountResp.Data)

		// when fetching created account
		fetchedAccount, err := s.accountApiClient.FetchAccount(context.Background(), account.Data.ID)

		// then
		s.Assert().NoError(err)
		s.assertCreatedAccount(account.Data, fetchedAccount.Data)

		// when deleting account
		err = s.accountApiClient.DeleteAccount(context.Background(), fetchedAccount.Data.ID, fetchedAccount.Data.Version)

		// then
		s.Require().NoError(err)
		fetchedAccount, err = s.accountApiClient.FetchAccount(context.Background(), fetchedAccount.Data.ID)
		var reqErr *RequestError
		s.Assert().ErrorAs(err, &reqErr)
		s.Assert().Equal(reqErr.StatusCode, http.StatusNotFound)
		s.Assert().Nil(fetchedAccount)
	})
}

func createAccountRequest() *models.CreateAccountRequest {
	accountID := uuid.New()
	organizationID := uuid.New()
	version := new(int64)
	*version = 0
	accountClassification := "Personal"
	accountMatchingOptOut := false
	country := "GB"
	jointAccount := false

	return &models.CreateAccountRequest{Data: &models.CreateAccountData{
		Attributes: &models.CreateAccountAttributes{
			AccountClassification:   &accountClassification, // enum ?
			AccountMatchingOptOut:   &accountMatchingOptOut, // deprecated
			AccountNumber:           "41426819",
			AlternativeNames:        []string{"Sam Holder"},
			BankID:                  "400300",
			BankIDCode:              "GBDSC",
			BaseCurrency:            "GBP",
			Bic:                     "NWBKGB22",
			Country:                 &country,
			Iban:                    "GB11NWBK40030041426819", // generated if not provided
			JointAccount:            &jointAccount,
			Name:                    []string{"Samantha Holder"},
			SecondaryIdentification: "A1B2C3D4",
			Status:                  nil, // Status of the account. pending and confirmed are set by Form3, closed can be set manually. Test creating closed account
			Switched:                nil, // deprecated, account switched away from organization
		},
		ID:             accountID,
		OrganisationID: organizationID,
		Type:           "accounts",
		Version:        version, // incremented witch each update, probably not needed in create
	}}
}

func (s *accountApiClientIntegrationSuite) assertCreatedAccount(expectedAccount *models.CreateAccountData, actualAccount *models.AccountDataResponse) {
	s.Assert().Equal(expectedAccount.ID, actualAccount.ID)
	s.Assert().Equal(expectedAccount.Type, actualAccount.Type)
	s.Assert().Equal(expectedAccount.Version, actualAccount.Version)
	s.Assert().Equal(expectedAccount.OrganisationID, actualAccount.OrganisationID)
	// asserting account attributes
	s.Assert().Equal(expectedAccount.Attributes.AccountClassification, actualAccount.Attributes.AccountClassification)
	s.Assert().Equal(expectedAccount.Attributes.AccountMatchingOptOut, actualAccount.Attributes.AccountMatchingOptOut)
	s.Assert().Equal(expectedAccount.Attributes.AccountNumber, actualAccount.Attributes.AccountNumber)
	s.Assert().Equal(expectedAccount.Attributes.AlternativeNames[0], actualAccount.Attributes.AlternativeNames[0])
	s.Assert().Equal(expectedAccount.Attributes.BankID, actualAccount.Attributes.BankID)
	s.Assert().Equal(expectedAccount.Attributes.BankIDCode, actualAccount.Attributes.BankIDCode)
	s.Assert().Equal(expectedAccount.Attributes.BaseCurrency, actualAccount.Attributes.BaseCurrency)
	s.Assert().Equal(expectedAccount.Attributes.Bic, actualAccount.Attributes.Bic)
	s.Assert().Equal(expectedAccount.Attributes.Country, actualAccount.Attributes.Country)

	if expectedAccount.Attributes.Iban != "" {
		s.Assert().Equal(expectedAccount.Attributes.Iban, actualAccount.Attributes.Iban)
	}

	s.Assert().Equal(expectedAccount.Attributes.JointAccount, actualAccount.Attributes.JointAccount)
	s.Assert().Equal(expectedAccount.Attributes.Name[0], actualAccount.Attributes.Name[0])
	s.Assert().Equal(expectedAccount.Attributes.SecondaryIdentification, actualAccount.Attributes.SecondaryIdentification)
	s.Assert().Equal(expectedAccount.Attributes.Status, actualAccount.Attributes.Status)
	s.Assert().Equal(expectedAccount.Attributes.Switched, actualAccount.Attributes.Switched)

	s.Assert().False(actualAccount.CreatedOn.IsZero())
	s.Assert().False(actualAccount.ModifiedOn.IsZero())
}
