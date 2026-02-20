# Getting Started

Welcome to PentLog! This guide will help you get up and running in under 5 minutes.

## What You'll Learn

<div class="grid cards" markdown>

-   :material-rocket-launch: __Quick Start__

    ---

    Get PentLog installed and running with your first engagement in just a few commands.

    [:octicons-arrow-right-24: Quick Start](quickstart.md)

-   :material-download: __Installation__

    ---

    Detailed installation instructions for all platforms and dependency setup.

    [:octicons-arrow-right-24: Installation](installation.md)

-   :material-lightbulb: __Core Concepts__

    ---

    Understand PentLog's evidence-first design, context modes, and session organization.

    [:octicons-arrow-right-24: Core Concepts](concepts.md)

</div>

## System Requirements

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| **OS** | macOS 10.15+ / Linux kernel 4.0+ | Latest macOS / Ubuntu LTS |
| **Go** | 1.24.0+ | 1.24.0+ |
| **Disk Space** | 100 MB | 1 GB+ |
| **Memory** | 512 MB RAM | 2 GB+ RAM |
| **Shell** | Bash / Zsh | Zsh with plugins |

## Quick Install

=== "macOS / Linux (Recommended)"

    ```bash
    curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh
    ```

    This installs the latest release to `~/.local/bin/pentlog`.

=== "Build from Source"

    ```bash
    git clone https://github.com/aancw/pentlog.git
    cd pentlog
    go build -o pentlog main.go
    sudo mv pentlog /usr/local/bin/
    ```

## Verify Installation

```bash
# Check version
pentlog version

# Verify dependencies
pentlog setup
```

## Next Steps

1. **[Quick Start](quickstart.md)** — Run through a complete example
2. **[Installation](installation.md)** — Detailed platform-specific instructions
3. **[Core Concepts](concepts.md)** — Understand how PentLog works
