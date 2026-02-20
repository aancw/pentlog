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
pentlog sessions
```

Shows:
- Active and completed sessions
- Session metadata (client, engagement, phase)
- File sizes and timestamps
- Session states (active/completed/crashed)

## Session States

| State | Description | Action Needed |
|-------|-------------|---------------|
| `active` | Currently recording | None |
| `completed` | Ended normally | None |
| `crashed` | Terminated unexpectedly | Run `pentlog recover` |

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
