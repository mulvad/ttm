.PHONY: build test clean install lint

# Build variables
BINARY_NAME=ttm
VERSION?=dev
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/ttm

# Run all tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install the binary to $GOPATH/bin
install:
	$(GOINSTALL) $(LDFLAGS) ./cmd/ttm

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

# Build for release
release:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/ttm
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/ttm
