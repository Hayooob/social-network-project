package db

import (
	"database/sql"
	"time"
)

type Session struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}

// inserts a new session into db for a user.
func CreateSession(db *sql.DB, userID int64, token string, expiresAt time.Time) error {
	_, err := db.Exec(`
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES (?, ?, ?)
	`, token, userID, expiresAt)
	return err
}

// fetches a session by its token.
func GetSession(db *sql.DB, token string) (*Session, error) {
	row := db.QueryRow(`
		SELECT token, user_id, expires_at, created_at
		FROM sessions
		WHERE token = ?
	`, token)

	var s Session
	if err := row.Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			// no session found
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

// remove a session by token.
func DeleteSession(db *sql.DB, token string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

// clean up expired sessions.
func DeleteExpiredSessions(db *sql.DB, now time.Time) (int64, error) {
	res, err := db.Exec(`
		DELETE FROM sessions
		WHERE expires_at <= ?
	`, now)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
