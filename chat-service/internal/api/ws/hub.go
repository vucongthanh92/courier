package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 45 * time.Second
	sendBufferSize = 32
)

type Hub struct {
	subscriber interfaces.RealtimeSubscriberI
	mu         sync.RWMutex
	clients    map[uint64]map[*Client]struct{}
}

func NewHub(subscriber interfaces.RealtimeSubscriberI) *Hub {
	return &Hub{
		subscriber: subscriber,
		clients:    make(map[uint64]map[*Client]struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	if h == nil || h.subscriber == nil {
		return
	}
	events, errs := h.subscriber.SubscribeMessageCreated(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-errs:
			if ok && err != nil {
				logger.Error("realtime subscription error", zap.Error(err))
			}
		case event, ok := <-events:
			if !ok {
				return
			}
			h.broadcast(event)
		}
	}
}

func (h *Hub) Register(userID uint64, conn *websocket.Conn) {
	client := &Client{
		hub:    h,
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, sendBufferSize),
	}

	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*Client]struct{})
	}
	h.clients[userID][client] = struct{}{}
	h.mu.Unlock()

	go client.writePump()
	client.readPump()
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userClients := h.clients[client.userID]
	if userClients == nil {
		return
	}
	if _, ok := userClients[client]; ok {
		delete(userClients, client)
		close(client.send)
	}
	if len(userClients) == 0 {
		delete(h.clients, client.userID)
	}
}

func (h *Hub) broadcast(event models.MessageCreatedEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		logger.Error("marshal realtime event failed", zap.Error(err))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userID := range event.RecipientUserIDs {
		for client := range h.clients[userID] {
			select {
			case client.send <- payload:
			default:
				go h.unregister(client)
			}
		}
	}
}

type Client struct {
	hub    *Hub
	userID uint64
	conn   *websocket.Conn
	send   chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := c.conn.NextReader(); err != nil {
			return
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func Upgrader(allowOrigins []string) websocket.Upgrader {
	allowed := make(map[string]struct{}, len(allowOrigins))
	for _, origin := range allowOrigins {
		allowed[origin] = struct{}{}
	}
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if len(allowed) == 0 {
				return true
			}
			origin := r.Header.Get("Origin")
			_, ok := allowed[origin]
			return ok
		},
	}
}
