# ./Makefile
BIN := ./bin/app
CMD := ./cmd/main.go

.PHONY: fmt build run clean tidy vet

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

build: tidy
	mkdir -p ./bin
	go build -o $(BIN) $(CMD)

run: build
	$(BIN)

clean:
	rm -rf ./bin
	go clean

.DEFAULT_GOAL := build
