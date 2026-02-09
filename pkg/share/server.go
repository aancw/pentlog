package share

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type Server struct {
	hub      *Hub
	token    string
	listener net.Listener
	server   *http.Server
}

func GenerateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func NewServer(hub *Hub, token string) *Server {
	return &Server{
		hub:   hub,
		token: token,
	}
}

func (s *Server) Start(addr string) (string, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/watch", s.handleWatch)
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/status", s.handleStatus)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = ln

	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		s.server.Serve(ln)
	}()

	return ln.Addr().String(), nil
}

func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}

func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}

func (s *Server) handleWatch(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") != s.token {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := strings.ReplaceAll(watchPageHTML, "__TOKEN__", s.token)
	w.Write([]byte(html))
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") != s.token {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	remoteAddr := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		remoteAddr = forwarded
	}

	client := &Client{
		hub:        s.hub,
		conn:       conn,
		send:       make(chan []byte, clientSendBufSize),
		remoteAddr: remoteAddr,
	}

	s.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") != s.token {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients":    s.hub.ClientCount(),
		"client_ips": s.hub.ClientIPs(),
	})
}

const watchPageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>PentLog Live Watch</title>
<style>
  * {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }
  
  html, body {
    width: 100%;
    height: 100%;
  }
  
  body {
    background: #0d1117;
    display: flex;
    flex-direction: column;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
    color: #e6edf3;
  }
  
  #header {
    background: #161b22;
    padding: 12px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid #30363d;
  }
  
  #header .left {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  
  #header .logo {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  #header .logo svg {
    width: 20px;
    height: 20px;
  }
  
  #header .title {
    font-weight: 600;
    font-size: 14px;
    color: #e6edf3;
    letter-spacing: 0.3px;
  }

  #header .separator {
    color: #30363d;
    font-size: 18px;
    font-weight: 300;
  }

  #header .subtitle {
    font-size: 13px;
    color: #8b949e;
    font-weight: 400;
  }
  
  #header .right {
    display: flex;
    gap: 10px;
    align-items: center;
  }
  
  .badge {
    background: rgba(187, 128, 9, 0.15);
    color: #d29922;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.5px;
    border: 1px solid rgba(187, 128, 9, 0.3);
  }
  
  #status {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 500;
    padding: 4px 10px;
    border-radius: 20px;
    background: #21262d;
    color: #8b949e;
    border: 1px solid #30363d;
    transition: all 0.3s ease;
  }
  
  #status.connecting {
    color: #8b949e;
    background: #21262d;
    border-color: #30363d;
  }
  
  #status.connected {
    color: #3fb950;
    background: rgba(46, 160, 67, 0.1);
    border-color: rgba(46, 160, 67, 0.3);
  }
  
  #status.disconnected {
    color: #f85149;
    background: rgba(248, 81, 73, 0.1);
    border-color: rgba(248, 81, 73, 0.3);
  }
  
  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
  }
  
  #status.connected .status-dot {
    animation: pulse 2s ease-in-out infinite;
  }
  
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }
  
  #terminal-wrapper {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: #0d1117;
    padding: 4px 0 0 0;
  }

  #terminal {
    flex: 1;
    overflow: hidden;
  }
  
  .xterm {
    height: 100%;
    padding: 0 8px;
  }
  
  .xterm-viewport {
    background: #0d1117 !important;
  }

  .xterm-viewport::-webkit-scrollbar {
    width: 8px;
  }

  .xterm-viewport::-webkit-scrollbar-track {
    background: #0d1117;
  }

  .xterm-viewport::-webkit-scrollbar-thumb {
    background: #30363d;
    border-radius: 4px;
  }

  .xterm-viewport::-webkit-scrollbar-thumb:hover {
    background: #484f58;
  }
  
  #footer {
    background: #161b22;
    padding: 6px 20px;
    border-top: 1px solid #30363d;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  #footer .info {
    font-size: 11px;
    color: #484f58;
  }

  #footer .brand {
    font-size: 11px;
    color: #484f58;
  }

  #footer .brand span {
    color: #6e7681;
    font-weight: 500;
  }
</style>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@xterm/xterm@5.5.0/css/xterm.min.css">
</head>
<body>
<div id="header">
  <div class="left">
    <div class="logo">
      <svg viewBox="0 0 24 24" fill="none" stroke="#3fb950" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="4 17 10 11 4 5"></polyline>
        <line x1="12" y1="19" x2="20" y2="19"></line>
      </svg>
      <span class="title">PentLog</span>
    </div>
    <span class="separator">/</span>
    <span class="subtitle">Live Watch</span>
  </div>
  <div class="right">
    <span class="badge">READ-ONLY</span>
    <div id="status" class="connecting">
      <span class="status-dot"></span>
      <span id="status-text">Connecting...</span>
    </div>
  </div>
</div>
<div id="terminal-wrapper">
  <div id="terminal"></div>
</div>
<div id="footer">
  <span class="info">View-only terminal session</span>
  <span class="brand">Powered by <span>PentLog</span></span>
</div>

<script src="https://cdn.jsdelivr.net/npm/@xterm/xterm@5.5.0/lib/xterm.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/@xterm/addon-fit@0.10.0/lib/addon-fit.min.js"></script>
<script>
const term = new Terminal({
  cursorBlink: true,
  cursorStyle: 'bar',
  disableStdin: true,
  theme: {
    background: '#0d1117',
    foreground: '#e6edf3',
    cursor: '#58a6ff',
    selectionBackground: '#264f78',
    selectionForeground: '#e6edf3',
    black: '#484f58',
    red: '#ff7b72',
    green: '#3fb950',
    yellow: '#d29922',
    blue: '#58a6ff',
    magenta: '#bc8cff',
    cyan: '#39d2c0',
    white: '#b1bac4',
    brightBlack: '#6e7681',
    brightRed: '#ffa198',
    brightGreen: '#56d364',
    brightYellow: '#e3b341',
    brightBlue: '#79c0ff',
    brightMagenta: '#d2a8ff',
    brightCyan: '#56d4dd',
    brightWhite: '#f0f6fc'
  },
  fontFamily: "'JetBrains Mono', 'Fira Code', Menlo, Monaco, 'Courier New', monospace",
  fontSize: 13,
  fontWeight: 400,
  lineHeight: 1.35,
  scrollback: 10000,
  allowTransparency: false,
  smoothScroll: true,
});

const fitAddon = new FitAddon.FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal'));

window.addEventListener('resize', () => {
  try { fitAddon.fit(); } catch(e) {}
});

const statusEl = document.getElementById('status');
const statusText = document.getElementById('status-text');

const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsURL = proto + '//' + location.host + '/ws?token=__TOKEN__';

let ws;
let reconnectTimer;
let reconnectAttempts = 0;
const maxReconnectAttempts = 10;

function connect() {
  if (reconnectAttempts >= maxReconnectAttempts) {
    statusEl.className = 'disconnected';
    statusText.textContent = 'Connection failed';
    return;
  }

  try {
    ws = new WebSocket(wsURL);
    ws.binaryType = 'arraybuffer';

    ws.onopen = () => {
      statusEl.className = 'connected';
      statusText.textContent = 'Connected';
      reconnectAttempts = 0;
      if (reconnectTimer) clearTimeout(reconnectTimer);
    };

    ws.onmessage = (e) => {
      term.write(new Uint8Array(e.data));
    };

    ws.onclose = (e) => {
      statusEl.className = 'disconnected';
      statusText.textContent = 'Reconnecting...';
      reconnectAttempts++;
      if (reconnectAttempts < maxReconnectAttempts) {
        reconnectTimer = setTimeout(connect, 2000);
      }
    };

    ws.onerror = (e) => {
      statusEl.className = 'disconnected';
      statusText.textContent = 'Connection error';
    };
  } catch(e) {
    statusEl.className = 'disconnected';
    statusText.textContent = 'Error';
    reconnectAttempts++;
    if (reconnectAttempts < maxReconnectAttempts) {
      reconnectTimer = setTimeout(connect, 2000);
    }
  }
}

setTimeout(() => {
  fitAddon.fit();
  connect();
}, 100);
</script>
</body>
</html>`
