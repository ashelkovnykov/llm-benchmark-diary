.PHONY: build run clean tidy fmt vet

BIN_DIR := bin
BINARY  := $(BIN_DIR)/tuidb

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd/

run: build
	./$(BINARY)

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)
	rm -rf db/
