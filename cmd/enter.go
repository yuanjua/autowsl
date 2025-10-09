package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/yuanjua/autowsl/internal/wsl"
)

var enterCmd = &cobra.Command{
	Use:   "enter [distro-name]",
	Short: "Enter a WSL distribution shell",
	Long: `Enter a WSL distribution shell interactively or by name.

Examples:
  # Interactive mode - select from installed distros
  autowsl enter

  # Enter a specific distribution by name
  autowsl enter ubuntu-2004-lts`,
	RunE: runEnter,
}

func init() {
	rootCmd.AddCommand(enterCmd)
}

func runEnter(cmd *cobra.Command, args []string) error {
	var distroName string
	var err error

	// Determine distro name - from args or interactive selection
	if len(args) > 0 {
		distroName = args[0]

		// Verify the distribution exists
		exists, err := wsl.IsDistroInstalled(distroName)
		if err != nil {
			return fmt.Errorf("failed to check distribution: %w", err)
		}
		if !exists {
			return fmt.Errorf("distribution '%s' is not installed", distroName)
		}
	} else {
		// Use interactive selection from installed distros
		distroName, err = selectInstalledDistroInteractive()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Entering '%s'...\n\n", distroName)

	// Execute wsl -d <distro-name>
	wslPath, err := exec.LookPath("wsl.exe")
	if err != nil {
		return fmt.Errorf("failed to find wsl.exe: %w", err)
	}

	// Create command to enter the WSL distribution
	wslCmd := exec.Command(wslPath, "-d", distroName)
	wslCmd.Stdin = os.Stdin
	wslCmd.Stdout = os.Stdout
	wslCmd.Stderr = os.Stderr

	err = wslCmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit with the same code as wsl
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to enter distribution: %w", err)
	}

	return nil
}
