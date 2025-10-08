package wsl

import (
	"errors"
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
	output, stderr, err := c.runner.Run("wsl.exe", "-l", "-v")
	if err == nil {
		return parseWSLList(output)
	}

	// Fallback: attempt legacy command without version column
	// This improves resilience on hosts where -v is unsupported or WSL is in a partially initialized state.
	fallbackOut, fallbackErrStr, fallbackErr := c.runner.Run("wsl.exe", "-l")
	if fallbackErr == nil {
		return parseWSLListBasic(fallbackOut), nil
	}

	// If both commands failed, decide whether to treat as empty (benign) or propagate error.
	// On some fresh systems a generic failure (exit status 0xffffffff) can occur before any distro is installed.
	// Combine stderr plus underlying error messages (in case stderr is empty but Go error contains exit status)
	combinedErrMsg := strings.Join([]string{stderr, fallbackErrStr, fmt.Sprint(err), fmt.Sprint(fallbackErr)}, "\n")
	lowered := strings.ToLower(combinedErrMsg)
	benignIndicators := []string{
		"0xffffffff",                 // generic WSL failure often seen pre-initialization
		"element not found",          // sometimes returned when feature pieces missing
		"no installed distributions", // hypothetical message
	}
	for _, indicator := range benignIndicators {
		if strings.Contains(lowered, indicator) {
			return []InstalledDistro{}, nil
		}
	}

	// Propagate a combined error for easier debugging.
	return nil, fmt.Errorf("failed to list WSL distributions: %w | fallback: %v", err, fallbackErr)
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

// parseWSLListBasic parses output of the legacy "wsl -l" (no -v) command.
// Expected format (example):
//
//	Windows Subsystem for Linux Distributions:
//	Ubuntu-22.04 (Default)
//	Debian
//
// A leading * may or may not appear depending on Windows build. We tolerate both.
func parseWSLListBasic(output string) []InstalledDistro {
	var distros []InstalledDistro
	output = strings.ReplaceAll(output, "\x00", "")
	output = strings.TrimPrefix(output, "\ufeff")

	lines := strings.Split(output, "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		l := strings.ToLower(line)
		if strings.HasPrefix(l, "windows subsystem for linux") || strings.Contains(l, "distributions:") || strings.HasPrefix(l, "the following") {
			continue // header lines
		}
		// Remove (Default) tag if present
		defaultFlag := false
		if strings.Contains(line, "(Default)") {
			defaultFlag = true
			line = strings.TrimSpace(strings.ReplaceAll(line, "(Default)", ""))
		}
		// Remove leading * if present
		line = strings.TrimLeft(line, "* ")
		if line == "" {
			continue
		}
		distros = append(distros, InstalledDistro{Name: line, Default: defaultFlag})
	}
	return distros
}

// Provide an exported helper only for tests (optional) - not exporting to avoid API surface increase.
var _ = errors.New // silence unused import if build tags exclude tests using errors

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
