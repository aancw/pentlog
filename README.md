# PentLog

![Version](https://img.shields.io/github/v/release/aancw/pentlog?style=for-the-badge)
![Go Version](https://img.shields.io/github/go-mod/go-version/aancw/pentlog?style=for-the-badge)
![Downloads](https://img.shields.io/github/downloads/aancw/pentlog/total?style=for-the-badge&color=blue)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

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


## Why PentLog?

Traditional logging (`script`, `tmux`) isn't built for professional engagements. PentLog fills the gap:

- **No Root Required**: Start recorded shells as a normal user; logs land in your home directory.
- **Context-Aware**: Tracks metadata and stamps every log. Flexible support for **Client Engagements** and **Exam/Labs** (OSCP, HTB, etc.).
- **Terminal-Perfect Logs**: Built-in **Virtual Terminal Emulator** guarantees that what you see in the search viewer matches exactly what you saw in your shell—preserving colors, handling overwrites/edits/redraws correctly, and eliminating ghost text.
- **Interactive Workflows**: Seamlessly create engagements, switch phases, and search logs using intuitive TUI menus.
- **Replayable**: Timing files enable faithful playback via `ttyplay`.
- **Export Friendly**: Export structured Markdown reports for any phase with an interactive preview/save menu.
- **Integrity Ready**: Freeze command hashes every log for evidence packaging.
- **AI Analysis**: Analyze your reports with AI to get a summary of the findings.
- **Shell Completion**: Generate and install shell completion scripts for bash and zsh.

> Used by professionals for **OSCP**, **HTB**, and **Real-World Engagements**.

<details>
<summary><strong>Command Reference</strong> (Click to expand)</summary>

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

</details>


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

Detailed guides are available in our **[Wiki](https://github.com/aancw/pentlog/wiki)** or locally in [`WIKI.md`](WIKI.md).

*   [**User Guide**](WIKI.md#user-guide): Deep dive into `switch`, `notes`, `freeze`, and more.
*   [**Modes**](WIKI.md#core-concepts): Learn about Client Mode vs. Exam/Lab Mode.
*   [**AI Analysis**](WIKI.md#ai-analysis): How to configure and use the AI summarizer.

### Quick Start

1.  **Initialize**: `pentlog create`
2.  **Start Shell**: `pentlog shell`
3.  **Search Logs**: `pentlog search`

Check out our [ROADMAP.md](ROADMAP.md) to see what features are currently implemented and what we have planned for the future.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a full list of changes.

## Contributing

Contributions are welcome! Checking out [ROADMAP.md](ROADMAP.md) for planned features and read [CONTRIBUTING.md](CONTRIBUTING.md) to get started.

## Acknowledgements

- Thanks to [roomkangali](https://github.com/roomkangali) for adding the [AI Summary feature](#ai-analysis) and the logo!
- Special thanks to the authors of `ttyrec/ttyplay` for the underlying recording technology.
- Special thanks to the authors of `ttygif` for the GIF export functionality.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.