APP := dbtui
BIN := ./bin/$(APP)

.PHONY: all fmt vet tidy build run clean

all: build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

tidy:
	go mod tidy

build: tidy fmt
	mkdir -p ./bin ./db
	go build -o $(BIN) ./cmd

run: build
	$(BIN)

clean:
	rm -rf ./bin
