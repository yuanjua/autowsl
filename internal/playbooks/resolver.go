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
