BIN_DIR := bin
BIN     := $(BIN_DIR)/dbtui

.PHONY: all build run fmt vet test tidy clean

all: fmt vet build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd

run: build
	./$(BIN)

fmt:
	gofmt -s -w .

vet:
	go vet ./...

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)
