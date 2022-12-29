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
		// given
		account := createAccount("test-account-id", "test-organization-id")

		// when
		err := s.accountApiClient.CreateAccount(context.Background(), account)

		// then
		s.Assert().NoError(err)
		fetchedAccount, err := s.accountApiClient.FetchAccount(context.Background(), account.ID)
		s.Assert().NoError(err)
		s.assertCreatedAccount(account, fetchedAccount)
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

func createAccount(accountID string, organizationID string) *models.AccountData {
	version := new(int64)
	*version = 0
	accountClassification := "Personal"
	accountMatchingOptOut := false
	country := "GB"
	jointAccount := false

	return &models.AccountData{
		Attributes: &models.AccountAttributes{
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
	}
}

func (s *accountApiClientSuite) assertCreatedAccount(expectedAccount *models.AccountData, actualAccount *models.AccountData) {
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
}
