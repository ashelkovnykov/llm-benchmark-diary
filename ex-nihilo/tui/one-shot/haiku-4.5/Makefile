.PHONY: fmt build run clean

fmt:
	go fmt ./...

build:
	mkdir -p bin
	go build -o bin/dbui ./cmd

run: build
	./bin/dbui

clean:
	rm -rf bin
