package app

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/internal/db"
	"social-network/internal/models"
)

func (s *Server) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
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

	notifications, err := db.GetNotifications(s.DB, int(user.ID), limit, offset)
	if err != nil {
		log.Println("GetNotifications error:", err)
		writeError(w, http.StatusInternalServerError, "failed to get notifications")
		return
	}

	if notifications == nil {
		notifications = []models.Notification{}
	}

	writeJSON(w, http.StatusOK, notifications)
}

func (s *Server) handleMarkNotificationRead(w http.ResponseWriter, r *http.Request) {
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
	trimmed := strings.TrimPrefix(path, "/api/notifications/")
	trimmed = strings.TrimSuffix(trimmed, "/read")

	notifID, err := strconv.Atoi(trimmed)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification ID")
		return
	}

	err = db.MarkNotificationAsRead(s.DB, notifID, int(user.ID))
	if err != nil {
		log.Println("MarkNotificationAsRead error:", err)
		writeError(w, http.StatusInternalServerError, "failed to mark notification as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "notification marked as read"})
}

func (s *Server) handleMarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err := db.MarkAllNotificationsAsRead(s.DB, int(user.ID))
	if err != nil {
		log.Println("MarkAllNotificationsAsRead error:", err)
		writeError(w, http.StatusInternalServerError, "failed to mark notifications as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "all notifications marked as read"})
}

func (s *Server) handleGetUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	count, err := db.GetUnreadNotificationCount(s.DB, int(user.ID))
	if err != nil {
		log.Println("GetUnreadNotificationCount error:", err)
		writeError(w, http.StatusInternalServerError, "failed to get unread count")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (s *Server) handleNotificationRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/notifications" || path == "/api/notifications/" {
		s.handleGetNotifications(w, r)
		return
	}

	if path == "/api/notifications/read-all" {
		s.handleMarkAllNotificationsRead(w, r)
		return
	}

	if path == "/api/notifications/unread-count" {
		s.handleGetUnreadNotificationCount(w, r)
		return
	}

	if strings.HasSuffix(path, "/read") {
		s.handleMarkNotificationRead(w, r)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}
