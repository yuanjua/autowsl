.PHONY: build run clean install test test-unit test-integration test-coverage test-verbose lint fmt vet help demo build-all

# Build the application
build:
	go build -o autowsl.exe .

# Build with version info
build-release:
	go build -ldflags="-s -w" -o autowsl.exe .

# Run the application
run:
	go run main.go

# Clean build artifacts
clean:
	rm -f autowsl.exe
	rm -rf .autowsl_tmp/
	rm -f *.appx *.appxbundle *.tar *.tar.gz
	rm -rf dist/

# Install dependencies
install:
	go mod download
	go mod tidy

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run unit tests only (fast)
test-unit:
	go test -short ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run integration tests (requires WSL)
test-integration:
	go test -v ./tests/...

# Lint code
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Run go vet
vet:
	go vet ./...

# Check for common issues
check: fmt vet
	go test -short ./...

# Show help for all commands
help:
	@./autowsl.exe --help
	@echo ""
	@./autowsl.exe install --help
	@echo ""
	@./autowsl.exe provision --help

# Quick demo - show all commands
demo:
	@echo "=== AutoWSL Commands ==="
	@echo ""
	@echo "1. List installed distributions:"
	@echo "   ./autowsl.exe list"
	@echo ""
	@echo "2. Install a distribution (interactive):"
	@echo "   ./autowsl.exe install"
	@echo ""
	@echo "3. Install a specific distribution:"
	@echo "   ./autowsl.exe install \"Ubuntu 22.04 LTS\" --name my-ubuntu --path ./wsl-distros/my-ubuntu"
	@echo ""
	@echo "4. Install with auto-provisioning:"
	@echo "   ./autowsl.exe install \"Ubuntu 22.04 LTS\" --playbooks curl,default"
	@echo ""
	@echo "5. Provision existing distribution:"
	@echo "   ./autowsl.exe provision my-ubuntu --playbooks default"
	@echo ""
	@echo "6. List playbook aliases:"
	@echo "   ./autowsl.exe aliases"
	@echo ""
	@echo "7. Download distribution package:"
	@echo "   ./autowsl.exe download \"Ubuntu 22.04 LTS\""
	@echo ""
	@echo "8. Backup a distribution:"
	@echo "   ./autowsl.exe backup <name>"
	@echo ""
	@echo "9. Remove a distribution:"
	@echo "   ./autowsl.exe remove <name>"

# Build for multiple platforms
build-all:
	@mkdir -p dist
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/autowsl-windows-amd64.exe .
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/autowsl-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/autowsl-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/autowsl-darwin-arm64 .
	@echo "Built binaries in dist/"

# Development workflow
dev: clean build test
	@echo "✓ Development build complete"

# Pre-commit checks
pre-commit: fmt vet test-unit
	@echo "✓ Pre-commit checks passed"

# CI/CD simulation
ci: install fmt vet test-coverage
	@echo "✓ CI checks complete"

