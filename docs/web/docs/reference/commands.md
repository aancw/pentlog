# CLI Commands Reference

Complete reference for all PentLog commands. Use this guide to explore the full capabilities of PentLog.

## Quick Reference Card

<div class="grid cards" markdown>

-   :material-console: __Session Management__

    ---

    `create`, `shell`, `switch`, `sessions`, `status`, `reset`

-   :material-magnify: __Analysis & Search__

    ---

    `search`, `timeline`, `dashboard`

-   :material-file-document: __Reporting__

    ---

    `export`, `analyze`, `serve`

-   :material-archive: __Data Management__

    ---

    `archive`, `import`, `freeze`, `recover`

</div>

---

## Session Management

### `pentlog create`
Initialize a new engagement context.

```bash
pentlog create [flags]
```

**Examples:**

```bash
# Interactive mode (recommended)
pentlog create

# Create with specific client/engagement
pentlog create -c "ACME Corp" -e "Internal Pentest"
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--client` | `-c` | Client name |
| `--engagement` | `-e` | Engagement name |
| `--phase` | `-p` | Phase (recon/exploit/post) |

---

### `pentlog shell`
Start a recorded shell session.

```bash
pentlog shell [flags]
```

**Examples:**

```bash
# Start recorded shell
pentlog shell

# Start with live sharing
pentlog shell --share

# Share on custom port
pentlog shell --share --share-port 8080
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--share` | `false` | Enable live browser sharing |
| `--share-port` | `0` | Port for sharing server |
| `--share-bind` | `""` | Bind address for sharing |

---

### `pentlog switch`
Switch between phases or targets.

```bash
pentlog switch [target]
```

**Examples:**

```bash
# Interactive switch
pentlog switch

# Toggle to previous session
pentlog switch -

# Switch to specific phase
pentlog switch exploitation
```

---

### `pentlog sessions`
List and manage recorded sessions.

```bash
pentlog sessions [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--client` | `-c` | Filter by client |
| `--engagement` | `-e` | Filter by engagement |
| `--all` | `-a` | Show all sessions including archived |

---

### `pentlog status`
Show current tool and engagement status.

```bash
pentlog status
```

---

### `pentlog reset`
Clear the current active engagement context.

```bash
pentlog reset
```

!!! warning "Caution"
    This clears the current context but does not delete any recorded sessions.

---

## Analysis & Search

### `pentlog search`
Search command history across all sessions.

```bash
pentlog search [query] [flags]
```

**Examples:**

```bash
# Interactive search TUI
pentlog search

# Search with query
pentlog search "nmap"

# Regex search
pentlog search "nmap.*-sV" --regex

# Boolean operators
pentlog search "sqlmap AND injection"
pentlog search "exploit OR payload"
pentlog search "scan NOT nmap"

# Filter by date
pentlog search "exploit" --after 20260115 --before 20260120

# Filter by client
pentlog search "exploit" --client ACME
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--regex` | `-r` | Use regex pattern |
| `--after` | | Date filter (DDMMYYYY) |
| `--before` | | Date filter (DDMMYYYY) |
| `--client` | `-c` | Filter by client |
| `--engagement` | `-e` | Filter by engagement |

---

### `pentlog timeline`
Extract and browse command timeline.

```bash
pentlog timeline [session-id] [flags]
```

**Examples:**

```bash
# Interactive timeline browser
pentlog timeline

# Export to JSON
pentlog timeline 42 -o timeline.json

# Specific session
pentlog timeline 42
```

---

### `pentlog dashboard`
Show interactive dashboard of pentest activity.

```bash
pentlog dashboard
```

---

## Notes & Vulnerabilities

### `pentlog note`
Manage session notes.

```bash
pentlog note add "note text"
pentlog note list [flags]
```

**Examples:**

```bash
# Add a note
pentlog note add "Found open port 8080"

# List notes for current session
pentlog note list

# List notes for specific session
pentlog note list --session 42
```

---

### `pentlog vuln`
Manage vulnerability findings.

```bash
pentlog vuln list [flags]
pentlog vuln add [flags]
```

**Examples:**

```bash
# List vulnerabilities
pentlog vuln list

# Filter by severity
pentlog vuln list --severity high
```

!!! tip "Quick Vuln Entry"
    Use `Ctrl+G` during a shell session for instant vulnerability logging.

---

## Reporting

### `pentlog export`
Generate reports from recorded sessions.

```bash
pentlog export [client] [flags]
```

**Examples:**

```bash
# Interactive export
pentlog export

# Export specific client/engagement
pentlog export ACME -e internal-pentest

# Export with AI analysis
pentlog export --analyze

# Export full report
pentlog export --analyze --full-report

# Specific format
pentlog export --format html
pentlog export --format markdown
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--client` | `-c` | Client name |
| `--engagement` | `-e` | Engagement name |
| `--phase` | `-p` | Phase to export |
| `--format` | `-f` | Output format (markdown/html) |
| `--output` | `-o` | Output file path |
| `--analyze` | | Run AI analysis |
| `--full-report` | | Full AI analysis (vs summarized) |

---

### `pentlog analyze`
Analyze a report with AI.

```bash
pentlog analyze <report-file> [flags]
```

**Examples:**

```bash
# Analyze existing report
pentlog analyze report.md

# Full analysis
pentlog analyze report.md --full-report
```

---

### `pentlog serve`
Start HTTP server to view HTML reports.

```bash
pentlog serve [flags]
```

**Examples:**

```bash
# Interactive server
pentlog serve

# Specific port
pentlog serve --port 8080

# Don't auto-open browser
pentlog serve --no-open
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `0` | Server port (0 = auto) |
| `--no-open` | `false` | Don't open browser |

---

## Replay & Export

### `pentlog replay`
Replay recorded sessions.

```bash
pentlog replay [session-id] [flags]
```

**Examples:**

```bash
# Interactive replay selection
pentlog replay

# Replay specific session
pentlog replay 42

# Replay at 2x speed
pentlog replay 42 -s 2.0

# Replay at half speed
pentlog replay 42 -s 0.5
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--speed` | `-s` | Playback speed multiplier |

---

### `pentlog gif`
Convert sessions to animated GIF.

```bash
pentlog gif [session-id] [flags]
```

**Examples:**

```bash
# Interactive GIF creation
pentlog gif

# Convert specific session
pentlog gif 42

# Custom speed (5x faster)
pentlog gif 42 -s 5

# Custom output
pentlog gif 42 -o demo.gif

# Specific resolution
pentlog gif --resolution 1080p
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--speed` | `-s` | Playback speed multiplier |
| `--output` | `-o` | Output filename |
| `--resolution` | `-r` | 720p or 1080p |
| `--cols` | | Terminal columns |
| `--rows` | | Terminal rows |

---

## Data Management

### `pentlog archive`
Archive old sessions with optional encryption.

```bash
pentlog archive [client] [flags]
```

**Examples:**

```bash
# Interactive archive
pentlog archive

# Archive specific client
pentlog archive ACME

# Archive old sessions (30+ days)
pentlog archive ACME --days 30

# Archive and delete originals
pentlog archive ACME --days 30 --delete

# Archive specific phase
pentlog archive ACME -e internal -p recon

# List archives
pentlog archive list
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--engagement` | `-e` | Filter by engagement |
| `--phase` | `-p` | Filter by phase |
| `--days` | `-d` | Archive sessions older than N days |
| `--delete` | | Delete original files after archiving |
| `--encrypt` | | Encrypt with password |

---

### `pentlog import`
Restore archived sessions.

```bash
pentlog import <archive> [flags]
```

**Examples:**

```bash
# Import archive
pentlog import archive.zip

# Import with password
pentlog import encrypted.zip --password mysecret

# Import to specific context
pentlog import archive.zip -c ACME -e Q1

# Preview contents
pentlog import list archive.zip

# Skip confirmation
pentlog import archive.zip -y
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--client` | `-c` | Import to client |
| `--engagement` | `-e` | Import to engagement |
| `--phase` | `-p` | Import to phase |
| `--password` | | Decryption password |
| `--overwrite` | | Overwrite existing |
| `--yes` | `-y` | Skip confirmation |

---

### `pentlog freeze`
Generate SHA256 hashes for integrity verification.

```bash
pentlog freeze [flags]
```

**Examples:**

```bash
# Generate hashes for all sessions
pentlog freeze

# Hash specific client
pentlog freeze -c ACME

# Output to specific file
pentlog freeze -o hashes.txt
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--client` | `-c` | Filter by client |
| `--engagement` | `-e` | Filter by engagement |
| `--output` | `-o` | Output file |

---

### `pentlog recover`
Recover crashed or stale sessions.

```bash
pentlog recover [flags]
```

**Examples:**

```bash
# Interactive recovery
pentlog recover

# List crashed sessions
pentlog recover --list

# Recover specific session
pentlog recover --recover 42

# Recover all crashed
pentlog recover --recover-all

# Mark stale as crashed
pentlog recover --mark-stale

# Clean orphaned entries
pentlog recover --clean-orphans
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--list` | List crashed/stale sessions |
| `--recover` | Recover specific session ID |
| `--recover-all` | Recover all crashed sessions |
| `--mark-stale` | Mark stale sessions as crashed |
| `--clean-orphans` | Remove orphaned entries |

---

## Live Sharing

### `pentlog share`
Share sessions in real-time.

```bash
pentlog share [session-id] [flags]
```

**Examples:**

```bash
# Share current live session
pentlog shell --share

# Share recorded session
pentlog share 42

# Custom port
pentlog share 42 --port 8080

# Check status
pentlog share status

# Stop sharing
pentlog share stop
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `0` | Server port |
| `--bind` | `""` | Bind address |

---

## Utilities

### `pentlog setup`
Verify dependencies and configure PentLog.

```bash
pentlog setup [flags]
```

**Examples:**

```bash
# Full setup
pentlog setup

# Configure AI only
pentlog setup ai
```

---

### `pentlog version`
Show version information.

```bash
pentlog version
```

---

### `pentlog update`
Update to the latest version.

```bash
pentlog update [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--dry-run` | Check for updates without installing |

---

### `pentlog completion`
Generate shell completion scripts.

```bash
pentlog completion
```

Follow the interactive prompts to install completions for Bash or Zsh.

---

## Global Flags

These flags work with any command:

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | | Path to config file |
| `--verbose` | `-v` | Enable verbose output |
| `--help` | `-h` | Show help |
| `--version` | `-V` | Show version |

---

## Keyboard Shortcuts

### Inside `pentlog shell`

| Hotkey | Action |
|--------|--------|
| `Ctrl+N` | Add quick note |
| `Ctrl+G` | Add vulnerability |

### Global Shortcuts

| Key | Action |
|-----|--------|
| `?` | Show keyboard shortcuts help |
| `/` | Focus search |

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PENTLOG_CONFIG` | Path to config file |
| `PENTLOG_HOME` | Base directory (default: `~/.pentlog`) |
| `PENTLOG_DEBUG` | Enable debug logging |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid usage |
| `3` | Not found |
| `4` | Permission denied |
| `5` | Already exists |
