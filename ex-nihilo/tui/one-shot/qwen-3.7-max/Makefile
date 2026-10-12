BINARY_NAME=db-explorer
BIN_DIR=./bin
CMD_DIR=./cmd

.PHONY: all build run clean fmt tidy

all: tidy fmt build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

run: build
	$(BIN_DIR)/$(BINARY_NAME)

clean:
	@rm -rf $(BIN_DIR)
	@rm -rf ./db

fmt:
	go fmt ./...

tidy:
	go mod tidy
