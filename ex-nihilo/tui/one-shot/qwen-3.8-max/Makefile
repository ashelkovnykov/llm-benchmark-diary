BINARY := bin/dbnav
PKG    := ./cmd

.PHONY: build run fmt vet clean

build:
	@mkdir -p bin
	go build -o $(BINARY) $(PKG)

run: build
	./$(BINARY)

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin/
