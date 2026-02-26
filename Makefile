BINARY   := gateway
BUILD_DIR := ./bin
CMD_PATH  := ./cmd/gateway

.PHONY: all run build tidy lint test clean

all: build

run:
	go run $(CMD_PATH)

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)

tidy:
	go mod tidy
	go mod verify

lint:
	golangci-lint run ./...

test:
	go test -race -cover ./...

vet:
	go vet ./...

build-image:
	podman build -t fpt-ojt/gateway:latest -f deployments/Dockerfile .

up:
	podman compose -f deployments/docker-compose.yml up -d --build
down:
	podman compose -f deployments/docker-compose.yml down
container-logs:
	podman compose -f deployments/docker-compose.yml logs -f gateway

clean:
	rm -rf $(BUILD_DIR)

help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //'
