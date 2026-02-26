BINARY   := gateway
BUILD_DIR := ./bin
CMD_PATH  := ./cmd/gateway

.PHONY: all run build tidy lint test clean

all: build

## run: Build and run the gateway in development mode.
run:
	go run $(CMD_PATH)

## build: Compile the binary to ./bin/gateway.
build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)

## tidy: Tidy and verify go modules.
tidy:
	go mod tidy
	go mod verify

## lint: Run golangci-lint (must be installed: https://golangci-lint.run/usage/install/).
lint:
	golangci-lint run ./...

## test: Run all tests with race detector and coverage.
test:
	go test -race -cover ./...

## vet: Run go vet static analysis.
vet:
	go vet ./...

## clean: Remove compiled binaries.
clean:
	rm -rf $(BUILD_DIR)

## help: Print this help message.
help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //'
