package ansible

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PlaybookOptions holds options for playbook execution
type PlaybookOptions struct {
	DistroName   string
	PlaybookPath string
	Tags         []string
	Verbose      bool
	ExtraVars    map[string]string
}

// ExecutePlaybook runs an Ansible playbook inside a WSL distribution
func ExecutePlaybook(opts PlaybookOptions) error {
	// Check if playbook file exists
	if _, err := os.Stat(opts.PlaybookPath); err != nil {
		return fmt.Errorf("playbook file '%s' not found: %w", opts.PlaybookPath, err)
	}

	fmt.Printf("Playbook: %s\n", filepath.Base(opts.PlaybookPath))
	fmt.Printf("Target:   %s\n", opts.DistroName)
	if len(opts.Tags) > 0 {
		fmt.Printf("Tags:     %s\n", strings.Join(opts.Tags, ", "))
	}
	fmt.Println()

	// First, ensure Ansible is installed in the WSL distribution
	fmt.Println("Checking Ansible installation...")
	if err := ensureAnsible(opts.DistroName); err != nil {
		return fmt.Errorf("failed to ensure Ansible is installed in distribution '%s': %w", opts.DistroName, err)
	}

	// Copy playbook to WSL filesystem for reliable execution
	wslPlaybookPath, err := copyPlaybookToWSL(opts.DistroName, opts.PlaybookPath)
	if err != nil {
		return fmt.Errorf("failed to copy playbook to WSL: %w", err)
	}

	// Build ansible-playbook command
	cmdArgs := []string{
		"-d", opts.DistroName,
		"bash", "-c",
		buildAnsibleCommand(wslPlaybookPath, opts),
	}

	fmt.Println("Executing playbook...")
	fmt.Println(strings.Repeat("-", 60))

	// Execute the command
	cmd := exec.Command("wsl.exe", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("playbook '%s' execution failed on distribution '%s': %w", filepath.Base(opts.PlaybookPath), opts.DistroName, err)
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("Playbook execution completed.")

	return nil
}

// ensureAnsible checks if Ansible is installed, and installs it if not
func ensureAnsible(distroName string) error {
	// Check if ansible is installed
	checkCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "which ansible-playbook")
	if err := checkCmd.Run(); err == nil {
		fmt.Println("Ansible is already installed")
		return nil
	}

	fmt.Println("Ansible not found, installing...")

	// Detect which package manager is available
	packageManagers := []struct {
		name        string
		checkCmd    string
		installCmd  string
		description string
	}{
		{"apt", "which apt-get", "sudo sed -i '/bullseye-backports/d' /etc/apt/sources.list 2>/dev/null; sudo apt-get update && sudo apt-get install -y ansible", "Ubuntu/Debian"},
		{"dnf", "which dnf", "sudo dnf install -y epel-release && sudo dnf install -y ansible", "Fedora/Oracle Linux/RHEL 8+"},
		{"yum", "which yum", "sudo yum install -y epel-release && sudo yum install -y ansible", "RHEL/CentOS/Oracle Linux"},
		{"zypper", "which zypper", "sudo zypper install -y ansible", "openSUSE"},
		{"pacman", "which pacman", "sudo pacman -S --noconfirm ansible", "Arch Linux"},
		{"apk", "which apk", "sudo apk add ansible", "Alpine Linux"},
	}

	var lastErr error
	for _, pm := range packageManagers {
		// Check if this package manager exists
		checkPMCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", pm.checkCmd)
		if checkPMCmd.Run() != nil {
			// Package manager not found, skip
			continue
		}

		fmt.Printf("Detected package manager: %s (%s)\n", pm.name, pm.description)

		// Try to install Ansible with this package manager
		installCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", pm.installCmd)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		if err := installCmd.Run(); err == nil {
			// Check if installation succeeded
			checkCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "which ansible-playbook")
			if checkCmd.Run() == nil {
				fmt.Println("Ansible installed successfully")
				return nil
			}
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("failed to install Ansible automatically (tried all available package managers): %w\nPlease install manually: wsl -d %s bash -c 'sudo dnf install ansible' or 'sudo yum install ansible'", lastErr, distroName)
}

// copyPlaybookToWSL copies a playbook from Windows to WSL filesystem
func copyPlaybookToWSL(distroName, windowsPlaybookPath string) (string, error) {
	// Target path in WSL filesystem
	wslPlaybookPath := "/tmp/autowsl-playbook.yml"

	// Read the playbook content from Windows
	content, err := os.ReadFile(windowsPlaybookPath)
	if err != nil {
		return "", fmt.Errorf("failed to read playbook '%s': %w", windowsPlaybookPath, err)
	}

	// Write content directly to WSL filesystem using bash
	// This avoids path conversion issues by piping the content
	writeCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c",
		fmt.Sprintf("cat > '%s' && chmod 644 '%s'", wslPlaybookPath, wslPlaybookPath))

	// Pipe the content to stdin
	writeCmd.Stdin = strings.NewReader(string(content))

	if output, err := writeCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to copy playbook to WSL filesystem: %s: %w", string(output), err)
	}

	return wslPlaybookPath, nil
}

// buildAnsibleCommand builds the ansible-playbook command string
func buildAnsibleCommand(playbookPath string, opts PlaybookOptions) string {
	cmd := fmt.Sprintf("ansible-playbook %s", playbookPath)

	// Add connection local flag (run on localhost)
	cmd += " --connection=local"

	// Add inventory (localhost)
	cmd += " -i localhost,"

	// Add tags if specified
	if len(opts.Tags) > 0 {
		cmd += fmt.Sprintf(" --tags %s", strings.Join(opts.Tags, ","))
	}

	// Add verbosity
	if opts.Verbose {
		cmd += " -vvv"
	}

	// Add extra vars
	if len(opts.ExtraVars) > 0 {
		vars := make([]string, 0, len(opts.ExtraVars))
		for k, v := range opts.ExtraVars {
			vars = append(vars, fmt.Sprintf("%s=%s", k, v))
		}
		cmd += fmt.Sprintf(" --extra-vars '%s'", strings.Join(vars, " "))
	}

	return cmd
}

// CloneGitRepo clones a git repository into a temporary directory in WSL
func CloneGitRepo(distroName, repoURL, destDir string) error {
	fmt.Printf("Cloning repository: %s\n", repoURL)

	// Ensure git is installed
	checkCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "which git")
	if err := checkCmd.Run(); err != nil {
		fmt.Println("Git not found, installing...")

		// Detect which package manager is available
		packageManagers := []struct {
			name       string
			checkCmd   string
			installCmd string
		}{
			{"apt", "which apt-get", "sudo apt-get update -qq && sudo apt-get install -y -qq git"},
			{"dnf", "which dnf", "sudo dnf install -y git"},
			{"yum", "which yum", "sudo yum install -y git"},
			{"zypper", "which zypper", "sudo zypper install -y git"},
			{"pacman", "which pacman", "sudo pacman -S --noconfirm git"},
			{"apk", "which apk", "sudo apk add git"},
		}

		var installed bool
		for _, pm := range packageManagers {
			// Check if this package manager exists
			checkPMCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", pm.checkCmd)
			if checkPMCmd.Run() != nil {
				// Package manager not found, skip
				continue
			}

			fmt.Printf("Detected package manager: %s\n", pm.name)

			// Try to install git with this package manager
			installCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", pm.installCmd)
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr

			if err := installCmd.Run(); err == nil {
				// Verify installation
				checkCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "which git")
				if checkCmd.Run() == nil {
					installed = true
					break
				}
			}
		}

		if !installed {
			return fmt.Errorf("failed to install git in distribution '%s' (tried all available package managers)", distroName)
		}
	}

	// Clone the repository
	cloneCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c",
		fmt.Sprintf("git clone %s %s", repoURL, destDir))
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr

	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository '%s' to '%s' in distribution '%s': %w", repoURL, destDir, distroName, err)
	}

	fmt.Println("Repository cloned successfully")
	return nil
}
