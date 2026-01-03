# Roadmap 🗺️

This document outlines the current status of `pentlog` and our plans for future development.

## ✅ Implemented Features

### Core
- [x] **Session Recording**
    - [x] Capture shell activity using `script` backend.
    - [x] Implement custom Virtual Terminal Emulator for accurate playback.
    - [x] Handle terminal overwrites, redraws, and ghost text elimination.
- [x] **Context Awareness**
    - [x] Track Client, Engagement, and Operator metadata.
    - [x] Support multiple phases (e.g., Recon, Exploit) per engagement.
    - [x] Timestamp every log entry.
- [x] **Data Integrity**
    - [x] Generate SHA256 hashes for all logs (`pentlog freeze`).
    - [x] Verify evidence chain-of-custody.
- [x] **Notes & Bookmarks**
    - [x] Add real-time notes via `pentlog note add`.
    - [x] List notes interactively with `pentlog note list`.
    - [x] Support offline note management for past sessions.

### User Interface (TUI)
- [x] **Dashboard**
    - [x] Interactive executive summary view.
    - [x] Display total sessions, notes, and clients.
    - [x] Show log size metrics per client/engagement.
    - [x] List recent findings.
- [x] **Unified Search**
    - [x] Interactive search across all logs and notes (`pentlog search`).
    - [x] Color-perfect pager integration (less).
    - [x] direct jump to relevant lines in the pager.
- [x] **Replay Engine**
    - [x] Replay sessions with faithful timing (`pentlog replay`).
    - [x] Interactive menu to select sessions for replay.
- [x] **Interactive Workflows**
    - [x] `pentlog create` wizard for starting engagements.
    - [x] `pentlog switch` menu for changing phases.
    - [x] Custom shell prompt (PS1) injection.

### Analysis & Reporting
- [x] **Export Engine**
    - [x] Generate Markdown reports.
    - [x] Generate HTML reports with styling.
    - [x] Organize reports by Client directory structure.
    - [x] Interactive preview before saving.
- [x] **AI Analysis**
    - [x] Integration with Google Gemini.
    - [x] Integration with Ollama for local LLMs.
    - [x] Interactive configuration wizard.
    - [x] Summarize findings from logs.
- [x] **Vulnerability Management**
    - [x] `pentlog vuln` command suite.
    - [x] Add vulnerabilities/findings interactively.
    - [x] Filter reports by findings.

### Maintenance
- [x] **Versioning & Updates**
    - [x] Semantic versioning support.
    - [x] `pentlog version` command.
    - [x] `pentlog update` self-updater via GitHub Releases.
    - [x] Installation script (`install.sh`).

---

## 🚀 Future Roadmap

### Security & Storage
- [ ] **Encryption at Rest**
    - [ ] AES-256 encryption for log files.
    - [ ] Secure key management integration.
    - [ ] Password-protected session access.
- [ ] **Compressed Storage**
    - [ ] Automatic GZIP/ZSTD compression for archived logs.
    - [ ] Transparent decompression during search/replay.
- [ ] **Remote Backup**
    - [ ] Backup to SFTP or S3-compatible storage (AWS, R2, MinIO).

### Advanced Analysis
- [ ] **Custom Report Templates**
    - [ ] Jinja2/Go template engine support.
    - [ ] User-defined export formats.
- [ ] **Advanced Search Syntax**
    - [ ] Regex support in search queries.
    - [ ] Boolean operators (AND, OR, NOT).
    - [ ] Date range filtering.
- [ ] **AI-Powered Querying**
    - [ ] Natural language to search query translation.
    - [ ] "Ask your logs" chat interface.

### Extensibility
- [ ] **Plugin System**
    - [ ] Hook system for log events.
    - [ ] Passive scanner plugin API.
    - [ ] Third-party metadata injection.
