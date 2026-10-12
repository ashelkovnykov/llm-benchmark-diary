BINARY  := bin/dbnav
PKG     := ./cmd
GOFLAGS :=

.PHONY: all build run fmt vet tidy deps test clean

all: build

## build: compile the binary into ./bin
build:
	@mkdir -p bin
	go build $(GOFLAGS) -o $(BINARY) $(PKG)

## run: run the app straight from source
run:
	go run $(PKG)

## fmt: format all Go source
fmt:
	gofmt -w .

## vet: run static analysis
vet:
	go vet ./...

## tidy: sync go.mod/go.sum
tidy:
	go mod tidy

## deps: alias for tidy (first-time setup)
deps: tidy

## test: run all tests
test:
	go test ./...

## clean: remove build artifacts
clean:
	rm -rf bin
