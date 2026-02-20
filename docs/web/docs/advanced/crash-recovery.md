# Crash Recovery

Protect your evidence from unexpected session terminations.

## How It Works

PentLog tracks session state with a heartbeat mechanism:

1. **Session State Tracking** — Each session is `active`, `completed`, or `crashed`
2. **Heartbeat** — Updated every 30 seconds during recording
3. **Stale Detection** — No heartbeat for 5+ minutes = crashed
4. **Startup Warning** — Any pentlog command warns about crashed sessions

## Session States

| State | Description | Indicator |
|-------|-------------|-----------|
| `active` | Currently recording | :green_circle: |
| `completed` | Ended normally | :white_check_mark: |
| `crashed` | Terminated unexpectedly | :warning: |

## Detecting Crashed Sessions

### Automatic Warning

On any pentlog command:

```bash
$ pentlog sessions

⚠️  Warning: 1 crashed session(s) detected.
   Run 'pentlog recover' to review and recover them.
```

### List Crashed Sessions

```bash
pentlog recover --list
```

## Recovery Options

### Interactive Recovery

```bash
pentlog recover
```

Menu options:
- List crashed/stale/orphaned sessions
- Recover specific session
- Recover all crashed sessions
- Mark stale sessions as crashed
- Clean up orphaned entries

### Recover Specific Session

```bash
pentlog recover --recover 42
```

### Recover All Crashed Sessions

```bash
pentlog recover --recover-all
```

### Mark Stale as Crashed

```bash
pentlog recover --mark-stale
```

### Clean Orphans

Remove database entries with missing files:

```bash
pentlog recover --clean-orphans
```

## Common Scenarios

### SSH Disconnect

```bash
# SSH drops during 4-hour exam
# Reconnect and run any pentlog command
$ pentlog sessions

⚠️  Warning: 1 crashed session(s) detected.

# Recover the session
$ pentlog recover
✓ Session 42 recovered successfully

# Session is now usable
$ pentlog replay 42
```

### System OOM Kill

```bash
# Process killed by out-of-memory
# On next pentlog command, session is marked crashed
$ pentlog recover --recover-all
✓ Recovered 1 crashed session(s)
```

### Power Failure

```bash
# Power loss during recording
# After reboot, session is marked crashed
$ pentlog recover
# Review and recover as needed
```

## Recovery Workflow

```
Crashed Session Detected
        ↓
   pentlog recover
        ↓
   Review Sessions
        ↓
   Recover Selected
        ↓
   Session Usable
```

## What Gets Recovered

Recovery ensures:
- :white_check_mark: TTY recording preserved
- :white_check_mark: Metadata intact
- :white_check_mark: Notes and vulnerabilities saved
- :white_check_mark: Searchable in database
- :white_check_mark: Exportable to reports

## Prevention Tips

!!! tip "Stable Connection"
    Use `tmux` or `screen` on remote systems to survive disconnects.

!!! tip "Regular Exports"
    Export reports periodically during long engagements.

!!! tip "Monitor Resources"
    Watch memory usage to avoid OOM kills.
