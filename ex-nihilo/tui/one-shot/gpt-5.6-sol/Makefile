BINARY := bin/sqlnav
PKG    := ./cmd

.PHONY: all build run fmt vet tidy clean test

all: build

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

build: tidy
	@mkdir -p bin
	go build -o $(BINARY) $(PKG)

run: build
	$(BINARY)

test:
	go test ./...

clean:
	rm -rf bin
