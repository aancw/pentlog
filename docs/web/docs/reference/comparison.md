# Tool Comparison

See how PentLog compares to traditional terminal logging tools and why professionals choose it for security engagements.

## Feature Comparison Matrix

<div class="grid cards" markdown>

-   :material-check-circle:{ .green } __Full Support__
-   :material-alert-circle:{ .yellow } __Partial Support__
-   :material-close-circle:{ .red } __Not Supported__

</div>

| Feature | `script` | `tmux` | **PentLog** |
|:--------|:--------:|:------:|:-----------:|
| **Terminal Fidelity** | :material-close-circle:{ .red } | :material-alert-circle:{ .yellow } | :material-check-circle:{ .green } |
| **Full-Text Search** | :material-close-circle:{ .red } | :material-close-circle:{ .red } | :material-check-circle:{ .green } |
| **Automatic Organization** | :material-close-circle:{ .red } | :material-close-circle:{ .red } | :material-check-circle:{ .green } |
| **Timestamps** | :material-alert-circle:{ .yellow } | :material-close-circle:{ .red } | :material-check-circle:{ .green } |
| **Compliance Ready** | :material-close-circle:{ .red } | :material-close-circle:{ .red } | :material-check-circle:{ .green } |
| **Session Replay** | :material-close-circle:{ .red } | :material-alert-circle:{ .yellow } | :material-check-circle:{ .green } |
| **Report Generation** | :material-close-circle:{ .red } | :material-close-circle:{ .red } | :material-check-circle:{ .green } |
| **Database Index** | :material-close-circle:{ .red } | :material-close-circle:{ .red } | :material-check-circle:{ .green } |
| **Live Sharing** | :material-close-circle:{ .red } | :material-close-circle:{ .red } | :material-check-circle:{ .green } |
| **Crash Recovery** | :material-close-circle:{ .red } | :material-alert-circle:{ .yellow } | :material-check-circle:{ .green } |
| **AI Analysis** | :material-close-circle:{ .red } | :material-close-circle:{ .red } | :material-check-circle:{ .green } |
| **GIF Export** | :material-close-circle:{ .red } | :material-close-circle:{ .red } | :material-check-circle:{ .green } |

## Detailed Breakdown

### :material-monitor: Terminal Fidelity

=== "`script`"

    - Captures raw output stream
    - **Breaks** on ANSI escape codes
    - Cursor movements not handled
    - Terminal artifacts in output

    !!! failure "Not suitable for evidence"
        Output contains garbled text from cursor movements and color codes.

=== "`tmux`"

    - Better than `script`
    - Captures pane content
    - **Missing** redraws and some sequences
    - Lossy capture

    !!! warning "Partial fidelity"
        Some terminal sequences are lost, especially during rapid output.

=== "PentLog"

    - Virtual Terminal Emulator
    - Perfect fidelity: colors, overwrites, redraws
    - **What you see is exactly what you get**

    !!! success "Evidence-grade quality"
        Every visual element is captured with perfect accuracy.

### :material-magnify: Search Capability

=== "`script`"

    - Manual `grep` only
    - No metadata or context
    - Session-by-session files
    - No cross-session search

=== "`tmux`"

    - Copy mode (current session only)
    - No search history
    - No regex support

=== "PentLog"

    - Full-text search across **all sessions**
    - Regex support: `pentlog search "nmap.*-sV" --regex`
    - Boolean operators: `AND`, `OR`, `NOT`
    - Filter by date, client, engagement
    - Interactive TUI with live results

### :material-folder-tree: Organization

=== "`script`"

    - Manual file naming
    - No structure or hierarchy
    - Easy to lose track
    - Flat file organization

=== "`tmux`"

    - Session names only
    - No hierarchy
    - Limited metadata

=== "PentLog"

    - **Client → Engagement → Phase** hierarchy
    - Automatic organization
    - Context-aware operations
    - SQLite database index

### :material-shield-check: Compliance

=== "`script` / `tmux`"

    - No integrity verification
    - No encryption
    - No audit trail
    - Not admissible as evidence

=== "PentLog"

    - SHA256 integrity hashes (`pentlog freeze`)
    - AES-256 encrypted archives
    - Detailed audit trails
    - Timestamped everything
    - **Compliance-ready**

## When to Use Each Tool

<div class="grid cards" markdown>

-   :material-console: __Use `script` when:__

    ---

    - Quick one-off logging
    - No compliance requirements
    - Simple debugging
    - Need universal availability (pre-installed)

-   :material-window-restore: __Use `tmux` when:__

    ---

    - Need session persistence
    - Multiple windows/panes required
    - Remote work with reconnect capability
    - Team screen sharing

-   :material-shield-star: __Use PentLog when:__

    ---

    - Professional penetration testing
    - Compliance requirements (SOC2, ISO27001)
    - Evidence documentation for legal
    - Client report generation
    - Long-term organization needed
    - Audit trail required

</div>

## Migration Guide

### From `script`

!!! example "Before vs After"

    === "Old Way (`script`)"
        ```bash
        # Start recording
        script session.log

        # Do your work...
        nmap -sV target.com

        # Stop recording
        exit

        # Search: manual grep
        grep "nmap" session.log

        # Report: manual copy/paste
        # No organization, no metadata
        ```

    === "New Way (PentLog)"
        ```bash
        # Setup once
        pentlog create

        # Start recording
        pentlog shell

        # Do your work...
        nmap -sV target.com

        # Stop recording
        exit

        # Search: powerful TUI
        pentlog search "nmap.*-sV" --regex

        # Report: auto-generated
        pentlog export --analyze
        ```

### From `tmux`

!!! example "Before vs After"

    === "Old Way (`tmux`)"
        ```bash
        # Create session
        tmux new -s pentest

        # Do your work...
        # Detach to save
        tmux detach

        # Later: reattach
        tmux attach -t pentest

        # No search, no export
        # No organization
        ```

    === "New Way (PentLog)"
        ```bash
        # Setup once
        pentlog create

        # Start recorded shell
        pentlog shell

        # Do your work...
        # Exit to save
        exit

        # Replay with exact timing
        pentlog replay

        # Search across sessions
        pentlog search

        # Export professional report
        pentlog export
        ```

## Performance Comparison

| Metric | `script` | `tmux` | PentLog |
|--------|----------|--------|---------|
| **Startup Time** | ~10ms | ~50ms | ~100ms |
| **Disk Usage** | Low | Medium | Medium |
| **CPU Overhead** | Minimal | Low | Low |
| **Search Speed** | Slow (grep) | N/A | Fast (SQLite) |
| **Concurrent Sessions** | 1 | Many | Many |

!!! info "Performance Note"
    PentLog's slight startup overhead is negligible compared to the productivity gains from search, organization, and report generation.

## Summary

| Use Case | Recommended Tool |
|----------|------------------|
| Quick debugging | `script` |
| Session persistence | `tmux` |
| Professional pentesting | **PentLog** |
| Compliance/auditing | **PentLog** |
| Evidence documentation | **PentLog** |
| Team collaboration | **PentLog** |
| Certification prep (OSCP) | **PentLog** |

---

**Ready to upgrade?** [Get Started with PentLog](../getting-started/quickstart.md){ .md-button .md-button--primary }
