MODULE  := sqlnav
BIN_DIR := bin
BIN     := $(BIN_DIR)/$(MODULE)

GO      ?= go
GOFLAGS ?=

.PHONY: all build run dev fmt vet tidy test clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN) ./cmd

run: build
	./$(BIN)

dev:
	$(GO) run ./cmd

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

clean:
	rm -rf $(BIN_DIR)
