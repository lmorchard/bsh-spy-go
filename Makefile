.PHONY: setup build run clean lint format test

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)

# Note: Uses modernc.org/sqlite (pure Go, no CGO) so builds can run with
# CGO_ENABLED=0 and produce statically linked binaries.

# Default target
all: build

# Install development tools
setup:
	@echo "Installing development tools..."
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools installed"

# Build the application
build:
	@echo "Building bsh-spy-go..."
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bsh-spy-go .
	@echo "✅ Built: bsh-spy-go"

# Run the application
run: build
	./bsh-spy-go

# Clean build artifacts
clean:
	@rm -f bsh-spy-go
	@rm -f *.db
	@echo "✅ Cleaned"

# Lint code
lint:
	@test -f $(HOME)/go/bin/golangci-lint || { \
		echo "❌ golangci-lint not found. Install with: make setup"; \
		exit 1; \
	}
	@echo "Running linters..."
	@$(HOME)/go/bin/golangci-lint run --timeout 5m
	@echo "✅ Lint complete"

# Format code
format:
	@go fmt ./...
	@test -f $(HOME)/go/bin/gofumpt || { \
		echo "❌ gofumpt not found. Install with: make setup"; \
		exit 1; \
	}
	@$(HOME)/go/bin/gofumpt -l -w .
	@echo "✅ Format complete"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...
	@echo "✅ Tests complete"
