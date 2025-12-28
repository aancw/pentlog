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

### 1. Start Engagement / Switch Context
```bash
# Run as your normal user
./pentlog start \
  --client "ACME Corp" \
  --engagement "Q1 PenTest" \
  --scope "10.10.10.0/24" \
  --operator "kali" \
  --phase "recon"

# Start a recorded shell with context (Recommended)
./pentlog shell

# OR: Apply context to current shell for scripting
eval $(./pentlog env)

# OR: Clear current active context
./pentlog reset
```

### 2. Check Status
```bash
./pentlog status
```

### 3. Manage & Replay
```bash
# List all recorded sessions
./pentlog sessions

# Replay a specific session ID
./pentlog replay <ID>
```

### 4. Extraction & Integrity
```bash
# Extract logs for a specific phase (Markdown output)
./pentlog extract recon > recon_report.md

# Generate SHA256 hashes of all logs
./pentlog freeze
```

## Storage Layout

- User Configuration & Context: `~/.pentlog/context.json`
- Manual Session Logs + Timing + Metadata: `~/.pentlog/logs/<client>/<engagement>/<phase>/manual-<operator>-<timestamp>.{log,timing,json}`
- Evidence Hashes: `~/.pentlog/hashes/sha256.txt`
- Extraction Reports: `~/.pentlog/extracts/`
