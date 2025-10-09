package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuanjua/autowsl/internal/ansible"
	"github.com/yuanjua/autowsl/internal/wsl"
)

var (
	provisionTags      []string
	provisionPlaybooks []string
	provisionExtraVars string
	provisionRepo      string
	provisionVerbose   bool
)

var provisionCmd = &cobra.Command{
	Use:   "provision [distro-name]",
	Short: "Provision a WSL distribution with Ansible",
	Long: `Provision a WSL distribution using Ansible playbooks.
Automatically installs Ansible if not present.

Examples:
  # Interactive mode - select distro and playbook
  autowsl provision

  # Specify distro, interactive playbook selection (default: curl)
  autowsl provision ubuntu-2204

  # Use specific playbook (file, URL, or alias)
  autowsl provision ubuntu-2204 --playbooks ./my-playbook.yml
  autowsl provision ubuntu-2204 --playbooks https://example.com/setup.yml
  autowsl provision ubuntu-2204 --playbooks curl

  # Use multiple playbooks
  autowsl provision ubuntu-2204 --playbooks curl,./dev.yml,https://example.com/extra.yml

  # Use playbook from Git repository
  autowsl provision ubuntu-2204 --repo https://github.com/user/ansible-playbooks

  # Run specific tags only
  autowsl provision ubuntu-2204 --playbooks ./setup.yml --tags docker,nodejs

  # Pass extra variables
  autowsl provision ubuntu-2204 --playbooks ./setup.yml --extra-vars "user=john env=dev"
  autowsl provision ubuntu-2204 --playbooks ./setup.yml --extra-vars "user=john,env=dev"

  # Verbose output
  autowsl provision ubuntu-2204 --verbose`,
	RunE: runProvision,
}

func init() {
	rootCmd.AddCommand(provisionCmd)
	provisionCmd.Flags().StringSliceVar(&provisionTags, "tags", nil, "Ansible tags to run (comma-separated)")
	provisionCmd.Flags().StringSliceVar(&provisionPlaybooks, "playbooks", []string{}, "Playbook files, URLs, or aliases (comma-separated or repeat flag)")
	provisionCmd.Flags().StringVar(&provisionExtraVars, "extra-vars", "", "Extra variables in key=val format (space or comma-separated)")
	provisionCmd.Flags().StringVar(&provisionRepo, "repo", "", "Git repository URL containing playbooks")
	provisionCmd.Flags().BoolVarP(&provisionVerbose, "verbose", "v", false, "Verbose output")
}

func runProvision(cmd *cobra.Command, args []string) error {
	var distroName string
	var playbookInputs []string

	// Get distribution name using shared helper
	if len(args) > 0 {
		distroName = args[0]
	} else {
		var err error
		distroName, err = selectInstalledDistroInteractive()
		if err != nil {
			return err
		}
	}

	// Check if distro exists
	exists, err := wsl.IsDistroInstalled(distroName)
	if err != nil {
		return fmt.Errorf("failed to check distribution: %w", err)
	}
	if !exists {
		// List available distros to help the user
		distros, listErr := wsl.ListInstalledDistros()
		if listErr == nil && len(distros) > 0 {
			fmt.Fprintf(os.Stderr, "\nError: distribution '%s' does not exist\n\n", distroName)
			fmt.Fprintln(os.Stderr, "Available installed distributions:")
			fmt.Fprintln(os.Stderr, strings.Repeat("-", 60))
			for _, d := range distros {
				marker := " "
				if d.Default {
					marker = "*"
				}
				status := d.State
				fmt.Fprintf(os.Stderr, "%s %-30s (%s)\n", marker, d.Name, status)
			}
			fmt.Fprintln(os.Stderr, strings.Repeat("-", 60))
			fmt.Fprintln(os.Stderr, "\nTip: Use 'autowsl list' to see all installed distributions")
			fmt.Fprintln(os.Stderr, "     Use 'autowsl install' to install a new distribution")
			return fmt.Errorf("distribution not found")
		}
		return fmt.Errorf("distribution '%s' does not exist", distroName)
	}

	// If no playbooks specified via flags, use interactive prompt
	if len(provisionPlaybooks) == 0 {
		// Interactive mode - prompt for playbooks (both with and without distro arg)
		playbookInputs, err = promptForPlaybooks()
		if err != nil {
			return err
		}
	} else {
		playbookInputs = provisionPlaybooks
	}

	// Create temp directory for downloads
	cwd, _ := os.Getwd()
	tempDir := filepath.Join(cwd, ".autowsl_tmp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Handle repo-based provisioning (legacy mode)
	if provisionRepo != "" {
		fmt.Printf("Using playbook from Git repository\n\n")

		tmpDir := "/tmp/autowsl-playbooks"
		if err := ansible.CloneGitRepo(distroName, provisionRepo, tmpDir); err != nil {
			return err
		}

		// Look for common playbook names
		commonNames := []string{"site.yml", "main.yml", "playbook.yml", "default.yml"}
		for _, name := range commonNames {
			testPath := filepath.Join(tmpDir, name)
			playbookInputs = []string{testPath}
			break
		}

		if len(playbookInputs) == 0 {
			return fmt.Errorf("no playbook found in repository (looked for: %s)", strings.Join(commonNames, ", "))
		}
	}

	// Process extra-vars
	var extraVarsSlice []string
	if provisionExtraVars != "" {
		// Replace commas with spaces, then split by spaces
		varsString := strings.ReplaceAll(provisionExtraVars, ",", " ")
		extraVarsSlice = strings.Fields(varsString)
	}

	// Use shared provisioning pipeline
	return runProvisioningPipeline(ProvisioningPipelineOptions{
		DistroName:     distroName,
		PlaybookInputs: playbookInputs,
		Tags:           provisionTags,
		Verbose:        provisionVerbose,
		ExtraVars:      extraVarsSlice,
		TempDir:        tempDir,
	})
}
