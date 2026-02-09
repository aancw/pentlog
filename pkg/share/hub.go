package share

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	clientSendBufSize  = 4096
	writeWait          = 10 * time.Second
	pongWait           = 60 * time.Second
	pingPeriod         = (pongWait * 9) / 10
	maxScrollbackBytes = 50 * 1024 * 1024
)

type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	remoteAddr string
}

type Hub struct {
	clients        map[*Client]struct{}
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	mu             sync.RWMutex
	done           chan struct{}
	scrollback     [][]byte
	scrollbackSize int
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	defer close(h.done)
	for {
		select {
		case client, ok := <-h.register:
			if !ok {
				return
			}
			h.mu.Lock()
			if len(h.scrollback) > 0 {
				combined := make([]byte, 0, h.scrollbackSize)
				for _, data := range h.scrollback {
					combined = append(combined, data...)
				}
				select {
				case client.send <- combined:
				default:
				}
			}
			h.clients[client] = struct{}{}
			h.mu.Unlock()

		case client, ok := <-h.unregister:
			if !ok {
				return
			}
			h.mu.Lock()
			if _, exists := h.clients[client]; exists {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case msg, ok := <-h.broadcast:
			if !ok {
				return
			}
			h.mu.Lock()
			h.scrollback = append(h.scrollback, msg)
			h.scrollbackSize += len(msg)
			for h.scrollbackSize > maxScrollbackBytes && len(h.scrollback) > 0 {
				h.scrollbackSize -= len(h.scrollback[0])
				h.scrollback = h.scrollback[1:]
			}
			h.mu.Unlock()

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					go func(c *Client) {
						h.unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Stop() {
	close(h.register)
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) ClientIPs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ips := make([]string, 0, len(h.clients))
	for client := range h.clients {
		ips = append(ips, client.remoteAddr)
	}
	return ips
}

func (h *Hub) Broadcast(data []byte) {
	select {
	case h.broadcast <- data:
	default:
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}
