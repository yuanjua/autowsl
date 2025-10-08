package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set during build time
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "autowsl",
	Short: "AutoWSL - Automatically download and manage WSL distributions",
	Long: `AutoWSL is a CLI tool to interactively select, download, and install 
WSL distributions from official sources.`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
}
