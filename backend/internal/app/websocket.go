package app

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"social-network/internal/db"
	"social-network/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:5173"
	},
}

type Client struct {
	UserID   int64
	UserName string
	Conn     *websocket.Conn
	Send     chan []byte
}

type Hub struct {
	Clients    map[int64]*Client
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
	DB         *sql.DB
}

func NewHub(database *sql.DB) *Hub {
	return &Hub{
		Clients:    make(map[int64]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		DB:         database,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("Client connected: %s (ID: %d)", client.UserName, client.UserID)
			h.broadcastOnlineUsers()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: %s (ID: %d)", client.UserName, client.UserID)
			h.broadcastOnlineUsers()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) broadcastOnlineUsers() {
	h.mu.RLock()
	var users []map[string]any
	for _, client := range h.Clients {
		users = append(users, map[string]any{
			"id":       client.UserID,
			"username": client.UserName,
		})
	}
	h.mu.RUnlock()

	msg := map[string]any{
		"type":  "online_users",
		"users": users,
	}

	data, _ := json.Marshal(msg)
	h.Broadcast <- data
}

func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.Clients[userID]; ok {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.Clients, userID)
		}
	}
}

func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.Clients[userID]
	return ok
}

type WSMessage struct {
	Type       string `json:"type"`
	To         int64  `json:"to,omitempty"`
	Content    string `json:"content,omitempty"`
	MessageID  int    `json:"message_id,omitempty"`
}

func (s *Server) HandleWebSocket(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := CurrentUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}

		client := &Client{
			UserID:   user.ID,
			UserName: user.FullName,
			Conn:     conn,
			Send:     make(chan []byte, 256),
		}

		hub.Register <- client

		go client.writePump()
		go client.readPump(hub, s.DB)
	}
}

func (c *Client) readPump(hub *Hub, database *sql.DB) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		switch wsMsg.Type {
		case "chat_message":
			handleChatMessage(c, hub, database, wsMsg)
		case "typing":
			handleTyping(c, hub, wsMsg)
		case "mark_read":
			handleMarkRead(c, database, wsMsg)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func handleChatMessage(sender *Client, hub *Hub, database *sql.DB, wsMsg WSMessage) {
	if wsMsg.To == 0 || wsMsg.Content == "" {
		return
	}

	msg, err := db.SendMessage(database, int(sender.UserID), int(wsMsg.To), wsMsg.Content)
	if err != nil {
		log.Println("Failed to save message:", err)
		return
	}

	response := map[string]any{
		"type":         "chat_message",
		"message":      msg,
		"from_user_id": sender.UserID,
		"from_user":    sender.UserName,
	}

	data, _ := json.Marshal(response)

	sender.Send <- data

	hub.SendToUser(wsMsg.To, data)

	if !hub.IsUserOnline(wsMsg.To) {
		fromUserID := int(sender.UserID)
		db.CreateNotification(database, int(wsMsg.To), models.NotificationTypeNewMessage, &msg.ID, &fromUserID, "sent you a message")
	}
}

func handleTyping(sender *Client, hub *Hub, wsMsg WSMessage) {
	if wsMsg.To == 0 {
		return
	}

	response := map[string]any{
		"type":         "typing",
		"from_user_id": sender.UserID,
		"from_user":    sender.UserName,
	}

	data, _ := json.Marshal(response)
	hub.SendToUser(wsMsg.To, data)
}

func handleMarkRead(client *Client, database *sql.DB, wsMsg WSMessage) {
	if wsMsg.To == 0 {
		return
	}

	err := db.MarkMessagesAsRead(database, int(client.UserID), int(wsMsg.To))
	if err != nil {
		log.Println("Failed to mark messages as read:", err)
	}
}