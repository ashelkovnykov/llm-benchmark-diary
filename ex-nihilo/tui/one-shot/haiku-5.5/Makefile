BINARY := bin/dbexplorer

.PHONY: all build run fmt vet tidy test clean clean-db

all: fmt vet build

build: tidy
	@mkdir -p bin
	go build -o $(BINARY) ./cmd

run: build
	./$(BINARY)

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

test:
	go test ./...

clean:
	rm -rf bin

clean-db:
	rm -f db/*.db
