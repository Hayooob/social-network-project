package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"social-network/internal/db"
)

type createEventRequest struct {
	GroupID     *int   `json:"group_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	EventDate   string `json:"event_date"` // ISO string
}

//type respondEventRequest struct {
//Response string `json:"response"`
//}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	uid := int(user.ID)

	switch r.Method {
	case http.MethodPost:
		var req createEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if strings.TrimSpace(req.Title) == "" ||
			strings.TrimSpace(req.Description) == "" ||
			strings.TrimSpace(req.Location) == "" ||
			strings.TrimSpace(req.EventDate) == "" {
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

		if req.GroupID != nil {
			isMember, err := db.IsGroupMember(s.DB, *req.GroupID, uid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "could not check membership")
				return
			}
			if !isMember {
				writeError(w, http.StatusForbidden, "not a group member")
				return
			}
		}

		ev, err := db.CreateEvent(s.DB, uid, req.GroupID, req.Title, req.Description, req.Location, eventDate)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create event")
			return
		}
		writeJSON(w, http.StatusCreated, ev)
		return

	case http.MethodGet:
		events, err := db.GetUpcomingEvents(s.DB, uid, 50)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not get events")
			return
		}
		writeJSON(w, http.StatusOK, events)
		return

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
}

// handleEventRoutes handles /api/events/{id}/...
func (s *Server) handleEventRoutes(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	uid := int(user.ID)

	path := strings.TrimPrefix(r.URL.Path, "/api/events/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	eventID, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		ev, err := db.GetEventByID(s.DB, eventID, uid)
		if err != nil {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}
		if ev.GroupID != nil {
			isMember, err := db.IsGroupMember(s.DB, *ev.GroupID, uid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "could not check membership")
				return
			}
			if !isMember {
				writeError(w, http.StatusForbidden, "not a group member")
				return
			}
		}
		writeJSON(w, http.StatusOK, ev)
		return
	}

	sub := parts[1]
	switch sub {
	case "respond":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		ev, err := db.GetEventByID(s.DB, eventID, uid)
		if err != nil {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}
		if ev.GroupID != nil {
			isMember, err := db.IsGroupMember(s.DB, *ev.GroupID, uid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "could not check membership")
				return
			}
			if !isMember {
				writeError(w, http.StatusForbidden, "not a group member")
				return
			}
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

		updated, _ := db.GetEventByID(s.DB, eventID, uid)
		writeJSON(w, http.StatusOK, updated)
		return

	case "responses":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		ev, err := db.GetEventByID(s.DB, eventID, uid)
		if err != nil {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}
		if ev.GroupID != nil {
			isMember, err := db.IsGroupMember(s.DB, *ev.GroupID, uid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "could not check membership")
				return
			}
			if !isMember {
				writeError(w, http.StatusForbidden, "not a group member")
				return
			}
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
}
