BINARY_NAME=sql-navigator
BIN_DIR=./bin
CMD_DIR=./cmd
DB_DIR=./db

.PHONY: all fmt build run clean

all: fmt build

fmt:
	go fmt ./...

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

run: build
	@mkdir -p $(DB_DIR)
	$(BIN_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BIN_DIR)
