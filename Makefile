.PHONY: build test clean install example

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o scnpm .

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f scnpm
	rm -f coverage.out coverage.html
	rm -f scnpm-*

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@which golangci-lint > /dev/null || echo "golangci-lint not installed. Install with: brew install golangci-lint"
	@which golangci-lint > /dev/null && golangci-lint run || true

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install the binary to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)"

# Run example with the provided example-package-lock.json
example:
	./scnpm --file example-package-lock.json react@18.2.0 lodash@4.17.21 @types/node@18.15.13

# Run example with JSON output
example-json:
	./scnpm --file example-package-lock.json --output json react@18.2.0 lodash@4.17.21

# Show help
help:
	./scnpm --help

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o scnpm-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o scnpm-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o scnpm-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o scnpm-windows-amd64.exe .
