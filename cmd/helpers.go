package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/yuanjua/autowsl/internal/ansible"
	"github.com/yuanjua/autowsl/internal/distro"
	"github.com/yuanjua/autowsl/internal/playbooks"
	"github.com/yuanjua/autowsl/internal/wsl"
)

// selectDistroInteractive handles interactive distribution selection with promptui
func selectDistroInteractive() (distro.Distro, error) {
	distros := distro.GetAllDistros()

	// Create selection prompt with colored templates
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Group | cyan }} - {{ .Version | yellow }} ({{ .Architecture | faint }})",
		Inactive: "  {{ .Group | white }} - {{ .Version | faint }} ({{ .Architecture | faint }})",
		Selected: "* {{ .Group | green }} - {{ .Version | green }}",
	}

	prompt := promptui.Select{
		Label:     "Select a WSL distribution",
		Items:     distros,
		Templates: templates,
		Size:      12,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return distro.Distro{}, fmt.Errorf("selection cancelled: %w", err)
	}

	return distros[idx], nil
}

// selectDistroByVersion finds a distribution by its version name or package ID
func selectDistroByVersion(versionName string) (distro.Distro, error) {
	distros := distro.GetAllDistros()

	// Try exact match on version name
	for _, d := range distros {
		if strings.EqualFold(d.Version, versionName) {
			return d, nil
		}
	}

	// Try match on package ID
	for _, d := range distros {
		if strings.EqualFold(d.PackageID, versionName) {
			return d, nil
		}
	}

	// If not found, show available options
	fmt.Fprintf(os.Stderr, "\nError: distribution '%s' not found\n\n", versionName)
	fmt.Fprintln(os.Stderr, "Available distributions:")
	fmt.Fprintln(os.Stderr, strings.Repeat("-", 80))

	currentGroup := ""
	for _, d := range distros {
		if d.Group != currentGroup {
			if currentGroup != "" {
				fmt.Fprintln(os.Stderr, "")
			}
			fmt.Fprintf(os.Stderr, "%s:\n", d.Group)
			currentGroup = d.Group
		}
		fmt.Fprintf(os.Stderr, "  %-30s %s\n", d.Version, d.PackageID)
	}
	fmt.Fprintln(os.Stderr, strings.Repeat("-", 80))
	fmt.Fprintln(os.Stderr, "\nTip: Use either the version name (e.g., 'Ubuntu 22.04 LTS') or package ID")

	return distro.Distro{}, fmt.Errorf("distribution '%s' not found", versionName)
}

// selectDistro selects a distribution either interactively or by version name
func selectDistro(args []string) (distro.Distro, error) {
	if len(args) == 0 {
		return selectDistroInteractive()
	}
	return selectDistroByVersion(args[0])
}

// selectInstalledDistroInteractive handles interactive selection from installed distros
func selectInstalledDistroInteractive() (string, error) {
	distros, err := wsl.ListInstalledDistros()
	if err != nil {
		return "", fmt.Errorf("failed to list distributions: %w", err)
	}

	if len(distros) == 0 {
		return "", fmt.Errorf("no WSL distributions found. Install one first with 'autowsl install'")
	}

	// Create selection prompt
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Name | cyan }} ({{ .State | yellow }})",
		Inactive: "  {{ .Name | white }} ({{ .State | faint }})",
		Selected: "* {{ .Name | green }}",
	}

	prompt := promptui.Select{
		Label:     "Select distribution",
		Items:     distros,
		Templates: templates,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("selection cancelled: %w", err)
	}

	return distros[idx].Name, nil
}

// selectInstalledDistro selects an installed distro either from args or interactively
// (selectInstalledDistro removed – previously unused; interactive selection handled where needed)

// promptForPlaybooks prompts user to enter playbooks interactively
func promptForPlaybooks() ([]string, error) {
	prompt := promptui.Prompt{
		Label:   "Enter playbook(s) (comma or space separated, aliases/files/URLs, or 'none' to skip)",
		Default: "",
	}

	result, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("input cancelled: %w", err)
	}

	result = strings.TrimSpace(result)
	if result == "" || strings.ToLower(result) == "none" {
		return []string{}, nil
	}

	// Split by comma or space
	var inputs []string
	// First try comma
	if strings.Contains(result, ",") {
		parts := strings.Split(result, ",")
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				inputs = append(inputs, trimmed)
			}
		}
	} else {
		// Split by space
		parts := strings.Fields(result)
		inputs = append(inputs, parts...)
	}

	if len(inputs) == 0 {
		return []string{"curl"}, nil
	}

	return inputs, nil
}

// ProvisioningPipelineOptions holds options for the provisioning pipeline
type ProvisioningPipelineOptions struct {
	DistroName     string
	PlaybookInputs []string
	Tags           []string
	ExtraVars      []string
	Verbose        bool
	TempDir        string
}

// runProvisioningPipeline executes the complete provisioning pipeline
func runProvisioningPipeline(opts ProvisioningPipelineOptions) error {
	fmt.Printf("\nProvisioning: %s\n", opts.DistroName)
	fmt.Println(strings.Repeat("=", 60))

	// Parse extra vars
	extraVarsMap := make(map[string]string)
	if len(opts.ExtraVars) > 0 {
		var err error
		extraVarsMap, err = playbooks.ParseExtraVars(opts.ExtraVars)
		if err != nil {
			return fmt.Errorf("invalid extra-vars: %w", err)
		}
	}

	// Ensure temp directory exists
	if opts.TempDir == "" {
		cwd, _ := os.Getwd()
		opts.TempDir = filepath.Join(cwd, ".autowsl_tmp")
	}
	if err := os.MkdirAll(opts.TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp dir '%s': %w", opts.TempDir, err)
	}

	// Resolve playbooks
	cwd, _ := os.Getwd()
	resolver := playbooks.NewResolver(opts.TempDir, cwd)
	playbookPaths, err := resolver.ResolveMultiple(opts.PlaybookInputs)
	if err != nil {
		return fmt.Errorf("failed to resolve playbooks: %w", err)
	}

	if len(playbookPaths) == 0 {
		return fmt.Errorf("no playbooks resolved")
	}

	// Execute playbooks with summary tracking
	summary := &ansible.ExecutionSummary{}

	for _, playbookPath := range playbookPaths {
		start := time.Now()

		fmt.Printf("\nRunning playbook: %s\n", filepath.Base(playbookPath))
		fmt.Println(strings.Repeat("-", 60))

		execOpts := ansible.PlaybookOptions{
			DistroName:   opts.DistroName,
			PlaybookPath: playbookPath,
			Tags:         opts.Tags,
			Verbose:      opts.Verbose,
			ExtraVars:    extraVarsMap,
		}

		err := ansible.ExecutePlaybook(execOpts)
		duration := time.Since(start)

		if err != nil {
			summary.Add(ansible.ExecutionResult{
				PlaybookName: filepath.Base(playbookPath),
				Status:       "failed",
				Duration:     duration,
				Error:        err,
			})
			fmt.Printf("\nPlaybook '%s' failed: %v\n", filepath.Base(playbookPath), err)
			break // Stop on first failure
		} else {
			summary.Add(ansible.ExecutionResult{
				PlaybookName: filepath.Base(playbookPath),
				Status:       "success",
				Duration:     duration,
			})
		}
	}

	// Print summary if multiple playbooks
	if len(playbookPaths) > 1 {
		summary.Print()
	}

	if summary.HasFailures() {
		return fmt.Errorf("provisioning completed with failures")
	}

	// Success message
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("SUCCESS: Distribution provisioned\n")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Distribution: %s\n", opts.DistroName)
	fmt.Printf("Playbooks:    %s\n", strings.Join(playbooks.BaseNames(playbookPaths), ", "))
	if len(opts.Tags) > 0 {
		fmt.Printf("Tags:         %s\n", strings.Join(opts.Tags, ", "))
	}
	if len(extraVarsMap) > 0 {
		fmt.Printf("Extra vars:   %d variables\n", len(extraVarsMap))
	}
	fmt.Println(strings.Repeat("=", 60))

	return nil
}
