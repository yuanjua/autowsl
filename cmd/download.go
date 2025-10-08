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
