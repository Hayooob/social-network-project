package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"social-network/internal/db"
)

type createGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

type createGroupPostRequest struct {
	Content string `json:"content"`
}

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	uid := int(user.ID)

	switch r.Method {
	case http.MethodPost:
		var req createGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Description) == "" {
			writeError(w, http.StatusBadRequest, "name and description are required")
			return
		}
		grp, err := db.CreateGroup(s.DB, uid, req.Name, req.Description, req.IsPrivate)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create group")
			return
		}
		writeJSON(w, http.StatusCreated, grp)
		return

	case http.MethodGet:
		groups, err := db.GetUserGroups(s.DB, uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not get groups")
			return
		}
		writeJSON(w, http.StatusOK, groups)
		return

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
}

// handleGroupRoutes handles /api/groups/{id}/...
func (s *Server) handleGroupRoutes(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	uid := int(user.ID)

	path := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	groupID, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	// /api/groups/{id}
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		grp, err := db.GetGroupByID(s.DB, groupID, uid)
		if err != nil {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}
		if grp.IsPrivate == 1 && grp.MyStatus != "accepted" && grp.CreatorID != uid {
			writeError(w, http.StatusForbidden, "private group")
			return
		}
		writeJSON(w, http.StatusOK, grp)
		return
	}

	sub := parts[1]

	switch sub {
	case "join":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		member, err := db.AddGroupMember(s.DB, groupID, uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not join group")
			return
		}
		writeJSON(w, http.StatusOK, member)
		return

	case "leave":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if err := db.RemoveGroupMember(s.DB, groupID, uid); err != nil {
			writeError(w, http.StatusInternalServerError, "could not leave group")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return

	case "members":
		// /api/groups/{id}/members (GET)
		if len(parts) == 2 {
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}

			grp, err := db.GetGroupByID(s.DB, groupID, uid)
			if err != nil {
				writeError(w, http.StatusNotFound, "group not found")
				return
			}
			if grp.IsPrivate == 1 && grp.MyStatus != "accepted" && grp.CreatorID != uid {
				writeError(w, http.StatusForbidden, "private group")
				return
			}

			members, err := db.GetGroupMembers(s.DB, groupID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "could not get members")
				return
			}
			writeJSON(w, http.StatusOK, members)
			return
		}

		// /api/groups/{id}/members/{userId}/{action}
		if len(parts) != 4 {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		targetID, err := strconv.Atoi(parts[2])
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		action := parts[3]

		isAdmin, err := db.IsGroupAdmin(s.DB, groupID, uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not check admin")
			return
		}
		if !isAdmin {
			writeError(w, http.StatusForbidden, "admin only")
			return
		}

		switch action {
		case "accept":
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			if err := db.UpdateMemberStatus(s.DB, groupID, targetID, "accepted"); err != nil {
				writeError(w, http.StatusInternalServerError, "could not accept member")
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return

		case "remove":
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			if err := db.RemoveGroupMember(s.DB, groupID, targetID); err != nil {
				writeError(w, http.StatusInternalServerError, "could not remove member")
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return

		default:
			writeError(w, http.StatusNotFound, "not found")
			return
		}

	case "posts":
		isMember, err := db.IsGroupMember(s.DB, groupID, uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not check membership")
			return
		}
		if !isMember {
			writeError(w, http.StatusForbidden, "not a group member")
			return
		}

		switch r.Method {
		case http.MethodGet:
			posts, err := db.GetGroupPosts(s.DB, groupID, 50)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "could not get group posts")
				return
			}
			writeJSON(w, http.StatusOK, posts)
			return

		case http.MethodPost:
			var req createGroupPostRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
			if strings.TrimSpace(req.Content) == "" {
				writeError(w, http.StatusBadRequest, "content is required")
				return
			}
			post, err := db.CreateGroupPost(s.DB, groupID, uid, req.Content)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "could not create post")
				return
			}
			writeJSON(w, http.StatusCreated, post)
			return

		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

	default:
		writeError(w, http.StatusNotFound, "not found")
		return
	}
}
