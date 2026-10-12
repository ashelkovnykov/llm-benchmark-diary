BINARY := bin/sqlnav
SRC    := ./cmd ./internal

.PHONY: all build run fmt fmt-check vet test deps clean help

all: build

## build: compile the project into ./bin/
build:
	@mkdir -p bin
	go build -o $(BINARY) ./cmd

## run: build and run the binary
run: build
	./$(BINARY)

## fmt: format all source files
fmt:
	gofmt -w $(SRC)

## fmt-check: fail if any source file is unformatted
fmt-check:
	@test -z "$$(gofmt -l $(SRC))" || (gofmt -l $(SRC) && exit 1)

## vet: run go vet
vet:
	go vet ./...

## test: run the test suite
test:
	go test ./...

## deps: sync module dependencies
deps:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -rf bin
	go clean

help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
