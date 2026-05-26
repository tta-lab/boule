.PHONY: help build install test lint fmt clean

help:
	@echo "Available commands:"
	@echo "  make build    - Build the bo binary"
	@echo "  make install  - Install bo to GOPATH/bin"
	@echo "  make test     - Run tests"
	@echo "  make lint     - Run golangci-lint"
	@echo "  make fmt      - Format code with gofmt"
	@echo "  make clean    - Remove built binaries"

build:
	@echo "Building bo..."
	@go build -o bo .
	@echo "Build complete: ./bo"

install:
	@echo "Installing bo..."
	@go build -o $(shell go env GOPATH)/bin/bo .
	@echo "Installed to $(shell go env GOPATH)/bin/bo"

test:
	@echo "Running tests..."
	@go test ./...

lint:
	@echo "Running linters..."
	@golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	@gofmt -w .

clean:
	@rm -f bo
	@echo "Clean complete"
