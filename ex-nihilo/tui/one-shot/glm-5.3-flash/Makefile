BIN_DIR  := bin
BINARY   := $(BIN_DIR)/sqlnav

.PHONY: all fmt vet build run clean

all: build

fmt:
	go fmt ./...

vet:
	go vet ./...

build: vet
	mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd

run: build
	./$(BINARY)

clean:
	rm -rf $(BIN_DIR)
