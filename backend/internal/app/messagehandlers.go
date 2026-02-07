package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/internal/db"
	"social-network/internal/models"
)

func (s *Server) handleGetConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversations, err := db.GetConversationList(s.DB, int(user.ID))
	if err != nil {
		log.Println("GetConversationList error:", err)
		writeError(w, http.StatusInternalServerError, "failed to get conversations")
		return
	}

	if conversations == nil {
		conversations = []models.Conversation{}
	}

	writeJSON(w, http.StatusOK, conversations)
}

func (s *Server) handleGetConversation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	path := r.URL.Path
	trimmed := strings.TrimPrefix(path, "/api/messages/")
	otherUserID, err := strconv.Atoi(trimmed)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	messages, err := db.GetConversation(s.DB, int(user.ID), otherUserID, limit, offset)
	if err != nil {
		log.Println("GetConversation error:", err)
		writeError(w, http.StatusInternalServerError, "failed to get messages")
		return
	}

	if messages == nil {
		messages = []models.Message{}
	}

	err = db.MarkMessagesAsRead(s.DB, int(user.ID), otherUserID)
	if err != nil {
		log.Println("MarkMessagesAsRead error:", err)
	}

	writeJSON(w, http.StatusOK, messages)
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	path := r.URL.Path
	trimmed := strings.TrimPrefix(path, "/api/messages/")
	receiverID, err := strconv.Atoi(trimmed)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content cannot be empty")
		return
	}

	msg, err := db.SendMessage(s.DB, int(user.ID), receiverID, req.Content)
	if err != nil {
		log.Println("SendMessage error:", err)
		writeError(w, http.StatusInternalServerError, "failed to send message")
		return
	}

	writeJSON(w, http.StatusCreated, msg)
}

func (s *Server) handleGetUnreadMessageCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	count, err := db.GetUnreadMessageCount(s.DB, int(user.ID))
	if err != nil {
		log.Println("GetUnreadMessageCount error:", err)
		writeError(w, http.StatusInternalServerError, "failed to get unread count")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (s *Server) handleMessageRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/messages" || path == "/api/messages/" {
		s.handleGetConversations(w, r)
		return
	}

	trimmed := strings.TrimPrefix(path, "/api/messages/")
	if trimmed == "unread-count" {
		s.handleGetUnreadMessageCount(w, r)
		return
	}

	if r.Method == http.MethodGet {
		s.handleGetConversation(w, r)
		return
	}

	if r.Method == http.MethodPost {
		s.handleSendMessage(w, r)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}