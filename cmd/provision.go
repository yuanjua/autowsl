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
	provisionExtraVars []string
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
  autowsl provision ubuntu-2204 --playbooks ./setup.yml --extra-vars user=john --extra-vars env=dev

  # Verbose output
  autowsl provision ubuntu-2204 --verbose`,
	RunE: runProvision,
}

func init() {
	rootCmd.AddCommand(provisionCmd)
	provisionCmd.Flags().StringSliceVar(&provisionTags, "tags", nil, "Ansible tags to run (comma-separated)")
	provisionCmd.Flags().StringSliceVar(&provisionPlaybooks, "playbooks", []string{}, "Playbook files, URLs, or aliases (comma-separated or repeat flag)")
	provisionCmd.Flags().StringArrayVar(&provisionExtraVars, "extra-vars", nil, "Extra variables in key=val format (repeatable)")
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
	os.MkdirAll(tempDir, 0755)

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

	// Use shared provisioning pipeline
	return runProvisioningPipeline(ProvisioningPipelineOptions{
		DistroName:     distroName,
		PlaybookInputs: playbookInputs,
		Tags:           provisionTags,
		Verbose:        provisionVerbose,
		ExtraVars:      provisionExtraVars,
		TempDir:        tempDir,
	})
}
