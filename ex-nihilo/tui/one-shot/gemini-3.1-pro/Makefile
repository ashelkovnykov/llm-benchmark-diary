APP_NAME := db-tui
BIN_DIR := ./bin
MAIN_PATH := ./cmd/main.go

.PHONY: all build run fmt clean

all: fmt build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(MAIN_PATH)

run: build
	@mkdir -p ./db
	$(BIN_DIR)/$(APP_NAME)

fmt:
	go fmt ./...

clean:
	rm -rf $(BIN_DIR)
