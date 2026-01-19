# Pentlog Wiki

Welcome to the official documentation for **pentlog**.

## 📚 Core Concepts

### Context Modes
Pentlog supports different workflows depending on your needs:

*   **Client Mode**: Best for professional engagements. Tracks Client, Engagement, Scope, etc.
*   **Exam/Lab Mode**: Optimized for CTFs and Certifications (OSCP, PNPT, etc.). Tracks Exam Name and Target IP.
*   **Log Only Mode**: Minimal setup. Just asks for a Project Name and starts logging immediately to a simplified path.

## 🚀 User Guide

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

### 5. Search & Export
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

# Export a report (Interactive Menu)
# - Select Phase
# - Preview in Pager or Save to File
pentlog export
```

### 6. Timeline Analysis
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

### 7. Replay
Replay recorded sessions with full fidelity.

```bash
# Lists recent sessions to pick from
pentlog replay

# Or specify ID directly (Linux Only)
pentlog replay 1 -s 2.0
```

### 8. Integrity
Generate SHA256 hashes of all logs for evidence integrity.

```bash
pentlog freeze
```

### 9. Dashboard
View an interactive executive summary of your engagement logic, including evidence size, recent findings, and statistical breakdowns.

```bash
pentlog dashboard
```

### 10. Versioning & Updates
Keep your tool up to date.

```bash
# Check version
pentlog version

# Update automatically
pentlog update
```

## 🧠 AI Analysis

Analyze your reports with AI to get a summary of the findings.

### Usage
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

## 🛠️ Advanced Configuration

### Shell Completion
Generate and install shell completion scripts for Zsh and Bash.

```bash
pentlog completion
```
Select your shell and follow the prompts.

### Archive Management
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

## 📦 Storage Layout

*   **User Configuration & Context**: `~/.pentlog/context.json`
*   **Database**: `~/.pentlog/pentlog.db` (SQLite session metadata)
*   **Manual Session Logs**: `~/.pentlog/logs/<client>/<engagement>/<phase>/manual-<operator>-<timestamp>.{tty,json}`
*   **Evidence Hashes**: `~/.pentlog/hashes/sha256.txt`
*   **Export Reports**: `~/.pentlog/reports/<client>/`
*   **Archives**: `~/.pentlog/archive/<client>/`
