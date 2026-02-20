# Live Sharing

Share your terminal session in real-time with teammates or reviewers.

## Starting a Shared Session

```bash
pentlog shell --share
```

The share URL is displayed in the shell banner:

```
╔══════════════════════════════════════════════════════════════╗
║  PentLog Live Share                                          ║
║  URL: http://192.168.1.13:44143/watch?token=abc123...       ║
╚══════════════════════════════════════════════════════════════╝
```

### Custom Port and Bind Address

```bash
pentlog shell --share --share-port 8080 --share-bind 0.0.0.0
```

## Sharing Existing Sessions

Share a recorded session for read-only viewing:

```bash
pentlog share <session-id>
```

### Share Status

Check connected viewers:

```bash
pentlog share status
```

Output:
```
PentLog Live Share (Active)
──────────────────────────────────────────
URL:     http://192.168.1.13:44143/watch?token=...
PID:     12345
Viewers: 2
         - 192.168.1.10:52341
         - 192.168.1.20:48823
──────────────────────────────────────────
```

### Stop Sharing

```bash
pentlog share stop
```

## Viewer Features

| Feature | Description |
|---------|-------------|
| **Dark Theme** | Modern dark-themed terminal viewer using xterm.js |
| **Scrollback** | Late-joining viewers see full session history |
| **Read-Only** | Viewers can only watch, no input accepted |
| **Auto-Reconnect** | Automatic reconnection on connection loss |
| **Token Auth** | Unique token for each session |

## Use Cases

### Remote Pairing

Share your session with a senior pentester for real-time guidance:

```bash
pentlog shell --share
# Share URL via secure channel
```

### Client Demonstrations

Show findings to clients in real-time:

```bash
pentlog shell --share --share-port 443
```

### Training

Instructor shares terminal with students:

```bash
pentlog shell --share --share-bind 0.0.0.0
```

## Security Considerations

!!! warning "Token Expiration"
    Share tokens are unique per session but don't expire automatically. Stop sharing when done.

!!! warning "Network Exposure"
    Using `--share-bind 0.0.0.0` exposes the share port on all interfaces. Use firewall rules to restrict access.

!!! tip "HTTPS Recommended"
    For production use, place a reverse proxy with HTTPS in front of the share server.

## Troubleshooting

### Port Already in Use

```bash
# Use a different port
pentlog shell --share --share-port 8081
```

### Connection Refused

Check firewall rules and ensure the bind address is correct:

```bash
# Bind to specific interface
pentlog shell --share --share-bind 192.168.1.10
```
