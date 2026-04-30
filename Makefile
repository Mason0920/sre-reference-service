# Hslee SRE Foundation — Makefile
# Common development tasks

GOLANGCI_LINT_VERSION := v1.64.8
GOLANGCI_LINT         := $(shell which golangci-lint 2>/dev/null)
GO                    := go
BINARY_NAME           := hslee
BUILD_DIR             := bin

.DEFAULT_GOAL := help

##@ Development

.PHONY: install
install: ## Install development tools (golangci-lint, etc.)
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: ## Run linter — fails on any error (mirrors CI behaviour)
	@if ! command -v golangci-lint &>/dev/null; then \
		echo "golangci-lint not found. Run 'make install' first."; \
		exit 1; \
	fi
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix: ## Run linter with auto-fix where possible
	golangci-lint run --fix ./...

.PHONY: fmt
fmt: ## Format all Go source files
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

##@ Testing

.PHONY: test
test: ## Run full test suite
	$(GO) test -race -count=1 ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	$(GO) test -race -count=1 -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

##@ Build

.PHONY: build
build: ## Build production binary
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./...

.PHONY: clean
clean: ## Remove build artifacts and coverage reports
	@rm -rf $(BUILD_DIR) coverage.out coverage.html

##@ Help

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)
