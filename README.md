# pentlog

Evidence-First Pentest Logging Tool.
Orchestrates `tlog` for secure, exam-safe terminal recording on Kali Linux / Debian.

## Features

- **Full Capture**: Records all terminal I/O via PAM integration (`pam_tlog`).
- **Resilient**: Works with local terminals, SSH, tmux. Survives disconnects.
- **Context-Aware**: Tags sessions with Client, Engagement, Operator, and Phase.
- **Exam-Safe**: No background daemons, no kernel modules, low overhead.

## Installation

```bash
# Build
go build -o pentlog main.go

# Install (requires root)
sudo ./pentlog setup
```

## Usage

### 1. Enable Logging
```bash
sudo ./pentlog enable --local --ssh
```

### 2. Start Engagement / Switch Context
```bash
./pentlog start \
  --client "ACME Corp" \
  --engagement "Q1 PenTest" \
  --scope "10.10.10.0/24" \
  --operator "user1" \
  --phase "recon"

# Copy-paste the export commands provided by the tool to tag current session.
```

### 3. Check Status
```bash
./pentlog status
```

### 4. Manage Sessions
```bash
# List sessions
./pentlog sessions

# Replay a session
./pentlog replay 1
```

### 5. Reporting
```bash
# Extract commands for a phase
./pentlog extract recon > recon_commands.md

# Freeze logs (hashing)
./pentlog freeze
```

## Directory Structure

- Configuration: `~/.pentlog/context.json`, `~/.pentlog/history.jsonl`
- Logs: `/var/log/tlog/`
- Hashes: `~/.pentlog/hashes/sha256.txt`
