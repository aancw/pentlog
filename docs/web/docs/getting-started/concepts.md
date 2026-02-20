# Core Concepts

Understanding PentLog's design philosophy and core concepts.

## Evidence-First Design

PentLog is built around one principle: **your logs are evidence**. Every feature supports this:

- Perfect terminal fidelity (what you see is what you get)
- Automatic integrity protection
- Compliance-ready export formats
- Encrypted archives for secure delivery

## Context Modes

PentLog adapts to your workflow with three context modes:

### Client Mode

For professional penetration testing engagements.

**Hierarchy:** `Client → Engagement → Scope → Phase`

```
Client: ACME Corp
  └── Engagement: Internal Pentest 2026
        └── Scope: 10.0.0.0/24
              └── Phase: exploitation
```

**Best for:**
- Real-world client engagements
- Compliance audits
- Long-term projects with multiple phases

### Exam/Lab Mode

Optimized for certifications and CTFs.

**Hierarchy:** `Exam Name → Target IP`

```
Exam: OSCP Lab
  └── Target: 10.10.10.5
```

**Best for:**
- OSCP, PNPT, eJPT preparation
- HackTheBox / TryHackMe
- CTF competitions

### Log Only Mode

Minimal setup for quick logging.

**Hierarchy:** `Project Name`

```
Project: Quick Research
```

**Best for:**
- Ad-hoc research
- Quick tests
- Personal notes

## Session Organization

### Directory Structure

```
~/.pentlog/
├── context.json          # Current active context
├── pentlog.db           # SQLite database (metadata)
├── logs/
│   └── CLIENT/
│       └── ENGAGEMENT/
│           └── PHASE/
│               └── manual-OPERATOR-TIMESTAMP.tty
├── reports/             # Generated reports
├── archive/             # Archived sessions
└── hashes/              # Integrity hashes
```

### Session Lifecycle

```
Create Context → Start Shell → Record Commands → Stop Shell → Search/Export
```

## Terminal Fidelity

Unlike `script` or `tmux`, PentLog uses a **Virtual Terminal Emulator** to capture:

- ANSI colors and formatting
- Cursor movements and redraws
- Terminal resizes
- Special characters

This means what you see in the viewer is *exactly* what you saw in your shell.

## Metadata Capture

Every session captures:

| Metadata | Description |
|----------|-------------|
| Timestamp | Every command timestamped |
| Operator | Who ran the command |
| Context | Client/Engagement/Phase |
| Working Directory | Where the command ran |
| Exit Code | Success/failure indication |

## Data Integrity

### SHA256 Hashes

```bash
pentlog freeze
```

Generates SHA256 hashes of all session logs for evidence integrity.

### Encryption

```bash
pentlog archive --encrypt
```

Creates AES-256 encrypted archives for secure client delivery.

## Search Architecture

PentLog indexes all content in SQLite for fast searching:

- Full-text search across all sessions
- Regex support
- Boolean operators (AND, OR, NOT)
- Filter by date, client, engagement

## Report Generation

Export formats:

| Format | Use Case |
|--------|----------|
| Markdown | GitHub, documentation |
| HTML | Client reports, browser viewing |
| JSON | Automation, integration |

All reports include:
- Command timeline
- Notes and bookmarks
- AI-generated summaries (optional)
- Integrity verification
