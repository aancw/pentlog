# Search & Analysis

Find any command across all your sessions with PentLog's powerful search.

## Interactive Search

```bash
pentlog search
```

Launches a TUI with:
- Live incremental search
- Regex support
- Boolean operators
- Result navigation

### Search Interface

```
Search: nmap.*-sV
─────────────────────────────────────────
[42] ACME/internal-pentest/recon
     14:32:15  nmap -sV -p- 10.0.0.5
     14:35:22  nmap -sV --script vuln 10.0.0.10

Result 1/15  (↑↓ Navigate, Enter to view)
```

## Command-Line Search

### Basic Query

```bash
pentlog search "vulnerability"
```

### Regex Search

```bash
pentlog search "nmap.*-sV" --regex
```

### Boolean Operators

```bash
# AND (both terms must appear)
pentlog search "sqlmap AND injection"

# OR (either term)
pentlog search "nmap OR masscan"

# NOT (exclude term)
pentlog search "exploit NOT metasploit"
```

### Date Filtering

```bash
# After specific date
pentlog search "exploit" --after 15012026

# Before specific date
pentlog search "recon" --before 31012026

# Date range
pentlog search "payload" --after 01012026 --before 31012026
```

## Dashboard

View an interactive executive summary:

```bash
pentlog dashboard
```

Shows:
- Evidence size and session count
- Recent findings
- Statistical breakdowns
- Activity timeline

## Search Scope

PentLog searches across:

| Content Type | Included |
|--------------|----------|
| Commands | :white_check_mark: |
| Command output | :white_check_mark: |
| Notes | :white_check_mark: |
| Vulnerability titles | :white_check_mark: |
| Vulnerability descriptions | :white_check_mark: |
| Session metadata | :white_check_mark: |

## Advanced Search Tips

### Find Specific Commands

```bash
# Find all curl commands
pentlog search "^curl"

# Find commands with specific flags
pentlog search "nmap.*-p-"

# Find SQL injection attempts
pentlog search "(sqlmap|sqlninja|sqldump)"
```

### Find by Context

```bash
# Search within specific client
pentlog search "exploit" --client ACME

# Search within specific engagement
pentlog search "payload" --engagement "Internal Pentest"

# Search within specific phase
pentlog search "scan" --phase reconnaissance
```

### Export Search Results

```bash
# Export matching sessions
pentlog search "critical" --export results.md
```

## Search Performance

PentLog uses SQLite full-text search (FTS5) for:
- Sub-second search across thousands of sessions
- Indexed content for fast retrieval
- Efficient regex matching

## Use Cases

### Compliance Audit

```bash
# Find all sudo commands for privilege escalation review
pentlog search "^sudo" --after 01012026
```

### Report Writing

```bash
# Find all exploitation commands
pentlog search "(exploit|payload|shell)" --phase exploitation
```

### Incident Response

```bash
# Find commands around a specific time
pentlog search "malware" --after 14022026 --before 15022026
```
