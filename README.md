# PentLog

**Evidence-First Pentest Logger — Capture every command, find anything, prove everything.**

High-fidelity terminal logs with AI analysis, searchable content, interactive timelines, and compliance-ready reports. Built on `ttyrec`.

Perfect for **Real-World Engagements**, **Compliance & Audits**, **OSCP**, and **HackTheBox**.

<p align="center">
  <img src="pentlog.png" width="500">
</p>

<p align="center">
  <a href="https://github.com/aancw/pentlog/releases"><img alt="Release" src="https://img.shields.io/github/v/release/aancw/pentlog?color=blue"></a>
  <a href="https://golang.org"><img alt="Go" src="https://img.shields.io/github/go-mod/go-version/aancw/pentlog?color=blue"></a>
  <a href="https://github.com/aancw/pentlog/releases"><img alt="Downloads" src="https://img.shields.io/github/downloads/aancw/pentlog/total?color=blue"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/badge/License-MIT-blue"></a>
  <a href="https://star-history.com/#aancw/pentlog&Date"><img alt="Star History" src="https://img.shields.io/github/stars/aancw/pentlog?style=social"></a>
</p>

<p align="center">
  <strong><a href="#quick-start">Quick Start</a> • <a href="#features">Features</a> • <a href="#common-workflows">Workflows</a> • <a href="#installation">Install</a> • <a href="#documentation">Docs</a></strong>
</p>

---

## The Problem with Traditional Logging

Using `script`, `tmux`, or basic shell redirection during pentests creates **fragmented, unsearchable, unmaintainable evidence**:

- **Lost commands** — Mixed with noise, impossible to extract context
- **No integrity** — How do you prove logs weren't tampered with?
- **Manual reports** — Hours spent copying/pasting into documents
- **Evidence gaps** — ANSI codes, terminal artifacts, overwrites break readability
- **Compliance nightmares** — No audit trails, no encrypted archives

---

## Quick Start

```bash
# Install (macOS/Linux) — see Installation section for more options
curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh
pentlog setup

# Create engagement and start recording
pentlog create && pentlog shell

# Search and export
pentlog search && pentlog export
```

After 5 minutes you get:
- Searchable terminal logs with perfect fidelity
- Timestamped commands organized by Client → Engagement → Phase
- Compliance-ready HTML reports
- Encrypted archives for client delivery

---

## Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **High-Fidelity Recording** | Every keystroke + output captured with perfect terminal accuracy using Virtual Terminal Emulator (ANSI colors, overwrites, redraws preserved) |
| **Interactive Search** | Find any command across all sessions instantly with regex and boolean operators |
| **Automatic Organization** | Commands timestamped and organized by Client → Engagement → Phase—no manual naming |
| **Compliance-Ready Export** | Generate Markdown/HTML reports with AI summaries, integrity hashes, encrypted archives |
| **Full Replay** | Faithful playback with `ttyplay` preserves exact timing |
| **Live Sharing** | Share terminal sessions in real-time via browser with dark-themed viewer |
| **AI Analysis** | Summarize findings with Google Gemini or Ollama (local LLM) |
| **Timeline Extraction** | Interactive timeline browser to reconstruct attack sequences |
| **Notes & Bookmarks** | Add timestamped annotations to sessions for later review |
| **AES-256 Encryption** | Password-protected encrypted archives for secure client delivery |
| **Crash Recovery** | Protect evidence from SSH disconnects, OOM kills, unexpected crashes |

### Comparison with Alternatives

| Feature | `script` | `tmux` | PentLog |
|---------|----------|--------|---------|
| **Terminal Fidelity** | ❌ Breaks on special chars | ⚠️ Lossy (missing redraws) | ✅ Perfect (Virtual Terminal Emulator) |
| **Searchable Logs** | ❌ Manual grep chaos | ❌ Session-by-session only | ✅ Full-text search + regex + boolean |
| **Automatic Organization** | ❌ Manual naming | ❌ Manual naming | ✅ Client → Engagement → Phase auto-organized |
| **Timestamps** | ⚠️ Only start/end | ❌ No timestamps | ✅ Every command timestamped |
| **Compliance Ready** | ❌ No integrity | ❌ No integrity | ✅ Hashes + encryption + audit trails |
| **Replay** | ❌ No timing info | ⚠️ Live sessions only | ✅ Faithful playback with `ttyplay` |
| **Reports** | ❌ Manual copy/paste | ❌ Manual copy/paste | ✅ Auto-generate Markdown/HTML + AI |
| **Database** | ❌ Just files | ❌ Just files | ✅ Indexed SQLite for fast searching |
| **Root Required** | ❌ Works as user | ⚠️ Often needs sudo | ✅ Works as normal user |
| **Live Sharing** | ❌ Not supported | ❌ Not supported | ✅ Real-time browser viewer |
| **Crash Recovery** | ❌ Logs lost | ⚠️ May lose session | ✅ Protected from SSH/OOM crashes |

---

## Common Workflows

### Starting a New Engagement
```bash
pentlog create    # Interactive wizard: Client, Engagement, Scope, Operator, Phase
pentlog shell     # Start recording with ttyrec
# Work normally in your shell...
# Press Ctrl+O to pause, Ctrl+T to resume
# Press Ctrl+N to add notes, Ctrl+G to add vulnerabilities
```

### Searching and Reporting
```bash
pentlog search              # Interactive search with regex + boolean operators
pentlog search "nmap"        # Find all nmap commands
pentlog search "exploit OR shell"  # Boolean search
pentlog export              # Generate Markdown/HTML report
pentlog export --analyze    # Include AI-powered summary
```

### Managing Evidence
```bash
pentlog timeline            # Extract command timeline from session
pentlog freeze              # Generate SHA256 hashes for integrity
pentlog archive             # Create encrypted ZIP archive
pentlog import archive.zip  # Restore archived sessions
```

### Analyzing and Sharing
```bash
pentlog analyze report.md   # AI analysis of your report
pentlog shell --share       # Live share session via browser
pentlog serve               # HTTP server for HTML reports with GIF players
pentlog gif session.tty     # Convert session to animated GIF
```

<details>
<summary><strong>View All Commands</strong></summary>

| Command | Description |
| :--- | :--- |
| **Session Management** ||
| `create` | Initialize a new engagement context (Interactive) |
| `shell` | Start a recorded shell with the engagement context loaded |
| `shell --share` | Start a recorded shell with live browser sharing enabled |
| `pause` | Pause the current recording session (Ctrl+O hotkey) |
| `resume` | Resume a paused recording session (Ctrl+T hotkey) |
| `share` | Share a live or recorded session for read-only viewing |
| `sessions` | List and manage recorded sessions |
| `switch` | Switch to a different pentest phase |
| **Analysis & Search** ||
| `search` | Search command history across all sessions (Regex & Boolean) |
| `timeline` | Interactive browser for command timeline extraction |
| `dashboard` | Show an interactive dashboard of your pentest activity |
| `note` | Manage session notes and bookmarks |
| **Reporting** ||
| `export` | Export commands for a specific phase (Markdown/HTML) |
| `analyze` | Analyze a report with an AI provider to summarize findings |
| `vuln` | Manage findings and vulnerabilities |
| **Data Management** ||
| `archive` | Archive old sessions with optional encryption |
| `import` | Restore archived sessions back into pentlog |
| `freeze` | Generate SHA256 hashes of all session logs for integrity |
| `gif` | Convert sessions to animated GIF (720p/1080p) |
| `serve` | Start HTTP server to view HTML reports with GIF players |
| `recover` | Recover and manage crashed or stale sessions |
| **Utilities** ||
| `replay` | Replay a recorded session with full fidelity |
| `status` | Show current tool and engagement status |
| `setup` | Verify dependencies and prepare local logging |
| `reset` | Clear the current active engagement context |
| `completion` | Generate auto-completion scripts for Zsh and Bash |
| `update` | Update pentlog to the latest version automatically |

</details>


## Installation

### Requirements

- **Go 1.24.0+** (if building from source)
- **ttyrec** (terminal recording tool)
- **ttyplay** (optional, for session replay)

### Quick Install

```bash
curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh
pentlog setup  # One-time dependency check and setup
```

### Build from Source

```bash
git clone https://github.com/aancw/pentlog.git
cd pentlog
go build -o pentlog main.go

# Or cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o pentlog main.go
```

### Install System Dependencies

**Automatic** (recommended):
```bash
pentlog setup  # Auto-installs on macOS, Ubuntu, Fedora, Alpine
```

**Manual Installation**:
- **macOS**: `brew install ttyrec`
- **Ubuntu/Debian/WSL**: `sudo apt-get install ttyrec`
- **Fedora**: `sudo dnf install https://github.com/ovh/ovh-ttyrec/releases/download/v1.1.7.1/ovh-ttyrec-1.1.7.1-1.x86_64.rpm`
- **Alpine**: `sudo apk add ttyrec`

### Security Best Practices

- **Password-Protected Archives**: Use interactive mode (`pentlog archive`) instead of `--password` flag to avoid storing passwords in shell history
- **Database Permissions**: Sensitive files are created with 0600 permissions automatically
- **Evidence Integrity**: Use `pentlog freeze` before archiving for compliance audits

---

## Documentation

### Getting Started
- **[Docs Home](https://pentlog.petruknisme.com/)** - Full documentation site
- **[Quick Start](https://pentlog.petruknisme.com/getting-started/quickstart/)** - Set up and run your first engagement
- **[User Guide](https://pentlog.petruknisme.com/guide/sessions/)** - Deep dive into all commands and features
- **[Core Concepts](https://pentlog.petruknisme.com/getting-started/concepts/)** - Client Mode vs. Exam/Lab Mode vs. Log-Only Mode

### Advanced Topics
- **[AI Analysis](https://pentlog.petruknisme.com/guide/ai-analysis/)** - Configure Gemini or Ollama for report summarization
- **[Export & Reporting](https://pentlog.petruknisme.com/guide/export/)** - Generate Markdown and HTML reports
- **[Archiving & Encryption](https://pentlog.petruknisme.com/advanced/archiving/)** - Create encrypted evidence packages

### Project Info
- **[Roadmap](ROADMAP.md)** - Implemented features and future plans
- **[Changelog](CHANGELOG.md)** - Version history and improvements
- **[Contributing](CONTRIBUTING.md)** - Help us improve PentLog

---

## Use Cases

**Penetration Testing Engagements** - Auto-capture everything with perfect terminal fidelity, organize by Client → Engagement → Phase automatically, export compliance-ready HTML reports with AI summaries.

**Compliance & Audits** - Generate integrity hashes with `pentlog freeze`, encrypt sessions with AES-256, maintain detailed audit trails with timestamps and operator tracking.

**Certifications (OSCP, HTB)** - Search across all sessions to find any command instantly, export clean Markdown reports, use timeline browser to reconstruct attack flows.

**Security Research & Red Teaming** - Record sessions with precise timing for faithful replay, extract command timelines for detailed analysis, generate GIF recordings for documentation.

---

## Contributing

We welcome contributions! Start by checking:
1. [Roadmap](ROADMAP.md) - See what's planned
2. [Contributing Guide](CONTRIBUTING.md) - Review guidelines
3. [Open Issues](https://github.com/aancw/pentlog/issues) - Find items to work on

---

## Acknowledgements

- **[roomkangali](https://github.com/roomkangali)** - AI Summary feature & logo design
- **ttyrec/ttyplay authors** - Underlying recording technology
- **Go community** - Bubble Tea, Cobra, and other excellent libraries

---

## License

MIT License - See [LICENSE](LICENSE) for details.

---

## Support & Sponsorship

If you find PentLog useful, consider supporting its development:

<a href="https://www.buymeacoffee.com/petruknisme" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 150px !important;width: 700px !important;" ></a>

Your support helps maintain and improve this tool for the security community.

**Resources:**
- **Documentation**: [docs/wiki/Home.md](docs/wiki/Home.md)
- **Issues**: [GitHub Issues](https://github.com/aancw/pentlog/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aancw/pentlog/discussions)

---

**Made for professionals. Evidence-first. No compromises.**
