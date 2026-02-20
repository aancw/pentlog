# Archiving

Manage disk usage by archiving old or completed sessions.

## Interactive Archive

```bash
pentlog archive
```

Wizard guides you through:
1. Select client/engagement
2. Choose archive options
3. Optional encryption
4. Confirm and execute

## Command-Line Archiving

### Archive by Client

```bash
# Backup mode (keeps originals)
pentlog archive acme

# Archive and delete originals
pentlog archive acme --delete
```

### Archive by Age

```bash
# Archive sessions older than 30 days
pentlog archive acme --days 30

# Archive and delete old sessions
pentlog archive acme --days 30 --delete
```

### Archive by Phase

```bash
pentlog archive acme -p recon
```

### Archive by Engagement

```bash
pentlog archive acme -e internal-audit
```

## Encryption

Create password-protected archives:

```bash
# Interactive (recommended — password not in history)
pentlog archive

# Command line (avoid — password in shell history)
pentlog archive acme --password mysecret
```

### Encryption Features

- **Algorithm**: AES-256
- **Format**: ZIP with encryption
- **Compatibility**: Standard ZIP tools

## Listing Archives

```bash
pentlog archive list
```

Shows:
- Archive location
- Size
- Creation date
- Encryption status

## Importing Archives

Restore archived sessions:

```bash
# Import with auto-detected metadata
pentlog import ~/.pentlog/archive/CLIENT/20260127-192108.zip

# Import encrypted archive
pentlog import encrypted.zip

# Import with specific password
pentlog import archive.zip --password mysecret

# Import to specific context
pentlog import archive.zip -c ACME -e Q1 -p Initial
```

### Import Preview

Preview archive contents without importing:

```bash
pentlog import list archive.zip
```

## Archive Structure

Archives contain:

```
archive.zip
├── session.tty           # Terminal recording
├── session.json          # Metadata
├── notes.json           # Notes and vulnerabilities
├── report.md            # Auto-generated report
└── MANIFEST.txt         # Archive contents list
```

## Best Practices

!!! tip "Archive Regularly"
    Archive completed engagements to free up disk space.

!!! tip "Use Encryption"
    Always encrypt archives containing client data.

!!! tip "Verify Before Delete"
    Test import an archive before deleting originals with `--delete`.

!!! warning "Backup Separately"
    Archives are not a backup solution. Keep separate backups of critical data.
