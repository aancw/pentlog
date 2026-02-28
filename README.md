# PentLog 🔐

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
</p>

<p align="center">
  <strong><a href="#-quick-start">Quick Start</a> • <a href="#-key-features">Features</a> • <a href="#️-commands">Commands</a> • <a href="#️-installation">Install</a> • <a href="#-documentation">Docs</a> • <a href="#-contributing">Contributing</a></strong>
</p>

---

## ✨ Why PentLog?

### The Problem with Traditional Logging

Using `script`, `tmux`, or basic shell redirection during pentests creates **fragmented, unsearchable, unmaintainable evidence**:

- 🔴 **Lost commands** — Mixed with noise, impossible to extract context
- 🔴 **No integrity** — How do you prove logs weren't tampered with?
- 🔴 **Manual reports** — Hours spent copying/pasting into documents
- 🔴 **Evidence gaps** — ANSI codes, terminal artifacts, overwrites break readability
- 🔴 **Compliance nightmares** — No audit trails, no encrypted archives

### How PentLog Solves It

- ✅ **Evidence-First Design** — Every command + output captured with perfect fidelity using Virtual Terminal Emulator (handles colors, overwrites, redraws—what you see matches what happened)
- ✅ **Context & Metadata** — Automatic timestamps, operator tracking, client/engagement organization from day one
- ✅ **Searchable Everything** — Find any command across all sessions with regex + boolean operators in seconds
- ✅ **Compliance-Ready** — Integrity hashes, AES-256 encrypted archives, detailed audit trails
- ✅ **Reports in Minutes** — Auto-generate Markdown/HTML with AI-powered summaries (no manual copy/paste)
- ✅ **No Root Required** — Works as normal user; logs land safely in `~/.pentlog/`
- ✅ **Interactive Workflows** — Intuitive TUI for engagement creation, phase switching, searching, and notes
- ✅ **Replayable** — Faithful session playback with `ttyplay` preserves exact timing
- ✅ **Integrity Protection** — `pentlog freeze` generates SHA256 hashes for evidence packaging

---

## 🚀 Quick Start

```bash
# 1. Install (macOS/Linux) — ~30 seconds
curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh

# 2. Setup (one-time) — Verify dependencies
pentlog setup

# 3. Create engagement — Interactive wizard
pentlog create

# 4. Start recording — Shell with ttyrec running
pentlog shell
# → You now have high-fidelity logs in ~/.pentlog/logs/
# → Logs are indexed in SQLite, ready to search

# 5. Search logs — Find commands across all sessions
pentlog search

# 6. Export report — Generate Markdown/HTML for client
pentlog export
```

**What you get after 5 minutes**:
- ✅ Searchable terminal logs with perfect fidelity (ANSI colors, overwrites, etc.)
- ✅ Timestamped commands organized by Client → Engagement → Phase
- ✅ Compliance-ready HTML reports with AI summaries
- ✅ Encrypted archives for secure client delivery

---

## 📋 Key Features

### 🌟 **Top 5 Features** (What sets PentLog apart)

| Feature | Why It Matters |
|---------|----------------|
| 🎬 **High-Fidelity Recording** | Every keystroke + output captured with perfect terminal accuracy (ANSI colors, overwrites, redraws preserved) |
| 🔍 **Interactive Search** | Find any command across all sessions instantly with regex and boolean operators—no more grep chaos |
| 📊 **Virtual Terminal Emulator** | What you see in the viewer is *exactly* what you saw in your shell (unlike `script` or `tmux` which break on special chars) |
| 💾 **Compliance-Ready Export** | Generate Markdown/HTML reports with AI summaries, integrity hashes, and encrypted archives in seconds |
| 📝 **Automatic Context** | Every command timestamped and organized by Client → Engagement → Phase—no manual naming or organizing |

### 📚 **Additional Features**

| Feature | Description |
|---------|-------------|
| 🤖 **AI Analysis** | Summarize findings with Google Gemini or Ollama (local LLM) |
| 🎯 **Timeline Extraction** | Interactive timeline browser to reconstruct your attack sequence |
| 📌 **Notes & Bookmarks** | Add timestamped annotations to sessions for later review |
| ⌨️ **Quick Hotkeys** | Ctrl+N for notes, Ctrl+G for vulns during shell sessions |
| 🔄 **Full Replay** | Faithful playback with `ttyplay` preserves exact timing |
| 🛡️ **Crash Recovery** | Protect evidence from SSH disconnects, OOM kills, and unexpected crashes |
| 📡 **Live Share** | Share terminal sessions in real-time via browser with dark-themed viewer |
| 🔐 **AES-256 Archive** | Password-protected encrypted archives for secure client delivery |

---

<details>
<summary><h2 style="display: inline;">⌨️ Commands</h2></summary>

| Command | Description |
| :--- | :--- |
| **Session Management** ||
| `create` | Initialize a new engagement context (Interactive) |
| `shell` | Start a recorded shell with the engagement context loaded |
| `shell --share` | Start a recorded shell with live browser sharing enabled |
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


## 🛠️ Installation

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

### ⚠️ Security Best Practices

- **Password-Protected Archives**: Use interactive mode (`pentlog archive`) instead of `--password` flag to avoid storing passwords in shell history
- **Database Permissions**: Sensitive files are created with 0600 permissions automatically
- **Evidence Integrity**: Use `pentlog freeze` before archiving for compliance audits

---

## 📊 Comparison: PentLog vs Alternatives

| Feature | `script` | `tmux` | PentLog |
|---------|----------|--------|---------|
| **Terminal Fidelity** | ❌ Breaks on special chars | ⚠️ Lossy (missing redraws) | ✅ Perfect (Virtual Terminal Emulator) |
| **Searchable Logs** | ❌ Manual grep chaos | ❌ Session-by-session only | ✅ Full-text search + regex + boolean operators |
| **Automatic Organization** | ❌ Manual naming | ❌ Manual naming | ✅ Client → Engagement → Phase auto-organized |
| **Timestamps** | ⚠️ Only start/end | ❌ No timestamps | ✅ Every command timestamped |
| **Compliance Ready** | ❌ No integrity | ❌ No integrity | ✅ Hashes + encryption + audit trails |
| **Replay** | ❌ No timing info | ⚠️ Live sessions only | ✅ Faithful playback with `ttyplay` |
| **Reports** | ❌ Manual copy/paste | ❌ Manual copy/paste | ✅ Auto-generate Markdown/HTML + AI summaries |
| **Database** | ❌ Just files | ❌ Just files | ✅ Indexed SQLite for fast searching |
| **Root Required** | ❌ Works as user | ⚠️ Often needs sudo | ✅ Works as normal user |
| **Live Sharing** | ❌ Not supported | ❌ Not supported | ✅ Real-time browser viewer with scrollback |
| **Crash Recovery** | ❌ Logs lost | ⚠️ May lose session | ✅ Protected from SSH/OOM crashes |

## 📖 Documentation

### Getting Started
- **[Docs Home](https://pentlog.petruknisme.com/)** - Full documentation site
- **[Quick Start](https://pentlog.petruknisme.com/getting-started/quickstart/)** - Set up and run your first engagement
- **[User Guide](https://pentlog.petruknisme.com/guide/sessions/)** - Deep dive into all commands and features
- **[Core Concepts](https://pentlog.petruknisme.com/getting-started/concepts/)** - Client Mode vs. Exam/Lab Mode vs. Log-Only Mode

### Advanced Topics
- **[AI Analysis](https://pentlog.petruknisme.com/guide/ai-analysis/)** - Configure Gemini or Ollama for report summarization
- **[Export & Reporting](https://pentlog.petruknisme.com/guide/export/)** - Generate Markdown and HTML reports
- **[Archiving & Encryption](https://pentlog.petruknisme.com/advanced/archiving/)** - Create encrypted evidence packages

### Local Docs
- Source files live in `docs/web/docs/` and are served with Zensical (`docs/web/README.md` has build instructions).

### Project Info
- **[Roadmap](ROADMAP.md)** - Implemented features and future plans
- **[Changelog](CHANGELOG.md)** - Version history and improvements
- **[Contributing](CONTRIBUTING.md)** - Help us improve PentLog

---

## 💡 Use Cases

### 🔴 **Penetration Testing Engagements** (Real-World)
**Problem**: Client demands evidence of every command. Your manual notes + tmux logs are a mess.

**Solution**:
- Auto-capture everything with perfect terminal fidelity
- Organize by Client → Engagement → Phase automatically
- Export compliance-ready HTML reports with AI summaries in minutes
- Archive with encryption for secure client delivery

### 🟡 **Compliance & Audits**
**Problem**: Regulators need tamper-proof logs, audit trails, and encryption. Your shell history isn't enough.

**Solution**:
- Generate integrity hashes with `pentlog freeze` before archiving
- Encrypt sessions with AES-256 for secure evidence packaging
- Maintain detailed audit trails with timestamps and operator tracking
- Create audit-ready reports with searchable content

### 🟢 **Certifications (OSCP, HTB, etc.)**
**Problem**: Need to document every step for writeups. Fighting with formatting and missing commands.

**Solution**:
- Search across all sessions to find any command instantly
- Export clean Markdown reports with proper formatting
- Perfect terminal fidelity means what you see is what you get
- Timeline browser helps reconstruct your attack flow

### 🔵 **Security Research & Red Teaming**
**Problem**: Need reproducible, timestamped terminal sessions for analysis and playback.

**Solution**:
- Record sessions with precise timing for faithful `ttyplay` replay
- Extract command timelines for detailed activity analysis
- Search across historical sessions for patterns and techniques
- Generate GIF recordings for documentation and presentations

---

## 🤝 Contributing

We welcome contributions! Start by checking:
1. [Roadmap](ROADMAP.md) - See what's planned
2. [Contributing Guide](CONTRIBUTING.md) - Review guidelines
3. [Open Issues](https://github.com/aancw/pentlog/issues) - Find items to work on

---

## 👏 Acknowledgements

- **[roomkangali](https://github.com/roomkangali)** - AI Summary feature & logo design
- **ttyrec/ttyplay authors** - Underlying recording technology
- **Go community** - Bubble Tea, Cobra, and other excellent libraries

---

## 📄 License

MIT License - See [LICENSE](LICENSE) for details.

---

## 🎯 Support & Community

- 📖 **Documentation**: [docs/wiki/Home.md](docs/wiki/Home.md)
- 🐛 **Issues**: [GitHub Issues](https://github.com/aancw/pentlog/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/aancw/pentlog/discussions)
- ⭐ **Star us on GitHub** if you find PentLog useful!

---

**Made for professionals. Evidence-first. No compromises.**
