.PHONY: build build-all clean test install fmt lint tidy

BINARY=confluence
DIST=dist

VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS=-ldflags "-s -w \
	-X github.com/kristofferrisa/confluence-cli/internal/commands.Version=$(VERSION) \
	-X github.com/kristofferrisa/confluence-cli/internal/commands.GitCommit=$(GIT_COMMIT) \
	-X github.com/kristofferrisa/confluence-cli/internal/commands.BuildDate=$(BUILD_DATE)"

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/confluence

install:
	go install $(LDFLAGS) ./cmd/confluence

test:
	go test -v ./...

build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-amd64 ./cmd/confluence
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-arm64 ./cmd/confluence

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-amd64 ./cmd/confluence
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-arm64 ./cmd/confluence

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-windows-amd64.exe ./cmd/confluence

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy
