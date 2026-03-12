# Session Management

Manage your pentest sessions with PentLog's powerful workflow tools.

## Creating an Engagement

```bash
pentlog create
```

Interactive wizard that guides you through:

1. **Select Context Mode** — Client, Exam/Lab, or Log Only
2. **Enter Metadata** — Client name, engagement details, scope
3. **Confirm** — Review and activate the context

### Example: Client Mode

```
$ pentlog create
? Select context type: Client Mode
? Client name: ACME Corp
? Engagement name: Internal Pentest 2026
? Scope: 10.0.0.0/24
? Initial phase: reconnaissance
✓ Context created and activated
```

### Example: Exam/Lab Mode

```
$ pentlog create
? Select context type: Exam/Lab Mode
? Exam name: OSCP Lab
? Target IP: 10.10.10.5
✓ Context created and activated
```

## Starting a Recorded Shell

```bash
pentlog shell
```

Features:
- Custom PS1 showing current context
- Automatic ttyrec recording
- Hotkeys for notes and vulnerabilities
- Graceful exit handling
- **Pause/Resume support** for breaks without session fragmentation

### Pausing and Resuming Sessions

Take breaks without creating multiple disjointed sessions:

```bash
# Pause recording (while staying in shell)
pentlog pause
```

Output:
```
⏸️  Session paused successfully!
   Time: 2026-03-12 23:21:38

   Recording is paused. The shell remains active.
   Run 'pentlog resume' to continue recording.
```

```bash
# Resume recording
pentlog resume
```

Output:
```
▶️  Session resumed successfully!
   Time: 2026-03-12 23:25:15
   Paused for: 3m37s

   Recording is now active.
```

**Use Cases:**
- **OSCP Exams**: Take breaks without creating multiple sessions
- **Client Engagements**: Pause before entering sensitive environments
- **Clean Evidence**: Single continuous session per engagement phase

**How It Works:**
- Pause/resume markers are embedded in the ttyrec file with timestamps
- Markers display as formatted banners during replay
- The `.pause_marker` file tracks pause state on disk
- Shell remains active during pause (you can still run commands, but they're not recorded)

### Custom PS1 Format

```
[ACME/internal-pentest/recon] user@host:~$
```

## Switching Contexts

### Interactive Switch

```bash
pentlog switch
```

Options:
- Select from recent sessions
- Enter new manual context
- Toggle to previous session

### Quick Toggle

```bash
pentlog switch -
```

Instantly switch to the previous session (great for multi-target engagements).

### Switch by Type

**Client Mode:** Switch between phases
```bash
pentlog switch
? Select phase: exploitation
```

**Exam/Lab Mode:** Switch to new target
```bash
pentlog switch
? New Target IP: 10.10.10.6
```

## Listing Sessions

```bash
pentlog sessions list
```

Shows:
- Active and completed sessions
- Session metadata (client, engagement, phase)
- File sizes and timestamps (human-readable: KB/MB/GB)
- Session states (active/completed/crashed)

### Pagination

```bash
# Show first 20 sessions
pentlog sessions list --limit 20

# Skip first 10 sessions
pentlog sessions list --offset 10
```

## Deleting Sessions

Remove sessions and their associated files:

```bash
# Delete by ID
pentlog sessions delete 65

# Interactive mode (shows list and prompts)
pentlog sessions delete
```

**What gets deleted:**
- `.tty` recording file
- `.json` metadata file  
- `.notes.json` notes file

**Safety features:**
- Confirmation prompt with session details
- Validates session ID exists
- Shows file size in confirmation

## Session States

| State | Description | Action Needed |
|-------|-------------|---------------|
| `active` | Currently recording | None |
| `completed` | Ended normally | None |
| `crashed` | Terminated unexpectedly | Run `pentlog recover` |

## Session Size Monitoring

PentLog monitors session file sizes to prevent performance issues with large recordings:

### How It Works

- **Background monitoring** every 30 seconds during shell sessions
- **Warning at 5MB**: "⚡ Session size: 5.0 MB - Approaching limit"
- **Alert at 10MB**: "⚠️ Session size: 10.0 MB - Consider splitting session"
- **5-minute cooldown** between alerts (prevents spam)

### Why These Thresholds?

The 5MB/10MB limits optimize performance for:
- **Replay**: Fast ttyplay parsing
- **Search**: Quick log scanning
- **Export**: Efficient report generation
- **GIF Generation**: Smooth frame extraction

### Managing Large Sessions

When you see the alert:
```bash
# Exit current session
exit

# Start a new session (creates separate .tty file)
pentlog shell
```

This keeps individual session files manageable and ensures fast processing.

## Resetting Context

Clear the current active engagement:

```bash
pentlog reset
```

This removes the active context but **preserves all recorded sessions**.

## Status Check

View current tool and engagement status:

```bash
pentlog status
```

Output:
```
PentLog v1.2.3
Active Context: ACME Corp / Internal Pentest 2026 / recon
Database: ~/.pentlog/pentlog.db
Storage: 2.3 GB used
```

## Best Practices

!!! tip "Naming Conventions"
    Use consistent naming for clients and engagements. This makes searching and archiving easier.

!!! tip "Phase Organization"
    In Client Mode, use phases to separate reconnaissance, exploitation, and post-exploitation activities.

!!! tip "Regular Exports"
    Export reports at the end of each day to avoid losing work.

!!! warning "Don't Forget to Exit"
    Always exit the shell properly (`exit` or Ctrl+D) to ensure the session is marked complete.
