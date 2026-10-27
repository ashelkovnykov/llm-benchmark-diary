BINARY := bin/dbtui

.PHONY: fmt tidy build run test clean

fmt:
	go fmt ./...

tidy:
	go mod tidy

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd

run: build
	./$(BINARY)

test:
	go test ./...

clean:
	rm -rf bin
