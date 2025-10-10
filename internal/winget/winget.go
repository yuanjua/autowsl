package winget

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WingetDownloader handles downloading WSL distributions using wingetcreate
type WingetDownloader struct {
	DownloadDir string
}

// NewWingetDownloader creates a new wingetcreate downloader
func NewWingetDownloader(downloadDir string) *WingetDownloader {
	return &WingetDownloader{
		DownloadDir: downloadDir,
	}
}

// Download downloads a package using wingetcreate
// packageID is the winget package identifier (e.g., "Canonical.Ubuntu.2204")
// Returns the path to the downloaded file
func (w *WingetDownloader) Download(packageID string) (string, error) {
	// Ensure download directory exists
	if err := os.MkdirAll(w.DownloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download directory: %w", err)
	}

	// Check if wingetcreate is available
	if !w.IsWingetAvailable() {
		return "", fmt.Errorf("wingetcreate is not available. Please install winget-create from https://github.com/microsoft/winget-create")
	}

	fmt.Printf("Downloading package: %s\n", packageID)
	fmt.Printf("Download directory: %s\n\n", w.DownloadDir)

	// Run wingetcreate download command
	// wingetcreate download <PackageId> --download-directory <PathToTempDir>
	cmd := exec.Command("wingetcreate", "download", packageID, "--download-directory", w.DownloadDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("wingetcreate download failed: %w", err)
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

// IsWingetAvailable checks if wingetcreate is installed and available
func (w *WingetDownloader) IsWingetAvailable() bool {
	cmd := exec.Command("wingetcreate", "--version")
	err := cmd.Run()
	return err == nil
}

// GetWingetVersion returns the installed wingetcreate version
func (w *WingetDownloader) GetWingetVersion() (string, error) {
	cmd := exec.Command("wingetcreate", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get wingetcreate version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
