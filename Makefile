# Build variables
BINARY_NAME=budgetassist
VERSION=$(shell git describe --tags --always --dirty)
COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
BUILD_USER=$(shell whoami)
LDFLAGS=-X 'github.com/lindehoff/Budget-Assist/cmd.Version=$(VERSION)' \
        -X 'github.com/lindehoff/Budget-Assist/cmd.CommitHash=$(COMMIT_HASH)' \
        -X 'github.com/lindehoff/Budget-Assist/cmd.BuildTime=$(BUILD_TIME)' \
        -X 'github.com/lindehoff/Budget-Assist/cmd.BuildUser=$(BUILD_USER)'

.PHONY: all build clean test

all: clean build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test ./... -v

# Run golangci-lint
lint:
	golangci-lint run

# Install development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 