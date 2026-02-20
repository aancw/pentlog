# Configuration

## Config File Location

PentLog looks for config in the following order:

1. `--config` flag (highest priority)
2. `~/.pentlog/config.yaml`
3. `/etc/pentlog/config.yaml`
4. Environment variables

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
  path: "~/.pentlog/sessions"
  max_size: "10GB"
  compress_after_days: 7

# Export settings
export:
  default_format: "json"
  include_timestamps: true
  include_metadata: true

# AI Analysis settings
analysis:
  enabled: false
  provider: "openai"  # or "local"
  api_key: ""         # or use PENTLOG_AI_API_KEY env var
  model: "gpt-4"

# Security settings
security:
  encryption_key_path: "~/.pentlog/key"
  key_derivation: "argon2id"
```

## Environment Variables

All config values can be overridden via environment variables:

```bash
export PENTLOG_STORAGE_PATH=/custom/path
export PENTLOG_ANALYSIS_ENABLED=true
export PENTLOG_AI_API_KEY=sk-...
```

## Per-Project Config

Create `.pentlog.yaml` in your project root:

```yaml
session:
  default_name_format: "{{ .Project }}-{{ .Time }}"
export:
  default_format: "markdown"
```
