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
