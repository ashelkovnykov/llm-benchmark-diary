BINARY := bin/dbnav

.PHONY: all deps fmt vet test build run clean

all: build

deps:
	go mod tidy

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

test:
	go test ./...

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd

run: build
	./$(BINARY)

clean:
	rm -rf bin
	go clean
