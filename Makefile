.PHONY: build test lint bench clean help

# Binary name
BINARY_NAME=yatisql
BIN_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

help:
	@echo "Makefile for yatisql - SQL Query Engine for CSV/TSV Files"
	@echo "Usage:"
	@echo "  make build       - Build the yatisql binary"
	@echo "  make test        - Run all tests"
	@echo "  make test-cover  - Run tests with coverage report"
	@echo "  make lint        - Run linters (golangci-lint)"
	@echo "  make bench       - Run benchmarks"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make deps        - Download dependencies"
	@echo "  make help        - Show this help message"

build: clean
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) -v ./cmd/yatisql
	@echo "✓ Binary built: $(BIN_DIR)/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-cover:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	@echo "Coverage report generated: coverage.out"
	$(GOTEST) -covermode=atomic -coverprofile=coverage.out ./...
	@echo "To view coverage: go tool cover -html=coverage.out"

lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "⚠ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./tests/unit

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out
	@echo "✓ Clean complete"

.DEFAULT_GOAL := help
