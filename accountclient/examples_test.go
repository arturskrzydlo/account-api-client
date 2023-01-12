package accountclient

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/arturskrzydlo/account-api-client/accountclient/models"
)

func Example() {
	// client initialization
	client, err := NewAccountClient("localhost:8080",
		WithRetriesOnDefaultRetryPolicy(3),
		WithLinearBackoffStrategy(time.Millisecond*100),
		WithCustomHTTPClient(&http.Client{Timeout: time.Second * 20}))
	if err != nil {
		log.Fatal(err)
	}

	// create account
	ukCountry := "UK"
	createAccountReq := &models.CreateAccountRequest{Data: &models.CreateAccountData{
		Attributes:     &models.CreateAccountAttributes{Name: []string{"some name"}, Country: &ukCountry},
		ID:             uuid.New(),
		OrganisationID: uuid.New(),
		Type:           "accounts",
		Version:        nil,
	}}

	accountResponse, err := client.CreateAccount(context.Background(), createAccountReq)
	if err != nil {
		log.Printf("failed to create a new account: %s", err.Error())
	}

	// fetch account
	accountResponse, err = client.FetchAccount(context.Background(), accountResponse.Data.ID)
	if err != nil {
		log.Printf("failed to fetch a new account: %s", err.Error())
	}

	// delete account
	err = client.DeleteAccount(context.Background(), accountResponse.Data.ID, accountResponse.Data.Version)
	if err != nil {
		log.Printf("failed to delete an account: %s", err.Error())
	}
}

func ExampleNewAccountClient() {
	NewAccountClient("localhost:8080")
}

func ExampleNewAccountClient_withOptions() {
	NewAccountClient("localhost:8080",
		WithRetriesOnDefaultRetryPolicy(3),
		WithLinearBackoffStrategy(time.Millisecond*100),
		WithCustomHTTPClient(&http.Client{Timeout: time.Second * 20}))
}
