.DEFAULT_GOAL := build

BINARY := bin/dbnav

.PHONY: deps fmt vet test build run clean

deps:
	go mod tidy

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

test:
	go test ./...

build: deps
	mkdir -p bin
	go build -o $(BINARY) ./cmd

run: build
	./$(BINARY)

clean:
	rm -rf bin
	go clean
