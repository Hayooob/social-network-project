package app

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"social-network/internal/db"
	"social-network/internal/models"
)

// handleGetGroupMessages GET /api/groups/{id}/messages
func (s *Server) handleGetGroupMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract group ID from path: /api/groups/{id}/messages
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/api/groups/")
	path = strings.TrimSuffix(path, "/messages")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	// Check if user is a member of the group
	isMember, err := db.IsGroupMember(s.DB, groupID, int(user.ID))
	if err != nil {
		log.Println("IsGroupMember error:", err)
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you are not a member of this group")
		return
	}

	// Get pagination params
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

	messages, err := db.GetGroupMessages(s.DB, groupID, limit, offset)
	if err != nil {
		log.Println("GetGroupMessages error:", err)
		writeError(w, http.StatusInternalServerError, "failed to get messages")
		return
	}

	if messages == nil {
		messages = []models.GroupMessage{}
	}

	writeJSON(w, http.StatusOK, messages)
}

// handleSendGroupMessage POST /api/groups/{id}/messages
func (s *Server) handleSendGroupMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract group ID from path
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/api/groups/")
	path = strings.TrimSuffix(path, "/messages")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	// Check membership
	isMember, err := db.IsGroupMember(s.DB, groupID, int(user.ID))
	if err != nil {
		log.Println("IsGroupMember error:", err)
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you are not a member of this group")
		return
	}

	// Parse request body
	var content string
	var imagePath string

	ct := r.Header.Get("Content-Type")

	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
			writeError(w, http.StatusBadRequest, "invalid multipart form")
			return
		}

		content = strings.TrimSpace(r.FormValue("content"))

		// image file: field name MUST be "image"
		file, header, err := r.FormFile("image")
		if err == nil && file != nil && header != nil {
			defer file.Close()

			_ = os.MkdirAll("uploads", 0755)

			ext := filepath.Ext(header.Filename)
			if ext == "" {
				ext = ".png"
			}

			filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
			dstPath := filepath.Join("uploads", filename)

			dst, err := os.Create(dstPath)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to save image")
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to save image")
				return
			}

			imagePath = "/uploads/" + filename
		}
	} else {
		var req struct {
			Content   string `json:"content"`
			ImagePath string `json:"image_path,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		content = strings.TrimSpace(req.Content)
		imagePath = req.ImagePath
	}

	if content == "" {
		writeError(w, http.StatusBadRequest, "content cannot be empty")
		return
	}

	// Save message
	msg, err := db.SendGroupMessage(s.DB, groupID, int(user.ID), content, imagePath)
	if err != nil {
		log.Println("SendGroupMessage error:", err)
		writeError(w, http.StatusInternalServerError, "failed to send message")
		return
	}

	// Broadcast to all group members via WebSocket
	if s.Hub != nil {
		memberIDs, err := db.GetGroupMemberIDs(s.DB, groupID)
		if err == nil {
			response := map[string]any{
				"type":      "group_message",
				"group_id":  groupID,
				"message":   msg,
				"from_user": user.FullName,
			}
			data, _ := json.Marshal(response)

			for _, memberID := range memberIDs {
				if memberID != user.ID { // Don't send to sender
					s.Hub.SendToUser(memberID, data)
				}
			}
		}
	}

	writeJSON(w, http.StatusCreated, msg)
}

// handleGroupMessageRoutes routes group message requests
func (s *Server) handleGroupMessageRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetGroupMessages(w, r)
	case http.MethodPost:
		s.handleSendGroupMessage(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
