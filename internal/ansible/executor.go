package ansible

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// packageManager contains information about available package managers.
type packageManager struct {
	name                   string
	checkCmd               string
	installCmd             string // %s will be replaced with the package name
	updateCmd              string
	preInstallSteps        []string // Commands to run before installing ANY package
	ansiblePostInstallCmds []string // Specific commands to run AFTER installing Ansible
	isAnsibleCore          bool     // True if the package manager installs ansible-core instead of ansible
	description            string
}

var (
	// memoizedPMs stores the detected package manager for each distro to avoid repeated detection.
	memoizedPMs = make(map[string]*packageManager)
	pmMutex     sync.Mutex

	// supportedPMs is a list of package managers the tool knows how to use.
	supportedPMs = []packageManager{
		{
			name:       "apt",
			checkCmd:   "command -v apt-get",
			updateCmd:  "sudo apt-get update",
			installCmd: "sudo apt-get install -y %s",
			preInstallSteps: []string{
				// This is required for Ansible's `apt` module to function correctly.
				"sudo apt-get install -y python3-apt",
			},
			description: "Ubuntu/Debian/Kali",
		},
		{
			name:            "dnf",
			checkCmd:        "command -v dnf",
			installCmd:      "sudo dnf install -y %s",
			preInstallSteps: []string{"sudo dnf install -y oracle-epel-release-el9 || true"}, // For Oracle/RHEL to get ansible
			isAnsibleCore:   true,
			description:     "Fedora/Oracle Linux/RHEL 8+",
		},
		{
			name:            "yum",
			checkCmd:        "command -v yum",
			installCmd:      "sudo yum install -y %s",
			preInstallSteps: []string{"sudo yum install -y epel-release || true"},
			description:     "RHEL/CentOS/Oracle Linux 7",
		},
		{
			name:       "zypper",
			checkCmd:   "command -v zypper",
			installCmd: "sudo zypper --non-interactive install -y %s",
			ansiblePostInstallCmds: []string{
				// This is required for Ansible's `systemd` module on SUSE systems.
				"sudo zypper --non-interactive install -y python3-dbus-python",
				// This is required for Ansible's `zypper` module (install without sudo for user).
				"ansible-galaxy collection install community.general --force",
			},
			description: "openSUSE",
		},
		{
			name:        "pacman",
			checkCmd:    "command -v pacman",
			installCmd:  "sudo pacman -S --noconfirm %s",
			description: "Arch Linux",
		},
		{
			name:        "apk",
			checkCmd:    "command -v apk",
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
	// Use POSIX sh to avoid reliance on bash (e.g., Alpine images)
	cmd := exec.Command("wsl.exe", "-d", distroName, "sh", "-c", command)
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
		// Use sh for robust availability across distros
		checkPMCmd := exec.Command("wsl.exe", "-d", distroName, "sh", "-c", pm.checkCmd)
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
	checkKaliCmd := exec.Command("wsl.exe", "-d", distroName, "sh", "-c", "grep -i kali /etc/os-release")
	if checkKaliCmd.Run() != nil {
		return nil // Not a Kali distribution, nothing to do.
	}

	fmt.Println("Detected Kali Linux, attempting to fix repositories...")

	// Step 1: Backup the original sources.list and create a new one with proper signed-by configuration
	backupCmd := "sudo cp /etc/apt/sources.list /etc/apt/sources.list.bak 2>/dev/null || true"
	if err := runWslCommand(distroName, backupCmd); err != nil {
		fmt.Printf("Warning: failed to backup sources.list: %v\n", err)
	}

	// Step 2: Comment out the old repositories and add the new signed repository
	updateSourcesCmd := `sudo sh -c 'sed -i "s/^deb/#deb/g" /etc/apt/sources.list && echo "deb [signed-by=/usr/share/keyrings/kali-archive-keyring.gpg] https://kali.download/kali kali-rolling main contrib non-free non-free-firmware" >> /etc/apt/sources.list'`
	if err := runWslCommand(distroName, updateSourcesCmd); err != nil {
		return fmt.Errorf("failed to update sources.list: %w", err)
	}

	// Step 3: Download the Kali archive keyring
	downloadKeyCmd := "wget -q https://archive.kali.org/archive-keyring.gpg -O /tmp/kali-archive-keyring.gpg && sudo mv /tmp/kali-archive-keyring.gpg /usr/share/keyrings/kali-archive-keyring.gpg"
	if err := runWslCommand(distroName, downloadKeyCmd); err != nil {
		return fmt.Errorf("failed to download Kali archive keyring: %w", err)
	}

	// Step 4: Update package lists with the new configuration
	updateCmd := "sudo apt-get update"
	if err := runWslCommand(distroName, updateCmd); err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}

	// Step 5: Install gnupg which is required for repository management
	installGnupgCmd := "sudo apt-get install -y gnupg"
	if err := runWslCommand(distroName, installGnupgCmd); err != nil {
		return fmt.Errorf("failed to install gnupg: %w", err)
	}

	fmt.Println("Kali repositories fixed and updated successfully.")
	return nil
}

// InstallPackage ensures a package is installed in the WSL distribution.
func InstallPackage(distroName, packageName string) error {
	pm, err := detectPackageManager(distroName)
	if err != nil {
		return err
	}

	// Run pre-installation steps if any (e.g., installing python3-apt).
	if len(pm.preInstallSteps) > 0 {
		fmt.Printf("Running pre-installation steps for %s...\n", pm.name)
		for _, step := range pm.preInstallSteps {
			if err := runWslCommand(distroName, step); err != nil {
				// A failure in a pre-install step is critical.
				return fmt.Errorf("pre-install step '%s' failed: %w", step, err)
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
	if err := runWslCommand(distroName, installCmdStr); err != nil {
		return err // The main install failed, so abort.
	}

	// Run post-installation steps specifically for Ansible, if defined.
	if packageName == "ansible" && len(pm.ansiblePostInstallCmds) > 0 {
		fmt.Printf("Running Ansible post-installation steps for %s...\n", pm.name)
		for _, step := range pm.ansiblePostInstallCmds {
			if err := runWslCommand(distroName, step); err != nil {
				// Post-install steps are critical for module functionality.
				return fmt.Errorf("ansible post-install step ('%s') failed: %w", step, err)
			}
		}
	}

	return nil
}

// ensurePackage checks if a command exists and installs the corresponding package if it doesn't.
func ensurePackage(distroName, commandName, packageName string) error {
	// Prefer POSIX 'command -v' over external 'which'
	checkCmd := exec.Command("wsl.exe", "-d", distroName, "sh", "-c", "command -v "+commandName)
	alreadyInstalled := checkCmd.Run() == nil

	if alreadyInstalled {
		fmt.Printf("Package '%s' is already installed.\n", packageName)

		// Even if Ansible is installed, ensure post-install steps have run (for SUSE, etc.)
		if packageName == "ansible" {
			pm, err := detectPackageManager(distroName)
			if err == nil && len(pm.ansiblePostInstallCmds) > 0 {
				// Check if community.general collection is installed (for SUSE)
				if pm.name == "zypper" {
					checkCollection := exec.Command("wsl.exe", "-d", distroName, "sh", "-c",
						"ansible-galaxy collection list | grep -q community.general")
					if checkCollection.Run() != nil {
						fmt.Println("Ansible collection 'community.general' not found, installing...")
						for _, step := range pm.ansiblePostInstallCmds {
							if err := runWslCommand(distroName, step); err != nil {
								return fmt.Errorf("ansible post-install step ('%s') failed: %w", step, err)
							}
						}
					}
				}
			}
		}
		return nil
	}

	// Handle repository preparation before trying to install.
	pm, err := detectPackageManager(distroName)
	if err != nil {
		return err // Could not detect a PM, cannot proceed.
	}

	if pm.name == "apt" {
		// This will fix Kali repos and run 'apt-get update'.
		if err := fixKaliRepositories(distroName); err != nil {
			return err
		}

		// Check if it's NOT Kali so we can run a standard update for Debian/Ubuntu.
		isKaliCmd := exec.Command("wsl.exe", "-d", distroName, "sh", "-c", "grep -i kali /etc/os-release")
		if isKaliCmd.Run() != nil {
			// It wasn't Kali, so no update has been run yet.
			fmt.Println("Running apt-get update...")
			if err := runWslCommand(distroName, pm.updateCmd); err != nil {
				// If apt-get update fails, try to fix broken repositories
				fmt.Printf("Warning: apt-get update failed, attempting to fix broken sources...\n")
				fixCmd := "sudo sed -i '/bullseye-backports/d' /etc/apt/sources.list /etc/apt/sources.list.d/* 2>/dev/null || true"
				_ = runWslCommand(distroName, fixCmd)

				// Try update again after fixing
				if err := runWslCommand(distroName, pm.updateCmd); err != nil {
					return fmt.Errorf("apt-get update failed even after attempting to fix broken sources: %w", err)
				}
				fmt.Println("Successfully fixed broken repositories and updated package lists.")
			}
		}
	}

	return InstallPackage(distroName, packageName)
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
	writeCmd := exec.Command("wsl.exe", "-d", distroName, "sh", "-c", writeCmdStr)
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
