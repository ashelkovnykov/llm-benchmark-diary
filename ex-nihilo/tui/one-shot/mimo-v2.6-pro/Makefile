BINARY  := sqlnav
BIN_DIR := bin

.PHONY: all deps fmt vet build run test clean

all: build

deps:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

build: deps
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) ./cmd

run: build
	./$(BIN_DIR)/$(BINARY)

test: deps
	go test ./...

clean:
	go clean
	rm -rf $(BIN_DIR)
