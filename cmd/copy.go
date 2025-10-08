package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/yuanjua/autowsl/internal/wsl"
)

var (
	copyName    string
	copyPath    string
	copyVersion int
)

var copyCmd = &cobra.Command{
	Use:   "copy [source-name]",
	Short: "Copy a WSL distribution with a new name",
	Long: `Copy an existing WSL distribution to a new distribution with a different name.
This exports the source distribution to a temporary tar file, then imports it
with the new name and location.

Examples:
	# Interactive copy (select source distro, prompt for name and path)
	autowsl copy

	# Copy specific distro (interactive name and path)
	autowsl copy ubuntu-2204-lts

	# Copy with all options specified
	autowsl copy ubuntu-2204-lts --name my-clone --path ./wsl-distros/my-clone`,
	RunE: runCopy,
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().StringVar(&copyName, "name", "", "Name for the new distribution")
	copyCmd.Flags().StringVar(&copyPath, "path", "", "Installation path for the new distribution")
	copyCmd.Flags().IntVar(&copyVersion, "version", 2, "WSL version to use (1 or 2)")
}

func runCopy(cmd *cobra.Command, args []string) error {
	// Check if WSL is installed
	if err := wsl.CheckWSLInstalled(); err != nil {
		return fmt.Errorf("WSL is not available: %w\nPlease install WSL first: https://docs.microsoft.com/en-us/windows/wsl/install", err)
	}

	isInteractive := len(args) == 0

	// Select source distribution
	var sourceDistro string
	var err error

	if isInteractive {
		sourceDistro, err = selectInstalledDistroInteractive()
		if err != nil {
			return err
		}
	} else {
		sourceDistro = args[0]
		// Verify source distro exists
		exists, err := wsl.IsDistroInstalled(sourceDistro)
		if err != nil {
			return fmt.Errorf("failed to check distribution: %w", err)
		}
		if !exists {
			return fmt.Errorf("source distribution '%s' does not exist", sourceDistro)
		}
	}

	// Determine new distribution name
	newName := copyName
	if newName == "" {
		defaultName := sourceDistro + "-copy"
		if isInteractive {
			namePrompt := promptui.Prompt{
				Label:   "New distribution name",
				Default: defaultName,
			}
			if customName, err := namePrompt.Run(); err != nil {
				return fmt.Errorf("failed to get distribution name: %w", err)
			} else {
				newName = customName
			}
		} else {
			newName = defaultName
		}
	}

	// Check if new distro name already exists
	exists, err := wsl.IsDistroInstalled(newName)
	if err != nil {
		return fmt.Errorf("failed to check existing distributions: %w", err)
	}
	if exists {
		return fmt.Errorf("distribution '%s' already exists", newName)
	}

	// Determine installation path
	newPath := copyPath
	if newPath == "" {
		cwd, _ := os.Getwd()
		newPath = filepath.Join(cwd, "wsl-distros", newName)
		if isInteractive {
			pathPrompt := promptui.Prompt{
				Label:   "Installation path",
				Default: newPath,
			}
			if customPath, err := pathPrompt.Run(); err != nil {
				return fmt.Errorf("failed to get installation path: %w", err)
			} else {
				newPath = customPath
			}
		}
	}

	// Validate WSL version
	if copyVersion != 1 && copyVersion != 2 {
		return fmt.Errorf("invalid --version %d (must be 1 or 2)", copyVersion)
	}

	// Display configuration
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Copy Configuration\n")
	fmt.Printf("%s\n", strings.Repeat("=", 60))
	fmt.Printf("Source:       %s\n", sourceDistro)
	fmt.Printf("New Name:     %s\n", newName)
	fmt.Printf("New Path:     %s\n", newPath)
	fmt.Printf("WSL Version:  %d\n", copyVersion)
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	// Create temporary directory for export
	cwd, _ := os.Getwd()
	tempDir := filepath.Join(cwd, ".autowsl_tmp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory '%s': %w", tempDir, err)
	}

	// Temporary tar file path
	tempTarPath := filepath.Join(tempDir, fmt.Sprintf("%s-export.tar", sourceDistro))

	// Export source distribution
	fmt.Printf("→ Exporting '%s' to temporary tar file...\n", sourceDistro)
	fmt.Println("  This may take a while depending on the size of your distribution...")

	if err := wsl.Export(sourceDistro, tempTarPath); err != nil {
		_ = os.RemoveAll(tempDir)
		return fmt.Errorf("failed to export distribution: %w", err)
	}

	// Get file size
	fileInfo, _ := os.Stat(tempTarPath)
	sizeInMB := float64(fileInfo.Size()) / 1024 / 1024
	fmt.Printf("  ✓ Export completed (%.2f MB)\n\n", sizeInMB)

	// Import to new name
	fmt.Printf("→ Importing to WSL as '%s'...\n", newName)
	importOpts := wsl.ImportOptions{
		Name:        newName,
		InstallPath: newPath,
		TarFilePath: tempTarPath,
		Version:     copyVersion,
	}

	if err := wsl.Import(importOpts); err != nil {
		_ = os.RemoveAll(tempDir)
		return fmt.Errorf("failed to import distribution '%s' to '%s': %w", newName, newPath, err)
	}

	fmt.Println("  ✓ Import completed successfully")

	// Cleanup temporary files
	fmt.Println("\n→ Cleaning up temporary files...")
	if err := os.RemoveAll(tempDir); err != nil {
		fmt.Printf("  ⚠ Warning: Failed to cleanup temp directory: %v\n", err)
	} else {
		fmt.Println("  ✓ Cleanup completed")
	}

	// Print success message
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("✓ SUCCESS: WSL distribution copied\n")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Source:   %s\n", sourceDistro)
	fmt.Printf("New Name: %s\n", newName)
	fmt.Printf("Location: %s\n", newPath)
	fmt.Printf("Version:  WSL %d\n", copyVersion)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\nLaunch with:  wsl -d %s\n", newName)
	fmt.Printf("List all:     autowsl list\n\n")

	return nil
}
