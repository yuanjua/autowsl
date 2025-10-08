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
