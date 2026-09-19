package db

import (
	"database/sql"
	"time"

	"social-network/internal/models"
)

func CreateEvent(db *sql.DB, creatorID int, groupID *int, title, description, location string, eventDate time.Time) (*models.Event, error) {
	res, err := db.Exec(`
		INSERT INTO events (group_id, creator_id, title, description, location, event_date)
		VALUES (?, ?, ?, ?, ?, ?)
	`, groupID, creatorID, title, description, location, eventDate)
	if err != nil {
		return nil, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetEventByID(db, int(id64), creatorID)
}

// GetEventByID returns event details plus response counts and viewer response.
func GetEventByID(db *sql.DB, eventID int, viewerID int) (*models.Event, error) {
	query := `
		SELECT
			e.id, e.group_id, e.creator_id, e.title, e.description, e.location, e.event_date, e.created_at,
			u.full_name AS creator_name,
			COALESCE(g.name, '') AS group_name,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'going') AS going_count,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'maybe') AS maybe_count,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'not_going') AS not_going_count,
			COALESCE((SELECT r2.response FROM event_responses r2 WHERE r2.event_id = e.id AND r2.user_id = ?), '') AS my_response
		FROM events e
		JOIN users u ON u.id = e.creator_id
		LEFT JOIN groups g ON g.id = e.group_id
		WHERE e.id = ?
	`

	var ev models.Event
	var groupName string
	var myResp string
	if err := db.QueryRow(query, viewerID, eventID).Scan(
		&ev.ID,
		&ev.GroupID,
		&ev.CreatorID,
		&ev.Title,
		&ev.Description,
		&ev.Location,
		&ev.EventDate,
		&ev.CreatedAt,
		&ev.CreatorName,
		&groupName,
		&ev.GoingCount,
		&ev.MaybeCount,
		&ev.NotGoingCount,
		&myResp,
	); err != nil {
		return nil, err
	}
	if groupName != "" {
		ev.GroupName = groupName
	}
	if myResp != "" {
		ev.MyResponse = myResp
	}
	return &ev, nil
}

// GetUpcomingEvents lists upcoming events for a user:
// - events with NULL group_id (global)
// - events in groups where the user is an accepted member
func GetUpcomingEvents(db *sql.DB, userID int, limit int) ([]models.Event, error) {
	query := `
		SELECT
			e.id, e.group_id, e.creator_id, e.title, e.description, e.location, e.event_date, e.created_at,
			u.full_name AS creator_name,
			COALESCE(g.name, '') AS group_name,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'going') AS going_count,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'maybe') AS maybe_count,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'not_going') AS not_going_count,
			COALESCE((SELECT r2.response FROM event_responses r2 WHERE r2.event_id = e.id AND r2.user_id = ?), '') AS my_response
		FROM events e
		JOIN users u ON u.id = e.creator_id
		LEFT JOIN groups g ON g.id = e.group_id
		WHERE e.event_date >= CURRENT_TIMESTAMP
			AND (
				e.group_id IS NULL
				OR EXISTS (
					SELECT 1 FROM group_members gm
					WHERE gm.group_id = e.group_id AND gm.user_id = ? AND gm.status = 'accepted'
				)
			)
		ORDER BY e.event_date ASC
		LIMIT ?
	`

	rows, err := db.Query(query, userID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var ev models.Event
		var groupName string
		var myResp string
		if err := rows.Scan(
			&ev.ID,
			&ev.GroupID,
			&ev.CreatorID,
			&ev.Title,
			&ev.Description,
			&ev.Location,
			&ev.EventDate,
			&ev.CreatedAt,
			&ev.CreatorName,
			&groupName,
			&ev.GoingCount,
			&ev.MaybeCount,
			&ev.NotGoingCount,
			&myResp,
		); err != nil {
			return nil, err
		}
		if groupName != "" {
			ev.GroupName = groupName
		}
		if myResp != "" {
			ev.MyResponse = myResp
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

// RespondToEvent sets the user's response (going/maybe/not_going).
func RespondToEvent(db *sql.DB, eventID int, userID int, response string) error {
	_, err := db.Exec(`
		INSERT INTO event_responses (event_id, user_id, response)
		VALUES (?, ?, ?)
		ON CONFLICT(event_id, user_id) DO UPDATE SET response = excluded.response, created_at = CURRENT_TIMESTAMP
	`, eventID, userID, response)
	return err
}

func GetEventResponses(db *sql.DB, eventID int) ([]models.EventResponse, error) {
	query := `
		SELECT r.id, r.event_id, r.user_id, r.response, r.created_at,
			u.full_name AS user_name
		FROM event_responses r
		JOIN users u ON u.id = r.user_id
		WHERE r.event_id = ?
		ORDER BY r.created_at DESC
	`
	rows, err := db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resps []models.EventResponse
	for rows.Next() {
		var r models.EventResponse
		if err := rows.Scan(&r.ID, &r.EventID, &r.UserID, &r.Response, &r.CreatedAt, &r.UserName); err != nil {
			return nil, err
		}
		resps = append(resps, r)
	}
	return resps, rows.Err()
}

func GetUserEvents(db *sql.DB, userID int, limit int) ([]models.Event, error) {
	query := `
		SELECT
			e.id, e.group_id, e.creator_id, e.title, e.description, e.location, e.event_date, e.created_at,
			u.full_name AS creator_name,
			COALESCE(g.name, '') AS group_name,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'going') AS going_count,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'maybe') AS maybe_count,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'not_going') AS not_going_count,
			COALESCE((SELECT r2.response FROM event_responses r2 WHERE r2.event_id = e.id AND r2.user_id = ?), '') AS my_response
		FROM event_responses ur
		JOIN events e ON e.id = ur.event_id
		JOIN users u ON u.id = e.creator_id
		LEFT JOIN groups g ON g.id = e.group_id
		WHERE ur.user_id = ?
		ORDER BY e.event_date DESC
		LIMIT ?
	`

	rows, err := db.Query(query, userID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var ev models.Event
		var groupName string
		var myResp string
		if err := rows.Scan(
			&ev.ID,
			&ev.GroupID,
			&ev.CreatorID,
			&ev.Title,
			&ev.Description,
			&ev.Location,
			&ev.EventDate,
			&ev.CreatedAt,
			&ev.CreatorName,
			&groupName,
			&ev.GoingCount,
			&ev.MaybeCount,
			&ev.NotGoingCount,
			&myResp,
		); err != nil {
			return nil, err
		}
		if groupName != "" {
			ev.GroupName = groupName
		}
		if myResp != "" {
			ev.MyResponse = myResp
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

// GetGroupEvents lists events for a group (member only).
func GetGroupEvents(dbConn *sql.DB, groupID int, viewerID int, limit int) ([]models.Event, error) {
	query := `
		SELECT
			e.id, e.group_id, e.creator_id, e.title, e.description, e.location, e.event_date, e.created_at,
			u.full_name AS creator_name,
			COALESCE(g.name, '') AS group_name,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'going') AS going_count,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'maybe') AS maybe_count,
			(SELECT COUNT(1) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'not_going') AS not_going_count,
			COALESCE((SELECT r2.response FROM event_responses r2 WHERE r2.event_id = e.id AND r2.user_id = ?), '') AS my_response
		FROM events e
		JOIN users u ON u.id = e.creator_id
		LEFT JOIN groups g ON g.id = e.group_id
		WHERE e.group_id = ?
		ORDER BY e.event_date ASC
		LIMIT ?
	`

	rows, err := dbConn.Query(query, viewerID, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var ev models.Event
		var groupName string
		var myResp string
		if err := rows.Scan(
			&ev.ID,
			&ev.GroupID,
			&ev.CreatorID,
			&ev.Title,
			&ev.Description,
			&ev.Location,
			&ev.EventDate,
			&ev.CreatedAt,
			&ev.CreatorName,
			&groupName,
			&ev.GoingCount,
			&ev.MaybeCount,
			&ev.NotGoingCount,
			&myResp,
		); err != nil {
			return nil, err
		}
		if groupName != "" {
			ev.GroupName = groupName
		}
		if myResp != "" {
			ev.MyResponse = myResp
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}
