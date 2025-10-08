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
