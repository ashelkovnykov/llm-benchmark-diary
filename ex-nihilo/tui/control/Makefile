APP_NAME := control
BIN_DIR := ./bin

.PHONY: fmt build run clean

fmt:
	go fmt ./...

build:
	go build -o ./${BIN_DIR}/${APP_NAME} ./cmd/main.go

run: build
	./${BIN_DIR}/${APP_NAME}

clean:
	rm -rf ./${BIN_DIR}
