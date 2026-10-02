package hub

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func NewClient(conn *websocket.Conn) *Client { return &Client{conn: conn} }

func (c *Client) ReadMessage() error {
	_, _, err := c.conn.ReadMessage()
	return err
}

func (c *Client) Write(messageType int, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}
	return c.conn.WriteMessage(messageType, payload)
}

func (c *Client) Close() error { return c.conn.Close() }

type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[*Client]struct{}
}

func New() *Hub { return &Hub{clients: make(map[int64]map[*Client]struct{})} }

func (h *Hub) Add(userID int64, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*Client]struct{})
	}
	h.clients[userID][client] = struct{}{}
}

func (h *Hub) Remove(userID int64, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients[userID], client)
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
	_ = client.Close()
}

func (h *Hub) Send(userID int64, payload []byte) error {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients[userID]))
	for client := range h.clients[userID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()
	for _, client := range clients {
		if err := client.Write(websocket.TextMessage, payload); err != nil {
			h.Remove(userID, client)
		}
	}
	return nil
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, clients := range h.clients {
		for client := range clients {
			_ = client.Close()
		}
	}
	clear(h.clients)
}
