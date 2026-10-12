BIN_DIR := ./bin
BINARY  := $(BIN_DIR)/sqlnav

.PHONY: all fmt vet tidy build run clean test

all: fmt build

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd

run:
	go run ./cmd

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
