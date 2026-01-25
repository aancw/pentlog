# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Crash Recovery Mechanism**: Protect evidence from unexpected session terminations
  - Session state tracking: `active`, `completed`, `crashed`
  - 30-second heartbeat during recording to track session health
  - New `pentlog recover` command for managing crashed/stale sessions
  - Automatic detection of stale sessions (no heartbeat for 5+ minutes)
  - Handle orphaned sessions (database entries with missing files)
  - Startup warning when crashed sessions are detected
  - Auto-run database migration on any pentlog command
- **Persistent Session Indicator for Bash**: Enhanced `pentlog shell` with bash-specific session indicator
  - Bash sessions now display a persistent indicator in the shell prompt (similar to zsh functionality)
  - Transient right prompt (rprompt) implementation for modern bash shells
  - Session indicator appears at the right side of the terminal, disappearing after each command execution
  - Automatic detection of bash version compatibility

### Fixed
- Fixed bash rprompt positioning to show correctly at the right bottom of every prompt
- Fixed transient rprompt implementation to behave like zsh (disappearing after command execution)

### Changed
- **Database Schema**: Added `state` and `last_sync_at` columns for crash recovery
- **Configuration Management Refactor**: Centralized ConfigManager singleton
  - Consolidated all config.GetXDir() functions into Manager().GetPaths()
  - Eliminated code duplication and improved consistency
  - Better environment variable override support (PENTLOG_HOME, PENTLOG_DB_PATH, etc.)
  - Improved test isolation with ResetManagerForTesting()
  - All configuration now has single source of truth

## [v0.12.0] - 2026-01-21
### Added
- **Incremental Search with Bubble Tea**: Refactored `pentlog search` command with modern TUI
  - Live search results as you type (background task execution)
  - Scrollable viewport showing 10 results at a time, navigate all matches
  - Smart scroll tracking keeps cursor always visible in viewport
  - Result counter showing current position (e.g., "Result 5/139")
  - Keyboard controls: ↑↓ navigate, Enter to open in pager, Home/End to jump
  - Prevents UI freezing with async search execution
  - Streamlined UI: query input, status bar, scrollable results, help footer
- **Dependency Management**: Smart dependency handling
  - Auto-installation support for `ttyrec` and `ttyplay` on macOS/Linux
  - Detailed health check via `pentlog status --dependencies`
  - Graceful degradation (tool warns but continues if optional deps are missing)
  - Updated `install.sh` to verify system requirements immediately
- **GIF Export (Stable)**: Convert sessions to animated GIFs using native Go rendering
  - Interactive resolution selection: 720p (1280×720) or 1080p (1920×1080)
  - Improved ANSI color palette for better Kali Linux terminal rendering
  - High-quality font rendering using Go Mono (gomono) font
  - Resolution-aware font sizing (12pt for 720p, 14pt for 1080p)
  - Support for single sessions, merged sessions, and direct file conversion
  - GIF output saved to `~/.pentlog/reports/`
- **Export Management**: Enhanced `pentlog export` workflow
  - **View Existing Reports**: Interactive menu to browse, select, and open previously generated reports for the current client.
  - **Overwrite Protection**: Automatically detects if a report already exists for the selected scope.
  - **Smart Prompt**: Show creation timestamp and ask for confirmation before regenerating a report.
  - **GIF Embedding**: Option to embed clickable GIF recordings directly into HTML reports using `--include-gifs`
  - **Template Updates**: New `pentlog update --template` command to refresh report templates from the repository


## [v0.11.0] - 2026-01-19
### Added
- **Interactive Timeline Browser**: Enhanced `pentlog timeline` command with interactive interface
  - Browse commands in scrollable list with timestamps
  - Boxed detail view separating command metadata from output
  - Search functionality within timeline (press `/`)
  - View full output in pager (less)
  - Smart output truncation with preview (first 10 lines)
  - Export timeline as JSON
  - Consistent UX with search view pattern
- **Timeline Preview Panel**: Restored post-selection preview display
  - Shows command details (timestamp, command, output excerpt)
  - Displays inline before action menu
  - No input blocking or hang issues

### Improved
- **Error Handling**: Added warnings to stderr when session files are missing for evidence integrity visibility
- **Archive Reliability**: Fixed incomplete cleanup on archive failures by properly closing resources before removing partial files
- **Timeline UX**: Updated prompt to indicate "Enter to view details" for better user guidance

### Changed
- **Code Modernization**: Replaced deprecated `ioutil.ReadFile` with `os.ReadFile` (Go 1.16+)
- **API Cleanup**: Removed deprecated `ExtractTarGz` function stub

### Security
- **Password Security**: Added password confirmation for archive encryption to prevent typos
- **User Awareness**: Added warning banner in shell about password logging
- **File Permissions**: Enforced `0600` permissions for sensitive AI config files (API keys)
- **Cross-Platform**: Fixed `SUDO_USER` home directory resolution for macOS compatibility
- **Input Sanitization**: Enhanced OSC sequence validation to block potential shell metacharacter injection

### Fixed
- **Timeline Hang**: Resolved critical hang issue in `pentlog timeline` where UI became unresponsive
  - Removed Details template that was causing rendering hangs on every keystroke
  - Added terminal clearing between loop iterations to prevent promptui state corruption
  - All functionality now works reliably: navigation, selection, exit, export, and pager view

## [v0.10.0] - 2026-01-17
### Added
- **Password Protection**: Added AES-256 encryption support for archives via `--password` flag or interactive prompt.
- **SQLite Backend**: Migrated session metadata to a local SQLite database for O(1) performance and robustness.
- **Log Only Mode**: Added a new mode for quick logging without complex metadata (`pentlog create` -> Log Only).
- **Security Hardening**: Enforced `0600` permissions on the local database file.
- **Robustness**: Added automatic "Legacy JSON" migration logic.
- **Documentation**: Refactored detailed guides into [WIKI.md](WIKI.md) and simplified `README.md`.

### Changed
- **Archive Format**: Switched default archive format from `tar.gz` to `zip` for better compatibility and encryption support.

### Performance & UX
- **Refactoring**: Split monolithic `shell` command logic for better maintainability.
- **Pagination**: Implemented efficient pagination for both session listing and search results.
- **Search UX**: Added "Infinite Scroll" and Dashboard-style result boxes with fixed alignment for wide characters.

### Fixed
- **AI**: Improved robustness of AI summarizer for ttyrec files using length-based chunking.

## [v0.9.0] - 2026-01-14
### Added
- **Dropdown Selection**: Added interactive dropdown for context selection in `create` command.
- **Progress Bars**: Added progress indicators to installer and updater.
- **Archive Enhancements**: Improved archive transparency (smart report inclusion).
- **Switch History**: Added interactive history selection and "switch back" (`-`) support to `switch` command.
- **Update Changelog**: `update` command now shows an interactive changelog.
- **Export Prompt**: Added a prompt to open the file immediately after export.

### Changed
- **Config**: Renamed `ai.yml` to `ai.yaml`.

## [v0.8.0] - 2026-01-12
### Added
- **Session Archiving**: Introduced `archive` command to bundle old sessions into `.tar.gz` files to save space (with filtering support).

## [v0.7.0] - 2026-01-11
### Added
- **Context Modes**: Added explicit support for "Client Mode" vs "Exam/Lab Mode".
- **Exam Mode**: Streamlined workflow for CTFs tracking `Exam Name` and `Target IP`.
- **Fast Switch**: Enable rapid target changes in Exam mode.
- **Search Filters**: Added date range filtering to `search`.

## [v0.6.0] - 2026-01-06
### Changed
- **Backend Migration**: Switched recording backend from `script` to `ttyrec` for better terminal fidelity and playback.

## [v0.5] - 2026-01-04
### Added
- **AI Feature**: Initial release of AI integration for report summarization.

## [v0.4] - 2026-01-03
### Changed
- **Style**: Updated styling in updater process.
- **Docs**: Updated documentation.

## [v0.1-v0.3] - 2026-01-03
### Added
- **Hierarchy Reporting**: Implemented hierarchical structure for reports (Client -> Engagement -> Phase).
- **Dashboard Metrics**: Added log size metrics breakdown per client and engagement.
- **Export Defaults**: Export command now defaults to organizing by Client/Engagement.
- **Interactive Search**: Made `search` fully interactive with pager integration (`less`).
- **Dashboard**: Added interactive dashboard command.
- **HTML Export**: Added support for exporting reports to HTML format.
- **Clean Output**: implemented terminal cleaning for search results.
- **Funding**: Added FUNDING.yml.
- **Updater**: Added automatic update logic (`pentlog update`).
- **Installation**: Created initial install script.
- **Versioning**: Added version info command.
- **License**: Added MIT License.
- **Renamed Command**: Changed `extract` command to `export` for clarity.
- **Export Cleanup**: Aligned export result cleaning logic with search result cleaner (stripping ANSI codes).
