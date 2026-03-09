# CLI Reference

## Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Path to config file |
| `--verbose` | Enable verbose output |
| `--help` | Show help message |

## Commands

### `pentlog start`

Start a new logging session.

```bash
pentlog start [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | auto-generated | Session name |
| `--encrypt` | false | Enable encryption |
| `--output` | `~/.pentlog/sessions` | Output directory |

### `pentlog sessions`

Manage recorded sessions.

```bash
# List sessions
pentlog sessions list

# Delete a session
pentlog sessions delete <id>
```

**Subcommands:**

| Command | Description |
|---------|-------------|
| `list` | List all recorded sessions |
| `delete` | Delete a session by ID |

**List Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--limit` | 0 | Maximum sessions to display |
| `--offset` | 0 | Number of sessions to skip |

### `pentlog export`

Export a session to various formats.

```bash
pentlog export --session-id <id> [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | json | Export format (json, html, markdown) |
| `--output` | stdout | Output file (default: print to stdout) |

### `pentlog analyze`

Run AI analysis on a session.

```bash
pentlog analyze --session-id <id> [flags]
```

### `pentlog config`

Manage configuration.

```bash
pentlog config init    # Create default config
pentlog config show    # Display current config
pentlog config edit    # Edit config in $EDITOR
```

## Examples

!!! tip "Pro Tip"
    Use session names that reflect the target or purpose for easier organization.

```bash
# Start named session with encryption
pentlog start --name "client-x-webapp" --encrypt

# Export to HTML report
pentlog export --session-id abc123 --format html --output report.html

# List all sessions as JSON
pentlog list --format json | jq '.[] | select(.duration > "1h")'
```
