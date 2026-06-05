# Changelog

All notable changes to this project will be documented in this file.

## [v0.19.1] - 2026-06-05
### Added
- **Archive Integrity Manifests**: Archives now include `manifest.json` with deterministic hashes and sizes
  - Import verifies the manifest when present to catch tampering
- **Explicit Session Sync**: Legacy filesystem-to-database sync is now a dedicated `pentlog sessions sync` migration step
- **Shell Guardrails**: `pentlog shell` now exposes `--phase`, `--target`, and stale-context confirmation before starting a recording
- **Shared AI Config Bootstrapping**: `analyze` and `export` now reuse the same AI configuration setup path
- **Web Search Triage Workflow**: Search results now support evidence-first drill-down from the browser dashboard
  - Search hits include transcript context lines and note timestamps
  - Result cards link directly to the matching session and matching hit
  - Session detail highlights the selected transcript line or operator note after drill-down
- **Search API Coverage**
  - Added handler tests for pagination and context-line search responses

### Changed
- **Web Defaults**: `pentlog web` now uses existing built assets by default and only rebuilds on `--rebuild` or when assets are missing
- **Web Search State**: Search filters now use URL-driven state for bookmarkable and shareable queries
  - Search page keeps `q`, `regex`, `from`, `to`, `limit`, and `page` in the URL
  - Search results paginate with stable `limit` / `offset` semantics and `has_more` metadata
- **Search Response Contract**: Web search now returns richer result metadata
  - Includes real `total_matches` counts instead of just the current page length
  - Includes per-hit context windows and note timestamps for downstream UI workflows

### Fixed
- **Evidence File Permissions**: Sensitive file writes now go through private helpers so evidence and config files use stricter permissions
- **Archive Import Hardening**: Import flow now validates manifest-backed archives and keeps the copy path aligned with the new integrity format
- **Search and Session UX**: Session detail and search pages now preserve context better when jumping between search hits and session content

## [v0.18.0] - 2026-04-22
### Added
- **Web Dashboard Mission Control**: Reworked browser dashboard surface for faster operational visibility
  - New at-a-glance layout for active context, session health, findings, and artifact readiness
  - Workflow quick actions linking directly to sessions, search, reports, and archives
- **Unified Dashboard API**: New `GET /api/dashboard/overview` endpoint
  - Consolidates stats, activity, clients, context, and artifact summary into one payload
- **Scoped Web Flow Prefill**: URL-based prefill support for web pages
  - Sessions page accepts `client`, `phase`, `state`, `tag`, and `q`
  - Search page accepts `q`, `regex`, `from`, `to`, and `limit`
  - Reports page accepts `client`, `engagement`, and `phase`
- **Dashboard Test Coverage**
  - Added handler tests for dashboard overview aggregation contract
  - Added session hydration tests for state/target fields

### Changed
- **Web Information Architecture**: Dashboard sections reordered to prioritize operational status before detail panels
- **Accessibility Improvements**:
  - Added skip-to-content link and stronger keyboard focus visibility
  - Improved active navigation semantics with `aria-current`
- **Dark/Light Theme Cohesion**:
  - Refined theme tokens for state blocks, secondary controls, and stat-card gradients
  - Added subtle dashboard section entrance animation for improved hierarchy

### Fixed
- **Dark Mode Contrast**: Resolved low-contrast empty/loading state text in the web dashboard (including live-share empty state)
- **Session State Reliability**:
  - `logs.ListSessions` and `logs.GetSession` now consistently hydrate `state`, `last_sync_at`, `target`, and `target_ip`
  - Improves reliability of state-based filtering and status rendering across web and API consumers

## [v0.17.0] - 2026-04-04
### Added
- **Target Management**: New `pentlog target` command for multi-target engagements
  - Add/list/switch/remove/clear targets within the active engagement
  - Targets stored in `~/.pentlog/targets.json` with name + optional IP/host
  - Switch prompts update active context with target + target IP

### Changed
- **Session Context & Metadata**: Targets now flow through context, sessions, and prompts
  - Context includes `target` and `target_ip` fields
  - Session metadata persists target fields in JSON and database
  - Shell prompt shows active target and session filenames include target slug

## [v0.16.0] - 2026-03-21
### Added
- **Session Pause/Resume**: Pause recording without creating new sessions
  - `pentlog pause` - Pause current recording session with visual marker
  - `pentlog resume` - Resume paused session, showing pause duration
  - Pause/resume markers written with proper ttyrec binary format
  - Visual banners display in replay showing pause/resume timestamps
  - Perfect for OSCP exams (take breaks without multiple sessions)
  - Ideal for client engagements (pause before sensitive operations)
  - Maintains single continuous session for clean evidence trails
- **Structured Logging**: New logging package with file output support
  - Logs written to `~/.pentlog/pentlog.log` automatically at startup
  - Uses Go's built-in `log/slog` for structured logging
  - Dual output: both stderr (terminal) and file simultaneously
  - JSON/text output options via `SetJSONOutput()`
  - User-friendly output wrappers: `Success()`, `Warning()`, `Error()`, `InfoMsg()`
- **Database Connection Pooling**: Added connection limits for SQLite
  - `SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(5min)`
  - Prevents resource exhaustion under heavy load
- **Session Tags**: Flexible tagging system for session organization
  - Add multiple tags to sessions: `pentlog sessions tag <id> "important" "dc-01"`
  - Remove tags: `pentlog sessions untag <id> "tag-name"`
  - View tags for session: `pentlog sessions tags <id>`
  - List all tags in use: `pentlog sessions tags`
  - Filter sessions by tag: `pentlog sessions list --tag "important"`
  - Useful for marking sessions (e.g., "contains-creds", "critical", "machine-1", "initial-access")
  - Automatic deduplication (duplicate tags ignored)
  - Tags cascade-deleted when session is deleted
  - Full database integration with indexes for fast queries
- **Session Size Monitoring and Alerts**: Background monitoring to prevent large session files
  - Monitors session .tty file size every 30 seconds during shell sessions
  - Warning alert at 5MB: "⚡ Session size: 5.0 MB - Approaching limit"
  - Critical alert at 10MB: "⚠️ Session size: 10.0 MB - Consider splitting session"
  - Alert cooldown (5 minutes) prevents notification spam
  - Works for both new and resumed sessions
  - Configurable thresholds via `config.Manager().GetMonitor()`
  - Optimized 5MB/10MB thresholds ensure fast .tty processing for replay, search, export, and GIF generation
- **Session Delete Command**: `pentlog sessions delete <id>` to remove sessions and associated files
  - Interactive mode: `pentlog sessions delete` shows list and prompts for ID
  - Deletes .tty, .json, and .notes.json files
  - Confirmation prompt with session details (path, size)
  - Validates session ID exists before deletion
- **Engagement Timeline Dashboard**: New `pentlog dashboard timeline` view
  - Single time axis for sessions, phase changes, notes, and vulnerabilities
  - Interactive scope prompt to select current context, client, or all engagements
  - Client grouping headers when showing all engagements

### Changed
- **Error Handling**: Consistent use of `errors.Fatal()` across shell command
  - Replaced mixed `fmt.Fprintf() + os.Exit(1)` with structured error handling
  - Better error context for debugging
- **Human-Readable Session Sizes**: `pentlog sessions` now displays sizes in KB/MB/GB
  - Replaces raw byte counts (e.g., 4660415) with human-readable format (e.g., 4.5 MB)
  - Uses standard KB/MB/GB notation instead of KiB/MiB/GiB
  - Improves readability when reviewing session list

## [v0.15.0] - 2026-02-24
### Added
- **Auto-Resume Crashed Sessions**: Interactive prompt to resume crashed sessions on shell startup
  - New `GetCrashedSessionsForContext()` to query crashed sessions by context
  - New `ResumeSession()` to transition session state from crashed to active
  - User can choose "Resume most recent" or "Start new session"
  - Resumed sessions append to existing .tty file using ttyrec -a flag
  - Prevents evidence fragmentation from network disconnects and crashes
- **Resume Marker in TTY Files**: Visual separator inserted when resuming crashed sessions
  - Inserts yellow banner "Session Resumed" into .tty file at resume point
  - Adjusts timestamp to skip idle time (prevents long waits during replay)
  - Timestamp set to 1 second after last frame instead of actual resume time
  - Clear visual indicator in replay showing where session was interrupted
- **Replay Pagination**: Browse session history beyond first 15 sessions
  - "Load More" option to load next 15 sessions
  - `--all/-a` flag to display all sessions without pagination
  - Session counter showing current range (e.g. "showing 1-15 of 86")
- **Live Share via Shell**: `pentlog shell --share` flag to start live sharing directly from a shell session
  - Embedded WebSocket share server runs in-process alongside the recording
  - Share URL displayed centered in shell banner on session start
  - Share session file saved so `pentlog share status` works with both `shell --share` and `pentlog share`
  - `--share-port` and `--share-bind` flags for network configuration
- **Share Status API**: New `/status` endpoint on share server for live session metadata
  - Returns connected viewer count and client IP addresses as JSON
  - Used by `pentlog share status` to display live viewer information
- **Viewer Tracking**: Track connected viewer IPs in share server
  - `pentlog share status` now shows viewer count and connected IP addresses
  - Supports `X-Forwarded-For` header for proxied connections
- **Session Scrollback Buffer**: Persist terminal output for late-joining viewers
  - New viewers and reconnecting clients receive full session history on connect
  - Scrollback capped at 50MB with automatic front-trimming
  - Data sent as single concatenated blob to preserve terminal escape sequence integrity

### Changed
- **Share HTML Viewer**: Complete dark theme redesign for the live watch page
  - GitHub-dark color scheme (#0d1117 background, #e6edf3 foreground)
  - Terminal-style SVG logo with green accent color
  - Pill-shaped status badges (connected/disconnected/read-only)
  - Dark xterm.js terminal theme with matching scrollbar styling
  - Footer with "Powered by PentLog" branding
  - JetBrains Mono / Fira Code font preference
- **Share Info Centering**: Fixed share info block centering in shell banner
  - Each line now centered individually using `runewidth.StringWidth` for proper Unicode width calculation
  - Fixes misalignment caused by multi-byte UTF-8 box-drawing characters (─)
- **CenterBlock ANSI Handling**: `utils.CenterBlock` now strips ANSI codes before measuring line width
- **Sessions Pagination Default**: `pentlog sessions` now shows the 20 most recent sessions and prompts to load more
  - Interactive "Show more? (Y/n)" prompt for pagination
  - `--limit` and `--offset` continue to provide non-interactive pagination

### Fixed
- **Replay Session Ordering**: Fixed bug where replay showed oldest 15 sessions instead of newest
  - Changed from `sessions[len-15:]` to `sessions[:15]` for DESC ordered results
  - Log Only sessions (phase=N/A) now appear correctly in replay list
- **Share Status Discovery**: `pentlog shell --share` now saves `.share_session` file so `pentlog share status` detects active sessions
- **Viewer Reconnect Alignment**: Fixed terminal output misalignment on browser refresh/reconnect
  - `fitAddon.fit()` now runs before WebSocket connection to ensure correct terminal dimensions
  - Scrollback replayed as single blob instead of individual frames to prevent dropped escape sequences
- **Resume Session Timestamp Normalization**: Fixed missing "Session Resumed" banner and timestamp gap
  - `InsertResumeMarker()` was never called during session resume (causing normalization to silently skip)
  - Added missing call in `startResumedSession()` to insert banner before recording resumes
  - `NormalizeResumedSession()` now correctly detects banner, removes it, and compresses the idle gap
  - Replay shows seamless 3-second transition instead of hours of idle time
- **Dependency Management**: Moved `github.com/gorilla/websocket` from indirect to direct dependency in go.mod
- **Import Session Sizes**: Imported sessions now update their recorded size after insertion
  - Prevents `pentlog sessions` from showing 0-byte sizes for imported `.tty` files

## [v0.14.0] - 2026-02-08
### Added
- **Report Server**: New `pentlog serve` command for viewing HTML reports with GIF players
  - Starts local HTTP server to serve reports directory
  - Solves CORS/file:// issues when loading embedded GIF recordings
  - Interactive report selection with auto-open in browser
  - Configurable port with `--port` flag (default: random available port)
- **HTTP Server in Export Flow**: Option to serve report via HTTP after saving HTML
  - New "Yes (via HTTP server)" option when opening HTML reports
  - Starts server on port 8080 and opens report in browser
- **GIF Regeneration Prompt**: Ask before regenerating existing GIFs during export
  - Detects existing GIF files and prompts: "No (use existing)" or "Yes (regenerate all)"
  - Saves time by reusing previously generated GIFs
- **Archive Import**: Restore archived sessions back into pentlog database
  - New `pentlog import <archive.zip>` command for session recovery
  - Auto-detect metadata from archive structure and directory hierarchy
  - Support encrypted archives with password prompt or `--password` flag
  - Granular import targeting with `-c/--client`, `-e/--engagement`, `-p/--phase`
  - Preview archive contents with `pentlog import list <archive>` before importing
  - `--overwrite` flag to replace existing files
  - `-y/--force` flag to skip confirmation prompts
  - Reverse operation of `pentlog archive` for complete session recovery
- **Database Backup Before Migration**: Automatic safety mechanism
  - Automatic backup of SQLite database before running migrations
  - Prevents data loss during schema updates
  - Backup stored with `.backup` suffix in `~/.pentlog/` directory

### Fixed
- **Shell Hang on Start (SIGTTIN)**: Fix `pentlog shell` hanging after banner on macOS and Linux
  - `Setpgid: true` without `Foreground: true` placed ttyrec in a background process group
  - Kernel sent SIGTTIN when the child shell tried to read stdin, suspending it
  - Added `Foreground: true` with `Ctty` to make the child the foreground process group
  - Regression introduced in `c20323c` (signal handling for graceful shutdown)
- **Signal Handling for Graceful Shutdown**: Properly forward SIGINT/SIGTERM/SIGHUP to subprocess
  - Subprocess now receives termination signals from parent
  - Recording files properly flushed before exit
  - Process group isolation prevents orphaned processes
  - Session state accurately reflects exit type (CRASHED vs COMPLETED)
  - Thread-safe signal reception tracking with mutex
  - All resources (signal channel) properly cleaned up

## [v0.13.0] - 2026-01-26
### Added
- **Quick Note Hotkey System**: Keyboard shortcuts for rapid note/vuln entry during shell sessions
  - `Ctrl+N`: Quick note entry with single-line prompt
  - `Ctrl+G`: Quick vulnerability entry with abbreviated severity input (c/h/m/l/i)
  - Works in both zsh and bash shells
  - Reads from `/dev/tty` for reliable input in keybinding context
  - Hotkey hints displayed in shell banner on session start
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
