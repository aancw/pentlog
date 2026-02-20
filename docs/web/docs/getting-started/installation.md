# Installation

Get PentLog installed and configured on your system. Choose your preferred installation method.

## System Requirements

Before installing, ensure your system meets these requirements:

| Component | Requirement |
|-----------|-------------|
| **Operating System** | macOS 10.15+ or Linux (Ubuntu, Fedora, Alpine, etc.) |
| **Go Version** | 1.24.0 or later (for building from source) |
| **Shell** | Bash, Zsh, or compatible POSIX shell |
| **Disk Space** | ~50 MB for installation, plus space for logs |
| **Dependencies** | `ttyrec` (auto-installed by `pentlog setup`) |

## Installation Methods

=== "Quick Install (Recommended)"

    The fastest way to install PentLog on macOS or Linux:

    ```bash
    curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh
    ```

    This script will:

    - Download the latest release binary
    - Install to `~/.local/bin/pentlog`
    - Add to your PATH if needed

    !!! tip "Verify the installation"
        After installation, run:
        ```bash
        pentlog version
        ```

=== "Homebrew (macOS)"

    Coming soon! For now, use the quick install method:

    ```bash
    curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh
    ```

=== "Build from Source"

    Build PentLog from the latest source code:

    ```bash
    # Clone the repository
    git clone https://github.com/aancw/pentlog.git
    cd pentlog

    # Build the binary
    go build -o pentlog main.go

    # Install system-wide (optional)
    sudo mv pentlog /usr/local/bin/
    ```

    !!! note "Go Version Required"
        Ensure Go 1.24.0+ is installed:
        ```bash
        go version
        ```

    ### Cross-Compile

    | Target | Command |
    |--------|---------|
    | Linux ARM64 | `GOOS=linux GOARCH=arm64 go build -o pentlog main.go` |
    | macOS AMD64 | `GOOS=darwin GOARCH=amd64 go build -o pentlog main.go` |
    | macOS ARM64 | `GOOS=darwin GOARCH=arm64 go build -o pentlog main.go` |

=== "Docker"

    Run PentLog in a container:

    ```bash
    docker pull aancw/pentlog:latest
    docker run -it --rm -v ~/.pentlog:/root/.pentlog aancw/pentlog
    ```

## Post-Installation Setup

After installing PentLog, run the setup command to:

- Verify system dependencies
- Install `ttyrec` if missing
- Create the `~/.pentlog/` directory structure

```bash
pentlog setup
```

You should see output like:

```
✓ PentLog Setup Complete
  • ttyrec: /usr/local/bin/ttyrec
  • ttyplay: /usr/local/bin/ttyplay
  • Database: ~/.pentlog/pentlog.db
  • Logs: ~/.pentlog/logs/
```

## Manual Dependency Installation

If automatic dependency installation fails, install `ttyrec` manually:

=== "macOS"

    ```bash
    brew install ttyrec
    ```

=== "Ubuntu / Debian"

    ```bash
    sudo apt-get update
    sudo apt-get install ttyrec
    ```

=== "Fedora"

    ```bash
    sudo dnf install ttyrec
    ```

=== "Alpine Linux"

    ```bash
    sudo apk add ttyrec
    ```

=== "Arch Linux"

    ```bash
    sudo pacman -S ttyrec
    ```

## Shell Completion

Generate auto-completion for your shell:

```bash
pentlog completion
```

Follow the prompts to install for Bash or Zsh.

## Verify Installation

Confirm PentLog is working correctly:

```bash
# Check version
pentlog version

# Show status
pentlog status

# Test help
pentlog --help
```

## Troubleshooting

### "command not found: pentlog"

The installation directory isn't in your PATH. Add it:

```bash
# For Bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For Zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### "ttyrec not found"

Install `ttyrec` manually using the commands above for your OS, then re-run:

```bash
pentlog setup
```

### Permission denied on macOS

If you see "cannot be opened because the developer cannot be verified":

```bash
# Remove the quarantine attribute
xattr -d com.apple.quarantine ~/.local/bin/pentlog
```

### Build errors from source

Ensure you have Go 1.24.0+:

```bash
# Check Go version
go version

# If needed, update Go (macOS example)
brew update
brew upgrade go
```

## Security Best Practices

!!! warning "Password Protection"
    Use interactive mode (`pentlog archive`) instead of `--password` flag to avoid storing passwords in shell history.

!!! info "File Permissions"
    Sensitive files are created with `0600` permissions automatically.

!!! tip "Evidence Integrity"
    Use `pentlog freeze` before archiving for compliance audits.

## Next Steps

Now that PentLog is installed, continue to the [Quick Start](quickstart.md) guide to create your first engagement and start logging.

[:octicons-arrow-right-24: Quick Start Guide](quickstart.md){ .md-button .md-button--primary }
