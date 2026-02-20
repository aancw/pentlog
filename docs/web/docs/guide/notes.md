# Notes & Bookmarks

Add timestamped annotations to your sessions for later review.

## Quick Notes During Shell Sessions

While inside `pentlog shell`, use keyboard shortcuts:

| Hotkey | Action | Description |
|--------|--------|-------------|
| `Ctrl+N` | Quick Note | Add a one-line note instantly |
| `Ctrl+G` | Quick Vuln | Log a vulnerability with severity |

### Adding a Quick Note

Press `Ctrl+N`:

```
📝 Quick note: Found open port 8080
✓ Note saved [14:05:43]
```

### Logging a Vulnerability

Press `Ctrl+G`:

```
🔓 Vuln title: SQL Injection in login form
Severity (c/h/m/l/i): h
Description (optional): POST /login endpoint vulnerable to blind SQLi
✓ Vuln saved: V-abc123 [High] SQL Injection in login form
```

Severity levels:

| Code | Level | Description |
|------|-------|-------------|
| `c` | Critical | Immediate action required |
| `h` | High | Significant risk |
| `m` | Medium | Moderate risk |
| `l` | Low | Minor issue |
| `i` | Info | Informational |

## Managing Notes

### List Notes

```bash
pentlog note list
```

Works both:
- **Inside shell**: Shows current session notes
- **Offline**: Interactive selector for past sessions

### Add Note Manually

```bash
pentlog note add "Found interesting config file"
```

### View Vulnerabilities

```bash
pentlog vuln list
```

Shows all logged vulnerabilities with:
- Severity indicators
- Timestamps
- Descriptions
- IDs for reference

## Note Format

Notes are stored with:

```json
{
  "timestamp": "2026-02-20T14:05:43Z",
  "session_id": 42,
  "type": "note",
  "content": "Found open port 8080",
  "command_context": "nmap -sV 10.0.0.5"
}
```

## Vulnerability Format

```json
{
  "id": "V-abc123",
  "timestamp": "2026-02-20T14:10:22Z",
  "session_id": 42,
  "type": "vulnerability",
  "title": "SQL Injection in login form",
  "severity": "high",
  "description": "POST /login endpoint vulnerable to blind SQLi",
  "evidence": "sqlmap output showing injection point"
}
```

## Best Practices

!!! tip "Note Early, Note Often"
    Add notes immediately when you find something interesting. Don't rely on memory.

!!! tip "Use Consistent Severity"
    Be consistent with severity ratings for easier prioritization later.

!!! tip "Include Context"
    When adding manual notes, include enough context to understand the finding later.

!!! example "Good Note Examples"
    - "Port 8080 open — Apache Tomcat 9.0"
    - "Found .git directory exposed"
    - "Admin panel at /admin with default creds"

## Exporting Notes

Notes are automatically included in reports:

```bash
pentlog export
```

The exported report includes:
- All notes with timestamps
- All vulnerabilities with severity
- Command context for each note
- Links to relevant session sections
