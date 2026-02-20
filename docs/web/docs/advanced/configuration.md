# Configuration

Customize PentLog to fit your workflow.

## Config File Location

PentLog looks for config in this order:

1. `--config` flag (highest priority)
2. `~/.pentlog/config.yaml`
3. Environment variables

## Default Configuration

```yaml
# ~/.pentlog/config.yaml

# Session settings
session:
  default_name_format: "session-{{ .Time }}"
  auto_encrypt: false
  retention_days: 90

# Storage settings
storage:
  path: "~/.pentlog/logs"
  max_size: "10GB"
  compress_after_days: 7

# Export settings
export:
  default_format: "markdown"
  include_timestamps: true
  include_metadata: true
  template_path: "~/.pentlog/templates"

# AI Analysis settings
analysis:
  enabled: false
  provider: "gemini"  # or "ollama"
  api_key: ""         # or use PENTLOG_AI_API_KEY env var
  model: "gemini-pro"

# Security settings
security:
  encryption_key_path: "~/.pentlog/key"
  key_derivation: "argon2id"

# UI settings
ui:
  theme: "dark"
  pager: "less -R"
  editor: "$EDITOR"
```

## Environment Variables

Override any config value:

```bash
export PENTLOG_STORAGE_PATH=/custom/path
export PENTLOG_ANALYSIS_ENABLED=true
export PENTLOG_AI_API_KEY=sk-...
export PENTLOG_RETENTION_DAYS=30
```

## Per-Project Config

Create `.pentlog.yaml` in your project root:

```yaml
session:
  default_name_format: "{{ .Project }}-{{ .Time }}"
export:
  default_format: "html"
analysis:
  enabled: true
```

## Shell Completion

Generate completion scripts:

```bash
pentlog completion
```

Select your shell and follow installation instructions.

## Template Customization

Report templates are stored in `~/.pentlog/templates/`:

### Markdown Template

```markdown
# {{ .Client }} - {{ .Engagement }} Report

**Date**: {{ .Date }}
**Operator**: {{ .Operator }}
**Scope**: {{ .Scope }}

## Executive Summary

{{ .Summary }}

## Findings

{{ .Findings }}

## Timeline

{{ .Timeline }}
```

### Custom Template Variables

| Variable | Description |
|----------|-------------|
| `{{ .Client }}` | Client name |
| `{{ .Engagement }}` | Engagement name |
| `{{ .Phase }}` | Current phase |
| `{{ .Operator }}` | Operator name |
| `{{ .Date }}` | Report date |
| `{{ .Scope }}` | Scope details |

## Advanced Settings

### Database Path

```yaml
storage:
  database_path: "~/.pentlog/pentlog.db"
```

### Log Rotation

```yaml
storage:
  max_file_size: "100MB"
  max_files: 100
```

### Custom Pager

```yaml
ui:
  pager: "bat --paging=always"
```
