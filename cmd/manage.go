package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/yuanjua/autowsl/internal/wsl"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed WSL distributions",
	RunE:  runList,
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a WSL distribution",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRemove,
}

var backupCmd = &cobra.Command{
	Use:   "backup <name>",
	Short: "Backup a WSL distribution",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runBackup,
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(backupCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	distros, err := wsl.ListInstalledDistros()
	if err != nil {
		return fmt.Errorf("failed to list distributions: %w", err)
	}

	if len(distros) == 0 {
		fmt.Println("No WSL distributions are currently installed.")
		fmt.Println("\nInstall one using: autowsl install")
		return nil
	}

	fmt.Println("\nInstalled WSL Distributions:")
	fmt.Println()

	// Create a tabwriter for nice formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "DEFAULT\tNAME\tSTATE\tVERSION")
	fmt.Fprintln(w, "-------\t----\t-----\t-------")

	for _, d := range distros {
		defaultMarker := " "
		if d.Default {
			defaultMarker = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", defaultMarker, d.Name, d.State, d.Version)
	}

	w.Flush()
	fmt.Println()

	return nil
}

func runRemove(cmd *cobra.Command, args []string) error {
	distroName := args[0]

	// Check if the distribution exists
	exists, err := wsl.IsDistroInstalled(distroName)
	if err != nil {
		return fmt.Errorf("failed to check distribution: %w", err)
	}
	if !exists {
		return fmt.Errorf("distribution '%s' does not exist", distroName)
	}

	// Confirmation prompt
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure you want to remove '%s'? This action cannot be undone", distroName),
		IsConfirm: true,
	}

	_, err = prompt.Run()
	if err != nil {
		fmt.Println("Removal cancelled")
		return nil
	}

	fmt.Printf("\nRemoving '%s'...\n", distroName)

	if err := wsl.Unregister(distroName); err != nil {
		return fmt.Errorf("failed to remove distribution: %w", err)
	}

	fmt.Printf("Successfully removed '%s'\n", distroName)

	return nil
}

func runBackup(cmd *cobra.Command, args []string) error {
	distroName := args[0]

	// Check if the distribution exists
	exists, err := wsl.IsDistroInstalled(distroName)
	if err != nil {
		return fmt.Errorf("failed to check distribution: %w", err)
	}
	if !exists {
		return fmt.Errorf("distribution '%s' does not exist", distroName)
	}

	// Generate default backup filename
	homeDir, _ := os.UserHomeDir()
	defaultBackupPath := filepath.Join(homeDir, "WSL-Backups", fmt.Sprintf("%s-backup.tar", distroName))

	// Prompt for backup location
	prompt := promptui.Prompt{
		Label:   "Backup file path",
		Default: defaultBackupPath,
	}

	backupPath, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("failed to get backup path: %w", err)
	}

	fmt.Printf("\nBacking up '%s' to %s...\n", distroName, backupPath)
	fmt.Println("This may take a while depending on the size of your distribution...")

	if err := wsl.Export(distroName, backupPath); err != nil {
		return fmt.Errorf("failed to backup distribution: %w", err)
	}

	// Get file size
	fileInfo, _ := os.Stat(backupPath)
	sizeInMB := float64(fileInfo.Size()) / 1024 / 1024

	fmt.Printf("\nSuccessfully backed up '%s'\n", distroName)
	fmt.Printf("Location: %s\n", backupPath)
	fmt.Printf("Size: %.2f MB\n", sizeInMB)

	return nil
}
