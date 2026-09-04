.PHONY: build test clean install lint bench release

BINARY_NAME=logparse
BUILD_DIR=bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: build

## build: Build the binary
build:
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/logparse

## test: Run all tests
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

## test-short: Run tests without race detector (faster)
test-short:
	go test -v ./...

## bench: Run benchmarks
bench:
	go test -bench=. -benchmem ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out dist/

## install: Install binary to GOPATH/bin
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format Go code
fmt:
	go fmt ./...

## vet: Run go vet
vet:
	go vet ./...

## tidy: Tidy go modules
tidy:
	go mod tidy

## release: Build release snapshots (requires goreleaser)
release:
	goreleaser release --snapshot --clean

## docker: Build Docker image
docker:
	docker build -t $(BINARY_NAME):$(VERSION) .

## demo: Run with sample data
demo: build
	$(BUILD_DIR)/$(BINARY_NAME) testdata/sample.json.log

demo-error: build
	$(BUILD_DIR)/$(BINARY_NAME) testdata/sample.json.log --level ERROR

demo-json: build
	$(BUILD_DIR)/$(BINARY_NAME) testdata/sample.json.log --output json

demo-csv: build
	$(BUILD_DIR)/$(BINARY_NAME) testdata/sample.json.log --output csv

demo-nginx: build
	$(BUILD_DIR)/$(BINARY_NAME) testdata/sample.nginx.log

demo-syslog: build
	$(BUILD_DIR)/$(BINARY_NAME) testdata/sample.syslog

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //g'

.DEFAULT_GOAL := help
