# Export & Reports

Generate professional reports from your pentest sessions.

## Export Formats

PentLog supports multiple export formats:

| Format | Extension | Best For |
|--------|-----------|----------|
| Markdown | `.md` | GitHub, documentation, editing |
| HTML | `.html` | Client reports, browser viewing |
| JSON | `.json` | Automation, integration |

## Interactive Export

```bash
pentlog export
```

Wizard guides you through:
1. Select client/engagement/phase
2. View existing reports
3. Choose format
4. Preview or save

### Export with Overwrite Protection

If a report already exists, PentLog warns you:

```
⚠️  Report already exists: ACME-internal-pentest-report.md
? Overwrite? (y/N)
```

## Command-Line Export

### Export Specific Engagement

```bash
pentlog export acme -e incident-response
```

### Export with AI Analysis

```bash
# Summarized analysis (default)
pentlog export --analyze

# Full detailed analysis
pentlog export --analyze --full-report
```

## Report Contents

Exported reports include:

### Header Section
- Client/engagement information
- Date range
- Operator name
- Scope details

### Executive Summary
- Session overview
- Key findings count
- Total commands executed
- Time spent

### Timeline
- Chronological command list
- Timestamps
- Working directories
- Exit codes

### Notes & Vulnerabilities
- All notes with timestamps
- Vulnerabilities with severity
- Evidence references

### AI Summary (if enabled)
- High-level findings summary
- Risk assessment
- Recommendations

## Viewing HTML Reports

HTML reports with embedded GIF players require HTTP access:

```bash
pentlog serve
```

This starts a local HTTP server and opens the report in your browser.

### Custom Port

```bash
pentlog serve --port 8080
```

## Report Templates

PentLog uses templates stored in `~/.pentlog/templates/`:

- `report.md` — Markdown template
- `report.html` — HTML template

Customize these to match your organization's branding.

## Best Practices

!!! tip "Export Regularly"
    Export reports at the end of each day to avoid losing work.

!!! tip "Use Descriptive Names"
    The default naming includes timestamps, but add context for clarity.

!!! tip "Include AI Analysis"
    Enable AI analysis for executive summaries in client reports.

!!! warning "Verify Before Sending"
    Always review exported reports before client delivery to ensure no sensitive data is exposed.
