# Storage Layout

Understanding PentLog's file organization.

## Directory Structure

```
~/.pentlog/
├── context.json              # Current active context
├── config.yaml              # User configuration
├── ai_config.yaml           # AI provider settings
├── pentlog.db               # SQLite database (metadata)
│
├── logs/                    # Session recordings
│   └── CLIENT/
│       └── ENGAGEMENT/
│           └── PHASE/
│               ├── manual-OPERATOR-TIMESTAMP.tty
│               └── manual-OPERATOR-TIMESTAMP.json
│
├── reports/                 # Generated reports
│   └── CLIENT/
│       ├── engagement-report-20260127.md
│       └── engagement-report-20260127.html
│
├── archive/                 # Archived sessions
│   └── CLIENT/
│       └── 20260127-192108.zip
│
├── templates/               # Report templates
│   ├── report.md
│   └── report.html
│
├── hashes/                  # Integrity hashes
│   └── sha256.txt
│
└── share/                   # Live share assets
    └── viewer.html
```

## File Types

### TTY Files

Terminal recordings in ttyrec format:

```
manual-operator-20260127-143022.tty
```

- Binary format with timing information
- Viewable with `ttyplay` or `pentlog replay`
- Convertible to GIF with `pentlog gif`

### JSON Metadata

Session metadata:

```json
{
  "session_id": 42,
  "client": "ACME Corp",
  "engagement": "Internal Pentest 2026",
  "phase": "reconnaissance",
  "operator": "operator",
  "start_time": "2026-01-27T14:30:22Z",
  "end_time": "2026-01-27T16:45:10Z",
  "commands": 156,
  "size_bytes": 2457600
}
```

### SQLite Database

Indexed metadata for fast searching:

| Table | Purpose |
|-------|---------|
| `sessions` | Session metadata |
| `commands` | Command history |
| `notes` | Notes and bookmarks |
| `vulnerabilities` | Vulnerability log |

## Context File

`context.json` stores the active engagement:

```json
{
  "mode": "client",
  "client": "ACME Corp",
  "engagement": "Internal Pentest 2026",
  "scope": "10.0.0.0/24",
  "phase": "exploitation",
  "operator": "operator"
}
```

## Permissions

Sensitive files are created with restricted permissions:

| Path | Permissions |
|------|-------------|
| `~/.pentlog/` | `0700` (owner only) |
| `*.tty` files | `0600` (owner read/write) |
| `*.db` | `0600` (owner read/write) |
| `context.json` | `0600` (owner read/write) |

## Backup Considerations

### Critical Files

Priority for backup:

1. `pentlog.db` — All metadata and search index
2. `logs/**/*.tty` — Session recordings
3. `archive/` — Archived sessions

### Excluded Files

Can be regenerated:

- `reports/` — Re-export from sessions
- `hashes/` — Re-run `pentlog freeze`
- `share/` — Temporary live share files

## Disk Usage Management

### Check Usage

```bash
du -sh ~/.pentlog
```

### Cleanup Strategies

1. **Archive old sessions**: `pentlog archive --days 30`
2. **Delete old reports**: Reports can be re-exported
3. **Compress logs**: Use `pentlog archive` with compression

## Migration

### Moving to New Machine

1. Copy `~/.pentlog/` directory
2. Ensure same permissions
3. Run `pentlog setup` to verify

### Importing from Archive

```bash
pentlog import archive.zip
```
