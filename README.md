# pentlog

Evidence-First Pentest Logging Tool.
Orchestrates `tlog` for secure terminal recording across Linux distributions.

## Features

- **Distro-Agnostic**: Supports any Linux with `tlog` (Kali, Debian, RHEL, CentOS, Fedora, Arch).
- **Full Capture**: Records all terminal I/O via managed PAM blocks.
- **SCP/SFTP Safe**: TTY-aware configuration prevents data corruption during file transfers.
- **Resilient**: Survives SSH disconnects and terminal crashes.
- **Context-Aware**: Tracks Engagement metadata and Pentest phases.

## Installation

```bash
# Build on Linux
go build -o pentlog main.go

# Cross-compile on Mac for Linux
GOOS=linux GOARCH=amd64 go build -o pentlog main.go

# Setup (Requires Root)
sudo ./pentlog setup
```

## Usage

### 1. Enable Logging
```bash
# Enable for local and SSH sessions (will restart SSH service)
sudo ./pentlog enable --local --ssh
```

### 2. Start Engagement / Switch Context
```bash
# Run as your normal user
./pentlog start \
  --client "ACME Corp" \
  --engagement "Q1 PenTest" \
  --scope "10.10.10.0/24" \
  --operator "kali" \
  --phase "recon"

# CRITICAL: Copy-paste the export commands provided to your terminal/tmux.
```

### 3. Check Status
```bash
./pentlog status
```

### 4. Manage & Replay
```bash
# List all recorded sessions
sudo ./pentlog sessions

# Replay a specific session ID
sudo ./pentlog replay <ID>
```

### 5. Extraction & Integrity
```bash
# Extract commands for a specific phase (Markdown output)
sudo ./pentlog extract recon > recon_report.md

# Generate SHA256 hashes of all logs
sudo ./pentlog freeze
```

## Security & Reliability

- **Managed Blocks**: PAM configurations are wrapped in `# BEGIN/END PENTLOG MANAGED BLOCK` for easy auditing.
- **Backups**: `pentlog` automatically creates timestamped `.bak` files before modifying any system configuration.
- **Root-Owned Logs**: Session logs in `/var/log/tlog/` are protected from non-root users.

## Directory Structure

- User Configuration: `~/.pentlog/`
- System Logs: `/var/log/tlog/`
- Evidence Hashes: `~/.pentlog/hashes/sha256.txt`