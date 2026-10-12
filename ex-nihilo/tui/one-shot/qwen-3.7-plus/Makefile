BINARY_NAME=dbnav
BIN_DIR=./bin

.PHONY: all build run fmt clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/main.go

run: build
	@$(BIN_DIR)/$(BINARY_NAME)

fmt:
	go fmt ./...

clean:
	rm -rf $(BIN_DIR)
