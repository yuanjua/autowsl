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
