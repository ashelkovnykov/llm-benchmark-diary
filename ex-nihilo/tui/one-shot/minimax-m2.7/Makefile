.PHONY: build run clean fmt lint test

BINARY=./bin/dbviewer
MAIN=./cmd/main.go

build:
	go build -o $(BINARY) $(MAIN)

run: build
	$(BINARY)

clean:
	rm -rf ./bin
	rm -rf ./db

fmt:
	go fmt ./...

lint:
	golangci-lint run

test:
	go test ./...
