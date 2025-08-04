# Go Claude Code SDK Makefile

# Variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)
GONAME := go-claude-code-sdk

# Go related variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# Build variables
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Test variables
TEST_TIMEOUT := 30m
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

.PHONY: all build clean test coverage deps lint fmt vet help

# Default target
all: test build

# Build the project
build:
	@echo "Building..."
	@$(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(GONAME) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(GOBIN)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race -timeout $(TEST_TIMEOUT) ./pkg/...
	@$(GOTEST) -v -race -timeout $(TEST_TIMEOUT) ./internal/...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./pkg/...
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run integration tests
integration-test:
	@echo "Running integration tests..."
	@$(GOTEST) -v -tags=integration -timeout $(TEST_TIMEOUT) ./tests/...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem -run=^$ ./pkg/... | tee benchmark.txt

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

# Run linter
lint:
	@echo "Running linter..."
	@$(GOLINT) run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) -s -w .
	@$(GOCMD) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...

# Check for security vulnerabilities
security:
	@echo "Checking for vulnerabilities..."
	@$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest
	@gosec -quiet ./...

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	@$(GOGET) -u ./...
	@$(GOMOD) tidy

# Generate documentation
docs:
	@echo "Generating documentation..."
	@$(GOCMD) install golang.org/x/tools/cmd/godoc@latest
	@echo "Run 'godoc -http=:6060' and visit http://localhost:6060/pkg/github.com/jonwraymond/go-claude-code-sdk/"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GOCMD) install golang.org/x/tools/cmd/godoc@latest
	@$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest
	@$(GOCMD) install github.com/google/go-licenses@latest

# Run all checks (lint, vet, fmt, test)
check: fmt vet lint test

# Run pre-commit checks
pre-commit: check security

# Create a new release
release:
	@echo "Creating release $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)

# Run examples
run-examples:
	@echo "Running examples..."
	@for dir in examples/*/; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "Running $$dir"; \
			$(GOCMD) run "$$dir/main.go" || true; \
		fi \
	done

# Help
help:
	@echo "Go Claude Code SDK Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all              Run tests and build"
	@echo "  build            Build the project"
	@echo "  clean            Clean build artifacts"
	@echo "  test             Run unit tests"
	@echo "  coverage         Run tests with coverage report"
	@echo "  integration-test Run integration tests"
	@echo "  bench            Run benchmarks"
	@echo "  deps             Download dependencies"
	@echo "  lint             Run linter"
	@echo "  fmt              Format code"
	@echo "  vet              Run go vet"
	@echo "  security         Check for vulnerabilities"
	@echo "  update-deps      Update dependencies"
	@echo "  docs             Generate documentation"
	@echo "  install-tools    Install development tools"
	@echo "  check            Run all checks (fmt, vet, lint, test)"
	@echo "  pre-commit       Run pre-commit checks"
	@echo "  release          Create a new release"
	@echo "  run-examples     Run all examples"
	@echo "  help             Show this help message"