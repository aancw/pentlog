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

> This ensures that the logging directory structure and dependencies are correctly initialized.
 
### 1. Initialize Engagement (Interactive)
The `create` command supports two modes:

- **Client Mode**: Best for professional engagements. Tracks Client, Engagement, Scope, etc.
- **Exam/Lab Mode**: Optimized for CTFs and Certifications (OSCP, PNPT, etc.). Tracks Exam Name and Target IP.

```bash
pentlog create
# Prompts for: Context Type (Client vs Exam/Lab)
# Then prompts for relevant details based on selection.
```

### 2. Enter Shell
Once initialized, start a recorded shell session.
```bash
pentlog shell
# Enters a recorded shell with custom PS1 and instant-logging.
```

### 3. Switch Phases / Targets
- **Client Mode**: Use `switch` to move between phases (e.g., recon -> exploit).
- **Exam/Lab Mode**: Use `switch` to quickly jump to a **New Target IP** without re-running the setup wizard.

```bash
pentlog switch
# Prompts for:
# - Select from History (Interactive list of recent sessions)
# - Enter Manual/New (Prompts for Client/Target + Phase)

# Or toggle quickly to the previous session:
pentlog switch -
```

### 4. Notes / Bookmarks
Add timestamped notes during your session without leaving the terminal.
```bash
# Add a note (e.g. "Found SQLi")
pentlog note add "Found SQLi"

# Review list of notes (Interactive)
# Works both inside a shell (current session) AND offline (select past session)
pentlog note list
```


### 5. Search & Export
All commands function interactively.
```bash
# Search logs and notes (Interactive Loop)
# - Select Client -> Engagement -> Query
# - View results in a color-perfect pager (less)
# - Jump straight to interesting lines of code
pentlog search

# Export a report (Interactive Menu)
# - Select Phase
# - Preview in Pager or Save to File
pentlog export
```

### 6. Replay (Interactive)
Replay recorded sessions with an interactive selection menu.
```bash
# Lists recent sessions to pick from
pentlog replay

# Or specify ID directly (Linux Only)
pentlog replay 1 -s 2.0
```

### 7. Integrity
```bash
# Generate SHA256 hashes of all logs
pentlog freeze
```

### 8. Dashboard
View an interactive executive summary of your engagement logic, including evidence size, recent findings, and statistical breakdowns.
```bash
pentlog dashboard
```

### 9. Versioning & Updates

Check your current version:
```bash
pentlog version
```

Update to the latest version automatically using the built-in update command:
```bash
# Update to the latest version automatically
pentlog update
```
The updater checks the upstream server, displays the new version, downloads the appropriate binary for your OS/Arch, and performs an in-place upgrade.

### 10. AI Analysis
Analyze your reports with AI to get a summary of the findings.

#### Usage
There are two ways to use the AI analysis feature:

1.  **Analyze an existing report:**
    ```bash
    # Summarized analysis (default)
    pentlog analyze <report_file>

    # Full analysis
    pentlog analyze --full-report <report_file>
    ```

2.  **Analyze a report during export:**
    ```bash
    # Summarized analysis (default)
    pentlog export --analyze

    # Full analysis
    pentlog export --analyze --full-report
    ```
### 11. Shell Completion

Pentlog provides an interactive setup to enable shell completion (suggestions) for Zsh and Bash.

```bash
pentlog completion
```
Select your shell and follow the prompts to automatically install the script and update your configuration.

Alternatively, you can manually source the script:
```bash
# Zsh
source <(pentlog completion zsh)

# Bash
source <(pentlog completion bash)
```

### 12. Archive
Manage disk usage by archiving old or completed sessions.
```bash
# Interactive Mode (Recommended)
pentlog archive

# Archive all 'acme' sessions (Backup mode - Keeps originals)
pentlog archive acme

# Archive 'acme' sessions older than 30 days and DELETE originals
pentlog archive acme --days 30 --delete

# Archive specific phase or engagement
pentlog archive acme -p recon
pentlog archive acme -e internal-audit

# List archives
pentlog archive list
```

## PentLog Demo

[![asciicast](https://asciinema.org/a/50dfZoej2Gy2oYKCTUWxwpMwb.svg)](https://asciinema.org/a/50dfZoej2Gy2oYKCTUWxwpMwb)

## Storage Layout

- User Configuration & Context: `~/.pentlog/context.json`
- Manual Session Logs + Timing + Metadata: `~/.pentlog/logs/<client>/<engagement>/<phase>/manual-<operator>-<timestamp>.{tty,json}`
- Evidence Hashes: `~/.pentlog/hashes/sha256.txt`
- Evidence Hashes: `~/.pentlog/hashes/sha256.txt`
- Export Reports: `~/.pentlog/reports/<client>/`
- Archives: `~/.pentlog/archive/<client>/`

## Roadmap
Check out our [ROADMAP.md](ROADMAP.md) to see what features are currently implemented and what we have planned for the future.

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Acknowledgements

- Thanks to [roomkangali](https://github.com/roomkangali) for adding the [AI Summary feature](#ai-analysis) and the logo!
- Special thanks to the authors of `ttyrec/ttyplay` for the underlying recording technology.
- Special thanks to the authors of `ttygif` for the GIF export functionality.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.