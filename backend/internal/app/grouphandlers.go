package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"social-network/internal/db"
)

type createGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

type inviteUserRequest struct {
	UserID int `json:"user_id"`
}

type createGroupPostRequest struct {
	Content string `json:"content"`
}

type createEventInGroupRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	EventDate   string `json:"event_date"` // RFC3339 or "YYYY-MM-DDTHH:mm"
}

type respondEventRequest struct {
	Response string `json:"response"`
}

// /api/groups  (GET visible groups, POST create group)
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
		groups, err := db.GetVisibleGroups(s.DB, uid)
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

// /api/groups/invitations (GET)
func (s *Server) handleGroupInvitations(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	uid := int(user.ID)

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	invites, err := db.GetMyPendingInvitations(s.DB, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not get invitations")
		return
	}
	writeJSON(w, http.StatusOK, invites)
}

// /api/groups/invitations/{id}/accept|decline
func (s *Server) handleGroupInvitationRoutes(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	uid := int(user.ID)

	rest := strings.TrimPrefix(r.URL.Path, "/api/groups/invitations/")
	rest = strings.Trim(rest, "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	inviteID, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invitation id")
		return
	}
	action := parts[1]

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	inv, err := db.GetInvitationByID(s.DB, inviteID)
	if err != nil {
		writeError(w, http.StatusNotFound, "invitation not found")
		return
	}
	if inv.InviteeID != uid {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	switch action {
	case "accept":
		// Add as accepted member
		_, err := s.DB.Exec(`
			INSERT OR IGNORE INTO group_members (group_id, user_id, role, status, joined_at)
			VALUES (?, ?, 'member', 'accepted', CURRENT_TIMESTAMP)
		`, inv.GroupID, uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not accept invitation")
			return
		}
		_ = db.DeleteInvitation(s.DB, inviteID)

		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return

	case "decline":
		_ = db.DeleteInvitation(s.DB, inviteID)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return

	default:
		writeError(w, http.StatusNotFound, "not found")
		return
	}
}

// /api/groups/{id}/...
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
		writeJSON(w, http.StatusOK, grp)
		return
	}

	sub := parts[1]

	switch sub {
	case "join":
		// request join (or auto-accept if public)
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		// If invited, UI should use accept/decline; but if they hit join anyway, we keep it safe:
		// join will behave as request based on privacy.
		var isPrivate int
		if err := s.DB.QueryRow(`SELECT is_private, creator_id FROM groups WHERE id = ?`, groupID).Scan(&isPrivate, new(int)); err != nil {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}

		status := "accepted"
		if isPrivate == 1 {
			status = "pending"
		}

		_, err := s.DB.Exec(`
			INSERT OR IGNORE INTO group_members (group_id, user_id, role, status, joined_at)
			VALUES (?, ?, 'member', ?, CURRENT_TIMESTAMP)
		`, groupID, uid, status)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not request join")
			return
		}

		// Notify creator on join request (private only)
		if isPrivate == 1 {
			var creatorID int
			_ = s.DB.QueryRow(`SELECT creator_id FROM groups WHERE id = ?`, groupID).Scan(&creatorID)
			ref := groupID
			from := uid
			_, _ = db.CreateNotification(s.DB, creatorID, "group_join_request", &ref, &from, "New group join request")
		}

		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": status})
		return

	case "invite":
		// invite a user (any accepted member can invite)
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		isMember, err := db.IsGroupMember(s.DB, groupID, uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not check membership")
			return
		}
		if !isMember {
			writeError(w, http.StatusForbidden, "not a group member")
			return
		}

		var req inviteUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID <= 0 {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if req.UserID == uid {
			writeError(w, http.StatusBadRequest, "cannot invite yourself")
			return
		}

		inviteID, err := db.CreateGroupInvitation(s.DB, groupID, uid, req.UserID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Notify invitee
		ref := groupID
		from := uid
		_, _ = db.CreateNotification(s.DB, req.UserID, "group_invite", &ref, &from, "You have been invited to a group")

		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "invitation_id": inviteID})
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
		// /api/groups/{id}/members (GET) OR admin actions (accept/remove)
		if len(parts) == 2 {
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
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

		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		switch action {
		case "accept":
			if err := db.UpdateMemberStatus(s.DB, groupID, targetID, "accepted"); err != nil {
				writeError(w, http.StatusInternalServerError, "could not accept member")
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return

		case "remove":
			// works as kick OR decline request
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
		// keep your group posts logic as-is (member only)
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

	case "events":
		// /api/groups/{id}/events (GET, POST)
		// /api/groups/{id}/events/{eventId}/respond (POST)
		// /api/groups/{id}/events/{eventId}/responses (GET)

		isMember, err := db.IsGroupMember(s.DB, groupID, uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not check membership")
			return
		}
		if !isMember {
			writeError(w, http.StatusForbidden, "not a group member")
			return
		}

		// /events
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				events, err := db.GetGroupEvents(s.DB, groupID, uid, 50)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "could not get events")
					return
				}
				writeJSON(w, http.StatusOK, events)
				return

			case http.MethodPost:
				var req createEventInGroupRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeError(w, http.StatusBadRequest, "invalid JSON body")
					return
				}
				if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Description) == "" || strings.TrimSpace(req.Location) == "" || strings.TrimSpace(req.EventDate) == "" {
					writeError(w, http.StatusBadRequest, "missing required fields")
					return
				}

				eventDate, err := time.Parse(time.RFC3339, req.EventDate)
				if err != nil {
					eventDate, err = time.Parse("2006-01-02T15:04", req.EventDate)
					if err != nil {
						writeError(w, http.StatusBadRequest, "invalid event_date")
						return
					}
				}

				gid := groupID
				ev, err := db.CreateEvent(s.DB, uid, &gid, req.Title, req.Description, req.Location, eventDate)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "could not create event")
					return
				}

				// Notify all accepted members in the group (except creator)
				memberIDs, _ := db.GetAcceptedGroupMemberIDs(s.DB, groupID)
				ref := groupID
				from := uid
				for _, mid := range memberIDs {
					if mid == uid {
						continue
					}
					_, _ = db.CreateNotification(s.DB, mid, "group_event_created", &ref, &from, "New event created in your group")
				}

				writeJSON(w, http.StatusCreated, ev)
				return

			default:
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
		}

		// /events/{eventId}/respond or /responses
		if len(parts) != 4 {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		eventID, err := strconv.Atoi(parts[2])
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid event id")
			return
		}
		action := parts[3]

		switch action {
		case "respond":
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			var req respondEventRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
			switch req.Response {
			case "going", "maybe", "not_going":
			default:
				writeError(w, http.StatusBadRequest, "invalid response")
				return
			}
			if err := db.RespondToEvent(s.DB, eventID, uid, req.Response); err != nil {
				writeError(w, http.StatusInternalServerError, "could not set response")
				return
			}
			ev, _ := db.GetEventByID(s.DB, eventID, uid)
			writeJSON(w, http.StatusOK, ev)
			return

		case "responses":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			resps, err := db.GetEventResponses(s.DB, eventID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "could not get responses")
				return
			}
			writeJSON(w, http.StatusOK, resps)
			return

		default:
			writeError(w, http.StatusNotFound, "not found")
			return
		}

	default:
		writeError(w, http.StatusNotFound, "not found")
		return
	}
}
