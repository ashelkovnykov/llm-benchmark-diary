BINARY_NAME := bin/sqlnav
MAIN := ./cmd/main.go

.PHONY: all fmt build run clean

all: fmt build

fmt:
	go fmt ./...

build:
	@mkdir -p bin
	go build -o $(BINARY_NAME) $(MAIN)

run: build
	./$(BINARY_NAME)

clean:
	rm -rf bin/
