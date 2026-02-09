# Pentlog Wiki

Welcome to the official documentation for **pentlog**.

## 📖 Table of Contents

- [Getting Started](#-getting-started)
- [Core Concepts](#-core-concepts)
- [User Guide](#-user-guide)
- [AI Analysis](#-ai-analysis)
- [Reporting](#-reporting)
- [Archiving](#-archiving)
- [Crash Recovery](#-crash-recovery)
- [Live Sharing](#-live-sharing)
- [Advanced Configuration](#️-advanced-configuration)
- [Storage Layout](#-storage-layout)

---

## 🚀 Getting Started

### 1. Quick Setup (3 steps)

```bash
# Install
curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh

# One-time setup (checks dependencies)
pentlog setup

# Create your first engagement
pentlog create
```

### 2. Choose Your Mode

When you run `pentlog create`, select a context mode:

- **Client Mode**: Full professional engagement (Client, Engagement, Scope, Phase)
- **Exam/Lab Mode**: Certifications & CTFs (Exam Name, Target IP)
- **Log Only Mode**: Quick logging (Project Name only)

### 3. Start Recording

```bash
pentlog shell
# Your shell is now recorded with perfect terminal fidelity
```

### 4. Search & Export

```bash
pentlog search       # Find commands across all sessions
pentlog export       # Generate Markdown/HTML reports
```

---

## 📚 Core Concepts

### Context Modes
Pentlog supports different workflows depending on your needs:

*   **Client Mode**: Best for professional engagements. Tracks Client, Engagement, Scope, etc.
*   **Exam/Lab Mode**: Optimized for CTFs and Certifications (OSCP, PNPT, etc.). Tracks Exam Name and Target IP.
*   **Log Only Mode**: Minimal setup. Just asks for a Project Name and starts logging immediately to a simplified path.

## 📖 User Guide

### 1. Initialize Engagement
Use the `create` command to start a new logging context.

```bash
pentlog create
# Prompts for: Context Type
# - Client Mode: Full metadata (Client, Engagement, Scope, Phase)
# - Exam/Lab Mode: Exam Name, Target IP
# - Log Only: Project Name (Defaults to "QuickLog")
```

### 2. Enter Shell
Once initialized, start a recorded shell session.

```bash
pentlog shell
# Enters a recorded shell with custom PS1 and instant-logging.
```

### 3. Switch Phases / Targets
*   **Client Mode**: Use `switch` to move between phases (e.g., recon -> exploit).
*   **Exam/Lab Mode**: Use `switch` to quickly jump to a **New Target IP** without re-running the setup wizard.

```bash
pentlog switch
# Prompts for:
# - Select from History (Interactive list of recent sessions)
# - Enter Manual/New (Prompts for Client/Target + Phase)

# Or toggle quickly to the previous session:
pentlog switch -
```

### 4. Notes & Bookmarks
Add timestamped notes during your session without leaving the terminal.

```bash
# Add a note (e.g. "Found SQLi")
pentlog note add "Found SQLi"

# Review list of notes (Interactive)
# Works both inside a shell (current session) AND offline (select past session)
pentlog note list
```

### Quick Hotkeys (Inside `pentlog shell`)

During a shell session, use keyboard shortcuts for rapid note/vuln entry:

| Hotkey | Action | Description |
|--------|--------|-------------|
| `Ctrl+N` | Quick Note | Prompts for a one-line note, saves instantly |
| `Ctrl+G` | Quick Vuln | Prompts for title, severity (c/h/m/l/i), and description |

**Example workflow:**
```bash
# Press Ctrl+N during shell session
📝 Quick note: Found open port 8080
✓ Note saved [14:05:43]

# Press Ctrl+G for vulnerability
🔓 Vuln title: SQL Injection in login form
Severity (c/h/m/l/i): h
Description (optional): POST /login endpoint vulnerable to blind SQLi
✓ Vuln saved: V-abc123 [High] SQL Injection in login form
```

### 5. Search
All commands function interactively.

```bash
# Search logs and notes (Live incremental search TUI)
# - Select Client -> Engagement
# - Type query to see results live (10 visible, scroll to all matches)
# - ↑↓ Navigate, Enter to open in pager, Home/End to jump
# - Shows "Result X/Y" counter for current position
pentlog search

# Search with query from command line
pentlog search "vulnerability" --regex --after 15012026
```

### 6. Integrity
Generate SHA256 hashes of all logs for evidence integrity.

```bash
pentlog freeze
```

### 7. Dashboard
View an interactive executive summary of your engagement logic, including evidence size, recent findings, and statistical breakdowns.

```bash
pentlog dashboard
```

### 8. Versioning & Updates
Keep your tool up to date.

```bash
# Check version
pentlog version

# Update automatically
pentlog update
```

---

## 💾 Reporting

### Export Reports
Generate structured Markdown and HTML reports from your sessions.

```bash
# Interactive mode (Recommended)
# - Select Phase
# - [View Existing Reports] to browse and open previous files (shows timestamps)
# - Generates report with overwrite protection check
# - Preview in Pager or Save to File
pentlog export

# Export specific client/engagement
pentlog export acme -e incident-response
```

### AI Analysis
Analyze your reports with AI to get a summary of the findings.

#### Usage

There are two ways to use the AI analysis feature:

1. **Analyze an existing report:**
    ```bash
    # Summarized analysis (default)
    pentlog analyze <report_file>

    # Full analysis
    pentlog analyze --full-report <report_file>
    ```

2. **Analyze a report during export:**
    ```bash
    # Summarized analysis (default)
    pentlog export --analyze

    # Full analysis
    pentlog export --analyze --full-report
    ```

#### Configuration

See the [AI Analysis](#-ai-analysis) section for setup instructions.

---

## 📊 Timeline Analysis
Analyze a terminal session recording and extract a chronological timeline of commands executed and their outputs.

```bash
# Interactive Viewer (Recommended)
# - Select from recent sessions
# - View commands in human-readable format (Loop mode)
# - Export to JSON option
pentlog timeline

# Direct JSON Export
pentlog timeline <session_id> -o output.json
```

### 9. Replay
Replay recorded sessions with full fidelity.

```bash
# Lists recent sessions to pick from
pentlog replay

# Or specify ID directly (Linux Only)
pentlog replay 1 -s 2.0
```

### 10. GIF Export
Convert sessions to animated GIFs with high-quality rendering.

```bash
# Interactive mode: select client, engagement, session, and resolution
pentlog gif

# Convert specific session
pentlog gif <session_id>

# Convert TTY file directly
pentlog gif session.tty

# Adjust playback speed
pentlog gif -s 5      # 5x speed (faster playback)

# Specify output filename
pentlog gif -o demo.gif

# Custom terminal dimensions (overrides resolution preset)
pentlog gif --cols 200 --rows 60
```

**Features**:
- **Resolution Selection**: Choose between 720p (1280×720) or 1080p (1920×1080)
- **High-Quality Font**: Uses Go Mono font for crisp, professional text rendering
- **Improved Colors**: Enhanced ANSI palette for better Kali Linux terminal rendering
- **Multiple Modes**: Single session, merged sessions, or direct file conversion
- **Native Rendering**: Pure Go implementation using `vt100` terminal emulator

### 11. Serve Reports
View HTML reports with embedded GIF players via local HTTP server.

```bash
# Interactive mode: select report to view
pentlog serve

# Use specific port
pentlog serve --port 8080

# Auto-opens browser with selected report
```

**Why use serve?**
HTML reports with embedded GIF players require HTTP access to load GIF files. Opening reports directly via `file://` protocol fails due to browser security restrictions. The serve command provides a simple solution by starting a local HTTP server.

---

## 🧠 AI Analysis

### Setup

Configure your AI provider before using the analyze feature.

```bash
# Configure AI provider (Gemini or Ollama)
# Creates ~/.pentlog/ai_config.yaml
pentlog setup ai

# View current configuration
cat ~/.pentlog/ai_config.yaml
```

### Supported Providers

- **Google Gemini** (cloud-based, requires API key)
- **Ollama** (local LLMs, self-hosted)

### Using AI Analysis

See the [Reporting](#-reporting) section for AI analysis usage with `pentlog export --analyze` and `pentlog analyze`.

---

## 💾 Archiving

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

**Features**:
- **Compression**: ZIP format for better compatibility
- **Encryption**: Optional AES-256 password protection
- **Selective**: Archive by Client, Engagement, or Phase
- **Evidence Ready**: Includes auto-generated reports in archives

---

## 📥 Importing Archives

Restore archived sessions back into your pentlog database.

```bash
# Import with auto-detected metadata
pentlog import ~/.pentlog/archive/CLIENT/20260127-192108.zip

# Import encrypted archive (will prompt for password)
pentlog import ~/.pentlog/archive/encrypted.zip

# Import with specific password
pentlog import archive.zip --password mysecret

# Import to specific client/engagement/phase (for generic archives)
pentlog import archive.zip -c ACME -e Q1 -p Initial

# Skip confirmation prompts
pentlog import archive.zip -y

# Overwrite existing files
pentlog import archive.zip --overwrite

# Preview archive contents without importing
pentlog import list archive.zip
```

**Notes**:
- Always use the **full path** to the archive file
- List available archives with `pentlog archive list`
- Encrypted archives will prompt for password automatically
- Supports ZIP format (`.tar.gz` archives need conversion)

---

## 🛡️ Crash Recovery

Pentlog protects your evidence from unexpected session terminations (SSH disconnects, OOM kills, SIGKILL, etc.) with automatic crash detection and recovery.

### How It Works

1. **Session State Tracking**: Each session is tracked as `active`, `completed`, or `crashed`
2. **Heartbeat Mechanism**: During recording, pentlog updates a heartbeat every 30 seconds
3. **Stale Detection**: Sessions with no heartbeat for 5+ minutes are marked as crashed
4. **Startup Warning**: Any pentlog command will warn you about crashed sessions

### Using Recovery

```bash
# Interactive mode (Recommended)
pentlog recover

# List crashed/stale/orphaned sessions
pentlog recover --list

# Recover a specific session by ID
pentlog recover --recover 42

# Recover all crashed sessions at once
pentlog recover --recover-all

# Mark stale active sessions as crashed
pentlog recover --mark-stale

# Clean up orphaned sessions (database entries with missing files)
pentlog recover --clean-orphans
```

### Session States

| State | Description |
|-------|-------------|
| `active` | Session currently recording (heartbeat within 5 min) |
| `completed` | Session ended normally (exit/Ctrl+D) |
| `crashed` | Session terminated unexpectedly |

### Common Scenarios

| Scenario | What Happens |
|----------|--------------|
| SSH disconnects | Session stays `active`, marked `crashed` after 5 min on next command |
| System OOM kills process | Session stays `active`, marked `crashed` after 5 min |
| Normal exit | Session marked `completed` immediately |
| Power failure | Session stays `active`, marked `crashed` after 5 min on reboot |

### Recovery Workflow

```bash
# 1. SSH disconnects during 4-hour exam
# 2. Reconnect and run any pentlog command
$ pentlog sessions

⚠️  Warning: 1 crashed session(s) detected.
   Run 'pentlog recover' to review and recover them.

# 3. Recover the session
$ pentlog recover
❌ Crashed Sessions (1):
  [42] ClientX/internal-pentest/exploitation
      File: session-operator-20260125-143022.tty (2.3 MiB)
      Crashed: 23 minutes ago

? What would you like to do?
> Recover a crashed session

✓ Session 42 recovered successfully

# 4. Session is now usable for replay, search, export
$ pentlog replay 42
```

**Note**: The recovery feature marks sessions as reviewed - your TTY recordings are preserved on disk even after crashes.

---

## 📡 Live Sharing

Share your terminal session in real-time with teammates or reviewers via a browser-based viewer.

### Starting a Shared Session

```bash
# Start shell with live sharing enabled
pentlog shell --share

# Custom port and bind address
pentlog shell --share --share-port 8080 --share-bind 0.0.0.0
```

The share URL is displayed in the shell banner. Viewers open it in any browser to watch your session in real-time (read-only).

### Sharing an Existing Session

```bash
# Share a recorded session by ID
pentlog share <session-id>

# Share with custom port
pentlog share <session-id> --port 8080

# Check share status (viewers, IPs)
pentlog share status

# Stop sharing
pentlog share stop
```

### Share Status

```bash
$ pentlog share status

  PentLog Live Share (Active)
  ──────────────────────────────────────────
  URL:     http://192.168.1.13:44143/watch?token=...
  PID:     12345
  Viewers: 2
           - 192.168.1.10:52341
           - 192.168.1.20:48823
  ──────────────────────────────────────────
```

### Features

| Feature | Description |
|---------|-------------|
| **Dark Theme Viewer** | Modern dark-themed terminal viewer using xterm.js |
| **Session Persistence** | Late-joining viewers see full session history (scrollback buffer) |
| **Viewer Tracking** | Monitor connected viewers and their IPs |
| **Read-Only** | Viewers can only watch, no input accepted |
| **Auto-Reconnect** | Browser viewer automatically reconnects on connection loss |
| **Token Auth** | Each session uses a unique token for access control |

---

## ⚙️ Advanced Configuration

### Shell Completion
Generate and install shell completion scripts for Zsh and Bash.

```bash
pentlog completion
```
Select your shell and follow the prompts.

## 📦 Storage Layout

*   **User Configuration & Context**: `~/.pentlog/context.json`
*   **Database**: `~/.pentlog/pentlog.db` (SQLite session metadata)
*   **Manual Session Logs**: `~/.pentlog/logs/<client>/<engagement>/<phase>/manual-<operator>-<timestamp>.{tty,json}`
*   **Evidence Hashes**: `~/.pentlog/hashes/sha256.txt`
*   **Export Reports**: `~/.pentlog/reports/<client>/`
*   **Templates**: `~/.pentlog/templates/`
*   **Archives**: `~/.pentlog/archive/<client>/`

Wiki Updated
