# pentlog

Evidence-First Pentest Logging Tool.
Captures shell activity as plain-text terminal logs backed by `script`/`scriptreplay`.

## Features

- **No Root Required**: Start recorded shells as a normal user; logs land in your home directory.
- **Context-Aware**: Tracks client/engagement/scope/operator/phase metadata and stamps every log.
- **Replayable**: Timing files enable faithful playback via `scriptreplay`.
- **Extraction Friendly**: Quickly dump per-phase command history to Markdown.
- **Integrity Ready**: Freeze command hashes every log for evidence packaging.

## Installation

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
./pentlog create
# Prompts for: Client, Engagement, Operator, etc.
```

### 2. Enter Shell
Once initialized, start a recorded shell session.
```bash
./pentlog shell
```

### 3. Switch Phases
When moving from one phase to another (e.g., recon -> exploit), use `switch`.
```bash
./pentlog switch
# Prompts for new phase (e.g., "exploit")
```

### 4. Search & Extract
All commands function interactively if arguments are omitted.
```bash
# Search logs (prompts for Regex query)
./pentlog search

# Extract a report (prompts for phase)
./pentlog extract > report.md
```

### 5. Replay (Interactive)
Replay recorded sessions with an interactive selection menu.
```bash
# Lists recent sessions to pick from
./pentlog replay

# Or specify ID directly (Linux Only)
./pentlog replay 1 -s 2.0
```

### 6. Integrity
```bash
# Generate SHA256 hashes of all logs
./pentlog freeze
```

## Storage Layout

- User Configuration & Context: `~/.pentlog/context.json`
- Manual Session Logs + Timing + Metadata: `~/.pentlog/logs/<client>/<engagement>/<phase>/manual-<operator>-<timestamp>.{log,timing,json}`
- Evidence Hashes: `~/.pentlog/hashes/sha256.txt`
- Extraction Reports: `~/.pentlog/extracts/`
