# Makefile for Recutils MCP Server

.PHONY: all build test clean fmt vet help

# Default target
all: test build

# Build binary
build:
	@echo "Building Recutils MCP Server..."
	go build -o recutils-mcp

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Static check
vet:
	@echo "Running static check..."
	go vet ./...

# Clean build files
clean:
	@echo "Cleaning build files..."
	rm -f recutils-mcp
	go clean

# Show help
help:
	@echo "Recutils MCP Server Build Tool"
	@echo ""
	@echo "Available commands:"
	@echo "  all     - Run tests and build"
	@echo "  build   - Build binary"
	@echo "  test    - Run tests"
	@echo "  fmt     - Format code"
	@echo "  vet     - Run static check"
	@echo "  clean   - Clean build files"
	@echo "  help    - Show this help message"
