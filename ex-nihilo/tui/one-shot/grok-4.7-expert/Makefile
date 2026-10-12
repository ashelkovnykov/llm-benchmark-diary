.PHONY: fmt build run clean tidy

fmt:
	go fmt ./...

tidy:
	go mod tidy

build: tidy
	mkdir -p bin
	go build -o bin/dbnav ./cmd
	chmod +x bin/dbnav

run: build
	./bin/dbnav

clean:
	rm -rf bin
	go clean
	rm -f go.sum
