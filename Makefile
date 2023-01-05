# Install all development tools and build artifacts to the project's
# `bin` directory.
export GOBIN=$(CURDIR)/bin

# Ensure that all dependencies are installed using vendored sources (/vendor).
export GOFLAGS=-mod=vendor

# Default to the system 'go'.
GO?=$(shell which go)

$(GOBIN):
	mkdir -p $(GOBIN)

.PHONY: setup
setup: ## Setting up local env
	cd $(GOBIN)
	wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.50.1


.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf $(GOBIN)

.PHONY: lint
lint: ## Lint the source code.
	$(GOBIN)/golangci-lint run --config $(shell pwd)/build/.golangci.yml --verbose ./...

.PHONY: prepare-integration-tests
prepare-integration-tests: ## Run fake account api server
	docker-compose -f docker-compose-fake-api.yml --project-name fake-account-api up --detach --remove-orphans

.PHONY: all-tests
all-tests: ## Run all test (unit + integration)
	$(GO) test -v -tags=integration ./...

.PHONY: cover
cover: ## Calculate coverage
	@go test -mod=vendor -coverprofile=coverage.out -tags=integration ./... ; \
	cat coverage.out | \
	awk 'BEGIN {cov=0; stat=0;} $$3!="" { cov+=($$3==1?$$2:0); stat+=$$2; } \
	END {printf("Total coverage: %.2f%% of statements\n", (cov/stat)*100);}'
	@go tool cover -html=coverage.out

.PHONY: tidy
tidy: ## Tidy go modules and re-vendor
	@go mod tidy
	@go mod vendor
