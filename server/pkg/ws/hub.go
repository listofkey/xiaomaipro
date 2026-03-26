package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const writeWait = 5 * time.Second

type Hub struct {
	upgrader websocket.Upgrader

	mu      sync.RWMutex
	clients map[*client]struct{}
}

type client struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	closed atomic.Bool
}

func NewHub() *Hub {
	return &Hub{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[*client]struct{}),
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &client{conn: conn}
	h.addClient(client)
	defer h.removeClient(client)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Hub) BroadcastJSON(payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return h.broadcast(websocket.TextMessage, data)
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}

func (h *Hub) Close() error {
	clients := h.snapshotClients()
	var errs []error
	for _, client := range clients {
		if err := client.close(); err != nil {
			errs = append(errs, err)
		}
	}

	h.mu.Lock()
	h.clients = make(map[*client]struct{})
	h.mu.Unlock()

	return errors.Join(errs...)
}

func (h *Hub) broadcast(messageType int, payload []byte) error {
	clients := h.snapshotClients()
	var errs []error

	for _, client := range clients {
		if err := client.write(messageType, payload); err != nil {
			errs = append(errs, err)
			h.removeClient(client)
		}
	}

	return errors.Join(errs...)
}

func (h *Hub) addClient(client *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = struct{}{}
}

func (h *Hub) removeClient(client *client) {
	h.mu.Lock()
	delete(h.clients, client)
	h.mu.Unlock()

	_ = client.close()
}

func (h *Hub) snapshotClients() []*client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}

	return clients
}

func (c *client) write(messageType int, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed.Load() {
		return netClosedError{}
	}

	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}

	return c.conn.WriteMessage(messageType, payload)
}

func (c *client) close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}

	return c.conn.Close()
}

type netClosedError struct{}

func (netClosedError) Error() string {
	return "websocket connection is closed"
}
