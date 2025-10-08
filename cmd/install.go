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
