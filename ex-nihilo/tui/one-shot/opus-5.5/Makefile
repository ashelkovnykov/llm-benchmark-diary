BINARY   := sqlnav
BIN_DIR  := ./bin
CMD_DIR  := ./cmd
DB_DIR   := ./db

.PHONY: all tidy fmt vet test build run clean clean-db

all: fmt vet build

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet: tidy
	go vet ./...

test: tidy
	go test ./...

build: tidy
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(CMD_DIR)

run: build
	$(BIN_DIR)/$(BINARY)

clean:
	rm -rf $(BIN_DIR)
	go clean

clean-db:
	rm -f $(DB_DIR)/*.db
