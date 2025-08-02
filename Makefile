.PHONY: all build test test-unit test-integration clean lint fmt vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Binary name
BINARY_NAME=claude-code-sdk

# Package lists
PACKAGES=$(shell go list ./... | grep -v /vendor/)
INTEGRATION_PACKAGES=./tests/integration/...

all: test build

build:
	$(GOBUILD) -v ./...

test: test-unit

test-unit:
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic $(PACKAGES)

test-integration:
	@echo "Running integration tests..."
	@if [ -z "$$ANTHROPIC_API_KEY" ]; then \
		echo "Warning: ANTHROPIC_API_KEY not set. Integration tests will be skipped."; \
	fi
	INTEGRATION_TESTS=true $(GOTEST) -v -tags=integration -timeout 15m $(INTEGRATION_PACKAGES)

test-all: test-unit test-integration

clean:
	$(GOCLEAN)
	rm -f coverage.txt

# Run go fmt
fmt:
	$(GOFMT) -s -w .

# Run go vet
vet:
	$(GOVET) $(PACKAGES)

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
deps-update:
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

# Run all checks
check: fmt vet lint test

# Install the library
install:
	$(GOGET) ./...

# Generate code coverage report
coverage: test-unit
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem $(PACKAGES)

# Quick test (no race detector, no coverage)
test-quick:
	$(GOTEST) $(PACKAGES)

# CI/CD pipeline command
ci: deps check test-all

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@$(GOMOD) download
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development environment ready!"

# Help
help:
	@echo "Available targets:"
	@echo "  make build           - Build the project"
	@echo "  make test            - Run unit tests"
	@echo "  make test-integration - Run integration tests (requires ANTHROPIC_API_KEY)"
	@echo "  make test-all        - Run all tests"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make fmt             - Format code"
	@echo "  make vet             - Run go vet"
	@echo "  make lint            - Run linter"
	@echo "  make deps            - Download dependencies"
	@echo "  make coverage        - Generate coverage report"
	@echo "  make bench           - Run benchmarks"
	@echo "  make ci              - Run CI pipeline"
	@echo "  make dev-setup       - Setup development environment"
	@echo "  make help            - Show this help message"