APP_NAME = sql-tui
BIN_DIR = bin
DB_DIR = db
SRC = ./cmd/main.go

.PHONY: fmt vet build run clean tidy

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

build: tidy fmt vet
	go build -o $(BIN_DIR)/$(APP_NAME) $(SRC)

run: build
	./$(BIN_DIR)/$(APP_NAME)

clean:
	rm -rf $(BIN_DIR)
	rm -rf $(DB_DIR)/*.db
