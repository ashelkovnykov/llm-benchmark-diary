BINARY := bin/dbtui
PKG    := ./cmd

.PHONY: all fmt vet tidy build run clean clean-db test

all: fmt vet build

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

build:
	@mkdir -p bin
	go build -o $(BINARY) $(PKG)

run:
	go run $(PKG)

test:
	go test ./...

clean:
	rm -rf bin

# Destructive: removes all generated databases.
clean-db:
	rm -rf db
