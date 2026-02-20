# Timeline & Replay

Analyze and replay your terminal sessions with perfect fidelity.

## Timeline Analysis

Extract a chronological timeline of commands from any session:

```bash
pentlog timeline
```

Interactive viewer that shows:
- Commands in execution order
- Timestamps
- Working directories
- Exit codes

### Direct JSON Export

```bash
pentlog timeline <session_id> -o output.json
```

### Timeline Output Format

```json
{
  "session_id": 42,
  "commands": [
    {
      "timestamp": "14:32:15",
      "command": "nmap -sV 10.0.0.5",
      "cwd": "/home/user",
      "exit_code": 0
    },
    {
      "timestamp": "14:35:22",
      "command": "curl http://10.0.0.5:80",
      "cwd": "/home/user",
      "exit_code": 0
    }
  ]
}
```

## Session Replay

Replay recorded sessions with exact timing:

```bash
pentlog replay
```

Interactive session selector, or specify directly:

```bash
pentlog replay 42
```

### Playback Speed

```bash
# 2x speed (faster)
pentlog replay 42 -s 2.0

# 0.5x speed (slower)
pentlog replay 42 -s 0.5
```

### Replay Features

- **Faithful playback** — Exact timing preserved
- **Terminal fidelity** — Colors, formatting, cursor movements
- **Pause/Resume** — Space to pause, any key to resume
- **Exit anytime** — `q` to quit

## GIF Export

Convert sessions to animated GIFs for documentation:

```bash
pentlog gif
```

Interactive mode lets you select:
- Client/engagement
- Session
- Resolution (720p or 1080p)

### Direct Conversion

```bash
# Convert specific session
pentlog gif <session_id>

# Convert TTY file directly
pentlog gif session.tty

# Adjust playback speed
pentlog gif -s 5  # 5x speed

# Custom output filename
pentlog gif -o demo.gif

# Custom terminal dimensions
pentlog gif --cols 200 --rows 60
```

### GIF Features

| Feature | Description |
|---------|-------------|
| Resolution | 720p (1280×720) or 1080p (1920×1080) |
| Font | Go Mono for crisp text |
| Colors | Enhanced ANSI palette |
| Speed | Adjustable playback speed |

## Use Cases

### Report Documentation

```bash
# Create GIF for report
pentlog gif 42 -o exploitation-demo.gif
```

### Training Materials

```bash
# Export at slower speed for clarity
pentlog gif 42 -s 0.5 -o training.gif
```

### Evidence Review

```bash
# Replay session for review
pentlog replay 42

# Extract timeline for analysis
pentlog timeline 42 -o timeline.json
```

## Comparison: Timeline vs Replay vs GIF

| Feature | Timeline | Replay | GIF |
|---------|----------|--------|-----|
| Format | JSON/Text | Terminal | Animated GIF |
| Timing | Timestamps | Real-time | Configurable |
| Interactive | No | Yes | No |
| Shareable | Yes | No | Yes |
| Size | Small | N/A | Medium |
| Best for | Analysis | Review | Documentation |
