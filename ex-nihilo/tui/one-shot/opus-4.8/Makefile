BINARY := bin/dbnav
PKG := ./cmd

.PHONY: all build run fmt vet tidy clean

all: build

build:
	@mkdir -p bin
	go build -o $(BINARY) $(PKG)

run: build
	./$(BINARY)

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin
