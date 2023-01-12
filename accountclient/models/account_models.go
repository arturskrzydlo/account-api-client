package models

import "time"

// Account represents an account in the form3 org section.
// See https://api-docs.form3.tech/api.html#organisation-accounts for
// more information about fields.

// Account models have been separated between create request and response
// to make it more extendable for fields which could exist in response
// but not in request (like created_on, modified_on dates and others)

type CreateAccountRequest struct {
	Data *CreateAccountData `json:"data,omitempty"`
}

type CreateAccountData struct {
	Attributes     *CreateAccountAttributes `json:"attributes,omitempty"`
	ID             string                   `json:"id,omitempty"`
	OrganisationID string                   `json:"organisation_id,omitempty"`
	Type           string                   `json:"type,omitempty"`
	Version        *int64                   `json:"version,omitempty"`
}

type CreateAccountAttributes struct {
	AccountClassification   *string  `json:"account_classification,omitempty"`
	AccountMatchingOptOut   *bool    `json:"account_matching_opt_out,omitempty"`
	AccountNumber           string   `json:"account_number,omitempty"`
	AlternativeNames        []string `json:"alternative_names,omitempty"`
	BankID                  string   `json:"bank_id,omitempty"`
	BankIDCode              string   `json:"bank_id_code,omitempty"`
	BaseCurrency            string   `json:"base_currency,omitempty"`
	Bic                     string   `json:"bic,omitempty"`
	Country                 *string  `json:"country,omitempty"`
	Iban                    string   `json:"iban,omitempty"`
	JointAccount            *bool    `json:"joint_account,omitempty"`
	Name                    []string `json:"name,omitempty"`
	SecondaryIdentification string   `json:"secondary_identification,omitempty"`
	Status                  *string  `json:"status,omitempty"`
	Switched                *bool    `json:"switched,omitempty"`
}

type AccountResponse struct {
	Data *AccountDataResponse `json:"data,omitempty"`
}

type AccountDataResponse struct {
	Attributes     *AccountAttributesResponse `json:"attributes,omitempty"`
	ID             string                     `json:"id,omitempty"`
	OrganisationID string                     `json:"organisation_id,omitempty"`
	Type           string                     `json:"type,omitempty"`
	Version        *int64                     `json:"version,omitempty"`
	CreatedOn      time.Time                  `json:"created_on"`
	ModifiedOn     time.Time                  `json:"modified_on"`
}

type AccountAttributesResponse struct {
	AccountClassification   *string  `json:"account_classification,omitempty"`
	AccountMatchingOptOut   *bool    `json:"account_matching_opt_out,omitempty"`
	AccountNumber           string   `json:"account_number,omitempty"`
	AlternativeNames        []string `json:"alternative_names,omitempty"`
	BankID                  string   `json:"bank_id,omitempty"`
	BankIDCode              string   `json:"bank_id_code,omitempty"`
	BaseCurrency            string   `json:"base_currency,omitempty"`
	Bic                     string   `json:"bic,omitempty"`
	Country                 *string  `json:"country,omitempty"`
	Iban                    string   `json:"iban,omitempty"`
	JointAccount            *bool    `json:"joint_account,omitempty"`
	Name                    []string `json:"name,omitempty"`
	SecondaryIdentification string   `json:"secondary_identification,omitempty"`
	Status                  *string  `json:"status,omitempty"`
	Switched                *bool    `json:"switched,omitempty"`
}
