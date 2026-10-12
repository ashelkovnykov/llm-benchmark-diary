GO := go
BINARY_NAME ?= sql-tui
BIN_DIR := ./bin

.PHONY: all tidy fmt vet test build run clean

all: tidy fmt vet build

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

vet: tidy
	$(GO) vet ./...

test: tidy
	$(GO) test ./...

build: tidy
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd

run: tidy
	CGO_ENABLED=0 $(GO) run ./cmd

clean:
	rm -rf $(BIN_DIR)
