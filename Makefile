.PHONY: build run clean install test test-unit test-integration test-coverage test-verbose lint fmt vet help demo build-all dev pre-commit ci test-ci coverage-html release-build release-checksums dist-clean build-matrix e2e

# ============================================================================
# Variables (override on command line, e.g. `make VERSION=v1.2.3 release-build`)
# ============================================================================
PKG             := github.com/yuanjua/autowsl
VERSION         ?= dev
LDFLAGS_BASE    := -s -w
LDFLAGS_VERSION := -X $(PKG)/cmd.Version=$(VERSION)
LDFLAGS         := $(LDFLAGS_BASE) $(LDFLAGS_VERSION)
GO              ?= go
RACE            ?= -race

# Detect host OS (used for naming)
HOST_OS := $(shell uname | tr '[:upper:]' '[:lower:]' 2>/dev/null || echo windows)
HOST_ARCH := $(shell uname -m 2>/dev/null || echo amd64)

# Build the application
build: ## Build (unversioned) for current OS
	$(GO) build -o autowsl.exe .

# Build with version info
build-release: ## Build (stripped) embedding VERSION
	$(GO) build -ldflags="$(LDFLAGS)" -o autowsl.exe .

# Run the application
run: ## Run main directly
	$(GO) run main.go

# Clean build artifacts
clean: ## Remove build artifacts
	rm -f autowsl autowsl.exe
	rm -rf .autowsl_tmp/ dist/ coverage.* coverage.out coverage.out coverage.html
	rm -f *.appx *.appxbundle *.tar *.tar.gz

dist-clean: ## Clean only dist directory
	rm -rf dist/

# Install dependencies
install: ## Sync and download modules
	$(GO) mod download
	$(GO) mod tidy

# Run all tests
test: ## Run all tests (no race / coverage)
	$(GO) test ./...

# Run tests with verbose output
test-verbose: ## Verbose tests
	$(GO) test -v ./...

# Run unit tests only (fast)
test-unit: ## Fast unit tests (short)
	$(GO) test -short ./...

test-ci: ## Mirrors CI test command (race + coverage on tests folder)
	$(GO) test -v $(RACE) -coverprofile=coverage.out -covermode=atomic ./tests/...

# Run tests with coverage
test-coverage: ## Full project coverage (includes all pkgs)
	$(GO) test -coverprofile=coverage.out -covermode=atomic ./...
	@echo "coverage.out written"
	$(GO) tool cover -func=coverage.out | grep total || true
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage: coverage.html"

coverage-html: ## Open (generate) HTML from existing coverage.out
	@[ -f coverage.out ] || { echo "coverage.out missing - run test-ci first"; exit 1; }
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage: coverage.html"

# Run integration tests (requires WSL)
test-integration: ## Alias to run extended tests (currently same set)
	$(GO) test -v ./tests/...

# Lint code
lint: ## Run golangci-lint (installs if missing)
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint"; $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run --timeout=5m

# Format code
fmt: ## Go format
	$(GO) fmt ./...
	gofmt -s -w .

# Run go vet
vet: ## Go vet
	$(GO) vet ./...

# Check for common issues
check: fmt vet ## Basic formatting + vet + short tests
	$(GO) test -short ./...

# Show help for all commands
help: ## Show CLI help for primary commands
	@./autowsl.exe --help || true
	@echo ""
	@./autowsl.exe install --help 2>/dev/null || true
	@echo ""
	@./autowsl.exe provision --help 2>/dev/null || true

# Quick demo - show all commands
demo: ## Show common CLI usage examples
	@echo "=== AutoWSL Quick Demo ==="
	@echo "List:              ./autowsl.exe list"
	@echo "Interactive install: ./autowsl.exe install"
	@echo "Specific install:   ./autowsl.exe install 'Ubuntu 22.04 LTS' --name my-ubuntu"
	@echo "Install + curl:     ./autowsl.exe install 'Ubuntu 22.04 LTS' --playbooks curl"
	@echo "Provision:          ./autowsl.exe provision my-ubuntu --playbooks curl"
	@echo "Aliases:            ./autowsl.exe aliases"
	@echo "Download only:      ./autowsl.exe download 'Ubuntu 22.04 LTS'"
	@echo "Backup:             ./autowsl.exe backup my-ubuntu"
	@echo "Remove:             ./autowsl.exe remove my-ubuntu"

# Build for multiple platforms
build-all: ## Cross-compile common platforms (non-arm linux only)
	@mkdir -p dist
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/autowsl-windows-amd64.exe .
	GOOS=linux   GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/autowsl-linux-amd64 .
	GOOS=darwin  GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/autowsl-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/autowsl-darwin-arm64 .
	@echo "Built binaries in dist/ (version: $(VERSION))"

build-matrix: build-all ## Alias (aligns with CI naming)

# Development workflow
dev: clean install fmt vet lint test ## Full local dev cycle
	@echo "✓ Development build complete (version $(VERSION))"

# Pre-commit checks
pre-commit: fmt vet lint test-unit ## Quick pre-push hygiene
	@echo "✓ Pre-commit checks passed"

# CI/CD simulation
ci: install lint test-ci build-release ## Simulate CI pipeline locally
	@echo "✓ CI checks complete (coverage in coverage.out)"

# Release helpers -----------------------------------------------------------
release-build: ## Build Windows binaries for release (amd64 + arm64) with VERSION
	@mkdir -p dist
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/autowsl-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o dist/autowsl-windows-arm64.exe .
	@echo "Built release binaries (version: $(VERSION)) in dist/"

release-checksums: ## Generate SHA256 checksums for dist binaries
	@cd dist && shasum -a 256 autowsl-* > checksums.txt || (cd dist && sha256sum autowsl-* > checksums.txt)
	@echo "checksums.txt written under dist/"

# Simple local E2E helper (WSL required). Usage:
#   make e2e DISTRO="Ubuntu 22.04 LTS" NAME=ubuntu-2204-test PLAYBOOKS=curl
DISTRO ?= Ubuntu 22.04 LTS
NAME   ?= $(shell echo $(DISTRO) | tr '[:upper:]' '[:lower:]' | tr ' ' '-' )
PLAYBOOKS ?=
e2e: build-release ## Install a distro (and optionally provision) locally
	@echo "-> Installing $(DISTRO) as $(NAME) (playbooks: $(PLAYBOOKS))"
	./autowsl.exe install "$(DISTRO)" --name "$(NAME)" $(if $(PLAYBOOKS),--playbooks $(PLAYBOOKS),)
	@echo "Run: wsl -d $(NAME) -- bash -l"

