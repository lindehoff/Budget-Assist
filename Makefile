# Build variables
BINARY_NAME=budgetassist
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
BUILD_USER=$(shell whoami)
LDFLAGS=-X 'github.com/lindehoff/Budget-Assist/cmd.Version=$(VERSION)' \
        -X 'github.com/lindehoff/Budget-Assist/cmd.CommitHash=$(COMMIT_HASH)' \
        -X 'github.com/lindehoff/Budget-Assist/cmd.BuildTime=$(BUILD_TIME)' \
        -X 'github.com/lindehoff/Budget-Assist/cmd.BuildUser=$(BUILD_USER)'

# Output directory for builds
BUILD_DIR=dist

.PHONY: all build clean test release-build

all: clean build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	go clean

test:
	go test ./... -v

# Run golangci-lint
lint:
	golangci-lint run

# Install development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 

# Build binaries for all supported platforms
release-build:
	mkdir -p $(BUILD_DIR)
	# Linux builds
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64
	# macOS builds
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
	# Windows builds
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	# Generate checksums
	cd $(BUILD_DIR) && sha256sum * > checksums.txt 