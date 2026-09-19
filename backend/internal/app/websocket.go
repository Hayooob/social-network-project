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

		// Same-origin requests (the nginx reverse proxy in docker-compose)
		// arrive with the Host as their Origin, so accept those too; anything
		// else has to be in ALLOWED_ORIGINS.
		if origin == "http://"+r.Host || origin == "https://"+r.Host {
			return true
		}

		if !isAllowedOrigin(origin) {
			log.Printf("WebSocket rejected origin: %q", origin)
			return false
		}
		return true
	},
}

type Client struct {
	UserID        int64
	UserName      string
	Conn          *websocket.Conn
	Send          chan []byte
	CurrentPage   string // Track which page the user is on
	ChatPartnerID int64  // If on messages page, track who they're chatting with
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
	log.Println("Hub started running")
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

	// Send directly to clients instead of using Broadcast channel to avoid deadlock
	h.mu.RLock()
	for _, client := range h.Clients {
		select {
		case client.Send <- data:
		default:
			// Skip if channel is full
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.Clients[userID]; ok {
		select {
		case client.Send <- message:
			log.Printf("Message sent to user %d", userID)
		default:
			close(client.Send)
			delete(h.Clients, userID)
		}
	} else {
		log.Printf("User %d is not online", userID)
	}
}

func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.Clients[userID]
	return ok
}

// IsUserViewingChatWith checks if a user is currently viewing a chat with a specific person
func (h *Hub) IsUserViewingChatWith(userID int64, chatPartnerID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.Clients[userID]; ok {
		return client.CurrentPage == "messages" && client.ChatPartnerID == chatPartnerID
	}
	return false
}

// UpdateClientPage updates the page status for a client
func (h *Hub) UpdateClientPage(userID int64, page string, chatPartnerID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, ok := h.Clients[userID]; ok {
		client.CurrentPage = page
		client.ChatPartnerID = chatPartnerID
		log.Printf("User %d page updated: %s (chatting with: %d)", userID, page, chatPartnerID)
	}
}

type WSMessage struct {
	Type          string `json:"type"`
	To            int64  `json:"to,omitempty"`
	Content       string `json:"content,omitempty"`
	MessageID     int    `json:"message_id,omitempty"`
	Page          string `json:"page,omitempty"`            // For page_status messages
	ChatPartnerID int64  `json:"chat_partner_id,omitempty"` // Who user is chatting with
}

func (s *Server) HandleWebSocket(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("WebSocket connection attempt...")

		user := CurrentUser(r.Context())
		if user == nil {
			log.Println("WebSocket: No user in context - unauthorized")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("WebSocket: User %s (ID: %d) attempting to connect", user.FullName, user.ID)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}

		log.Printf("WebSocket: Connection upgraded for user %s", user.FullName)

		client := &Client{
			UserID:        user.ID,
			UserName:      user.FullName,
			Conn:          conn,
			Send:          make(chan []byte, 256),
			CurrentPage:   "other", // Default to not on messages page
			ChatPartnerID: 0,
		}

		hub.Register <- client

		go client.writePump()
		go client.readPump(hub, s.DB)
	}
}

func (c *Client) readPump(hub *Hub, database *sql.DB) {
	defer func() {
		log.Printf("readPump ending for user %d", c.UserID)
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(120 * time.Second)) // Increased to 120 seconds
	c.Conn.SetPongHandler(func(string) error {
		log.Printf("Pong received from user %d", c.UserID)
		c.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})

	log.Printf("readPump started for user %d", c.UserID)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error for user %d: %v", c.UserID, err)
			} else {
				log.Printf("WebSocket closed for user %d: %v", c.UserID, err)
			}
			break
		}

		log.Printf("Message received from user %d: %s", c.UserID, string(message))

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		log.Printf("Parsed message - Type: %s, To: %d, Content: %s", wsMsg.Type, wsMsg.To, wsMsg.Content)

		switch wsMsg.Type {
		case "chat_message":
			log.Println("Handling chat_message...")
			handleChatMessage(c, hub, database, wsMsg)
		case "typing":
			handleTyping(c, hub, wsMsg)
		case "mark_read":
			handleMarkRead(c, database, wsMsg)
		case "page_status":
			handlePageStatus(c, hub, wsMsg)
		default:
			log.Printf("Unknown message type: %s", wsMsg.Type)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		log.Printf("writePump ended for user %d", c.UserID)
	}()

	log.Printf("writePump started for user %d", c.UserID)

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
			log.Printf("Sending ping to user %d", c.UserID)
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping failed for user %d: %v", c.UserID, err)
				return
			}
		}
	}
}

func handlePageStatus(client *Client, hub *Hub, wsMsg WSMessage) {
	log.Printf("handlePageStatus: user=%d, page=%s, chatPartner=%d", client.UserID, wsMsg.Page, wsMsg.ChatPartnerID)
	hub.UpdateClientPage(client.UserID, wsMsg.Page, wsMsg.ChatPartnerID)
}

func handleChatMessage(sender *Client, hub *Hub, database *sql.DB, wsMsg WSMessage) {
	log.Printf("handleChatMessage: from=%d, to=%d, content=%s", sender.UserID, wsMsg.To, wsMsg.Content)

	if wsMsg.To == 0 || wsMsg.Content == "" {
		log.Println("handleChatMessage: Missing 'to' or 'content'")
		return
	}

	// Most messages arrive over this socket rather than the REST endpoint, so
	// the same permission rule has to be applied here or the check is trivially
	// bypassed by talking to the WebSocket directly.
	allowed, err := db.CanMessage(database, int(sender.UserID), int(wsMsg.To))
	if err != nil {
		log.Println("handleChatMessage: CanMessage error:", err)
		return
	}
	if !allowed {
		log.Printf("handleChatMessage: user %d not allowed to message %d", sender.UserID, wsMsg.To)
		sendWSError(sender, "you cannot message this user")
		return
	}

	log.Println("Saving message to database...")
	msg, err := db.SendMessage(database, int(sender.UserID), int(wsMsg.To), wsMsg.Content)
	if err != nil {
		log.Println("Failed to save message:", err)
		return
	}
	log.Printf("Message saved with ID: %d", msg.ID)

	response := map[string]any{
		"type":         "chat_message",
		"message":      msg,
		"from_user_id": sender.UserID,
		"from_user":    sender.UserName,
	}

	data, _ := json.Marshal(response)
	log.Printf("Sending response: %s", string(data))

	// Send to sender
	sender.Send <- data
	log.Println("Sent to sender")

	// Send to receiver
	hub.SendToUser(wsMsg.To, data)

	// Create notification if:
	// 1. Receiver is offline, OR
	// 2. Receiver is online but NOT viewing chat with sender
	shouldNotify := !hub.IsUserOnline(wsMsg.To) || !hub.IsUserViewingChatWith(wsMsg.To, sender.UserID)

	if shouldNotify {
		log.Printf("Creating notification for user %d (offline or not viewing chat)", wsMsg.To)
		fromUserID := int(sender.UserID)
		db.CreateNotification(database, int(wsMsg.To), models.NotificationTypeNewMessage, &msg.ID, &fromUserID, "sent you a message")
	} else {
		log.Printf("User %d is viewing chat with sender, skipping notification", wsMsg.To)
	}

	log.Println("handleChatMessage completed")
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

// sendWSError delivers a non-fatal error back to a single client without
// dropping the connection.
func sendWSError(c *Client, message string) {
	payload, err := json.Marshal(map[string]any{
		"type":  "error",
		"error": message,
	})
	if err != nil {
		return
	}

	select {
	case c.Send <- payload:
	default:
		// Client's buffer is full; dropping the notice is preferable to
		// blocking the hub.
	}
}
