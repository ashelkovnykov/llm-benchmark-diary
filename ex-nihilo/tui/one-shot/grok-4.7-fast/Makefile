.PHONY: fmt build run clean tidy

BIN_DIR := bin
BINARY := $(BIN_DIR)/dbtui

fmt:
	go fmt ./...

tidy:
	go mod tidy

build: tidy
	mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd

run: build
	./$(BINARY)

clean:
	rm -rf $(BIN_DIR)
	rm -f $(BINARY)

.PHONY: all
all: fmt build
