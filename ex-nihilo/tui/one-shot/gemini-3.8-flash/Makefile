APP_NAME := sqlnav
BIN_DIR := ./bin
CMD_DIR := ./cmd

.PHONY: all build run fmt clean tidy

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)/main.go

run: build
	@$(BIN_DIR)/$(APP_NAME)

fmt:
	go fmt ./...

tidy:
	go mod tidy

clean:
	@rm -rf $(BIN_DIR)
