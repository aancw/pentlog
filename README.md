# pentlog

![GitHub release (latest by date)](https://img.shields.io/github/v/release/aancw/pentlog)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/aancw/pentlog)
![GitHub all releases](https://img.shields.io/github/downloads/aancw/pentlog/total)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Evidence-First Pentest Logging Tool.
Captures shell activity as high-fidelity terminal logs backed by `ttyrec`.

<p align="center">
  <img src="pentlog.png" width="500">
</p>

## 📜 Table of Contents

- [✨ Features](#features)
- [⌨️ Command Reference](#command-reference)
- [🛠️ Installation](#installation)
- [🚀 Usage](#usage)
- [🧠 AI Analysis](#ai-analysis)
- [📦 Storage Layout](#storage-layout)
- [🗺️ Roadmap](#roadmap)
- [🤝 Contributing](#contributing)
- [📝 License](#license)


## Features

- **No Root Required**: Start recorded shells as a normal user; logs land in your home directory.
- **Context-Aware**: Tracks metadata and stamps every log. Flexible support for **Client Engagements** and **Exam/Labs** (OSCP, HTB, etc.).
- **Terminal-Perfect Logs**: Built-in **Virtual Terminal Emulator** guarantees that what you see in the search viewer matches exactly what you saw in your shell—preserving colors, handling overwrites/edits/redraws correctly, and eliminating ghost text.
- **Interactive Workflows**: Seamlessly create engagements, switch phases, and search logs using intuitive TUI menus.
- **Replayable**: Timing files enable faithful playback via `ttyplay`.
- **Export Friendly**: Export structured Markdown reports for any phase with an interactive preview/save menu.
- **Integrity Ready**: Freeze command hashes every log for evidence packaging.
- **AI Analysis**: Analyze your reports with AI to get a summary of the findings.
- **Shell Completion**: Generate and install shell completion scripts for bash and zsh.

## Command Reference

| Command | Description |
| :--- | :--- |
| `analyze` | Analyze a report with an AI provider to summarize findings. |
| `archive` | Archive old sessions to save space (Interactive). |
| `completion` | Generate auto-completion scripts for Zsh and Bash. |
| `create` | Initialize a new engagement context (Interactive). |
| `dashboard` | Show an interactive dashboard of your pentest activity. |
| `export` | Export commands for a specific phase (recon, exploit, etc.). |
| `freeze` | Generate SHA256 hashes of all session logs. |
| `note` | Manage session notes and bookmarks. |
| `replay` | Replay a recorded session with full fidelity (Interactive). |
| `reset` | Clear the current active engagement context. |
| `search` | Search command history across all sessions (supports Regex). |
| `sessions` | List and manage recorded sessions. |
| `setup` | Verify dependencies and prepare local logging. |
| `shell` | Start a recorded shell with the engagement context loaded. |
| `status` | Show current tool and engagement status. |
| `switch` | Switch to a different pentest phase (Interactive/History). |
| `update` | Update pentlog to the latest version automatically. |
| `vuln` | Manage findings and vulnerabilities.


## Installation

### Quick Install (Linux & macOS)

```bash
curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh
```

### Build from Source

```bash
# Build on Linux
go build -o pentlog main.go

# Cross-compile on Mac for Linux
GOOS=linux GOARCH=amd64 go build -o pentlog main.go

# Initial setup (checks deps, creates ~/.pentlog/logs)
# ⚠️ REQUIRED before first use!
pentlog setup
```

## Usage

For detailed usage instructions, including advanced features and workflows, please refer to the [**GitHub Wiki**](https://github.com/aancw/pentlog/wiki) or the local [**WIKI.md**](WIKI.md).

### Quick Start

1.  **Initialize**: `pentlog create`
2.  **Start Shell**: `pentlog shell`
3.  **Search Logs**: `pentlog search`
Check out our [ROADMAP.md](ROADMAP.md) to see what features are currently implemented and what we have planned for the future.

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Acknowledgements

- Thanks to [roomkangali](https://github.com/roomkangali) for adding the [AI Summary feature](#ai-analysis) and the logo!
- Special thanks to the authors of `ttyrec/ttyplay` for the underlying recording technology.
- Special thanks to the authors of `ttygif` for the GIF export functionality.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.