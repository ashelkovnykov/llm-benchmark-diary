.PHONY: all build run fmt test tidy clean

all: build

build:
	mkdir -p bin
	go build -o bin/dbtui ./cmd

run: build
	./bin/dbtui

fmt:
	go fmt ./...

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf bin
