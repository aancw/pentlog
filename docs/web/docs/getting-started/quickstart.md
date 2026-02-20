# Quick Start

Get PentLog running in under 5 minutes with this step-by-step guide.

## Prerequisites

Before you begin, ensure you have:

- macOS 10.15+ or a Linux distribution
- Internet connection for installation
- A terminal with Bash or Zsh

## 5-Minute Setup

### Step 1: Install PentLog

=== "macOS / Linux (Quick)"

    ```bash
    curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh
    ```

    This downloads and installs the latest release to `~/.local/bin/pentlog`.

=== "Build from Source"

    ```bash
    git clone https://github.com/aancw/pentlog.git
    cd pentlog
    go build -o pentlog main.go
    ```

    !!! note "Go Required"
        Requires Go 1.24.0+. Install from [golang.org](https://golang.org).

### Step 2: Setup Dependencies

```bash
pentlog setup
```

This command will:

- Check for `ttyrec` and `ttyplay`
- Auto-install missing dependencies (on supported platforms)
- Create the `~/.pentlog/` directory structure
- Set up the SQLite database

Expected output:

```
✓ PentLog Setup Complete
  • ttyrec: /usr/local/bin/ttyrec
  • ttyplay: /usr/local/bin/ttyplay
  • Database: ~/.pentlog/pentlog.db
  • Config: ~/.pentlog/context.json
```

### Step 3: Create Your First Engagement

```bash
pentlog create
```

You'll be guided through an interactive wizard to choose your **Context Mode**:

<div class="grid cards" markdown>

-   :material-office-building: __Client Mode__

    ---

    For professional penetration testing engagements.

    **Hierarchy:** `Client → Engagement → Scope → Phase`

    **Example:** ACME Corp → Internal Pentest → 10.0.0.0/24 → Recon

-   :material-school: __Exam/Lab Mode__

    ---

    For certifications and CTFs.

    **Hierarchy:** `Exam Name → Target IP`

    **Example:** OSCP Lab → 10.10.10.5

-   :material-note: __Log Only Mode__

    ---

    For quick logging without organization.

    **Hierarchy:** `Project Name`

    **Example:** Quick Research

</div>

!!! tip "Choose Wisely"
    Pick Client Mode for professional work. Exam Mode is optimized for OSCP/HTB. Log Only is great for quick tests.

### Step 4: Start Recording

```bash
pentlog shell
```

Your terminal is now being recorded with **perfect fidelity**. You'll see a custom prompt showing your current context:

```
┌─[ACME/internal/recon]─[~]
└──>
```

!!! info "What's Being Captured?"
    - Every keystroke and output
    - ANSI colors and formatting
    - Cursor movements and redraws
    - Terminal resizes
    - Working directory changes
    - Timestamp for every command

### Step 5: Add Notes & Vulnerabilities

While in a recorded shell, use these hotkeys:

| Hotkey | Action | Example Output |
|--------|--------|----------------|
| `Ctrl+N` | Add note | `📝 Note saved [14:05:43]` |
| `Ctrl+G` | Add vulnerability | `🔓 Vuln saved: V-abc123 [High]` |

**Example workflow:**

```bash
# Press Ctrl+N
📝 Quick note: Found open port 8080
✓ Note saved [14:05:43]

# Press Ctrl+G
🔓 Vuln title: SQL Injection in login form
Severity (c/h/m/l/i): h
Description: POST /login vulnerable to blind SQLi
✓ Vuln saved: V-abc123 [High] SQL Injection in login form
```

### Step 6: Search Your Sessions

Exit the shell (type `exit` or press `Ctrl+D`), then search:

```bash
pentlog search
```

The interactive search TUI lets you:

- Type queries to see live results
- Use regex patterns: `pentlog search "nmap.*-sV" --regex`
- Boolean operators: `pentlog search "sqlmap AND injection"`
- Filter by date: `pentlog search "exploit" --after 15012026`

### Step 7: Export Your Report

Generate a professional report:

```bash
pentlog export
```

This interactive wizard will:

- Let you select the phase to export
- Choose format (Markdown or HTML)
- Optionally run AI analysis
- Generate integrity hashes

!!! tip "AI Analysis"
    Add `--analyze` to automatically summarize findings with your configured AI provider.

## What You've Accomplished

In just 5 minutes, you've:

✅ **Installed PentLog** — Professional terminal logging tool
✅ **Created an engagement** — Organized by Client → Engagement → Phase
✅ **Recorded a session** — High-fidelity capture of all commands
✅ **Added notes** — Timestamped annotations for important findings
✅ **Searched sessions** — Found commands across all your history
✅ **Exported a report** — Client-ready Markdown or HTML

## Your First Session Files

After recording, your files are organized as:

```
~/.pentlog/
├── context.json          # Current active context
├── pentlog.db           # SQLite database (search index)
└── logs/
    └── ACME/
        └── internal/
            └── recon/
                ├── manual-operator-20260127-143022.tty    # Recording
                └── manual-operator-20260127-143022.json   # Metadata
```

## Next Steps

<div class="grid cards" markdown>

-   :material-lightbulb: __Core Concepts__

    ---

    Understand PentLog's evidence-first design and session organization.

    [:octicons-arrow-right-24: Learn More](concepts.md)

-   :material-console: __Session Management__

    ---

    Learn advanced workflows: switching phases, managing multiple engagements.

    [:octicons-arrow-right-24: Explore](guide/sessions.md)

-   :material-robot: __AI Analysis__

    ---

    Configure Google Gemini or Ollama for automated report summaries.

    [:octicons-arrow-right-24: Set Up AI](guide/ai-analysis.md)

</div>

## Common Issues

??? question "Command not found after install"
    Add `~/.local/bin` to your PATH:
    ```bash
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
    source ~/.zshrc
    ```

??? question "ttyrec not found error"
    Install manually for your OS:
    ```bash
    # macOS
    brew install ttyrec

    # Ubuntu/Debian
    sudo apt-get install ttyrec

    # Then re-run setup
    pentlog setup
    ```

??? question "Permission denied on macOS"
    Remove the quarantine attribute:
    ```bash
    xattr -d com.apple.quarantine ~/.local/bin/pentlog
    ```

## Need Help?

- :material-github: [GitHub Issues](https://github.com/aancw/pentlog/issues)
- :material-forum: [GitHub Discussions](https://github.com/aancw/pentlog/discussions)
- :material-file-document: [Full Documentation](https://pentlog.pages.dev)
