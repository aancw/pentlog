# AI Analysis

Summarize your findings with AI-powered analysis using Google Gemini or Ollama.

## Supported Providers

| Provider | Type | Requirements |
|----------|------|--------------|
| **Google Gemini** | Cloud | API key |
| **Ollama** | Local | Self-hosted |

## Setup

### Configure AI Provider

```bash
pentlog setup ai
```

Interactive setup creates `~/.pentlog/ai_config.yaml`:

```yaml
provider: gemini
api_key: your-api-key-here
model: gemini-pro
```

### View Configuration

```bash
cat ~/.pentlog/ai_config.yaml
```

## Using AI Analysis

### Analyze Existing Report

```bash
# Summarized analysis (default)
pentlog analyze report.md

# Full detailed analysis
pentlog analyze --full-report report.md
```

### Analyze During Export

```bash
# Export with AI summary
pentlog export --analyze

# Export with full analysis
pentlog export --analyze --full-report
```

## Analysis Output

### Summarized Analysis

- High-level findings overview
- Key vulnerabilities identified
- Risk assessment
- Quick recommendations

### Full Report Analysis

- Detailed technical findings
- Step-by-step attack chain
- Comprehensive risk analysis
- Actionable remediation steps
- Compliance implications

## Example Output

```markdown
## AI Analysis Summary

### Executive Summary
This penetration test identified 3 critical vulnerabilities and 5 high-severity issues in the ACME Corp internal network.

### Key Findings
1. **SQL Injection** (Critical) — Login form vulnerable to blind SQL injection
2. **Unencrypted Database** (Critical) — Customer PII stored in plaintext
3. **Default Credentials** (High) — Admin panel accessible with default passwords

### Risk Assessment
- **Overall Risk**: Critical
- **Likelihood of Exploitation**: High
- **Business Impact**: Severe

### Recommendations
1. Implement parameterized queries for all database interactions
2. Enable encryption at rest for sensitive data
3. Enforce strong password policies
```

## Ollama Setup (Local LLM)

For offline/air-gapped environments:

1. Install Ollama: https://ollama.ai
2. Pull a model:
   ```bash
   ollama pull llama2
   ```
3. Configure PentLog:
   ```yaml
   provider: ollama
   endpoint: http://localhost:11434
   model: llama2
   ```

## Privacy Considerations

!!! warning "Cloud Provider"
    Using Google Gemini sends report data to Google's servers. Ensure this complies with your client's data handling requirements.

!!! tip "Local Alternative"
    Use Ollama for completely offline analysis. No data leaves your machine.

## Cost Considerations

| Provider | Cost | Notes |
|----------|------|-------|
| Google Gemini | Free tier available | Rate limits apply |
| Ollama | Free | Requires local compute resources |
