# AutoWSL 🐧✨

A powerful CLI tool to automate the installation, management, and provisioning of your WSL environment.

This tool simplifies the entire WSL lifecycle, from installing a new distribution to a custom location, to automatically configuring it with Ansible playbooks.

## Installation

Make sure WSL is already installed.
```bash
wsl --install --no-distribution
```

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

**Installation**: You will be prompted to install a distribution to a custom installation path.

```bash
./autowsl.exe install
```

**Provisioning**: Setup environment with Ansible playbooks. You will be prompted to select an installed WSL distribution and enter playbook(s) to run (e.g. ssh)

```bash
./autowsl.exe provision
```

### Non-Interactive Install

For scripting or power users, you can provide all the details as command-line flags. This allows for one-command, end-to-end setup and provisioning. 

```bash
# Example: Install Ubuntu 22.04, name it "dev-box", install to D:\WSL, and run the selected playbooks
./autowsl.exe install "Ubuntu" --name ub-dev --path ./wsl-distros/ub-dev --playbooks systemd,ssh --extra-vars ssh_port=2224
```

Need to restart in order to enable systemd:
```bash
wsl --terminate <distro>
```

After running the above example, SSH service is installed by the playbook:

```bash
root@yuanzhoupc:/mnt/e/work/projects/autowsl# systemctl status ssh.service 
● ssh.service - OpenBSD Secure Shell server
     Loaded: loaded (/lib/systemd/system/ssh.service; enabled; vendor preset: enabled)
     Active: active (running) since Thu 2025-10-09 18:15:18 CST; 11s ago
       Docs: man:sshd(8)
             man:sshd_config(5)
   Main PID: 163 (sshd)
      Tasks: 1 (limit: 19078)
     Memory: 5.3M
        CPU: 22ms
     CGroup: /system.slice/ssh.service
             └─163 "sshd: /usr/sbin/sshd -D [listener] 0 of 10-100 startups"

Oct 09 18:15:18 yuanzhoupc systemd[1]: Starting OpenBSD Secure Shell server...
Oct 09 18:15:18 yuanzhoupc sshd[163]: Server listening on 0.0.0.0 port 2224.
Oct 09 18:15:18 yuanzhoupc sshd[163]: Server listening on :: port 2224.
Oct 09 18:15:18 yuanzhoupc systemd[1]: Started OpenBSD Secure Shell server.
```

Install from file:

```bash
./autowsl.exe install --from welcome-to-docker.tar --name docker-welcome --path ./wsl-distros/docker-test
```
This command shows using `--from` to create distribution from local `.tar` file.

**Provision existing distribution:**

```bash
# Provision with built-in alias with extra variables
./autowsl.exe provision ubuntu-2204 --playbooks curl,systemd,ssh --extra-vars ssh_port=2224

# Provision with local playbook file
./autowsl.exe provision debian --playbooks ./my-setup.yml

# Provision with Ansible tags
./autowsl.exe provision ubuntu-2204 --playbooks ./setup.yml --tags docker,nodejs
```

### Other Commands

- `autowsl list`: See all your installed WSL distributions
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
