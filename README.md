# pentlog

![GitHub release (latest by date)](https://img.shields.io/github/v/release/aancw/pentlog)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/aancw/pentlog)
![GitHub all releases](https://img.shields.io/github/downloads/aancw/pentlog/total)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Evidence-First Pentest Logging Tool.
Captures shell activity as plain-text terminal logs backed by `script`/`scriptreplay`.

## Features

- **No Root Required**: Start recorded shells as a normal user; logs land in your home directory.
- **Context-Aware**: Tracks client/engagement/scope/operator/phase metadata and stamps every log.
- **Terminal-Perfect Logs**: Built-in **Virtual Terminal Emulator** guarantees that what you see in the search viewer matches exactly what you saw in your shell—preserving colors, handling overwrites/edits/redraws correctly, and eliminating ghost text.
- **Interactive Workflows**: Seamlessly create engagements, switch phases, and search logs using intuitive TUI menus.
- **Replayable**: Timing files enable faithful playback via `scriptreplay`.
- **Export Friendly**: Export structured Markdown reports for any phase with an interactive preview/save menu.
- **Integrity Ready**: Freeze command hashes every log for evidence packaging.

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
./pentlog setup
```

## Usage

### 1. Initialize Engagement (Interactive)
The recommended way to start is using the interactive `create` mode.
```bash
pentlog create
# Prompts for: Client, Engagement, Operator, etc.
```

### 2. Enter Shell
Once initialized, start a recorded shell session.
```bash
pentlog shell
# Enters a recorded shell with custom PS1 and instant-logging.
```

### 3. Switch Phases
When moving from one phase to another (e.g., recon -> exploit), use `switch`.
```bash
pentlog switch
# Prompts for new phase (e.g., "exploit")
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
# Public repository (authenticated via GITHUB_TOKEN if set, otherwise anonymous)
pentlog update

# For private repositories, set GH_TOKEN or GITHUB_TOKEN
export GH_TOKEN="your_personal_access_token"
pentlog update
```
The updater checks the upstream server, displays the new version, downloads the appropriate binary for your OS/Arch, and performs an in-place upgrade.

## Storage Layout

- User Configuration & Context: `~/.pentlog/context.json`
- Manual Session Logs + Timing + Metadata: `~/.pentlog/logs/<client>/<engagement>/<phase>/manual-<operator>-<timestamp>.{log,timing,json}`
- Evidence Hashes: `~/.pentlog/hashes/sha256.txt`
- Export Reports: Saved to current directory
