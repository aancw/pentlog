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
    - [x] Global Setup Validation.
    - [x] Migration: `script` to `ttyrec` backend.
    - [x] **Dependency Management**: Auto-installation of `ttyrec`/`ttyplay` and graceful degradation.
- [x] **Data Integrity**
    - [x] Generate SHA256 hashes for all logs (`pentlog freeze`).
    - [x] Verify evidence chain-of-custody.
    - [x] Warn on missing session files during listing.
    - [x] Enforce secure file permissions (0600) for sensitive configs.
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
    - [x] **Live Incremental Search** with Bubble Tea TUI.
    - [x] Scrollable viewport (10 results visible, navigate all matches).
    - [x] Smart scroll tracking keeps cursor always visible.
    - [x] Result counter showing current position (Result X/Y).
    - [x] Color-perfect pager integration (less).
    - [x] Direct jump to relevant lines in the pager.
- [x] **Replay Engine**
    - [x] Replay sessions with faithful timing (`pentlog replay`).
    - [x] Interactive menu to select sessions for replay.
- [x] **Interactive Workflows**
    - [x] `pentlog create` wizard for starting engagements.
    - [x] `pentlog switch` menu with History selection.
    - [x] `pentlog switch -` for quick previous session toggle.
    - [x] Custom shell prompt (PS1) injection.
    - [x] Aesthetic Shell UI (Boxed Banner).
    - [x] **Flexible Contexts**: Support for both Client Engagements and Exam/Labs (with fast target switching).
    - [x] **Log Only Mode**: Quick-start logging with minimized metadata and simplified path structure.

### Analysis & Reporting
- [x] **Export Engine**
    - [x] Generate Markdown reports.
    - [x] Generate HTML reports with styling.
    - [x] Organize reports by Client directory structure.
    - [x] Interactive preview before saving.
    - [x] **Clean Exports**: ANSI stripping for Markdown reports.
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
    - [x] **Shell Completion**
        - [x] Interactive setup for Zsh and Bash.
        - [x] Automatic configuration update.

---

## 🚀 Future Roadmap

### Security & Storage
- [x] **Encryption at Rest**
    - [x] AES-256 encryption for log files (Archives).
    - [ ] Secure key management integration.
    - [x] Password-protected session access (Archives).
- [x] **Compressed Storage**
    - [x] Automatic ZIP compression for archived logs.
    - [x] Granular archiving (Client, Engagement, Phase).
    - [x] **Smart Archive**: Auto-include MD/HTML reports and reuse existing ones.
    - [ ] Transparent decompression during search/replay.
- [ ] **Remote Backup**
    - [ ] Backup to SFTP or S3-compatible storage (AWS, R2, MinIO).

### Advanced Analysis & Export
- [x] **Command Timeline**
    - [x] Extract chronological timeline from session recordings
    - [x] Interactive browser with command list and detail views
    - [x] Boxed detail display with command/output separation
    - [x] Search within timeline
    - [x] View full output in pager
    - [x] JSON export with accurate timestamps
    - [x] Fixed critical hang issue (removed Details template, added terminal clearing)
    - [x] Restored preview panel (post-selection display with no hang)
- [ ] **Export to Visuals**
    - [ ] Export session to GIF (via `ttygif`).
    - [ ] Export to MP4/WebM.
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

### Bastion Mode
- [ ] **Unified Access Logging**
    - [ ] Record SSH logins (Bastion Shell).
    - [ ] Record `su` actions.
    - [ ] Record `sudo` actions.
