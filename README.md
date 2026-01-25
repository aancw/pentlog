# PentLog 🔐

**Evidence-First Penetration Testing Logging Tool**

Capture shell activity as high-fidelity terminal logs backed by `ttyrec`. Perfect for **OSCP**, **HTB**, **Real-World Engagements**, and compliance audits.

<p align="center">
  <img src="pentlog.png" width="500">
</p>

<p align="center">
  <a href="https://github.com/aancw/pentlog/releases"><img alt="Release" src="https://img.shields.io/github/v/release/aancw/pentlog?style=flat-square&color=blue"></a>
  <a href="https://golang.org"><img alt="Go" src="https://img.shields.io/github/go-mod/go-version/aancw/pentlog?style=flat-square&color=blue"></a>
  <a href="https://github.com/aancw/pentlog/releases"><img alt="Downloads" src="https://img.shields.io/github/downloads/aancw/pentlog/total?style=flat-square&color=blue"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/badge/License-MIT-blue?style=flat-square"></a>
</p>

<p align="center">
  <strong><a href="#-quick-start">Quick Start</a> • <a href="#-key-features">Features</a> • <a href="#️-commands">Commands</a> • <a href="#️-installation">Install</a> • <a href="#-documentation">Docs</a> • <a href="#-contributing">Contributing</a></strong>
</p>

---

## ✨ Why PentLog?

Traditional logging (`script`, `tmux`) isn't built for professional engagements. PentLog fills the gap:

- **No Root Required**: Start recorded shells as a normal user; logs land in your home directory.
- **Context-Aware**: Tracks metadata and stamps every log. Flexible support for **Client Engagements** and **Exam/Labs** (OSCP, HTB, etc.).
- **Terminal-Perfect Logs**: Built-in **Virtual Terminal Emulator** guarantees that what you see in the search viewer matches exactly what you saw in your shell—preserving colors, handling overwrites/edits/redraws correctly, and eliminating ghost text.
- **Interactive Workflows**: Seamlessly create engagements, switch phases, and search logs using intuitive TUI menus.
- **Replayable**: Timing files enable faithful playback via `ttyplay`.
- **Export Friendly**: Export structured Markdown and **customizable HTML** reports for any phase with an interactive preview/save menu.
- **Integrity Ready**: Freeze command hashes every log for evidence packaging.
- **AI Analysis**: Analyze your reports with AI to get a summary of the findings.
- **Shell Completion**: Generate and install shell completion scripts for bash and zsh.

---

## 🚀 Quick Start

```bash
# 1. Install (macOS/Linux)
curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh

# 2. Setup (one-time)
pentlog setup

# 3. Create engagement
pentlog create

# 4. Start recording
pentlog shell

# 5. Search logs
pentlog search
```

---

## 📋 Key Features

| Feature | Description |
|---------|-------------|
| 🎬 **High-Fidelity Recording** | Captures full terminal output with timing using `ttyrec` |
| 🔍 **Interactive Search** | Search logs with regex and boolean operators across all sessions |
| 📊 **Virtual Terminal Emulator** | Guarantees what you see matches what happened (handles colors, overwrites, etc.) |
| 📝 **Context Awareness** | Tracks Client, Engagement, Phase, Operator, Timestamp automatically |
| 💾 **Structured Export** | Export to Markdown and customizable HTML reports |
| 🔐 **AES-256 Archive** | Compress and encrypt sessions for evidence packaging |
| 🤖 **AI Analysis** | Summarize findings with Google Gemini or Ollama |
| 🎯 **Timeline Extraction** | Browse command history with interactive timeline browser |
| 📌 **Notes & Bookmarks** | Add timestamped notes to sessions |
| 🔄 **Full Replay** | Faithful playback with `ttyplay` |
| 🛡️ **Crash Recovery** | Protect evidence from SSH disconnects, OOM kills, and unexpected crashes |

---

<details>
<summary><h2 style="display: inline;">⌨️ Commands</h2></summary>

| Command | Description |
| :--- | :--- |
| **Session Management** ||
| `create` | Initialize a new engagement context (Interactive) |
| `shell` | Start a recorded shell with the engagement context loaded |
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
| `freeze` | Generate SHA256 hashes of all session logs for integrity |
| `gif` | Convert sessions to animated GIF (720p/1080p) |
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
- **Fedora**: `sudo dnf install ttyrec`
- **Alpine**: `sudo apk add ttyrec`

### ⚠️ Security Best Practices

- **Password-Protected Archives**: Use interactive mode (`pentlog archive`) instead of `--password` flag to avoid storing passwords in shell history
- **Database Permissions**: Sensitive files are created with 0600 permissions automatically
- **Evidence Integrity**: Use `pentlog freeze` before archiving for compliance audits

## 📖 Documentation

### Getting Started
- **[Quick Start Guide](WIKI.md#getting-started)** - Set up and run your first engagement
- **[User Guide](WIKI.md#user-guide)** - Deep dive into all commands and features
- **[Modes Guide](WIKI.md#core-concepts)** - Client Mode vs. Exam/Lab Mode vs. Log-Only Mode

### Advanced Topics
- **[AI Analysis Setup](WIKI.md#ai-analysis)** - Configure Gemini or Ollama for report summarization
- **[Export & Reporting](WIKI.md#reporting)** - Generate Markdown and HTML reports
- **[Archiving & Encryption](WIKI.md#archiving)** - Create encrypted evidence packages

### Project Info
- **[Roadmap](ROADMAP.md)** - Implemented features and future plans
- **[Changelog](CHANGELOG.md)** - Version history and improvements
- **[Contributing](CONTRIBUTING.md)** - Help us improve PentLog

---

## 💡 Use Cases

### Penetration Testing Engagements
- Document every command and output for professional reports
- Maintain metadata and context throughout the engagement
- Generate evidence-ready documentation with AI summaries

### Certifications (OSCP, HTB)
- Track all activity for writeups with perfect terminal fidelity
- Search across all sessions to find specific commands
- Export clean Markdown reports for documentation

### Compliance & Audits
- Create tamper-proof logs with SHA256 integrity verification
- Archive evidence with AES-256 encryption
- Maintain detailed audit trails with timestamps

### Security Research
- Record terminal sessions with precise timing for reproducibility
- Extract command timelines for analysis
- Replay sessions exactly as they happened

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

- 📖 **Documentation**: [WIKI.md](WIKI.md)
- 🐛 **Issues**: [GitHub Issues](https://github.com/aancw/pentlog/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/aancw/pentlog/discussions)
- ⭐ **Star us on GitHub** if you find PentLog useful!

---

**Made for professionals. Evidence-first. No compromises.**