package ansible

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// playbookManager contains information about available package managers.
type packageManager struct {
	name             string
	checkCmd         string
	installCmd       string // %s will be replaced with the package name
	updateCmd        string
	preInstallSteps  []string // Commands to run before installing a package
	isAnsibleCore    bool     // True if the package manager installs ansible-core instead of ansible
	description      string
}

var (
	// memoizedPMs stores the detected package manager for each distro to avoid repeated detection.
	memoizedPMs = make(map[string]*packageManager)
	pmMutex     sync.Mutex

	// supportedPMs is a list of package managers the tool knows how to use.
	supportedPMs = []packageManager{
		{
			name:        "apt",
			checkCmd:    "which apt-get",
			updateCmd:   "sudo apt-get update",
			installCmd:  "sudo apt-get install -y %s",
			description: "Ubuntu/Debian/Kali",
		},
		{
			name:            "dnf",
			checkCmd:        "which dnf",
			installCmd:      "sudo dnf install -y %s",
			preInstallSteps: []string{"sudo dnf install -y oracle-epel-release-el9 || true"}, // For Oracle/RHEL to get ansible
			isAnsibleCore:   true,
			description:     "Fedora/Oracle Linux/RHEL 8+",
		},
		{
			name:            "yum",
			checkCmd:        "which yum",
			installCmd:      "sudo yum install -y %s",
			preInstallSteps: []string{"sudo yum install -y epel-release || true"},
			description:     "RHEL/CentOS/Oracle Linux 7",
		},
		{
			name:        "zypper",
			checkCmd:    "which zypper",
			installCmd:  "sudo zypper install -y %s",
			description: "openSUSE",
		},
		{
			name:        "pacman",
			checkCmd:    "which pacman",
			installCmd:  "sudo pacman -S --noconfirm %s",
			description: "Arch Linux",
		},
		{
			name:        "apk",
			checkCmd:    "which apk",
			installCmd:  "sudo apk add %s",
			description: "Alpine Linux",
		},
	}
)

// PlaybookOptions holds options for playbook execution.
type PlaybookOptions struct {
	DistroName   string
	PlaybookPath string
	Tags         []string
	Verbose      bool
	ExtraVars    map[string]string
}

// runWslCommand executes a command within a specified WSL distribution and streams its output.
func runWslCommand(distroName, command string) error {
	cmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command '%s' failed: %w", command, err)
	}
	return nil
}

// detectPackageManager identifies the package manager used by the distribution.
func detectPackageManager(distroName string) (*packageManager, error) {
	pmMutex.Lock()
	defer pmMutex.Unlock()

	if pm, ok := memoizedPMs[distroName]; ok {
		return pm, nil
	}

	for i := range supportedPMs {
		pm := &supportedPMs[i]
		checkPMCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", pm.checkCmd)
		if checkPMCmd.Run() == nil {
			fmt.Printf("Detected package manager: %s (%s)\n", pm.name, pm.description)
			memoizedPMs[distroName] = pm
			return pm, nil
		}
	}

	return nil, fmt.Errorf("could not detect a supported package manager in distribution '%s'", distroName)
}

// fixKaliRepositories handles the specific GPG key issue in new Kali Linux instances.
func fixKaliRepositories(distroName string) error {
	checkKaliCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "grep -i kali /etc/os-release")
	if checkKaliCmd.Run() != nil {
		return nil // Not a Kali distribution, nothing to do.
	}

	fmt.Println("Detected Kali Linux, attempting to fix repositories...")

	// Install the keyring, allowing it to be unauthenticated since that's the problem we're solving.
	keyringCmd := "sudo apt-get install -y --allow-unauthenticated kali-archive-keyring"
	if err := runWslCommand(distroName, keyringCmd); err != nil {
		return fmt.Errorf("failed to install kali-archive-keyring: %w", err)
	}

	// Update repositories again with the correct keys in place.
	if err := runWslCommand(distroName, "sudo apt-get update"); err != nil {
		return fmt.Errorf("failed to update repositories after fixing keyring: %w", err)
	}

	fmt.Println("Kali repositories fixed successfully.")
	return nil
}

// InstallPackage ensures a package is installed in the WSL distribution.
func InstallPackage(distroName, packageName string) error {
	pm, err := detectPackageManager(distroName)
	if err != nil {
		return err
	}

	// Run pre-installation steps if any (e.g., enabling EPEL repo).
	if len(pm.preInstallSteps) > 0 {
		fmt.Printf("Running pre-installation steps for %s...\n", pm.name)
		for _, step := range pm.preInstallSteps {
			if err := runWslCommand(distroName, step); err != nil {
				fmt.Printf("Warning: pre-install step failed, continuing anyway: %v\n", err)
			}
		}
	}

	// Special handling for Ansible package name variations
	installPkgName := packageName
	if packageName == "ansible" && pm.isAnsibleCore {
		installPkgName = "ansible-core"
	}

	// Run the installation command.
	installCmdStr := fmt.Sprintf(pm.installCmd, installPkgName)
	fmt.Printf("Installing '%s' with %s...\n", installPkgName, pm.name)
	return runWslCommand(distroName, installCmdStr)
}

// ensurePackage checks if a command exists and installs the corresponding package if it doesn't.
func ensurePackage(distroName, commandName, packageName string) error {
	checkCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", "which "+commandName)
	if err := checkCmd.Run(); err == nil {
		fmt.Printf("Package '%s' is already installed.\n", packageName)
		return nil
	}

	// Handle special cases before generic installation.
	if pm, err := detectPackageManager(distroName); err == nil && pm.name == "apt" {
		if err := fixKaliRepositories(distroName); err != nil {
			fmt.Printf("Warning: Failed to fix Kali repositories, installation may fail: %v\n", err)
		}
		// Always run an update for apt-based systems before first install.
		fmt.Println("Running apt-get update...")
		_ = runWslCommand(distroName, "sudo apt-get update")
	}

	if err := InstallPackage(distroName, packageName); err != nil {
		return fmt.Errorf("failed to install package '%s': %w", packageName, err)
	}
	return nil
}

// ExecutePlaybook runs an Ansible playbook inside a WSL distribution.
func ExecutePlaybook(opts PlaybookOptions) error {
	if _, err := os.Stat(opts.PlaybookPath); err != nil {
		return fmt.Errorf("playbook file '%s' not found: %w", opts.PlaybookPath, err)
	}

	fmt.Printf("Playbook: %s\n", filepath.Base(opts.PlaybookPath))
	fmt.Printf("Target:   %s\n", opts.DistroName)
	if len(opts.Tags) > 0 {
		fmt.Printf("Tags:     %s\n", strings.Join(opts.Tags, ", "))
	}
	fmt.Println()

	if err := ensurePackage(opts.DistroName, "ansible-playbook", "ansible"); err != nil {
		return fmt.Errorf("failed to ensure Ansible is installed: %w", err)
	}

	wslPlaybookPath, err := copyPlaybookToWSL(opts.DistroName, opts.PlaybookPath)
	if err != nil {
		return fmt.Errorf("failed to copy playbook to WSL: %w", err)
	}

	ansibleCmd := buildAnsibleCommand(wslPlaybookPath, opts)
	fmt.Println("Executing playbook...")
	fmt.Println(strings.Repeat("-", 60))

	if err := runWslCommand(opts.DistroName, ansibleCmd); err != nil {
		return fmt.Errorf("playbook '%s' execution failed: %w", filepath.Base(opts.PlaybookPath), err)
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("Playbook execution completed.")
	return nil
}

// copyPlaybookToWSL copies a playbook from Windows to the WSL filesystem.
func copyPlaybookToWSL(distroName, windowsPlaybookPath string) (string, error) {
	wslPlaybookPath := "/tmp/autowsl-playbook.yml"
	content, err := os.ReadFile(windowsPlaybookPath)
	if err != nil {
		return "", fmt.Errorf("failed to read playbook '%s': %w", windowsPlaybookPath, err)
	}

	writeCmdStr := fmt.Sprintf("cat > '%s' && chmod 644 '%s'", wslPlaybookPath, wslPlaybookPath)
	writeCmd := exec.Command("wsl.exe", "-d", distroName, "bash", "-c", writeCmdStr)
	writeCmd.Stdin = strings.NewReader(string(content))

	if output, err := writeCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to copy playbook to WSL filesystem: %s: %w", string(output), err)
	}

	return wslPlaybookPath, nil
}

// buildAnsibleCommand constructs the full ansible-playbook command string.
func buildAnsibleCommand(playbookPath string, opts PlaybookOptions) string {
	var cmd strings.Builder
	cmd.WriteString(fmt.Sprintf("ansible-playbook %s --connection=local -i localhost,", playbookPath))

	if len(opts.Tags) > 0 {
		cmd.WriteString(fmt.Sprintf(" --tags %s", strings.Join(opts.Tags, ",")))
	}

	if opts.Verbose {
		cmd.WriteString(" -vvv")
	}

	if len(opts.ExtraVars) > 0 {
		var vars []string
		for k, v := range opts.ExtraVars {
			vars = append(vars, fmt.Sprintf("%s=%s", k, v))
		}
		// Use single quotes to handle spaces and other special characters in values.
		cmd.WriteString(fmt.Sprintf(" --extra-vars '%s'", strings.Join(vars, " ")))
	}

	return cmd.String()
}

// CloneGitRepo clones a git repository into a specified directory in the WSL distribution.
func CloneGitRepo(distroName, repoURL, destDir string) error {
	fmt.Printf("Cloning repository: %s\n", repoURL)
	if err := ensurePackage(distroName, "git", "git"); err != nil {
		return fmt.Errorf("failed to ensure git is installed: %w", err)
	}

	cloneCmdStr := fmt.Sprintf("git clone %s %s", repoURL, destDir)
	if err := runWslCommand(distroName, cloneCmdStr); err != nil {
		return fmt.Errorf("failed to clone repository '%s': %w", repoURL, err)
	}

	fmt.Println("Repository cloned successfully.")
	return nil
}
