package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var aliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "List available playbook aliases",
	Long: `List all available playbook aliases that can be used with --playbooks.
These are the built-in playbooks located in the playbooks/ directory.

Examples:
  autowsl aliases
  autowsl install "Ubuntu 22.04 LTS" --playbooks curl
  autowsl provision ubuntu-2204 --playbooks curl,default`,
	RunE: runAliases,
}

func init() {
	rootCmd.AddCommand(aliasesCmd)
}

func runAliases(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	playbooksDir := filepath.Join(cwd, "playbooks")

	// Check if playbooks directory exists
	if _, err := os.Stat(playbooksDir); os.IsNotExist(err) {
		fmt.Println("No playbooks directory found.")
		fmt.Printf("Create one at: %s\n", playbooksDir)
		return nil
	}

	// List all .yml and .yaml files
	entries, err := os.ReadDir(playbooksDir)
	if err != nil {
		return fmt.Errorf("failed to read playbooks directory: %w", err)
	}

	var playbooks []struct {
		Alias string
		Path  string
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			alias := strings.TrimSuffix(strings.TrimSuffix(name, ".yml"), ".yaml")
			playbooks = append(playbooks, struct {
				Alias string
				Path  string
			}{
				Alias: alias,
				Path:  filepath.Join(playbooksDir, name),
			})
		}
	}

	if len(playbooks) == 0 {
		fmt.Println("No playbook aliases found in playbooks/ directory.")
		return nil
	}

	fmt.Println("\nAvailable Playbook Aliases:")
	fmt.Println(strings.Repeat("=", 60))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ALIAS\tFILE")
	fmt.Fprintln(w, strings.Repeat("-", 20)+"\t"+strings.Repeat("-", 35))

	for _, pb := range playbooks {
		fmt.Fprintf(w, "%s\t%s\n", pb.Alias, filepath.Base(pb.Path))
	}

	w.Flush()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total: %d playbook(s)\n\n", len(playbooks))
	fmt.Println("Usage:")
	fmt.Println("  autowsl install \"Ubuntu 22.04 LTS\" --playbooks <alias>")
	fmt.Println("  autowsl provision <distro> --playbooks <alias1>,<alias2>")
	fmt.Println()

	return nil
}
