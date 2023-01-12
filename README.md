# account-api-client

This is a [take home exercise](https://github.com/form3tech-oss/interview-accountapi/blob/master/README.md)
implementation for Form3 interview written by Artur Skrzydlo

This library consumes fake (and limited) account api delevered with `docker-compose.yml`. Documentation for real api can
be found [here](https://www.api-docs.form3.tech/api/tutorials/getting-started/create-an-account)

## Getting started

### Testing

Verification if everything works can be done by running tests from repo main directory:

```shell
docker compose up --build --abort-on-container-exit && docker compose logs accountapitests -t
```

Latter part of the command it just for convenience - logs can be mixed so it's good to see logs only for tests which
have been run

### Installation

In your project using go modules just run

```shell
go get github.com/arturskrzydlo/account-api-client
```

### Usage

Example of how to use library can be found in [documentation](#documentation).
This is basic usage example taken from documentation :

```go
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
```

### Documentation

To generate documentation install `godoc` tool:

```shell
go install golang.org/x/tools/cmd/godoc@latest
```

and then run:

```shell
godoc -http :8080 
```

Browsing localhost:8080 should direct you to generated documentation for this library

### Development

For developing purposes only I've added few Makefile commands which I was using actively during development:

* **setup** - only to download linter binary useful for linting during development
* **clean** - delete binary downloaded in setup
* **lint** - running set of linters to check code style etc
* **prepare-integration-test** running delivered `docker-compose.yml` with fake api to be able to run integration tests.
* **all-tests** runs all tests unit and integration ones - **prepare-integration-test** is needed to be run before. This
  command is also used in docker file to run tests
* **tests** running all tests without integration tests. Doesn't need preparation step
* **cover** runs code coverage to have an overview of which code parts have been tested
* **tidy** runs go tidy and vendor commands

## Remarks & possible improvements

First and most important remark. I've realized at the end of development that I have used one library which was not
standard `http/net` library.
I'm talking about `github.com/afex/hystrix-go`. It would take me quite a lot of time to remove it and write simple
equivalent, so I've decided to leave it in this repo.
If it would be a reason of rejection please just give me short note, and I'll try to find more time to do so

### Remarks:

* **Testing** - I've created a different type of the tests. First of all I've made integration tests which hits running
  server with api. This checks major behaviours of library.
  However, there were some issues with testing some non-functional requirements. Testing if retries have been applied
  correctly or backoff strategy works I've used tests which use `httptest` package
  and mock server. At very last end I've written small amount of unit test where I would be struggling testing it higher
  level tests.

  Also having docker-compose file affected my decisions here. If I were testing pulic api via some provided sandbox then
  I would be limiting such tests to bare minimum and focus mostly on `httptest` tests
* **Error handling** - I've created some specific error type which would be helpful for passing error down the stream
  and make decision upon this. This error struct, called `RequestError` contains status code and error message
  Because library doesn't return only this type of error it might be a bit confusing (bolier plate code for checking
  error type) but still I think it might be better than parsing error string to get these values

### Possible improvements:

* **Versioning support** - With current approach adding new version handles won't be super smooth. I can add new client
  methods, but it could be done on client api level
* **Logging** - Logging could be added and be configurable. We could use some loggers like `uber-go/zap` and make the
  logging conditional. Same apply for tracing
* **Hystrix** - Could be more configurable and fine-grained, currently in library there is one hystrix command for all
  methods, but we might want to treat them individually, even with individual circuit breaker rules
* **Contract testing** - It depends on who would be the ownership of the service with account api, but assuming that it
  will be all in Form3 company contract testing would be crucial to verify changes in api
* **Thread safe** - Current implementation is probably not a thread safe. I've not verified it though. It has minimal
  state in it but such verification should be made and adjustments made to be able to run multiple requests over this
  library
* **Rest client separation** - generic rest client could be separated from the code for better re-usage. It could be
  even extracted to it's very own rest client package
* **Retrier** - there is a field to merge backoff and retryPolicy into one object. Currently, Retrier is a separate type
  but it's nothing more than wrapper, but it could be used better (i.e. it should manage maximum number of retries)


