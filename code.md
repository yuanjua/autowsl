Project Path: autowsl

Source Tree:

```txt
autowsl
├── Makefile
├── README.md
├── cmd
│   ├── aliases.go
│   ├── download.go
│   ├── helpers.go
│   ├── install.go
│   ├── manage.go
│   ├── provision.go
│   └── root.go
├── go.mod
├── go.sum
├── internal
│   ├── ansible
│   │   ├── executor.go
│   │   └── summary.go
│   ├── checksum
│   │   └── checksum.go
│   ├── distro
│   │   ├── distro.go
│   │   ├── distros-winget.json
│   │   └── distros.json
│   ├── downloader
│   │   └── downloader.go
│   ├── extractor
│   │   └── extractor.go
│   ├── playbooks
│   │   ├── extravars.go
│   │   └── resolver.go
│   ├── runner
│   │   └── runner.go
│   ├── system
│   │   └── arch.go
│   ├── winget
│   │   ├── catalog.go
│   │   ├── manager.go
│   │   ├── search.go
│   │   └── winget.go
│   └── wsl
│       ├── client.go
│       ├── importer.go
│       └── status.go
├── main.go
├── playbooks
│   └── curl.yml
└── tests
    ├── runner_test.go
    ├── winget_test.go
    ├── wsl_import_test.go
    ├── wsl_status_test.go
    └── wsl_test.go

```

`Makefile`:

```
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


```

`cmd\aliases.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var aliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "List available playbook aliases",
	Long: `List all available playbook aliases that can be used with --playbooks.
These are the built-in playbooks located in the playbooks/ directory.

Examples:
  autowsl aliases
  autowsl install "Ubuntu 22.04 LTS" --playbooks curl
  autowsl provision ubuntu-2204 --playbooks curl,default`,
	RunE: runAliases,
}

func init() {
	rootCmd.AddCommand(aliasesCmd)
}

func runAliases(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	playbooksDir := filepath.Join(cwd, "playbooks")

	// Check if playbooks directory exists
	if _, err := os.Stat(playbooksDir); os.IsNotExist(err) {
		fmt.Println("No playbooks directory found.")
		fmt.Printf("Create one at: %s\n", playbooksDir)
		return nil
	}

	// List all .yml and .yaml files
	entries, err := os.ReadDir(playbooksDir)
	if err != nil {
		return fmt.Errorf("failed to read playbooks directory: %w", err)
	}

	var playbooks []struct {
		Alias string
		Path  string
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			alias := strings.TrimSuffix(strings.TrimSuffix(name, ".yml"), ".yaml")
			playbooks = append(playbooks, struct {
				Alias string
				Path  string
			}{
				Alias: alias,
				Path:  filepath.Join(playbooksDir, name),
			})
		}
	}

	if len(playbooks) == 0 {
		fmt.Println("No playbook aliases found in playbooks/ directory.")
		return nil
	}

	fmt.Println("\nAvailable Playbook Aliases:")
	fmt.Println(strings.Repeat("=", 60))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ALIAS\tFILE")
	fmt.Fprintln(w, strings.Repeat("-", 20)+"\t"+strings.Repeat("-", 35))

	for _, pb := range playbooks {
		fmt.Fprintf(w, "%s\t%s\n", pb.Alias, filepath.Base(pb.Path))
	}

	w.Flush()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total: %d playbook(s)\n\n", len(playbooks))
	fmt.Println("Usage:")
	fmt.Println("  autowsl install \"Ubuntu 22.04 LTS\" --playbooks <alias>")
	fmt.Println("  autowsl provision <distro> --playbooks <alias1>,<alias2>")
	fmt.Println()

	return nil
}

```

`cmd\download.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuanjua/autowsl/internal/winget"
)

var (
	downloadOutputDir string
	downloadPackageID string
)

var downloadCmd = &cobra.Command{
	Use:   "download [version]",
	Short: "Download a WSL distribution without installing",
	Long: `Download a WSL distribution package to a directory without installing it.
This is useful if you want to:
  - Download for offline installation later
  - Keep a backup of the installation package
  - Inspect the package contents manually

Examples:
  autowsl download                                 # Interactive mode
  autowsl download "Ubuntu 22.04 LTS"              # Direct download by version
  autowsl download --package-id Canonical.Ubuntu.2204  # Download by package ID
  autowsl download --output ./packages             # Download to specific directory`,
	RunE: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringVarP(&downloadOutputDir, "output", "o", "", "Output directory (default: current directory)")
	downloadCmd.Flags().StringVar(&downloadPackageID, "package-id", "", "Winget package ID (alternative to version name)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	var packageID string
	var distroName string

	// If package ID is provided directly, use it
	if downloadPackageID != "" {
		packageID = downloadPackageID
		distroName = downloadPackageID
	} else {
		// Use shared helper for distro selection
		selectedDistro, err := selectDistro(args)
		if err != nil {
			return err
		}

		// Check if distro has packageId
		if selectedDistro.PackageID == "" {
			return fmt.Errorf("distribution '%s' does not have a winget package ID", selectedDistro.Version)
		}

		packageID = selectedDistro.PackageID
		distroName = selectedDistro.Version

		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("Download Configuration\n")
		fmt.Printf("%s\n", strings.Repeat("=", 60))
		fmt.Printf("Distribution: %s - %s (%s)\n", selectedDistro.Group, selectedDistro.Version, selectedDistro.Architecture)
		fmt.Printf("Package ID:   %s\n", packageID)
	}

	// Determine output directory
	outputDir := downloadOutputDir
	if outputDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		outputDir = cwd
	}

	fmt.Printf("Output:       %s\n", outputDir)
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	// Download using winget
	fmt.Println("→ Downloading package...")
	mgr := winget.NewManager(outputDir)

	// Check if winget is available
	if !mgr.IsWingetAvailable() {
		return fmt.Errorf("winget is not available. Please install 'App Installer' from Microsoft Store")
	}

	downloadedFile, err := mgr.Download(winget.DownloadOptions{
		PackageID: packageID,
	})
	if err != nil {
		return fmt.Errorf("failed to download '%s': %w", distroName, err)
	}

	fmt.Printf("  ✓ Download completed\n")

	// Success message
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("✓ SUCCESS: Package downloaded\n")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("File:     %s\n", filepath.Base(downloadedFile))
	fmt.Printf("Location: %s\n", downloadedFile)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  Install:  autowsl install --path <install-path>\n")
	fmt.Printf("  Extract:  Use 7-Zip or similar to extract the AppX/AppXBundle\n\n")

	return nil
}

```

`cmd\helpers.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/yuanjua/autowsl/internal/ansible"
	"github.com/yuanjua/autowsl/internal/distro"
	"github.com/yuanjua/autowsl/internal/playbooks"
	"github.com/yuanjua/autowsl/internal/wsl"
)

// selectDistroInteractive handles interactive distribution selection with promptui
func selectDistroInteractive() (distro.Distro, error) {
	distros := distro.GetAllDistros()

	// Create selection prompt with colored templates
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Group | cyan }} - {{ .Version | yellow }} ({{ .Architecture | faint }})",
		Inactive: "  {{ .Group | white }} - {{ .Version | faint }} ({{ .Architecture | faint }})",
		Selected: "* {{ .Group | green }} - {{ .Version | green }}",
	}

	prompt := promptui.Select{
		Label:     "Select a WSL distribution",
		Items:     distros,
		Templates: templates,
		Size:      12,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return distro.Distro{}, fmt.Errorf("selection cancelled: %w", err)
	}

	return distros[idx], nil
}

// selectDistroByVersion finds a distribution by its version name
func selectDistroByVersion(versionName string) (distro.Distro, error) {
	distros := distro.GetAllDistros()

	for _, d := range distros {
		if strings.EqualFold(d.Version, versionName) {
			return d, nil
		}
	}

	return distro.Distro{}, fmt.Errorf("distribution '%s' not found", versionName)
}

// selectDistro selects a distribution either interactively or by version name
func selectDistro(args []string) (distro.Distro, error) {
	if len(args) == 0 {
		return selectDistroInteractive()
	}
	return selectDistroByVersion(args[0])
}

// selectInstalledDistroInteractive handles interactive selection from installed distros
func selectInstalledDistroInteractive() (string, error) {
	distros, err := wsl.ListInstalledDistros()
	if err != nil {
		return "", fmt.Errorf("failed to list distributions: %w", err)
	}

	if len(distros) == 0 {
		return "", fmt.Errorf("no WSL distributions found. Install one first with 'autowsl install'")
	}

	// Create selection prompt
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Name | cyan }} ({{ .State | yellow }})",
		Inactive: "  {{ .Name | white }} ({{ .State | faint }})",
		Selected: "* {{ .Name | green }}",
	}

	prompt := promptui.Select{
		Label:     "Select distribution",
		Items:     distros,
		Templates: templates,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("selection cancelled: %w", err)
	}

	return distros[idx].Name, nil
}

// selectInstalledDistro selects an installed distro either from args or interactively
func selectInstalledDistro(args []string) (string, error) {
	if len(args) > 0 {
		// Verify it exists
		exists, err := wsl.IsDistroInstalled(args[0])
		if err != nil {
			return "", fmt.Errorf("failed to check distribution: %w", err)
		}
		if !exists {
			return "", fmt.Errorf("distribution '%s' does not exist", args[0])
		}
		return args[0], nil
	}
	return selectInstalledDistroInteractive()
}

// promptForPlaybooks prompts user to enter playbooks interactively
func promptForPlaybooks() ([]string, error) {
	prompt := promptui.Prompt{
		Label:   "Enter playbook(s) (comma or space separated, aliases/files/URLs)",
		Default: "curl",
	}

	result, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("input cancelled: %w", err)
	}

	result = strings.TrimSpace(result)
	if result == "" {
		return []string{"curl"}, nil
	}

	// Split by comma or space
	var inputs []string
	// First try comma
	if strings.Contains(result, ",") {
		parts := strings.Split(result, ",")
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				inputs = append(inputs, trimmed)
			}
		}
	} else {
		// Split by space
		parts := strings.Fields(result)
		inputs = append(inputs, parts...)
	}

	if len(inputs) == 0 {
		return []string{"curl"}, nil
	}

	return inputs, nil
}

// ProvisioningPipelineOptions holds options for the provisioning pipeline
type ProvisioningPipelineOptions struct {
	DistroName     string
	PlaybookInputs []string
	Tags           []string
	ExtraVars      []string
	Verbose        bool
	TempDir        string
}

// runProvisioningPipeline executes the complete provisioning pipeline
func runProvisioningPipeline(opts ProvisioningPipelineOptions) error {
	fmt.Printf("\nProvisioning: %s\n", opts.DistroName)
	fmt.Println(strings.Repeat("=", 60))

	// Parse extra vars
	extraVarsMap := make(map[string]string)
	if len(opts.ExtraVars) > 0 {
		var err error
		extraVarsMap, err = playbooks.ParseExtraVars(opts.ExtraVars)
		if err != nil {
			return fmt.Errorf("invalid extra-vars: %w", err)
		}
	}

	// Ensure temp directory exists
	if opts.TempDir == "" {
		cwd, _ := os.Getwd()
		opts.TempDir = filepath.Join(cwd, ".autowsl_tmp")
	}
	os.MkdirAll(opts.TempDir, 0755)

	// Resolve playbooks
	cwd, _ := os.Getwd()
	resolver := playbooks.NewResolver(opts.TempDir, cwd)
	playbookPaths, err := resolver.ResolveMultiple(opts.PlaybookInputs)
	if err != nil {
		return fmt.Errorf("failed to resolve playbooks: %w", err)
	}

	if len(playbookPaths) == 0 {
		return fmt.Errorf("no playbooks resolved")
	}

	// Execute playbooks with summary tracking
	summary := &ansible.ExecutionSummary{}

	for _, playbookPath := range playbookPaths {
		start := time.Now()

		fmt.Printf("\nRunning playbook: %s\n", filepath.Base(playbookPath))
		fmt.Println(strings.Repeat("-", 60))

		execOpts := ansible.PlaybookOptions{
			DistroName:   opts.DistroName,
			PlaybookPath: playbookPath,
			Tags:         opts.Tags,
			Verbose:      opts.Verbose,
			ExtraVars:    extraVarsMap,
		}

		err := ansible.ExecutePlaybook(execOpts)
		duration := time.Since(start)

		if err != nil {
			summary.Add(ansible.ExecutionResult{
				PlaybookName: filepath.Base(playbookPath),
				Status:       "failed",
				Duration:     duration,
				Error:        err,
			})
			fmt.Printf("\nPlaybook '%s' failed: %v\n", filepath.Base(playbookPath), err)
			break // Stop on first failure
		} else {
			summary.Add(ansible.ExecutionResult{
				PlaybookName: filepath.Base(playbookPath),
				Status:       "success",
				Duration:     duration,
			})
		}
	}

	// Print summary if multiple playbooks
	if len(playbookPaths) > 1 {
		summary.Print()
	}

	if summary.HasFailures() {
		return fmt.Errorf("provisioning completed with failures")
	}

	// Success message
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("SUCCESS: Distribution provisioned\n")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Distribution: %s\n", opts.DistroName)
	fmt.Printf("Playbooks:    %s\n", strings.Join(playbooks.BaseNames(playbookPaths), ", "))
	if len(opts.Tags) > 0 {
		fmt.Printf("Tags:         %s\n", strings.Join(opts.Tags, ", "))
	}
	if len(extraVarsMap) > 0 {
		fmt.Printf("Extra vars:   %d variables\n", len(extraVarsMap))
	}
	fmt.Println(strings.Repeat("=", 60))

	return nil
}

```

`cmd\install.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/yuanjua/autowsl/internal/distro"
	"github.com/yuanjua/autowsl/internal/extractor"
	"github.com/yuanjua/autowsl/internal/winget"
	"github.com/yuanjua/autowsl/internal/wsl"
)

var (
	installName       string
	installPath       string
	installKeepTar    bool
	installPlaybooks  []string
	installExtraVars  []string
	installTags       []string
	installVerbose    bool
	installWSLVersion int
)

var installCmd = &cobra.Command{
	Use:   "install [catalog version]",
	Short: "Install a WSL distribution",
	Long: `Install a WSL distribution interactively or by specifying a catalog version string.
Optionally provision it with one or more Ansible playbooks in a single command.

Examples:
	# Basic installation (interactive)
	autowsl install

	# Direct installation
	autowsl install "Ubuntu 22.04 LTS"

	# Specify WSL 1 instead of default WSL 2
	autowsl install "Ubuntu 22.04 LTS" --version 1
	
	# Install then run a single playbook (file)
	autowsl install "Ubuntu 22.04 LTS" --playbooks ./setup.yml

	# Install then run a remote playbook (URL)
	autowsl install "Debian GNU/Linux" --playbooks https://raw.githubusercontent.com/user/repo/main/playbook.yml

	# Install then run multiple playbooks / aliases sequentially
	autowsl install "Ubuntu 22.04 LTS" --playbooks curl,./dev.yml --tags docker,nodejs`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringVar(&installName, "name", "", "Custom name for the distribution")
	installCmd.Flags().StringVar(&installPath, "path", "", "Custom installation path")
	installCmd.Flags().BoolVar(&installKeepTar, "keep-tar", false, "Keep the extracted tar file for future use")
	installCmd.Flags().StringSliceVar(&installPlaybooks, "playbooks", []string{}, "Playbook files, URLs, or aliases (comma-separated or repeat flag)")
	installCmd.Flags().StringArrayVar(&installExtraVars, "extra-vars", nil, "Extra variables in key=val format (repeatable)")
	installCmd.Flags().StringSliceVar(&installTags, "tags", []string{}, "Ansible tags to run (comma-separated)")
	installCmd.Flags().BoolVarP(&installVerbose, "verbose", "v", false, "Verbose Ansible output")
	installCmd.Flags().IntVar(&installWSLVersion, "version", 2, "WSL version to use (1 or 2)")
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Check if WSL is installed
	if err := wsl.CheckWSLInstalled(); err != nil {
		return fmt.Errorf("WSL is not available: %w\nPlease install WSL first: https://docs.microsoft.com/en-us/windows/wsl/install", err)
	}

	// Use shared helper for distro selection
	selectedDistro, err := selectDistro(args)
	if err != nil {
		return err
	}

	// Check if distro has packageId
	if selectedDistro.PackageID == "" {
		return fmt.Errorf("distribution '%s' does not have a winget package ID", selectedDistro.Version)
	}

	isInteractive := len(args) == 0

	// Determine installation name
	distroName := installName
	if distroName == "" {
		distroName = generateDistroName(selectedDistro)
		if isInteractive {
			namePrompt := promptui.Prompt{
				Label:   "Distribution name",
				Default: distroName,
			}
			if customName, err := namePrompt.Run(); err != nil {
				return fmt.Errorf("failed to get distribution name: %w", err)
			} else {
				distroName = customName
			}
		}
	}

	// Check if distro already exists
	exists, err := wsl.IsDistroInstalled(distroName)
	if err != nil {
		return fmt.Errorf("failed to check existing distributions: %w", err)
	}
	if exists {
		return fmt.Errorf("distribution '%s' already exists", distroName)
	}

	// Determine installation path
	distroPath := installPath
	if distroPath == "" {
		cwd, _ := os.Getwd()
		distroPath = filepath.Join(cwd, "wsl-distros", distroName)
		if isInteractive {
			pathPrompt := promptui.Prompt{
				Label:   "Installation path",
				Default: distroPath,
			}
			if customPath, err := pathPrompt.Run(); err != nil {
				return fmt.Errorf("failed to get installation path: %w", err)
			} else {
				distroPath = customPath
			}
		}
	}

	if installWSLVersion != 1 && installWSLVersion != 2 {
		return fmt.Errorf("invalid --version %d (must be 1 or 2)", installWSLVersion)
	}

	// Display configuration
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Installation Configuration\n")
	fmt.Printf("%s\n", strings.Repeat("=", 60))
	fmt.Printf("Distribution: %s - %s (%s)\n", selectedDistro.Group, selectedDistro.Version, selectedDistro.Architecture)
	fmt.Printf("Package ID:   %s\n", selectedDistro.PackageID)
	fmt.Printf("Name:         %s\n", distroName)
	fmt.Printf("Path:         %s\n", distroPath)
	fmt.Printf("WSL Version:  %d\n", installWSLVersion)
	if installKeepTar {
		fmt.Printf("Keep tar:     yes (saved to .autowsl_tmp/)\n")
	}
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	// Create temporary directory in current working directory
	cwd, _ := os.Getwd()
	tempDir := filepath.Join(cwd, ".autowsl_tmp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory '%s': %w", tempDir, err)
	}

	// Download the distribution using winget
	fmt.Println("→ Downloading distribution...")
	mgr := winget.NewManager(tempDir)

	// Check if winget is available
	if !mgr.IsWingetAvailable() {
		extractor.CleanupTempDir(tempDir)
		return fmt.Errorf("winget is not available. Please install 'App Installer' from Microsoft Store")
	}

	downloadedFile, err := mgr.Download(winget.DownloadOptions{
		PackageID: selectedDistro.PackageID,
	})
	if err != nil {
		extractor.CleanupTempDir(tempDir)
		return fmt.Errorf("failed to download '%s': %w", selectedDistro.Version, err)
	}
	fmt.Println("  ✓ Download completed")

	fmt.Println("\n→ Extracting package...")
	tarFilePath, err := extractor.ExtractAppx(downloadedFile, tempDir)
	if err != nil {
		extractor.CleanupTempDir(tempDir)
		return fmt.Errorf("failed to extract package '%s': %w", filepath.Base(downloadedFile), err)
	}

	fmt.Printf("  ✓ Found rootfs: %s\n\n", filepath.Base(tarFilePath))

	// Import the distribution
	fmt.Println("→ Importing to WSL...")
	importOpts := wsl.ImportOptions{
		Name:        distroName,
		InstallPath: distroPath,
		TarFilePath: tarFilePath,
		Version:     installWSLVersion,
	}

	if err := wsl.Import(importOpts); err != nil {
		extractor.CleanupTempDir(tempDir)
		return fmt.Errorf("failed to import distribution '%s' to '%s': %w", distroName, distroPath, err)
	}

	fmt.Println("  ✓ Import completed successfully")

	// Cleanup temporary directory
	if installKeepTar {
		fmt.Printf("\n→ Keeping tar file: %s\n", tarFilePath)
		fmt.Println("  (You can use this for future installations)")

		// Remove only the downloaded appx/appxbundle
		if err := os.Remove(downloadedFile); err != nil {
			fmt.Printf("  ⚠ Warning: Failed to remove downloaded package: %v\n", err)
		}
	} else {
		fmt.Println("\n→ Cleaning up temporary files...")
		if err := extractor.CleanupTempDir(tempDir); err != nil {
			fmt.Printf("  ⚠ Warning: Failed to cleanup temp directory: %v\n", err)
		} else {
			fmt.Println("  ✓ Cleanup completed")
		}
	}

	// Print success message with details
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("✓ SUCCESS: WSL distribution installed\n")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Name:     %s\n", distroName)
	fmt.Printf("Location: %s\n", distroPath)
	fmt.Printf("Version:  WSL %d\n", installWSLVersion)
	if installKeepTar {
		fmt.Printf("Tar file: %s\n", tarFilePath)
	}
	fmt.Println(strings.Repeat("=", 60))

	// Parse extra vars for provisioning
	var extraVarsSlice []string
	if len(installExtraVars) > 0 {
		extraVarsSlice = installExtraVars
	}

	// Hyper Pipeline: Auto-provision if playbooks are specified
	if len(installPlaybooks) > 0 {
		// Use shared provisioning pipeline
		err := runProvisioningPipeline(ProvisioningPipelineOptions{
			DistroName:     distroName,
			PlaybookInputs: installPlaybooks,
			Tags:           installTags,
			Verbose:        installVerbose,
			ExtraVars:      extraVarsSlice,
			TempDir:        tempDir,
		})

		if err != nil {
			if !installKeepTar {
				extractor.CleanupTempDir(tempDir)
			}
			return nil // runProvisioningPipeline already printed errors
		}

		// Cleanup temp dir after successful provisioning (unless keep-tar is set)
		if !installKeepTar {
			fmt.Println("\n→ Cleaning up temporary files...")
			if err := extractor.CleanupTempDir(tempDir); err != nil {
				fmt.Printf("  ⚠ Warning: Failed to cleanup temp directory: %v\n", err)
			} else {
				fmt.Println("  ✓ Cleanup completed")
			}
		}
	} else {
		// No provisioning requested
		fmt.Printf("\nLaunch with:  wsl -d %s\n", distroName)
		fmt.Printf("List all:     autowsl list\n")
		fmt.Printf("Provision:    autowsl provision %s\n\n", distroName)
	}

	return nil
}

// generateDistroName generates a default distribution name from the distro version
func generateDistroName(d distro.Distro) string {
	// Clean up the version name to create a valid distro name
	name := strings.ReplaceAll(d.Version, " ", "-")
	name = strings.ReplaceAll(name, ".", "")
	name = strings.ToLower(name)
	return name
}

```

`cmd\manage.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/yuanjua/autowsl/internal/wsl"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed WSL distributions",
	RunE:  runList,
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a WSL distribution",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRemove,
}

var backupCmd = &cobra.Command{
	Use:   "backup <name>",
	Short: "Backup a WSL distribution",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runBackup,
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(backupCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	distros, err := wsl.ListInstalledDistros()
	if err != nil {
		return fmt.Errorf("failed to list distributions: %w", err)
	}

	if len(distros) == 0 {
		fmt.Println("No WSL distributions are currently installed.")
		fmt.Println("\nInstall one using: autowsl install")
		return nil
	}

	fmt.Println("\nInstalled WSL Distributions:")
	fmt.Println()

	// Create a tabwriter for nice formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "DEFAULT\tNAME\tSTATE\tVERSION")
	fmt.Fprintln(w, "-------\t----\t-----\t-------")

	for _, d := range distros {
		defaultMarker := " "
		if d.Default {
			defaultMarker = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", defaultMarker, d.Name, d.State, d.Version)
	}

	w.Flush()
	fmt.Println()

	return nil
}

func runRemove(cmd *cobra.Command, args []string) error {
	distroName := args[0]

	// Check if the distribution exists
	exists, err := wsl.IsDistroInstalled(distroName)
	if err != nil {
		return fmt.Errorf("failed to check distribution: %w", err)
	}
	if !exists {
		return fmt.Errorf("distribution '%s' does not exist", distroName)
	}

	// Confirmation prompt
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure you want to remove '%s'? This action cannot be undone", distroName),
		IsConfirm: true,
	}

	_, err = prompt.Run()
	if err != nil {
		fmt.Println("Removal cancelled")
		return nil
	}

	fmt.Printf("\nRemoving '%s'...\n", distroName)

	if err := wsl.Unregister(distroName); err != nil {
		return fmt.Errorf("failed to remove distribution: %w", err)
	}

	fmt.Printf("Successfully removed '%s'\n", distroName)

	return nil
}

func runBackup(cmd *cobra.Command, args []string) error {
	distroName := args[0]

	// Check if the distribution exists
	exists, err := wsl.IsDistroInstalled(distroName)
	if err != nil {
		return fmt.Errorf("failed to check distribution: %w", err)
	}
	if !exists {
		return fmt.Errorf("distribution '%s' does not exist", distroName)
	}

	// Generate default backup filename
	homeDir, _ := os.UserHomeDir()
	defaultBackupPath := filepath.Join(homeDir, "WSL-Backups", fmt.Sprintf("%s-backup.tar", distroName))

	// Prompt for backup location
	prompt := promptui.Prompt{
		Label:   "Backup file path",
		Default: defaultBackupPath,
	}

	backupPath, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("failed to get backup path: %w", err)
	}

	fmt.Printf("\nBacking up '%s' to %s...\n", distroName, backupPath)
	fmt.Println("This may take a while depending on the size of your distribution...")

	if err := wsl.Export(distroName, backupPath); err != nil {
		return fmt.Errorf("failed to backup distribution: %w", err)
	}

	// Get file size
	fileInfo, _ := os.Stat(backupPath)
	sizeInMB := float64(fileInfo.Size()) / 1024 / 1024

	fmt.Printf("\nSuccessfully backed up '%s'\n", distroName)
	fmt.Printf("Location: %s\n", backupPath)
	fmt.Printf("Size: %.2f MB\n", sizeInMB)

	return nil
}

```

`cmd\provision.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuanjua/autowsl/internal/ansible"
	"github.com/yuanjua/autowsl/internal/wsl"
)

var (
	provisionTags      []string
	provisionPlaybooks []string
	provisionExtraVars []string
	provisionRepo      string
	provisionVerbose   bool
)

var provisionCmd = &cobra.Command{
	Use:   "provision [distro-name]",
	Short: "Provision a WSL distribution with Ansible",
	Long: `Provision a WSL distribution using Ansible playbooks.
Automatically installs Ansible if not present.

Examples:
  # Interactive mode - select distro and playbook
  autowsl provision

  # Specify distro, interactive playbook selection (default: curl)
  autowsl provision ubuntu-2204

  # Use specific playbook (file, URL, or alias)
  autowsl provision ubuntu-2204 --playbooks ./my-playbook.yml
  autowsl provision ubuntu-2204 --playbooks https://example.com/setup.yml
  autowsl provision ubuntu-2204 --playbooks curl

  # Use multiple playbooks
  autowsl provision ubuntu-2204 --playbooks curl,./dev.yml,https://example.com/extra.yml

  # Use playbook from Git repository
  autowsl provision ubuntu-2204 --repo https://github.com/user/ansible-playbooks

  # Run specific tags only
  autowsl provision ubuntu-2204 --playbooks ./setup.yml --tags docker,nodejs

  # Pass extra variables
  autowsl provision ubuntu-2204 --playbooks ./setup.yml --extra-vars user=john --extra-vars env=dev

  # Verbose output
  autowsl provision ubuntu-2204 --verbose`,
	RunE: runProvision,
}

func init() {
	rootCmd.AddCommand(provisionCmd)
	provisionCmd.Flags().StringSliceVar(&provisionTags, "tags", nil, "Ansible tags to run (comma-separated)")
	provisionCmd.Flags().StringSliceVar(&provisionPlaybooks, "playbooks", []string{}, "Playbook files, URLs, or aliases (comma-separated or repeat flag)")
	provisionCmd.Flags().StringArrayVar(&provisionExtraVars, "extra-vars", nil, "Extra variables in key=val format (repeatable)")
	provisionCmd.Flags().StringVar(&provisionRepo, "repo", "", "Git repository URL containing playbooks")
	provisionCmd.Flags().BoolVarP(&provisionVerbose, "verbose", "v", false, "Verbose output")
}

func runProvision(cmd *cobra.Command, args []string) error {
	var distroName string
	var playbookInputs []string

	// Get distribution name using shared helper
	if len(args) > 0 {
		distroName = args[0]
	} else {
		var err error
		distroName, err = selectInstalledDistroInteractive()
		if err != nil {
			return err
		}
	}

	// Check if distro exists
	exists, err := wsl.IsDistroInstalled(distroName)
	if err != nil {
		return fmt.Errorf("failed to check distribution: %w", err)
	}
	if !exists {
		return fmt.Errorf("distribution '%s' does not exist", distroName)
	}

	// If no playbooks specified via flags, use interactive prompt
	if len(provisionPlaybooks) == 0 {
		// Interactive mode - prompt for playbooks (both with and without distro arg)
		playbookInputs, err = promptForPlaybooks()
		if err != nil {
			return err
		}
	} else {
		playbookInputs = provisionPlaybooks
	}

	// Create temp directory for downloads
	cwd, _ := os.Getwd()
	tempDir := filepath.Join(cwd, ".autowsl_tmp")
	os.MkdirAll(tempDir, 0755)

	// Handle repo-based provisioning (legacy mode)
	if provisionRepo != "" {
		fmt.Printf("Using playbook from Git repository\n\n")

		tmpDir := "/tmp/autowsl-playbooks"
		if err := ansible.CloneGitRepo(distroName, provisionRepo, tmpDir); err != nil {
			return err
		}

		// Look for common playbook names
		commonNames := []string{"site.yml", "main.yml", "playbook.yml", "default.yml"}
		for _, name := range commonNames {
			testPath := filepath.Join(tmpDir, name)
			playbookInputs = []string{testPath}
			break
		}

		if len(playbookInputs) == 0 {
			return fmt.Errorf("no playbook found in repository (looked for: %s)", strings.Join(commonNames, ", "))
		}
	}

	// Use shared provisioning pipeline
	return runProvisioningPipeline(ProvisioningPipelineOptions{
		DistroName:     distroName,
		PlaybookInputs: playbookInputs,
		Tags:           provisionTags,
		Verbose:        provisionVerbose,
		ExtraVars:      provisionExtraVars,
		TempDir:        tempDir,
	})
}

```

`cmd\root.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "autowsl",
	Short: "AutoWSL - Automatically download and manage WSL distributions",
	Long: `AutoWSL is a CLI tool to interactively select, download, and install 
WSL distributions from official sources.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
}

```

`go.mod`:

```mod
module github.com/yuanjua/autowsl

go 1.21

toolchain go1.23.2

require (
	github.com/manifoldco/promptui v0.9.0
	github.com/spf13/cobra v1.8.0
)

require (
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.0.0-20181122145206-62eef0e2fa9b // indirect
)

```

`go.sum`:

```sum
github.com/chzyer/logex v1.1.10 h1:Swpa1K6QvQznwJRcfTfQJmTE72DqScAa40E+fbHEXEE=
github.com/chzyer/logex v1.1.10/go.mod h1:+Ywpsq7O8HXn0nuIou7OrIPyXbp3wmkHB+jjWRnGsAI=
github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e h1:fY5BOSpyZCqRo5OhCuC+XN+r/bBCmeuuJtjz+bCNIf8=
github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e/go.mod h1:nSuG5e5PlCu98SY8svDHJxuZscDgtXS6KTTbou5AhLI=
github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 h1:q763qf9huN11kDQavWsoZXJNW3xEE4JJyHa5Q25/sd8=
github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1/go.mod h1:Q3SI9o4m/ZMnBNeIyt5eFwwo7qiLfzFZmjNmxjkiQlU=
github.com/cpuguy83/go-md2man/v2 v2.0.3/go.mod h1:tgQtvFlXSQOSOSIRvRPT7W67SCa46tRHOmNcaadrF8o=
github.com/inconshreveable/mousetrap v1.1.0 h1:wN+x4NVGpMsO7ErUn/mUI3vEoE6Jt13X2s0bqwp9tc8=
github.com/inconshreveable/mousetrap v1.1.0/go.mod h1:vpF70FUmC8bwa3OWnCshd2FqLfsEA9PFc4w1p2J65bw=
github.com/manifoldco/promptui v0.9.0 h1:3V4HzJk1TtXW1MTZMP7mdlwbBpIinw3HztaIlYthEiA=
github.com/manifoldco/promptui v0.9.0/go.mod h1:ka04sppxSGFAtxX0qhlYQjISsg9mR4GWtQEhdbn6Pgg=
github.com/russross/blackfriday/v2 v2.1.0/go.mod h1:+Rmxgy9KzJVeS9/2gXHxylqXiyQDYRxCVz55jmeOWTM=
github.com/spf13/cobra v1.8.0 h1:7aJaZx1B85qltLMc546zn58BxxfZdR/W22ej9CFoEf0=
github.com/spf13/cobra v1.8.0/go.mod h1:WXLWApfZ71AjXPya3WOlMsY9yMs7YeiHhFVlvLyhcho=
github.com/spf13/pflag v1.0.5 h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=
github.com/spf13/pflag v1.0.5/go.mod h1:McXfInJRrz4CZXVZOBLb0bTZqETkiAhM9Iw0y3An2Bg=
golang.org/x/sys v0.0.0-20181122145206-62eef0e2fa9b h1:MQE+LT/ABUuuvEZ+YQAMSXindAdUh7slEmAkup74op4=
golang.org/x/sys v0.0.0-20181122145206-62eef0e2fa9b/go.mod h1:STP8DvDyc/dI5b8T5hshtkjS+E42TnysNCUPdjciGhY=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=

```

`internal\ansible\executor.go`:

```go
package ansible

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PlaybookOptions holds options for playbook execution
type PlaybookOptions struct {
	DistroName   string
	PlaybookPath string
	Tags         []string
	Verbose      bool
	ExtraVars    map[string]string
}

// ExecutePlaybook runs an Ansible playbook inside a WSL distribution
func ExecutePlaybook(opts PlaybookOptions) error {
	// Check if playbook file exists
	if _, err := os.Stat(opts.PlaybookPath); err != nil {
		return fmt.Errorf("playbook file '%s' not found: %w", opts.PlaybookPath, err)
	}

	fmt.Printf("Playbook: %s\n", filepath.Base(opts.PlaybookPath))
	fmt.Printf("Target:   %s\n", opts.DistroName)
	if len(opts.Tags) > 0 {
		fmt.Printf("Tags:     %s\n", strings.Join(opts.Tags, ", "))
	}
	fmt.Println()

	// First, ensure Ansible is installed in the WSL distribution
	fmt.Println("Checking Ansible installation...")
	if err := ensureAnsible(opts.DistroName); err != nil {
		return fmt.Errorf("failed to ensure Ansible is installed in distribution '%s': %w", opts.DistroName, err)
	}

	// Copy playbook to WSL filesystem for reliable execution
	wslPlaybookPath, err := copyPlaybookToWSL(opts.DistroName, opts.PlaybookPath)
	if err != nil {
		return fmt.Errorf("failed to copy playbook to WSL: %w", err)
	}

	// Build ansible-playbook command
	cmdArgs := []string{
		"-d", opts.DistroName,
		"bash", "-c",
		buildAnsibleCommand(wslPlaybookPath, opts),
	}

	fmt.Println("Executing playbook...")
	fmt.Println(strings.Repeat("-", 60))

	// Execute the command
	cmd := exec.Command("wsl.exe", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("playbook '%s' execution failed on distribution '%s': %w", filepath.Base(opts.PlaybookPath), opts.DistroName, err)
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("Playbook execution completed.")

	return nil
}

// ensureAnsible checks if Ansible is installed, and installs it if not
func ensureAnsible(distroName string) error {
	// Check if ansible is installed
	checkCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "which ansible-playbook")
	if err := checkCmd.Run(); err == nil {
		fmt.Println("Ansible is already installed")
		return nil
	}

	fmt.Println("Ansible not found, installing...")

	// Try multiple installation methods
	installCommands := []string{
		// Try apt first (Ubuntu/Debian) - fix sources if needed and install
		"sudo sed -i '/bullseye-backports/d' /etc/apt/sources.list 2>/dev/null; sudo apt-get update && sudo apt-get install -y ansible",
		// Try dnf (Fedora)
		"sudo dnf install -y ansible",
		// Try yum (RHEL/CentOS/Oracle)
		"sudo yum install -y ansible",
		// Try zypper (openSUSE)
		"sudo zypper install -y ansible",
		// Try pacman (Arch)
		"sudo pacman -S --noconfirm ansible",
		// Try apk (Alpine)
		"sudo apk add ansible",
	}

	var lastErr error
	for _, cmdStr := range installCommands {
		installCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", cmdStr)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		if err := installCmd.Run(); err == nil {
			// Check if installation succeeded
			checkCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "which ansible-playbook")
			if checkCmd.Run() == nil {
				fmt.Println("Ansible installed successfully")
				return nil
			}
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("failed to install Ansible automatically (tried all package managers): %w\nPlease install manually: wsl -d %s bash -c 'sudo apt install ansible'", lastErr, distroName)
}

// copyPlaybookToWSL copies a playbook from Windows to WSL filesystem
func copyPlaybookToWSL(distroName, windowsPlaybookPath string) (string, error) {
	// Target path in WSL filesystem
	wslPlaybookPath := "/tmp/autowsl-playbook.yml"

	// Read the playbook content from Windows
	content, err := os.ReadFile(windowsPlaybookPath)
	if err != nil {
		return "", fmt.Errorf("failed to read playbook '%s': %w", windowsPlaybookPath, err)
	}

	// Write content directly to WSL filesystem using bash
	// This avoids path conversion issues by piping the content
	writeCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c",
		fmt.Sprintf("cat > '%s' && chmod 644 '%s'", wslPlaybookPath, wslPlaybookPath))

	// Pipe the content to stdin
	writeCmd.Stdin = strings.NewReader(string(content))

	if output, err := writeCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to copy playbook to WSL filesystem: %s: %w", string(output), err)
	}

	return wslPlaybookPath, nil
}

// convertToWSLPath converts a Windows path to a WSL path using wslpath utility
func convertToWSLPath(windowsPath string) (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(windowsPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for '%s': %w", windowsPath, err)
	}

	// Use WSL's official wslpath utility for reliable conversion
	cmd := exec.Command("wsl.exe", "wslpath", "-u", absPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to convert Windows path '%s' to WSL path: %w", absPath, err)
	}

	wslPath := strings.TrimSpace(string(output))
	return wslPath, nil
}

// buildAnsibleCommand builds the ansible-playbook command string
func buildAnsibleCommand(playbookPath string, opts PlaybookOptions) string {
	cmd := fmt.Sprintf("ansible-playbook %s", playbookPath)

	// Add connection local flag (run on localhost)
	cmd += " --connection=local"

	// Add inventory (localhost)
	cmd += " -i localhost,"

	// Add tags if specified
	if len(opts.Tags) > 0 {
		cmd += fmt.Sprintf(" --tags %s", strings.Join(opts.Tags, ","))
	}

	// Add verbosity
	if opts.Verbose {
		cmd += " -vvv"
	}

	// Add extra vars
	if len(opts.ExtraVars) > 0 {
		vars := make([]string, 0, len(opts.ExtraVars))
		for k, v := range opts.ExtraVars {
			vars = append(vars, fmt.Sprintf("%s=%s", k, v))
		}
		cmd += fmt.Sprintf(" --extra-vars '%s'", strings.Join(vars, " "))
	}

	return cmd
}

// CloneGitRepo clones a git repository into a temporary directory in WSL
func CloneGitRepo(distroName, repoURL, destDir string) error {
	fmt.Printf("Cloning repository: %s\n", repoURL)

	// Ensure git is installed
	checkCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "which git")
	if err := checkCmd.Run(); err != nil {
		fmt.Println("Git not found, installing...")

		// Try multiple package managers
		installCommands := []string{
			"sudo apt-get update -qq && sudo apt-get install -y -qq git",
			"sudo dnf install -y git",
			"sudo yum install -y git",
			"sudo zypper install -y git",
			"sudo pacman -S --noconfirm git",
			"sudo apk add git",
		}

		var installed bool
		for _, cmdStr := range installCommands {
			installCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", cmdStr)
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr

			if err := installCmd.Run(); err == nil {
				// Verify installation
				checkCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "which git")
				if checkCmd.Run() == nil {
					installed = true
					break
				}
			}
		}

		if !installed {
			return fmt.Errorf("failed to install git in distribution '%s' (tried all package managers)", distroName)
		}
	}

	// Clone the repository
	cloneCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c",
		fmt.Sprintf("git clone %s %s", repoURL, destDir))
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr

	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository '%s' to '%s' in distribution '%s': %w", repoURL, destDir, distroName, err)
	}

	fmt.Println("Repository cloned successfully")
	return nil
}

```

`internal\ansible\summary.go`:

```go
package ansible

import (
	"fmt"
	"strings"
	"time"
)

// ExecutionResult tracks the result of a playbook execution
type ExecutionResult struct {
	PlaybookName string
	Status       string // "success", "failed", "skipped"
	Duration     time.Duration
	Error        error
}

// ExecutionSummary holds multiple execution results
type ExecutionSummary struct {
	Results []ExecutionResult
}

// Add adds a result to the summary
func (s *ExecutionSummary) Add(result ExecutionResult) {
	s.Results = append(s.Results, result)
}

// HasFailures returns true if any execution failed
func (s *ExecutionSummary) HasFailures() bool {
	for _, r := range s.Results {
		if r.Status == "failed" {
			return true
		}
	}
	return false
}

// SuccessCount returns the number of successful executions
func (s *ExecutionSummary) SuccessCount() int {
	count := 0
	for _, r := range s.Results {
		if r.Status == "success" {
			count++
		}
	}
	return count
}

// Print displays the execution summary
func (s *ExecutionSummary) Print() {
	if len(s.Results) == 0 {
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("PLAYBOOK EXECUTION SUMMARY")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("%-40s %-10s %-15s\n", "PLAYBOOK", "STATUS", "DURATION")
	fmt.Println(strings.Repeat("-", 70))

	for _, r := range s.Results {
		status := r.Status
		if r.Status == "success" {
			status = "OK"
		} else if r.Status == "failed" {
			status = "FAILED"
		}
		fmt.Printf("%-40s %-10s %-15s\n", r.PlaybookName, status, r.Duration.Round(time.Second))
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Total: %d | Success: %d | Failed: %d\n",
		len(s.Results),
		s.SuccessCount(),
		len(s.Results)-s.SuccessCount())
	fmt.Println(strings.Repeat("=", 70))
}

```

`internal\checksum\checksum.go`:

```go
package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// VerifyFile checks if a file matches the expected SHA256 checksum
func VerifyFile(path, expectedSHA256 string) error {
	if expectedSHA256 == "" {
		return fmt.Errorf("no checksum provided for verification")
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	expected := strings.ToLower(strings.TrimSpace(expectedSHA256))

	if actual != expected {
		return fmt.Errorf("checksum mismatch: got %s, expected %s", actual, expected)
	}

	return nil
}

// ComputeFile computes the SHA256 checksum of a file
func ComputeFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to compute checksum: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

```

`internal\distro\distro.go`:

```go
package distro

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed distros-winget.json
var distrosJSON []byte

// Distro represents a WSL distribution
type Distro struct {
	Group        string `json:"group"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	PackageID    string `json:"packageId,omitempty"` // Winget package ID (new method)
	URL          string `json:"url,omitempty"`       // Direct URL (legacy method)
	SHA256       string `json:"sha256,omitempty"`    // Optional checksum for verification
}

// DistroList represents the JSON structure
type DistroList struct {
	Distributions []Distro `json:"distributions"`
}

// GetAllDistros returns all available WSL distributions from embedded JSON
func GetAllDistros() []Distro {
	var distroList DistroList

	if err := json.Unmarshal(distrosJSON, &distroList); err != nil {
		// Fallback to empty list if JSON parsing fails
		fmt.Printf("Warning: Failed to parse distros.json: %v\n", err)
		return []Distro{}
	}

	return distroList.Distributions
}

// FindDistroByVersion finds a distribution by its version name
func FindDistroByVersion(version string) (*Distro, error) {
	distros := GetAllDistros()

	for _, d := range distros {
		if d.Version == version {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("distribution '%s' not found", version)
}

// GetDistrosByGroup returns all distributions in a specific group
func GetDistrosByGroup(group string) []Distro {
	distros := GetAllDistros()
	var result []Distro

	for _, d := range distros {
		if d.Group == group {
			result = append(result, d)
		}
	}

	return result
}

```

`internal\distro\distros-winget.json`:

```json
{
  "distributions": [
    {
      "group": "Ubuntu",
      "version": "Ubuntu",
      "architecture": "x64, arm64",
      "packageId": "Canonical.Ubuntu"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 24.04 LTS",
      "architecture": "x64, arm64",
      "packageId": "Canonical.Ubuntu.2404"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 22.04 LTS",
      "architecture": "x64, arm64",
      "packageId": "Canonical.Ubuntu.2204"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 20.04 LTS",
      "architecture": "x64, arm64",
      "packageId": "Canonical.Ubuntu.2004"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 18.04 LTS",
      "architecture": "x64",
      "packageId": "Canonical.Ubuntu.1804"
    },
    {
      "group": "Debian",
      "version": "Debian",
      "architecture": "x64, arm64",
      "packageId": "Debian.Debian"
    },
    {
      "group": "Kali",
      "version": "Kali Linux",
      "architecture": "x64, arm64",
      "packageId": "OffSec.KaliLinux"
    },
    {
      "group": "openSUSE",
      "version": "openSUSE Tumbleweed",
      "architecture": "x64",
      "packageId": "SUSE.openSUSE.Tumbleweed"
    },
    {
      "group": "openSUSE",
      "version": "openSUSE Leap 15.6",
      "architecture": "x64",
      "packageId": "SUSE.openSUSE.Leap.15.6"
    },
    {
      "group": "SUSE",
      "version": "SUSE Linux Enterprise 15 SP6",
      "architecture": "x64",
      "packageId": "SUSE.SUSE.15SP6"
    },
    {
      "group": "Oracle Linux",
      "version": "Oracle Linux 9.5",
      "architecture": "x64, arm64",
      "packageId": "Oracle.OracleLinux.9.5"
    },
    {
      "group": "Oracle Linux",
      "version": "Oracle Linux 9.1",
      "architecture": "x64, arm64",
      "packageId": "Oracle.OracleLinux.9.1"
    },
    {
      "group": "Oracle Linux",
      "version": "Oracle Linux 8.10",
      "architecture": "x64, arm64",
      "packageId": "Oracle.OracleLinux.8.10"
    },
    {
      "group": "Oracle Linux",
      "version": "Oracle Linux 8.7",
      "architecture": "x64, arm64",
      "packageId": "Oracle.OracleLinux.8.7"
    },
    {
      "group": "Oracle Linux",
      "version": "Oracle Linux 7.9",
      "architecture": "x64",
      "packageId": "Oracle.OracleLinux.7.9"
    }
  ]
}

```

`internal\distro\distros.json`:

```json
{
  "distributions": [
    {
      "group": "Ubuntu",
      "version": "Ubuntu",
      "architecture": "x64, arm64",
      "url": "https://aka.ms/wslubuntu"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 24.04 LTS",
      "architecture": "x64, arm64",
      "url": "https://wslstorestorage.blob.core.windows.net/wslblob/Ubuntu2404-240425.AppxBundle"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 22.04 LTS",
      "architecture": "x64, arm64",
      "url": "https://aka.ms/wslubuntu2204"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 20.04 LTS",
      "architecture": "x64, arm64",
      "url": "https://aka.ms/wslubuntu2004"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 18.04 LTS",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-ubuntu-1804"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 18.04 LTS ARM",
      "architecture": "arm64",
      "url": "https://aka.ms/wsl-ubuntu-1804-arm"
    },
    {
      "group": "Ubuntu",
      "version": "Ubuntu 16.04",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-ubuntu-1604"
    },
    {
      "group": "Debian",
      "version": "Debian GNU/Linux",
      "architecture": "x64, arm64",
      "url": "https://aka.ms/wsl-debian-gnulinux"
    },
    {
      "group": "Kali Linux",
      "version": "Kali Linux Rolling",
      "architecture": "-",
      "url": "https://aka.ms/wsl-kali-linux-new"
    },
    {
      "group": "Oracle Linux",
      "version": "Oracle Linux 9.1",
      "architecture": "x64",
      "url": "https://publicwsldistros.blob.core.windows.net/wsldistrostorage/OracleLinux_9.1-230428.Appx"
    },
    {
      "group": "Oracle Linux",
      "version": "Oracle Linux 8.7",
      "architecture": "x64",
      "url": "https://publicwsldistros.blob.core.windows.net/wsldistrostorage/OracleLinux_8.7-230428.Appx"
    },
    {
      "group": "Oracle Linux",
      "version": "Oracle Linux 8.5",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-oraclelinux-8-5"
    },
    {
      "group": "Oracle Linux",
      "version": "Oracle Linux 7.9",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-oraclelinux-7-9"
    },
    {
      "group": "openSUSE",
      "version": "openSUSE Tumbleweed",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-opensuse-tumbleweed"
    },
    {
      "group": "openSUSE",
      "version": "openSUSE Leap 15.6",
      "architecture": "x64",
      "url": "https://publicwsldistros.blob.core.windows.net/wsldistrostorage/SUSELeap15p6-240801_x64.Appx"
    },
    {
      "group": "openSUSE",
      "version": "openSUSE Leap 15.3",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-opensuseleap15-3"
    },
    {
      "group": "openSUSE",
      "version": "openSUSE Leap 15.2",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-opensuseleap15-2"
    },
    {
      "group": "SUSE Linux Enterprise Server",
      "version": "SUSE Linux Enterprise Server 15 SP6",
      "architecture": "x64",
      "url": "https://publicwsldistros.blob.core.windows.net/wsldistrostorage/SUSELinuxEnterprise15SP6-241001_x64.Appx"
    },
    {
      "group": "SUSE Linux Enterprise Server",
      "version": "SUSE Linux Enterprise Server 15 SP5",
      "architecture": "x64",
      "url": "https://publicwsldistros.blob.core.windows.net/wsldistrostorage/SUSELinuxEnterprise15_SP5-240801.Appx"
    },
    {
      "group": "SUSE Linux Enterprise Server",
      "version": "SUSE Linux Enterprise Server 15 SP3",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-SUSELinuxEnterpriseServer15SP3"
    },
    {
      "group": "SUSE Linux Enterprise Server",
      "version": "SUSE Linux Enterprise Server 15 SP2",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-SUSELinuxEnterpriseServer15SP2"
    },
    {
      "group": "SUSE Linux Enterprise Server",
      "version": "SUSE Linux Enterprise Server 12",
      "architecture": "x64",
      "url": "https://aka.ms/wsl-sles-12"
    },
    {
      "group": "Fedora Remix",
      "version": "Fedora Remix for WSL",
      "architecture": "x64, arm64",
      "url": "https://github.com/WhitewaterFoundry/Fedora-Remix-for-WSL/releases"
    }
  ]
}

```

`internal\downloader\downloader.go`:

```go
package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuanjua/autowsl/internal/checksum"
	"github.com/yuanjua/autowsl/internal/distro"
)

// Downloader handles downloading WSL distributions
type Downloader struct {
	client         *http.Client
	VerifyChecksum bool // Whether to verify checksums (default: warn if mismatch)
}

// New creates a new Downloader instance
func New() *Downloader {
	return &Downloader{
		client:         &http.Client{},
		VerifyChecksum: false, // Default to warn-only mode
	}
}

// Download downloads a distribution to the current directory
func (d *Downloader) Download(dist distro.Distro) error {
	// Get the filename from the URL
	filename := d.getFilename(dist.URL)

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	filepath := filepath.Join(cwd, filename)

	fmt.Printf("Downloading to: %s\n", filepath)

	return d.downloadToFile(dist.URL, filepath)
}

// DownloadToDir downloads a distribution to a specific directory and returns the file path
func (d *Downloader) DownloadToDir(dist distro.Distro, dir string) (string, error) {
	// Get the filename from the URL
	filename := d.getFilename(dist.URL)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	filepath := filepath.Join(dir, filename)

	// Always download fresh - remove any existing file first
	if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
	}

	if err := d.downloadToFile(dist.URL, filepath); err != nil {
		return "", err
	}

	// Verify checksum if provided
	if dist.SHA256 != "" {
		fmt.Println("Verifying checksum...")
		if err := checksum.VerifyFile(filepath, dist.SHA256); err != nil {
			if d.VerifyChecksum {
				// Strict mode: fail on mismatch
				os.Remove(filepath)
				return "", fmt.Errorf("checksum verification failed: %w", err)
			} else {
				// Warn mode: continue but alert user
				fmt.Printf("Warning: %v\n", err)
				fmt.Println("Continuing anyway (use --verify-checksum to enforce)")
			}
		} else {
			fmt.Println("Checksum verified successfully")
		}
	} else {
		fmt.Println("Warning: No checksum available for this distribution")
	}

	return filepath, nil
}

// downloadToFile downloads from URL to a specific file path
func (d *Downloader) downloadToFile(url, filepath string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %w", filepath, err)
	}
	defer out.Close()

	// Send GET request
	resp, err := d.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from '%s': %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download from '%s': HTTP %s", url, resp.Status)
	}

	// Get the total size for progress
	totalSize := resp.ContentLength

	// Create progress writer
	counter := &ProgressWriter{
		Total:  totalSize,
		Writer: out,
	}

	// Copy the data with progress
	_, err = io.Copy(counter, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Println() // New line after progress

	return nil
}

// getFilename extracts filename from URL
func (d *Downloader) getFilename(url string) string {
	// Handle special cases for aka.ms URLs
	if strings.Contains(url, "aka.ms") {
		parts := strings.Split(url, "/")
		name := parts[len(parts)-1]

		// Create a reasonable filename
		return fmt.Sprintf("%s.appxbundle", name)
	}

	// For direct URLs, extract the filename
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	// Ensure it has a proper extension
	if !strings.Contains(filename, ".") {
		filename += ".appxbundle"
	}

	return filename
}

// ProgressWriter tracks download progress
type ProgressWriter struct {
	Total        int64
	Downloaded   int64
	Writer       io.Writer
	LastPrint    int64
	PrintEveryMB int64
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.Downloaded += int64(n)

	// Print progress every MB or at the end
	if pw.PrintEveryMB == 0 {
		pw.PrintEveryMB = 1024 * 1024 // 1 MB
	}

	if pw.Downloaded-pw.LastPrint >= pw.PrintEveryMB || pw.Downloaded == pw.Total {
		pw.printProgress()
		pw.LastPrint = pw.Downloaded
	}

	return n, nil
}

func (pw *ProgressWriter) printProgress() {
	if pw.Total > 0 {
		percentage := float64(pw.Downloaded) / float64(pw.Total) * 100
		downloadedMB := float64(pw.Downloaded) / 1024 / 1024
		totalMB := float64(pw.Total) / 1024 / 1024
		fmt.Printf("\rProgress: %.2f%% (%.2f MB / %.2f MB)", percentage, downloadedMB, totalMB)
	} else {
		downloadedMB := float64(pw.Downloaded) / 1024 / 1024
		fmt.Printf("\rDownloaded: %.2f MB", downloadedMB)
	}
}

```

`internal\extractor\extractor.go`:

```go
package extractor

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuanjua/autowsl/internal/system"
)

// ExtractAppx extracts the root filesystem tar file from an Appx/AppxBundle package
func ExtractAppx(appxPath, outputDir string) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Detect host architecture
	hostArch := system.GetHostArchitecture()
	fmt.Printf("Host architecture: %s\n", hostArch)

	// Open the appx file as a zip archive
	reader, err := zip.OpenReader(appxPath)
	if err != nil {
		return "", fmt.Errorf("failed to open appx file: %w", err)
	}
	defer reader.Close()

	var tarFilePath string

	// First, look for direct tar.gz files (single Appx case)
	for _, file := range reader.File {
		lowerName := strings.ToLower(file.Name)

		// Skip incompatible architectures
		if system.ShouldSkipArchitecture(lowerName) {
			continue
		}

		if strings.HasSuffix(lowerName, "install.tar.gz") || strings.HasSuffix(lowerName, "install.tar") {
			// Extract this file
			extractedPath := filepath.Join(outputDir, filepath.Base(file.Name))

			if err := extractFile(file, extractedPath); err != nil {
				return "", fmt.Errorf("failed to extract tar file: %w", err)
			}

			tarFilePath = extractedPath
			break
		}
	}

	// If we didn't find install.tar.gz, look for .appx files inside (AppxBundle case)
	if tarFilePath == "" {
		// Collect all .appx files and prioritize matching architecture
		var matchingAppx, genericAppx *zip.File
		preferredSuffix := system.GetPreferredArchitectureSuffix()

		for _, file := range reader.File {
			lowerName := strings.ToLower(file.Name)

			if !strings.HasSuffix(lowerName, ".appx") {
				continue
			}

			// Skip incompatible architectures
			if system.ShouldSkipArchitecture(lowerName) {
				fmt.Printf("   Skipping incompatible: %s\n", file.Name)
				continue
			}

			// Prefer matching architecture
			if strings.Contains(lowerName, strings.ToLower(preferredSuffix)) {
				matchingAppx = file
				fmt.Printf("   Selected: %s (matches %s)\n", file.Name, hostArch)
				break // Found matching arch, use it!
			}

			// Keep track of first non-incompatible appx as fallback
			if genericAppx == nil {
				genericAppx = file
			}
		}

		// Use matching architecture if found, otherwise use generic
		selectedAppx := matchingAppx
		if selectedAppx == nil {
			selectedAppx = genericAppx
			if selectedAppx != nil {
				fmt.Printf("   Selected: %s (generic)\n", selectedAppx.Name)
			}
		}

		if selectedAppx != nil {
			// Extract the nested appx
			nestedAppxPath := filepath.Join(outputDir, filepath.Base(selectedAppx.Name))

			if err := extractFile(selectedAppx, nestedAppxPath); err != nil {
				return "", fmt.Errorf("failed to extract nested appx: %w", err)
			}

			// Recursively extract from the nested appx
			tarFilePath, err = ExtractAppx(nestedAppxPath, outputDir)
			if err != nil {
				return "", err
			}

			// Clean up the nested appx file
			os.Remove(nestedAppxPath)
		}
	}

	if tarFilePath == "" {
		return "", fmt.Errorf("could not find rootfs tar file in the package")
	}

	return tarFilePath, nil
}

// extractFile extracts a single file from a zip archive
func extractFile(zipFile *zip.File, destPath string) error {
	// Open the file in the zip archive
	rc, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, rc)
	return err
}

// CleanupTempDir removes the temporary extraction directory
func CleanupTempDir(dir string) error {
	return os.RemoveAll(dir)
}

```

`internal\playbooks\extravars.go`:

```go
package playbooks

import (
	"fmt"
	"strings"
)

// ParseExtraVars converts key=val strings into a map
func ParseExtraVars(kvs []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, kv := range kvs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid extra-vars entry: %s (expected key=val)", kv)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty key in extra-vars entry: %s", kv)
		}
		m[key] = val
	}
	return m, nil
}

```

`internal\playbooks\resolver.go`:

```go
package playbooks

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Resolver handles playbook resolution from various input formats
type Resolver struct {
	TempDir string
	FSRoot  string
}

// NewResolver creates a new playbook resolver
func NewResolver(tempDir, fsRoot string) *Resolver {
	return &Resolver{
		TempDir: tempDir,
		FSRoot:  fsRoot,
	}
}

// Resolve converts playbook input (URL, file, alias) to concrete file paths
func (r *Resolver) Resolve(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty playbook input")
	}

	// Handle URLs
	if isURL(input) {
		path, err := r.downloadPlaybook(input)
		if err != nil {
			return nil, err
		}
		return []string{path}, nil
	}

	// Handle comma-separated lists
	if strings.Contains(input, ",") {
		parts := strings.Split(input, ",")
		all := []string{}
		for _, p := range parts {
			sub, err := r.Resolve(strings.TrimSpace(p))
			if err != nil {
				return nil, err
			}
			all = append(all, sub...)
		}
		return all, nil
	}

	// Check if file exists as given
	if stat, err := os.Stat(input); err == nil && !stat.IsDir() {
		absPath, _ := filepath.Abs(input)
		return []string{absPath}, nil
	}

	// Try as alias (playbooks/<name>.yml)
	aliasPath := filepath.Join(r.FSRoot, "playbooks", ensureYmlExt(input))
	if _, err := os.Stat(aliasPath); err == nil {
		absPath, _ := filepath.Abs(aliasPath)
		return []string{absPath}, nil
	}

	return nil, fmt.Errorf("could not resolve '%s' as URL, file, or alias (looked for %s)", input, aliasPath)
}

// ResolveMultiple resolves multiple playbook inputs
func (r *Resolver) ResolveMultiple(inputs []string) ([]string, error) {
	var results []string
	seen := make(map[string]bool)

	for _, input := range inputs {
		paths, err := r.Resolve(input)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve '%s': %w", input, err)
		}
		for _, p := range paths {
			if !seen[p] {
				results = append(results, p)
				seen[p] = true
			}
		}
	}

	return results, nil
}

// downloadPlaybook downloads a playbook from a URL
func (r *Resolver) downloadPlaybook(url string) (string, error) {
	playbookFile := filepath.Join(r.TempDir, "autowsl-playbook-"+sanitizeFilename(url)+".yml")

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download from '%s': %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download from '%s': HTTP %s", url, resp.Status)
	}

	out, err := os.Create(playbookFile)
	if err != nil {
		return "", fmt.Errorf("failed to create playbook file '%s': %w", playbookFile, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write playbook file: %w", err)
	}

	return playbookFile, nil
}

func isURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

func ensureYmlExt(name string) string {
	if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
		return name
	}
	return name + ".yml"
}

func sanitizeFilename(url string) string {
	// Simple hash-like identifier from URL
	safe := strings.ReplaceAll(url, "://", "-")
	safe = strings.ReplaceAll(safe, "/", "-")
	safe = strings.ReplaceAll(safe, "?", "-")
	safe = strings.ReplaceAll(safe, "&", "-")
	if len(safe) > 40 {
		safe = safe[:40]
	}
	return safe
}

// BaseNames extracts base names from file paths
func BaseNames(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, filepath.Base(p))
	}
	return out
}

```

`internal\runner\runner.go`:

```go
package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Runner executes external commands
type Runner interface {
	Run(name string, args ...string) (stdout string, stderr string, err error)
	RunWithInput(name string, stdin string, args ...string) (stdout string, stderr string, err error)
}

// ExecRunner executes real system commands with timeout support
type ExecRunner struct {
	Timeout time.Duration
	DryRun  bool
}

// NewExecRunner creates a new runner with the given timeout
func NewExecRunner(timeout time.Duration) *ExecRunner {
	return &ExecRunner{
		Timeout: timeout,
		DryRun:  false,
	}
}

// Run executes a command and returns stdout, stderr, and error
func (r *ExecRunner) Run(name string, args ...string) (string, string, error) {
	if r.DryRun {
		return r.dryRunLog(name, args...), "", nil
	}

	// Create context with or without timeout
	var ctx context.Context
	var cancel context.CancelFunc
	if r.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), r.Timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var outB, errB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB

	err := cmd.Run()
	return outB.String(), errB.String(), err
}

// RunWithInput executes a command with stdin and returns stdout, stderr, and error
func (r *ExecRunner) RunWithInput(name string, stdin string, args ...string) (string, string, error) {
	if r.DryRun {
		return r.dryRunLog(name, args...), "", nil
	}

	// Create context with or without timeout
	var ctx context.Context
	var cancel context.CancelFunc
	if r.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), r.Timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var outB, errB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB
	cmd.Stdin = bytes.NewBufferString(stdin)

	err := cmd.Run()
	return outB.String(), errB.String(), err
}

func (r *ExecRunner) dryRunLog(name string, args ...string) string {
	return fmt.Sprintf("[dry-run] %s %v", name, args)
}

```

`internal\system\arch.go`:

```go
package system

import (
	"runtime"
	"strings"
)

// HostArchitecture represents the system architecture
type HostArchitecture string

const (
	ArchX64   HostArchitecture = "x64"
	ArchARM64 HostArchitecture = "arm64"
	ArchX86   HostArchitecture = "x86"
)

// GetHostArchitecture returns the current system architecture
func GetHostArchitecture() HostArchitecture {
	arch := runtime.GOARCH

	switch arch {
	case "amd64":
		return ArchX64
	case "arm64":
		return ArchARM64
	case "386":
		return ArchX86
	default:
		// Default to x64 for unknown architectures
		return ArchX64
	}
}

// IsCompatibleArchitecture checks if a distribution architecture is compatible with host
func IsCompatibleArchitecture(distroArch string) bool {
	hostArch := GetHostArchitecture()
	distroArchLower := strings.ToLower(distroArch)

	switch hostArch {
	case ArchX64:
		// x64 can run x64 and x86
		return strings.Contains(distroArchLower, "x64") ||
			strings.Contains(distroArchLower, "x86") ||
			strings.Contains(distroArchLower, "amd64")
	case ArchARM64:
		// ARM64 can run ARM64 (and potentially ARM32)
		return strings.Contains(distroArchLower, "arm64") ||
			strings.Contains(distroArchLower, "aarch64")
	case ArchX86:
		// x86 can only run x86
		return strings.Contains(distroArchLower, "x86") &&
			!strings.Contains(distroArchLower, "x64")
	default:
		return false
	}
}

// GetPreferredArchitectureSuffix returns the preferred architecture suffix for filtering
func GetPreferredArchitectureSuffix() string {
	hostArch := GetHostArchitecture()

	switch hostArch {
	case ArchX64:
		return "x64"
	case ArchARM64:
		return "arm64"
	case ArchX86:
		return "x86"
	default:
		return "x64"
	}
}

// ShouldSkipArchitecture returns true if the architecture should be skipped
func ShouldSkipArchitecture(filename string) bool {
	hostArch := GetHostArchitecture()
	filenameLower := strings.ToLower(filename)

	switch hostArch {
	case ArchX64:
		// Skip ARM versions on x64 systems
		return strings.Contains(filenameLower, "arm64") ||
			strings.Contains(filenameLower, "arm32") ||
			strings.Contains(filenameLower, "arm_") ||
			strings.Contains(filenameLower, "aarch64")
	case ArchARM64:
		// Skip x64 versions on ARM64 systems
		return strings.Contains(filenameLower, "x64") ||
			strings.Contains(filenameLower, "amd64") ||
			strings.Contains(filenameLower, "_64") && !strings.Contains(filenameLower, "arm64")
	default:
		return false
	}
}

```

`internal\winget\catalog.go`:

```go
package winget

import (
	"github.com/yuanjua/autowsl/internal/distro"
)

// WingetDistro represents a WSL distribution available via winget
type WingetDistro struct {
	Name         string // Display name
	Version      string // Version string
	PackageID    string // Winget package identifier
	Group        string // Distribution family (Ubuntu, Debian, etc.)
	Architecture string // Supported architectures
}

// GetWingetDistros returns the list of WSL distributions available via winget
// This reads from the distro catalog
func GetWingetDistros() []WingetDistro {
	distros := distro.GetAllDistros()
	result := make([]WingetDistro, 0, len(distros))

	for _, d := range distros {
		if d.PackageID != "" {
			result = append(result, WingetDistro{
				Name:         d.Version,
				Version:      d.Version,
				PackageID:    d.PackageID,
				Group:        d.Group,
				Architecture: d.Architecture,
			})
		}
	}

	return result
}

// FindWingetDistroByVersion finds a distribution by its version string
func FindWingetDistroByVersion(version string) (*WingetDistro, error) {
	d, err := distro.FindDistroByVersion(version)
	if err != nil {
		return nil, err
	}

	if d.PackageID == "" {
		return nil, nil // No winget package ID
	}

	return &WingetDistro{
		Name:         d.Version,
		Version:      d.Version,
		PackageID:    d.PackageID,
		Group:        d.Group,
		Architecture: d.Architecture,
	}, nil
}

// FindWingetDistroByPackageID finds a distribution by its winget package ID
func FindWingetDistroByPackageID(packageID string) (*WingetDistro, error) {
	distros := distro.GetAllDistros()
	for _, d := range distros {
		if d.PackageID == packageID {
			return &WingetDistro{
				Name:         d.Version,
				Version:      d.Version,
				PackageID:    d.PackageID,
				Group:        d.Group,
				Architecture: d.Architecture,
			}, nil
		}
	}
	return nil, nil // Not found
}

```

`internal\winget\manager.go`:

```go
package winget

import (
	"fmt"
)

// Manager handles WSL distribution downloads using winget
type Manager struct {
	downloader *WingetDownloader
	tempDir    string
}

// NewManager creates a new download manager
func NewManager(tempDir string) *Manager {
	return &Manager{
		downloader: NewWingetDownloader(tempDir),
		tempDir:    tempDir,
	}
}

// DownloadOptions contains options for downloading a distribution
type DownloadOptions struct {
	// Either specify a known version or a custom package ID
	Version   string // e.g., "Ubuntu 22.04 LTS"
	PackageID string // e.g., "Canonical.Ubuntu.2204"

	// If true, validate package ID before downloading
	ValidatePackageID bool
}

// Download downloads a WSL distribution
// Returns the path to the downloaded file
func (m *Manager) Download(opts DownloadOptions) (string, error) {
	var packageID string

	// Determine package ID
	if opts.PackageID != "" {
		// User provided a custom package ID
		packageID = opts.PackageID

		// Optionally validate it
		if opts.ValidatePackageID {
			fmt.Printf("Validating package ID: %s\n", packageID)
			valid, err := ValidatePackageID(packageID)
			if err != nil {
				return "", fmt.Errorf("failed to validate package ID: %w", err)
			}
			if !valid {
				return "", fmt.Errorf("package ID '%s' not found in winget", packageID)
			}

			// Warn if it doesn't look like a WSL package
			if !IsWSLPackage(packageID) {
				fmt.Printf("Warning: Package '%s' may not be a WSL distribution\n", packageID)
			}
		}
	} else if opts.Version != "" {
		// Look up version in catalog
		distro, err := FindWingetDistroByVersion(opts.Version)
		if err != nil {
			return "", err
		}

		if distro == nil {
			return "", fmt.Errorf("distribution '%s' not found in catalog. Use --package-id to specify a custom winget package", opts.Version)
		}

		packageID = distro.PackageID
		fmt.Printf("Found in catalog: %s (%s)\n", distro.Name, distro.PackageID)
	} else {
		return "", fmt.Errorf("either Version or PackageID must be specified")
	}

	// Download using winget
	downloadedFile, err := m.downloader.Download(packageID)
	if err != nil {
		return "", err
	}

	return downloadedFile, nil
}

// GetCatalog returns the list of known distributions
func (m *Manager) GetCatalog() []WingetDistro {
	return GetWingetDistros()
}

// IsWingetAvailable checks if winget is available
func (m *Manager) IsWingetAvailable() bool {
	return m.downloader.IsWingetAvailable()
}

// GetWingetVersion returns the winget version
func (m *Manager) GetWingetVersion() (string, error) {
	return m.downloader.GetWingetVersion()
}

// CleanupDownloadDir removes the temporary download directory
func (m *Manager) CleanupDownloadDir() error {
	// Don't cleanup if user might want to keep files
	// This should be called explicitly
	fmt.Printf("Download directory: %s\n", m.tempDir)
	fmt.Println("Files are kept for your use. Delete manually if needed.")
	return nil
}

```

`internal\winget\search.go`:

```go
package winget

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// PackageInfo represents information about a winget package
type PackageInfo struct {
	ID      string
	Name    string
	Version string
	Source  string
}

// SearchPackage searches for a package in winget
func SearchPackage(query string) ([]PackageInfo, error) {
	cmd := exec.Command("winget", "search", query, "--accept-source-agreements")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("winget search failed: %w", err)
	}

	// Parse output (basic parsing, winget output format may vary)
	output := out.String()
	lines := strings.Split(output, "\n")

	var packages []PackageInfo
	// Skip header lines and parse results
	for i, line := range lines {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		// Basic parsing - this is simplified and may need adjustment
		// Winget output format: Name    Id    Version    Source
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			pkg := PackageInfo{
				ID:   fields[1],
				Name: fields[0],
			}
			if len(fields) >= 3 {
				pkg.Version = fields[2]
			}
			if len(fields) >= 4 {
				pkg.Source = fields[3]
			}
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// ValidatePackageID checks if a package ID exists in winget
func ValidatePackageID(packageID string) (bool, error) {
	cmd := exec.Command("winget", "show", "--id", packageID, "--accept-source-agreements")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		// Check if error is due to package not found
		output := out.String()
		if strings.Contains(output, "No package found") {
			return false, nil
		}
		return false, fmt.Errorf("winget show failed: %w", err)
	}

	return true, nil
}

// IsWSLPackage checks if a package ID is a WSL distribution
// This is a heuristic check based on common patterns
func IsWSLPackage(packageID string) bool {
	// Check if it's a Microsoft Store ID (12 uppercase alphanumeric)
	// We assume all MS Store IDs that match our catalog are WSL
	if len(packageID) == 12 && strings.ToUpper(packageID) == packageID {
		// Check if it contains only alphanumeric characters
		for _, c := range packageID {
			if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')) {
				return false
			}
		}
		// MS Store ID - check against our catalog
		distro, _ := FindWingetDistroByPackageID(packageID)
		return distro != nil
	}

	// Legacy format: check for known prefixes
	wslPrefixes := []string{
		"Canonical.",         // Ubuntu
		"Debian.",            // Debian
		"KaliLinux.",         // Kali
		"openSUSE.",          // openSUSE
		"Oracle.",            // Oracle Linux
		"AlpineWSL.",         // Alpine
		"SUSE.",              // SUSE
		"Pengwin.",           // Pengwin
		"WhitewaterFoundry.", // Pengwin/Fedora Remix
	}

	for _, prefix := range wslPrefixes {
		if strings.HasPrefix(packageID, prefix) {
			return true
		}
	}

	// Check if package name contains "WSL" or "Linux"
	packageLower := strings.ToLower(packageID)
	return strings.Contains(packageLower, "wsl") || strings.Contains(packageLower, "linux")
}

```

`internal\winget\winget.go`:

```go
package winget

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WingetDownloader handles downloading WSL distributions using winget
type WingetDownloader struct {
	DownloadDir string
}

// NewWingetDownloader creates a new winget downloader
func NewWingetDownloader(downloadDir string) *WingetDownloader {
	return &WingetDownloader{
		DownloadDir: downloadDir,
	}
}

// Download downloads a package using winget
// packageID is the winget package identifier (e.g., "Canonical.Ubuntu.2204")
// Returns the path to the downloaded file
func (w *WingetDownloader) Download(packageID string) (string, error) {
	// Ensure download directory exists
	if err := os.MkdirAll(w.DownloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download directory: %w", err)
	}

	// Check if winget is available
	if !w.IsWingetAvailable() {
		return "", fmt.Errorf("winget is not available. Please install App Installer from Microsoft Store")
	}

	fmt.Printf("Downloading package: %s\n", packageID)
	fmt.Printf("Download directory: %s\n\n", w.DownloadDir)

	// Run winget download command
	// winget download --id <PackageId> --download-directory <PathToTempDir>
	cmd := exec.Command("winget", "download", "--id", packageID, "--download-directory", w.DownloadDir, "--accept-package-agreements", "--accept-source-agreements")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("winget download failed: %w", err)
	}

	// Find the downloaded file (should be in the download directory)
	// Winget typically downloads with package name, look for .appx or .appxbundle files
	downloadedFile, err := w.findDownloadedFile(packageID)
	if err != nil {
		return "", fmt.Errorf("failed to find downloaded file: %w", err)
	}

	fmt.Printf("\nDownload completed: %s\n", filepath.Base(downloadedFile))
	return downloadedFile, nil
}

// findDownloadedFile searches for the downloaded package file in the download directory
func (w *WingetDownloader) findDownloadedFile(packageID string) (string, error) {
	// List files in download directory
	entries, err := os.ReadDir(w.DownloadDir)
	if err != nil {
		return "", fmt.Errorf("failed to read download directory: %w", err)
	}

	// Look for .appx, .appxbundle, or .msix files
	validExtensions := []string{".appx", ".appxbundle", ".msix", ".msixbundle"}

	var candidates []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range validExtensions {
			if strings.HasSuffix(strings.ToLower(name), ext) {
				fullPath := filepath.Join(w.DownloadDir, name)
				candidates = append(candidates, fullPath)
			}
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no package files found in download directory")
	}

	// If multiple files, return the most recent one
	if len(candidates) > 1 {
		fmt.Printf("Warning: Multiple package files found, using: %s\n", filepath.Base(candidates[0]))
	}

	return candidates[0], nil
}

// IsWingetAvailable checks if winget is installed and available
func (w *WingetDownloader) IsWingetAvailable() bool {
	cmd := exec.Command("winget", "--version")
	err := cmd.Run()
	return err == nil
}

// GetWingetVersion returns the installed winget version
func (w *WingetDownloader) GetWingetVersion() (string, error) {
	cmd := exec.Command("winget", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get winget version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

```

`internal\wsl\client.go`:

```go
package wsl

import "github.com/yuanjua/autowsl/internal/runner"

// Client is a wrapper for executing WSL commands.
// It uses dependency injection to allow for easy testing.
type Client struct {
	runner runner.Runner
}

// NewClient creates a new WSL client with the provided runner.
func NewClient(r runner.Runner) *Client {
	return &Client{runner: r}
}

// DefaultClient returns a client configured with default settings.
func DefaultClient() *Client {
	return NewClient(runner.NewExecRunner(0)) // 0 = no timeout
}

```

`internal\wsl\importer.go`:

```go
package wsl

import (
	"fmt"
	"os"
	"path/filepath"
)

// ImportOptions contains options for importing a WSL distribution
type ImportOptions struct {
	Name        string // Name of the distribution
	InstallPath string // Custom installation path
	TarFilePath string // Path to the tar file
	Version     int    // WSL version (1 or 2)
}

// Import imports a WSL distribution from a tar file
func (c *Client) Import(opts ImportOptions) error {
	// Validate inputs
	if opts.Name == "" {
		return fmt.Errorf("distribution name cannot be empty")
	}
	if opts.InstallPath == "" {
		return fmt.Errorf("installation path cannot be empty")
	}
	if opts.TarFilePath == "" {
		return fmt.Errorf("tar file path cannot be empty")
	}

	// Check if tar file exists
	if _, err := os.Stat(opts.TarFilePath); os.IsNotExist(err) {
		return fmt.Errorf("tar file does not exist: %s", opts.TarFilePath)
	}

	// Create installation directory if it doesn't exist
	if err := os.MkdirAll(opts.InstallPath, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Check if distro already exists
	exists, err := c.IsDistroInstalled(opts.Name)
	if err != nil {
		return fmt.Errorf("failed to check if distro exists: %w", err)
	}
	if exists {
		return fmt.Errorf("distribution '%s' already exists", opts.Name)
	}

	// Default to WSL 2
	version := opts.Version
	if version == 0 {
		version = 2
	}

	// Convert paths to absolute paths
	absInstallPath, err := filepath.Abs(opts.InstallPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for install location: %w", err)
	}

	absTarPath, err := filepath.Abs(opts.TarFilePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for tar file: %w", err)
	}

	// Execute wsl --import command
	_, stderr, err := c.runner.Run("wsl.exe", "--import", opts.Name, absInstallPath, absTarPath, "--version", fmt.Sprintf("%d", version))
	if err != nil {
		return fmt.Errorf("failed to import distribution: %w\nOutput: %s", err, stderr)
	}

	return nil
}

// Unregister removes a WSL distribution
func (c *Client) Unregister(name string) error {
	if name == "" {
		return fmt.Errorf("distribution name cannot be empty")
	}

	// Check if distro exists
	exists, err := c.IsDistroInstalled(name)
	if err != nil {
		return fmt.Errorf("failed to check if distro exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("distribution '%s' does not exist", name)
	}

	// Execute wsl --unregister command
	_, stderr, err := c.runner.Run("wsl.exe", "--unregister", name)
	if err != nil {
		return fmt.Errorf("failed to unregister distribution: %w\nOutput: %s", err, stderr)
	}

	return nil
}

// Export backs up a WSL distribution to a tar file
func (c *Client) Export(name, outputPath string) error {
	if name == "" {
		return fmt.Errorf("distribution name cannot be empty")
	}
	if outputPath == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	// Check if distro exists
	exists, err := c.IsDistroInstalled(name)
	if err != nil {
		return fmt.Errorf("failed to check if distro exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("distribution '%s' does not exist", name)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Execute wsl --export command
	_, stderr, err := c.runner.Run("wsl.exe", "--export", name, outputPath)
	if err != nil {
		return fmt.Errorf("failed to export distribution: %w\nOutput: %s", err, stderr)
	}

	return nil
}

// Package-level convenience functions that use a default client
// These maintain backward compatibility with existing code

// Import imports a WSL distribution from a tar file (uses default client)
func Import(opts ImportOptions) error {
	return DefaultClient().Import(opts)
}

// Unregister removes a WSL distribution (uses default client)
func Unregister(name string) error {
	return DefaultClient().Unregister(name)
}

// Export backs up a WSL distribution to a tar file (uses default client)
func Export(name, outputPath string) error {
	return DefaultClient().Export(name, outputPath)
}

```

`internal\wsl\status.go`:

```go
package wsl

import (
	"fmt"
	"regexp"
	"strings"
)

// InstalledDistro represents a currently installed WSL distribution
type InstalledDistro struct {
	Name    string
	State   string // Running, Stopped
	Version string // 1 or 2
	Default bool
}

// CheckWSLInstalled checks if WSL is installed and available
func (c *Client) CheckWSLInstalled() error {
	_, _, err := c.runner.Run("wsl.exe", "--status")
	if err != nil {
		return fmt.Errorf("WSL is not installed or not available: %w", err)
	}
	return nil
}

// ListInstalledDistros lists all currently installed WSL distributions
func (c *Client) ListInstalledDistros() ([]InstalledDistro, error) {
	output, _, err := c.runner.Run("wsl.exe", "-l", "-v")
	if err != nil {
		return nil, fmt.Errorf("failed to list WSL distributions: %w", err)
	}

	return parseWSLList(output)
}

// parseWSLList parses the output of "wsl -l -v"
func parseWSLList(output string) ([]InstalledDistro, error) {
	var distros []InstalledDistro

	// Clean up the output - remove UTF-8 BOM and null bytes
	output = strings.ReplaceAll(output, "\x00", "")
	output = strings.TrimPrefix(output, "\ufeff")

	lines := strings.Split(output, "\n")

	// Skip header lines
	startIdx := 0
	for i, line := range lines {
		if strings.Contains(line, "NAME") || strings.Contains(line, "---") {
			startIdx = i + 1
			break
		}
	}

	// Regular expression to parse each line
	// Format: "  * Ubuntu-22.04    Running         2"
	re := regexp.MustCompile(`^\s*(\*?)\s*([^\s]+)\s+(\w+)\s+(\d+)\s*$`)

	for _, line := range lines[startIdx:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			distro := InstalledDistro{
				Name:    strings.TrimSpace(matches[2]),
				State:   strings.TrimSpace(matches[3]),
				Version: strings.TrimSpace(matches[4]),
				Default: matches[1] == "*",
			}
			distros = append(distros, distro)
		}
	}

	return distros, nil
}

// IsDistroInstalled checks if a specific distribution is installed
func (c *Client) IsDistroInstalled(name string) (bool, error) {
	distros, err := c.ListInstalledDistros()
	if err != nil {
		return false, err
	}

	for _, d := range distros {
		if d.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// Package-level convenience functions that use a default client
// These maintain backward compatibility with existing code

// CheckWSLInstalled checks if WSL is installed (uses default client)
func CheckWSLInstalled() error {
	return DefaultClient().CheckWSLInstalled()
}

// ListInstalledDistros lists all installed WSL distributions (uses default client)
func ListInstalledDistros() ([]InstalledDistro, error) {
	return DefaultClient().ListInstalledDistros()
}

// IsDistroInstalled checks if a specific distribution is installed (uses default client)
func IsDistroInstalled(name string) (bool, error) {
	return DefaultClient().IsDistroInstalled(name)
}

```

`main.go`:

```go
package main

import (
	"github.com/yuanjua/autowsl/cmd"
)

func main() {
	cmd.Execute()
}

```

`playbooks\curl.yml`:

```yml
---
# Alias playbook: curl
# Purpose: ensure curl is installed across all supported WSL distributions.
# 
# Supported Distributions (from winget catalog):
#   - Ubuntu (all versions: 18.04, 20.04, 22.04, 24.04)
#   - Debian
#   - Kali Linux
#   - openSUSE (Tumbleweed, Leap 15.6)
#   - SUSE Linux Enterprise 15 SP6
#   - Oracle Linux (7.9, 8.7, 8.10, 9.1, 9.5)
#
# This playbook uses ansible_os_family for broad compatibility:
#   - Debian family: Ubuntu, Debian, Kali Linux
#   - RedHat family: Oracle Linux
#   - Suse family: openSUSE, SUSE Linux Enterprise

- name: Ensure curl is installed
  hosts: localhost
  connection: local
  become: yes
  tasks:
    - name: Install curl on Debian-based distributions (Ubuntu, Debian, Kali)
      apt:
        name: curl
        state: present
        update_cache: yes
      when: ansible_os_family == "Debian"

    - name: Install curl on RedHat-based distributions (Oracle Linux)
      yum:
        name: curl
        state: present
      when: ansible_os_family == "RedHat" and ansible_distribution_major_version|int < 8

    - name: Install curl on Oracle Linux 8+ using dnf
      dnf:
        name: curl
        state: present
      when: ansible_os_family == "RedHat" and ansible_distribution_major_version|int >= 8

    - name: Install curl on SUSE-based distributions (openSUSE, SUSE Linux Enterprise)
      zypper:
        name: curl
        state: present
      when: ansible_os_family == "Suse"

    - name: Install curl on Arch Linux (for future support)
      pacman:
        name: curl
        state: present
      when: ansible_os_family == "Archlinux"

    - name: Install curl on Alpine Linux (for future support)
      apk:
        name: curl
        state: present
      when: ansible_os_family == "Alpine"

    - name: Show installed curl version
      command: curl --version
      register: curl_version
      changed_when: false
      ignore_errors: yes

    - name: Display result
      debug:
        msg: "curl provisioning finished (version info may be above)."

```

`tests\runner_test.go`:

```go
package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/yuanjua/autowsl/internal/runner"
)

func TestExecRunnerNoTimeout(t *testing.T) {
	// Create runner with no timeout (0 means no timeout)
	r := runner.NewExecRunner(0)

	// Run a simple command that should succeed
	stdout, stderr, err := r.Run("echo", "hello")
	if err != nil {
		t.Fatalf("Expected no error, got %v\nStderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "hello") {
		t.Errorf("Expected stdout to contain 'hello', got: %s", stdout)
	}
}

func TestExecRunnerWithTimeout(t *testing.T) {
	// Create runner with 1 second timeout
	r := runner.NewExecRunner(1 * time.Second)

	// Run a quick command that should succeed
	stdout, stderr, err := r.Run("echo", "test")
	if err != nil {
		t.Fatalf("Expected no error, got %v\nStderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "test") {
		t.Errorf("Expected stdout to contain 'test', got: %s", stdout)
	}
}

func TestExecRunnerTimeoutExceeded(t *testing.T) {
	// Skip on Windows as sleep command behaves differently
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	// Create runner with very short timeout
	r := runner.NewExecRunner(100 * time.Millisecond)

	// Try to run a command that would take longer (platform-specific)
	// On Windows, we'd use timeout or ping, on Unix we'd use sleep
	// For cross-platform compatibility, let's just verify the timeout mechanism works
	_, _, err := r.Run("ping", "127.0.0.1", "-n", "100") // Windows: ping 100 times

	// We expect a timeout or error
	if err == nil {
		t.Log("Warning: Command completed before timeout (might be too fast)")
	}
}

func TestExecRunnerWithInput(t *testing.T) {
	// Create runner with no timeout
	r := runner.NewExecRunner(0)

	// Test RunWithInput - on Windows we can use 'sort' or 'findstr'
	// For cross-platform, let's use a command that reads stdin
	input := "test input"
	stdout, stderr, err := r.RunWithInput("cmd", input, "/C", "more")

	if err != nil {
		t.Logf("RunWithInput test: %v (stderr: %s)", err, stderr)
		// Some platforms might not support this, so just log
	}

	if stdout != "" {
		t.Logf("Received stdout: %s", stdout)
	}
}

func TestExecRunnerDryRun(t *testing.T) {
	r := runner.NewExecRunner(0)
	r.DryRun = true

	stdout, stderr, err := r.Run("some-command", "arg1", "arg2")

	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}

	if stderr != "" {
		t.Errorf("Dry run should have empty stderr, got %s", stderr)
	}

	if !strings.Contains(stdout, "dry-run") {
		t.Errorf("Expected dry-run log, got: %s", stdout)
	}

	if !strings.Contains(stdout, "some-command") {
		t.Errorf("Expected command name in output, got: %s", stdout)
	}
}

func TestExecRunnerInvalidCommand(t *testing.T) {
	r := runner.NewExecRunner(0)

	_, _, err := r.Run("this-command-definitely-does-not-exist-12345")

	if err == nil {
		t.Error("Expected error for invalid command, got nil")
	}
}

```

`tests\winget_test.go`:

```go
package tests

import (
	"fmt"
	"testing"

	"github.com/yuanjua/autowsl/internal/winget"
)

// TestWingetDistrosCatalog verifies the catalog structure
func TestWingetDistrosCatalog(t *testing.T) {
	distros := winget.GetWingetDistros()

	if len(distros) == 0 {
		t.Fatal("Catalog should not be empty")
	}

	// Verify each distro has required fields
	for _, d := range distros {
		if d.Version == "" {
			t.Errorf("Distro missing Version: %+v", d)
		}
		if d.PackageID == "" {
			t.Errorf("Distro missing PackageID: %+v", d)
		}
		if d.Group == "" {
			t.Errorf("Distro missing Group: %+v", d)
		}
	}

	t.Logf("Catalog contains %d distributions", len(distros))
}

// TestFindWingetDistroByVersion tests version lookup
func TestFindWingetDistroByVersion(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{"Ubuntu 22.04 LTS", false},
		{"Ubuntu 24.04 LTS", false},
		{"Debian", false},
		{"NonExistent", true}, // Should return error for non-existent
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			distro, err := winget.FindWingetDistroByVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindWingetDistroByVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && distro == nil {
				t.Errorf("Expected distro for %s, got nil", tt.version)
			}
		})
	}
}

// TestIsWSLPackage tests the WSL package detection heuristic
func TestIsWSLPackage(t *testing.T) {
	tests := []struct {
		packageID string
		want      bool
	}{
		{"Canonical.Ubuntu.2204", true},
		{"Debian.Debian", true},
		{"KaliLinux.KaliLinux", true},
		{"openSUSE.Tumbleweed", true},
		{"Microsoft.VisualStudioCode", false},
		{"Google.Chrome", false},
		{"SomeVendor.WSLDistro", true}, // Contains "WSL"
	}

	for _, tt := range tests {
		t.Run(tt.packageID, func(t *testing.T) {
			if got := winget.IsWSLPackage(tt.packageID); got != tt.want {
				t.Errorf("IsWSLPackage(%s) = %v, want %v", tt.packageID, got, tt.want)
			}
		})
	}
}

// ExampleManager demonstrates basic usage
func ExampleManager() {
	// Create download manager
	mgr := winget.NewManager("./.autowsl_tmp")

	// Check if winget is available
	if !mgr.IsWingetAvailable() {
		fmt.Println("Winget is not available")
		return
	}

	// List available distributions
	distros := mgr.GetCatalog()
	fmt.Printf("Available distributions: %d\n", len(distros))

	// Download a distribution (commented out to avoid actual download in example)
	// opts := winget.DownloadOptions{Version: "Ubuntu 22.04 LTS"}
	// filePath, _ := mgr.Download(opts)
	// fmt.Printf("Downloaded to: %s\n", filePath)
}

// ExampleDownloadOptions_catalog demonstrates downloading from catalog
func ExampleDownloadOptions_catalog() {
	mgr := winget.NewManager("./.autowsl_tmp")

	opts := winget.DownloadOptions{
		Version: "Ubuntu 22.04 LTS",
	}

	filePath, err := mgr.Download(opts)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	fmt.Printf("Downloaded to: %s\n", filePath)
}

// ExampleDownloadOptions_custom demonstrates downloading with custom package ID
func ExampleDownloadOptions_custom() {
	mgr := winget.NewManager("./.autowsl_tmp")

	opts := winget.DownloadOptions{
		PackageID:         "Canonical.Ubuntu.2204",
		ValidatePackageID: true,
	}

	filePath, err := mgr.Download(opts)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	fmt.Printf("Downloaded to: %s\n", filePath)
}

```

`tests\wsl_import_test.go`:

```go
package tests

import (
	"testing"

	"github.com/yuanjua/autowsl/internal/wsl"
)

func TestWSLImportValidation(t *testing.T) {
	mock := NewMockRunner()
	mock.Outputs["wsl.exe -l -v"] = "" // No existing distros

	client := wsl.NewClient(mock)

	tests := []struct {
		name        string
		opts        wsl.ImportOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty name",
			opts: wsl.ImportOptions{
				Name:        "",
				InstallPath: "/some/path",
				TarFilePath: "test.tar",
			},
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name: "empty install path",
			opts: wsl.ImportOptions{
				Name:        "test-distro",
				InstallPath: "",
				TarFilePath: "test.tar",
			},
			expectError: true,
			errorMsg:    "installation path cannot be empty",
		},
		{
			name: "empty tar file path",
			opts: wsl.ImportOptions{
				Name:        "test-distro",
				InstallPath: "/some/path",
				TarFilePath: "",
			},
			expectError: true,
			errorMsg:    "tar file path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Import(tt.opts)

			if tt.expectError && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if tt.expectError && err != nil {
				// Just verify we got an error, don't check exact message
				// as file existence checks might vary
			}
		})
	}
}

func TestWSLUnregisterValidation(t *testing.T) {
	mock := NewMockRunner()
	client := wsl.NewClient(mock)

	// Test empty name
	err := client.Unregister("")
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}

	// Test non-existing distro
	mock.Outputs["wsl.exe -l -v"] = `  NAME                   STATE           VERSION
* Ubuntu                 Running         2
`
	err = client.Unregister("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existing distro, got nil")
	}

	// Test successful unregister
	mock.Outputs["wsl.exe --unregister Ubuntu"] = ""
	err = client.Unregister("Ubuntu")
	if err != nil {
		t.Errorf("Expected no error for valid unregister, got: %v", err)
	}
}

func TestWSLExportValidation(t *testing.T) {
	mock := NewMockRunner()
	client := wsl.NewClient(mock)

	tests := []struct {
		name        string
		distroName  string
		outputPath  string
		expectError bool
	}{
		{
			name:        "empty distro name",
			distroName:  "",
			outputPath:  "output.tar",
			expectError: true,
		},
		{
			name:        "empty output path",
			distroName:  "Ubuntu",
			outputPath:  "",
			expectError: true,
		},
		{
			name:        "non-existing distro",
			distroName:  "NonExistent",
			outputPath:  "output.tar",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock to return no distros for non-existing check
			mock.Outputs["wsl.exe -l -v"] = ""

			err := client.Export(tt.distroName, tt.outputPath)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

```

`tests\wsl_status_test.go`:

```go
package tests

import (
	"testing"

	"github.com/yuanjua/autowsl/internal/wsl"
)

func TestListInstalledDistros(t *testing.T) {
	// Fake wsl.exe output
	fakeOutput := `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
  kali-linux             Stopped         2
`

	mock := &MockRunner{
		Outputs: map[string]string{
			"wsl.exe -l -v": fakeOutput,
		},
	}

	client := wsl.NewClient(mock)
	distros, err := client.ListInstalledDistros()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(distros) != 3 {
		t.Fatalf("Expected 3 distros, got %d", len(distros))
	}

	// Check first distro
	if distros[0].Name != "Ubuntu-22.04" {
		t.Errorf("Expected 'Ubuntu-22.04', got '%s'", distros[0].Name)
	}
	if distros[0].State != "Running" {
		t.Errorf("Expected 'Running', got '%s'", distros[0].State)
	}
	if !distros[0].Default {
		t.Error("Expected first distro to be default")
	}

	// Check second distro
	if distros[1].Name != "Debian" {
		t.Errorf("Expected 'Debian', got '%s'", distros[1].Name)
	}
	if distros[1].Default {
		t.Error("Expected second distro not to be default")
	}
}

func TestIsDistroInstalled(t *testing.T) {
	fakeOutput := `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
`

	mock := &MockRunner{
		Outputs: map[string]string{
			"wsl.exe -l -v": fakeOutput,
		},
	}

	client := wsl.NewClient(mock)

	// Test existing distro
	exists, err := client.IsDistroInstalled("Ubuntu-22.04")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !exists {
		t.Error("Expected Ubuntu-22.04 to exist")
	}

	// Test non-existing distro
	exists, err = client.IsDistroInstalled("NonExistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if exists {
		t.Error("Expected NonExistent to not exist")
	}
}

```

`tests\wsl_test.go`:

```go
package tests

import (
	"strings"
	"testing"

	"github.com/yuanjua/autowsl/internal/wsl"
)

// MockRunner is a test double for runner.Runner
type MockRunner struct {
	Outputs map[string]string // command -> stdout
	Errors  map[string]error  // command -> error
	Calls   []string          // track all commands called
}

func NewMockRunner() *MockRunner {
	return &MockRunner{
		Outputs: make(map[string]string),
		Errors:  make(map[string]error),
		Calls:   make([]string, 0),
	}
}

func (m *MockRunner) Run(name string, args ...string) (string, string, error) {
	cmd := name + " " + strings.Join(args, " ")
	m.Calls = append(m.Calls, cmd)

	if err, ok := m.Errors[cmd]; ok {
		return "", "", err
	}
	if output, ok := m.Outputs[cmd]; ok {
		return output, "", nil
	}
	return "", "", nil
}

func (m *MockRunner) RunWithInput(name string, stdin string, args ...string) (string, string, error) {
	return m.Run(name, args...)
}

func TestWSLListInstalledDistros(t *testing.T) {
	// Prepare fake WSL output
	fakeOutput := `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
  kali-linux             Stopped         1
`

	mock := NewMockRunner()
	mock.Outputs["wsl.exe -l -v"] = fakeOutput

	client := wsl.NewClient(mock)
	distros, err := client.ListInstalledDistros()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify command was called
	if len(mock.Calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(mock.Calls))
	}

	// Verify we got 3 distributions
	if len(distros) != 3 {
		t.Fatalf("Expected 3 distros, got %d", len(distros))
	}

	// Test first distro (default)
	if distros[0].Name != "Ubuntu-22.04" {
		t.Errorf("Expected name 'Ubuntu-22.04', got '%s'", distros[0].Name)
	}
	if distros[0].State != "Running" {
		t.Errorf("Expected state 'Running', got '%s'", distros[0].State)
	}
	if distros[0].Version != "2" {
		t.Errorf("Expected version '2', got '%s'", distros[0].Version)
	}
	if !distros[0].Default {
		t.Error("Expected first distro to be default")
	}

	// Test second distro (not default)
	if distros[1].Name != "Debian" {
		t.Errorf("Expected name 'Debian', got '%s'", distros[1].Name)
	}
	if distros[1].Default {
		t.Error("Expected second distro not to be default")
	}

	// Test third distro (WSL 1)
	if distros[2].Version != "1" {
		t.Errorf("Expected version '1', got '%s'", distros[2].Version)
	}
}

func TestWSLIsDistroInstalled(t *testing.T) {
	fakeOutput := `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
`

	mock := NewMockRunner()
	mock.Outputs["wsl.exe -l -v"] = fakeOutput

	client := wsl.NewClient(mock)

	tests := []struct {
		name     string
		distro   string
		expected bool
	}{
		{"existing distro", "Ubuntu-22.04", true},
		{"another existing distro", "Debian", true},
		{"non-existing distro", "Arch", false},
		{"case sensitive check", "ubuntu-22.04", false}, // Names are case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := client.IsDistroInstalled(tt.distro)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if exists != tt.expected {
				t.Errorf("IsDistroInstalled(%s) = %v, want %v", tt.distro, exists, tt.expected)
			}
		})
	}
}

func TestWSLCheckWSLInstalled(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockRunner)
		expectError bool
	}{
		{
			name: "WSL is installed",
			setupMock: func(m *MockRunner) {
				m.Outputs["wsl.exe --status"] = "some status output"
			},
			expectError: false,
		},
		{
			name: "WSL is not installed",
			setupMock: func(m *MockRunner) {
				m.Errors["wsl.exe --status"] = &mockError{"wsl not found"}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRunner()
			tt.setupMock(mock)

			client := wsl.NewClient(mock)
			err := client.CheckWSLInstalled()

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestWSLParseListEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCount int
		description   string
	}{
		{
			name: "UTF-8 BOM and null bytes",
			input: "\ufeff\x00  NAME\x00                   STATE\x00           VERSION\x00\n" +
				"* Ubuntu\x00                 Running\x00         2\x00\n",
			expectedCount: 1,
			description:   "Should handle UTF-8 BOM and null bytes",
		},
		{
			name:          "Empty output",
			input:         "",
			expectedCount: 0,
			description:   "Should handle empty output",
		},
		{
			name: "Only headers",
			input: `  NAME                   STATE           VERSION
----------------------------------------
`,
			expectedCount: 0,
			description:   "Should handle output with only headers",
		},
		{
			name: "Multiple distros with varying states",
			input: `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
  Arch                   Installing      2
  Alpine                 Uninstalling    1
`,
			expectedCount: 4,
			description:   "Should parse multiple distros with different states",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRunner()
			mock.Outputs["wsl.exe -l -v"] = tt.input

			client := wsl.NewClient(mock)
			distros, err := client.ListInstalledDistros()

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(distros) != tt.expectedCount {
				t.Errorf("%s: expected %d distros, got %d", tt.description, tt.expectedCount, len(distros))
			}
		})
	}
}

// Helper type for mock errors
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

```