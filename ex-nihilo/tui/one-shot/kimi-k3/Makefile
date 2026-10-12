# Makefile

BINARY_NAME=sqltui
BIN_DIR=./bin
BINARY=$(BIN_DIR)/$(BINARY_NAME)
CMD_DIR=./cmd

.PHONY: all build run clean fmt vet tidy test

all: fmt vet build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) $(CMD_DIR)

run: build
	$(BINARY)

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
