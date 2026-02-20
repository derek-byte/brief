# Brief - Makefile for development and releases

.PHONY: help setup build test tag publish clean

help: ## Show this help message
	@echo "Brief - Development and Release Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: ## One-time setup for releases (install svu and authenticate gh)
	@echo "Installing svu (semantic version utility)..."
	@go install github.com/caarlos0/svu@latest
	@echo ""
	@echo "Checking gh authentication..."
	@gh auth status || gh auth login
	@echo ""
	@echo "Setup complete! You can now use 'make tag' and 'make publish'"

build: ## Build the binary locally
	@echo "Building brief..."
	@go build -o brief ./cmd/brief
	@echo "Binary created: ./brief"

test: ## Run all tests
	@echo "Running tests..."
	@go test ./...

test-golden-update: ## Update golden test files
	@UPDATE_GOLDEN=1 go test ./internal/render/...

fmt: ## Format all Go files
	@gofmt -w .

lint: ## Run golangci-lint
	@golangci-lint run ./...

check: fmt ## Run format, tests, and linter
	@go test ./...
	@golangci-lint run ./...

tag: ## Calculate next version from commits and create git tag
	@echo "Calculating next semantic version..."
	@if ! command -v svu &> /dev/null; then \
		echo "Error: svu not found. Run 'make setup' first."; \
		exit 1; \
	fi
	$(eval NEXT_VERSION := $(shell svu next))
	@echo "Next version: $(NEXT_VERSION)"
	@echo ""
	@echo "Creating and pushing tag..."
	@git tag -a $(NEXT_VERSION) -m "Release $(NEXT_VERSION)"
	@git push origin $(NEXT_VERSION)
	@echo ""
	@echo "✓ Tag $(NEXT_VERSION) created and pushed"
	@echo "  Run 'make publish' to create the GitHub release"

publish: ## Run goreleaser to create GitHub release
	@echo "Publishing release with GoReleaser..."
	@if ! command -v goreleaser &> /dev/null; then \
		echo "Error: goreleaser not found. Install with: brew install goreleaser"; \
		exit 1; \
	fi
	@goreleaser release --clean
	@echo ""
	@echo "✓ Release published!"

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f brief
	@rm -rf dist/
	@echo "✓ Clean complete"

# Development helpers
dev-build: build ## Alias for build

dev-install: ## Build and install to ~/bin
	@echo "Building and installing to ~/bin..."
	@go build -o ~/bin/brief ./cmd/brief
	@echo "✓ Installed to ~/bin/brief"
