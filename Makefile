GOLANGCI_LINT_VERSION := v2.10.1

.PHONY: all test vet build lint cover clean setup

all: vet lint test build

## Install development tools
setup:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	}

## Run all tests
test:
	go test ./... -count=1

## Run tests with verbose output
test-v:
	go test ./... -v -count=1

## Run go vet
vet:
	go vet ./...

## Run golangci-lint
lint: setup
	golangci-lint run ./...

## Build all packages
build:
	go build ./...

## Run tests with coverage report
cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Remove build artifacts
clean:
	rm -f coverage.out coverage.html
