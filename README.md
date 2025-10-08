# AutoWSL 🐧✨

A powerful CLI tool to automate the installation, management, and provisioning of your WSL environment.

This tool simplifies the entire WSL lifecycle, from installing a new distribution to a custom location, to automatically configuring it with Ansible playbooks.

## Installation

To get started, you'll need Go (version 1.21+) installed.

**Clone the repository:**

```bash
git clone https://github.com/yuanjua/autowsl.git
```

**Build the executable:**

```bash
cd autowsl
go build -o autowsl.exe .
```

**(Optional)** Move `autowsl.exe` to a directory in your system's PATH for easy access.

## Usage

### Simple Interactive Install

**Installation**: This is the easiest way to install a new WSL distribution. The tool will guide you through every step.

```bash
./autowsl.exe install
```

You will be prompted to:
- Select a distribution from the official catalog
- Provide a custom name for your new environment
- Choose a custom installation path (perfect for non-C: drives)

**Provisioning**: Configure an already-installed distribution with Ansible playbooks. The tool will guide you through selecting a distribution and playbook.

```bash
./autowsl.exe provision
```

You will be prompted to:
- Select an installed WSL distribution
- Enter playbook(s) to run (default: `curl`)

Playbooks can be specified as:
- **Built-in alias**: `curl` (installs curl on all supported distros)
- **Local file**: `./my-setup.yml` or `/full/path/to/playbook.yml`
- **URL**: `https://example.com/playbook.yml`
- **Git repository**: Use `--repo` flag to clone and run playbooks
- **Multiple**: `curl,./dev.yml` (comma-separated)

### Non-Interactive Install

For scripting or power users, you can provide all the details as command-line flags. This allows for one-command, end-to-end setup and provisioning.

```bash
# Example: Install Ubuntu 22.04, name it "dev-box", install to D:\WSL, and run the 'curl' playbook
./autowsl.exe install "Ubuntu 22.04 LTS" --name dev-box --path D:\WSL\dev --playbooks curl
```

- `"Ubuntu 22.04 LTS"`: The distro to install from the catalog
- `--name`: Your custom name for the WSL instance
- `--path`: The custom installation directory
- `--playbooks`: A comma-separated list of built-in (aliases) or custom playbooks to run after installation

```bash
./autowsl.exe install --from welcome-to-docker.tar --name docker-welcome --path ./wsl-distros/docker-test
```
This command shows using `--from` to create distribution from local `.tar` file.

**Provision existing distribution:**

```bash
# Provision with built-in alias
./autowsl.exe provision ubuntu-2204 --playbooks curl

# Provision with local playbook file
./autowsl.exe provision debian --playbooks ./my-setup.yml

# Provision with remote playbook URL
./autowsl.exe provision kali-linux --playbooks https://raw.githubusercontent.com/user/repo/main/setup.yml

# Provision from Git repository
./autowsl.exe provision ubuntu-2204 --repo https://github.com/user/ansible-playbooks

# Provision with multiple playbooks
./autowsl.exe provision ubuntu-2204 --playbooks curl,./dev.yml,./docker.yml

# Provision with Ansible tags
./autowsl.exe provision ubuntu-2204 --playbooks ./setup.yml --tags docker,nodejs

# Provision with extra variables
./autowsl.exe provision debian --playbooks ./setup.yml --extra-vars user=john --extra-vars env=dev
```

### Other Commands

- `autowsl list`: See all your installed WSL distributions
- `autowsl provision <name>`: Configure an existing distro with Ansible playbooks
- `autowsl aliases`: List available built-in playbooks (PRs are welcome)
- `autowsl -h`: For more details

## For Developers

Interested in improving autowsl? Here's how to get started.

**Build the project:**

```bash
go build -o autowsl.exe .
```

**Run pre-commit:**

```bash
make pre-commit
```

## Contributing

Contributions are welcome!

1. Fork the repository
2. Create a new feature branch (`git checkout -b feature/my-new-feature`)
3. Make your changes
4. Run the tests to ensure everything still works
5. Commit your changes and open a Pull Request

## License

This project is licensed under the MIT License.
