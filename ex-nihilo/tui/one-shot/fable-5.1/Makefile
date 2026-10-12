BIN_DIR := bin
BINARY  := $(BIN_DIR)/sqltui
PKG     := ./cmd

.PHONY: all fmt vet tidy build run clean

all: fmt vet build

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) $(PKG)

run: build
	./$(BINARY)

clean:
	rm -rf $(BIN_DIR)
	go clean
