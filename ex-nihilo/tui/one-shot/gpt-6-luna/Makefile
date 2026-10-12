.PHONY: fmt build run clean

BINARY := bin/sqlnav

fmt:
	gofmt -w ./cmd ./internal

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd

run: build
	./$(BINARY)

clean:
	rm -rf bin
