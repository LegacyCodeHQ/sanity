.PHONY: test test-coverage coverage coverage-html clean help build build-version

# Version information (can be overridden via command line)
# Try to get version from git tag, otherwise use "dev"
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || git describe --tags 2>/dev/null || echo "")
VERSION ?= $(if $(GIT_TAG),$(GIT_TAG),dev)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Default target
help:
	@echo "Available targets:"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage percentage"
	@echo "  coverage      - Generate coverage profile (coverage.out)"
	@echo "  coverage-html - Generate HTML coverage report (coverage.html)"
	@echo "  build         - Build the binary (default version: dev)"
	@echo "  build-version - Build the binary with version info from git"
	@echo "  clean         - Remove coverage files and binary"

# Run all tests
test:
	go test ./...

# Run tests with coverage percentage (exclude cmd package as it has no tests)
test-coverage:
	@go list ./... | grep -v '/cmd$$' | xargs go test -cover

# Alternative: test all packages including cmd (may fail on Go 1.25+)
test-coverage-all:
	go test -cover ./...

# Generate coverage profile (exclude cmd package as it has no tests)
coverage:
	@echo "mode: atomic" > coverage.out
	@go list ./... | grep -v '/cmd$$' | while read pkg; do \
		go test -coverprofile=coverage.tmp -covermode=atomic $$pkg || true; \
		if [ -f coverage.tmp ]; then \
			tail -n +2 coverage.tmp >> coverage.out; \
			rm coverage.tmp; \
		fi; \
	done

# Generate HTML coverage report (requires coverage.out)
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build the binary (default version)
build:
	go build -o sanity ./main.go

# Build the binary with version information from git
build-version:
	@echo "Building with version: $(VERSION), commit: $(COMMIT), date: $(BUILD_DATE)"
	go build -ldflags "-X sanity/cmd.version=$(VERSION) -X sanity/cmd.buildDate=$(BUILD_DATE) -X sanity/cmd.commit=$(COMMIT)" -o sanity ./main.go

# Clean coverage files and binary
clean:
	rm -f coverage.out coverage.html coverage.tmp *.coverprofile *.cover sanity
