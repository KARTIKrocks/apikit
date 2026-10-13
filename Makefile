include common.mk

# Go modules in this repo; the nested ones are not covered by ./... at the root.
MODULES := . ./otel ./examples/otel

.PHONY: fmt all test vet build lint vuln cover clean setup ci

all: fmt vet lint test build

## Format code
fmt:
	gofmt -w .
	goimports -w .

## Run all tests with race detector
test:
	go test -race -count=1 ./...

## Run tests with verbose output
test-v:
	go test -race -v -count=1 ./...

## Run go vet
vet:
	go vet ./...

## Run golangci-lint
lint: setup
	golangci-lint run ./...

## Scan every module (root, otel, examples/otel) for known vulnerabilities,
## filtered to advisories this code actually reaches.
##
## Needs network access — the advisory database is fetched on every run.
vuln: setup
	@for mod in $(MODULES); do \
		echo "==> Scanning $$mod"; \
		(cd $$mod && govulncheck ./...) || exit 1; \
	done

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

## CI pipeline: vet, lint, test, vulnerability-scan
ci: vet lint test vuln
