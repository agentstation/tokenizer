MAKEFLAGS += --no-print-directory

# Default target when running just 'make'
.DEFAULT_GOAL := help

# Build variables
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | cut -d ' ' -f 3)

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE) -X main.goVersion=$(GO_VERSION)"

.PHONY: all
all: test lint generate build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display the list of targets and their descriptions
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1mUsage:\033[0m\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } \
		/^###/ { printf "  \033[90m%s\033[0m\n", substr($$0, 4) }' $(MAKEFILE_LIST)

##@ Tooling 

.PHONY: install-devbox
install-devbox: ## Install Devbox
	@echo "Installing Devbox..."
	@curl -fsSL https://get.jetify.dev | bash

.PHONY: devbox-update
devbox-update: ## Update Devbox
	@devbox update

.PHONY: devbox
devbox: ## Run Devbox shell
	@devbox shell

.PHONY: install-air
install-air: ## Install Air for hot-reload development
	@echo "Installing Air..."
	@go install github.com/air-verse/air@latest

.PHONY: install-pre-commit
install-pre-commit: ## Install pre-commit hooks
	@echo "Installing pre-commit..."
	@pip install pre-commit || pip3 install pre-commit
	@pre-commit install

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/tetafro/godot/cmd/godot@latest
	@go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
	@echo "Tools installed successfully"

##@ Install Dependencies

.PHONY: deps
deps: ## Download go modules
	@echo "Downloading go modules..."
	go mod download

.PHONY: install
install: ## Install the tokenizer binary
	@echo "Installing tokenizer binary..."
	go install ./cmd/tokenizer

##@ Development
### Use 'make dev' for hot-reload development mode

.PHONY: dev
dev: ## Start development server with hot-reloading (requires Air)
	@echo "Starting development server with hot reload..."
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Air is not installed. Install with: make install-air"; \
		exit 1; \
	fi
	@air

.PHONY: fmt
fmt: ## Run go fmt
	@echo "Running go fmt..."
	@go fmt ./...
	@go mod tidy

.PHONY: fmt-all
fmt-all: ## Comprehensive formatting with all tools
	@echo "Running comprehensive formatting..."
	@echo "  → Running gofmt..."
	@gofmt -w .
	@echo "  → Running goimports..."
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w -local "github.com/agentstation/tokenizer" .; \
	else \
		echo "    goimports not installed, skipping..."; \
	fi
	@echo "  → Running godot..."
	@if command -v godot >/dev/null 2>&1; then \
		godot -w .; \
	else \
		echo "    godot not installed, skipping..."; \
	fi
	@echo "  → Running golangci-lint with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix; \
	else \
		echo "    golangci-lint not installed, skipping..."; \
	fi
	@echo "  → Running go mod tidy..."
	@go mod tidy
	@echo "Formatting complete!"

.PHONY: generate
generate: ## Generate and embed go documentation into README.md
	@echo "Generating and embedding go documentation into README.md..."
	go generate ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	golangci-lint run ./...

##@ Benchmarking, Testing, & Coverage

.PHONY: bench
bench: ## Run Go benchmarks
	@echo "Running go benchmarks..."
	go test ./... -tags=bench -bench=.

.PHONY: test
test: ## Run Go tests
	@echo "Running go tests..."
	@go test -v ./...

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -race -v ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -tags=integration -v ./... -run "Integration"

.PHONY: test-e2e
test-e2e: build ## Run end-to-end tests
	@echo "Running end-to-end tests..."
	@go test -tags=e2e -v ./cmd/tokenizer -run "E2E"

.PHONY: test-all
test-all: test test-race test-integration test-e2e ## Run all tests
	@echo "All tests completed!"

.PHONY: test-cli
test-cli: ## Run CLI-specific tests
	@echo "Running CLI tests..."
	@go test -v ./cmd/tokenizer/...

.PHONY: coverage
coverage: ## Run tests and generate coverage report
	@echo "Running tests and generating coverage report..."
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

##@ Build & Release

.PHONY: build
build: ## Build the CLI binary
	@echo "Building tokenizer binary..."
	@mkdir -p dist
	@go build $(LDFLAGS) -o dist/tokenizer ./cmd/tokenizer
	@echo "Binary built: dist/tokenizer"

.PHONY: build-all
build-all: ## Build binaries for all platforms
	@echo "Building binaries for all platforms..."
	goreleaser build --snapshot --clean

.PHONY: release
release: ## Create a new release (requires version tag)
	@echo "Creating release..."
	@if [ -z "$$(git describe --tags --exact-match 2>/dev/null)" ]; then \
		echo "Error: No tag found. Please create a tag first using 'make tag VERSION=v1.2.3'"; \
		exit 1; \
	fi
	goreleaser release --clean

.PHONY: release-snapshot
release-snapshot: ## Test release process locally (doesn't publish)
	@echo "Testing release process locally..."
	goreleaser release --snapshot --clean

.PHONY: tag
tag: ## Create and push a new version tag (usage: make tag VERSION=v1.2.3)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make tag VERSION=v1.2.3"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tag $(VERSION) created. Push with: git push origin $(VERSION)"