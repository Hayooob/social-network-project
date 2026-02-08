package db

import (
	"database/sql"
	"errors"

	"social-network/internal/models"
)

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// CreateUser inserts a new user into the users table.
func CreateUser(db *sql.DB, u *models.User) error {
	stmt := `
        INSERT INTO users (
            uuid,
            email,
            password_hash,
            full_name,
            date_of_birth,
            avatar_url,
            nickname,
            about_me,
            is_private
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := db.Exec(stmt,
		u.UUID,
		u.Email,
		u.PasswordHash,
		u.FullName,
		u.DateOfBirth,
		u.AvatarURL,
		u.Nickname,
		u.AboutMe,
		boolToInt(u.IsPrivate),
	)
	return err
}

// GetUserByEmail returns a user by email, or (nil, nil) if not found.
func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	stmt := `
        SELECT
            id,
            uuid,
            email,
            password_hash,
            full_name,
            date_of_birth,
            avatar_url,
            nickname,
            about_me,
            is_private,
            created_at
        FROM users
        WHERE email = ?
    `

	row := db.QueryRow(stmt, email)

	var u models.User
	var isPrivateInt int

	err := row.Scan(
		&u.ID,
		&u.UUID,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.DateOfBirth,
		&u.AvatarURL,
		&u.Nickname,
		&u.AboutMe,
		&isPrivateInt,
		&u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	u.IsPrivate = isPrivateInt == 1
	return &u, nil
}

// GetUserByID returns a user by numeric ID, or (nil, nil) if not found.
func GetUserByID(db *sql.DB, id int64) (*models.User, error) {
	stmt := `
        SELECT
            id,
            uuid,
            email,
            password_hash,
            full_name,
            date_of_birth,
            avatar_url,
            nickname,
            about_me,
            is_private,
            created_at
        FROM users
        WHERE id = ?
    `

	row := db.QueryRow(stmt, id)

	var u models.User
	var isPrivateInt int

	err := row.Scan(
		&u.ID,
		&u.UUID,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.DateOfBirth,
		&u.AvatarURL,
		&u.Nickname,
		&u.AboutMe,
		&isPrivateInt,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	u.IsPrivate = isPrivateInt == 1
	return &u, nil
}

// UpdateUserPrivacy updates the is_private field for a user.
func UpdateUserPrivacy(db *sql.DB, userID int64, isPrivate bool) error {
	stmt := `
        UPDATE users
        SET is_private = ?
        WHERE id = ?
    `
	_, err := db.Exec(stmt, boolToInt(isPrivate), userID)
	return err
}

// GetSuggestedUsers returns users that the current user is NOT following (accepted or pending).
func GetSuggestedUsers(db *sql.DB, currentUserID int64, limit int) ([]models.User, error) {
	query := `
        SELECT 
            id, uuid, email, full_name, date_of_birth, 
            avatar_url, nickname, about_me, is_private, created_at
        FROM users
        WHERE id != ?
        AND id NOT IN (
            SELECT following_id 
            FROM followers 
            WHERE follower_id = ? 
            AND (status = 'accepted' OR status = 'pending')
        )
        ORDER BY created_at DESC
        LIMIT ?
    `

	rows, err := db.Query(query, currentUserID, currentUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var isPrivateInt int

		err := rows.Scan(
			&u.ID,
			&u.UUID,
			&u.Email,
			&u.FullName,
			&u.DateOfBirth,
			&u.AvatarURL,
			&u.Nickname,
			&u.AboutMe,
			&isPrivateInt,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		u.IsPrivate = isPrivateInt == 1
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// SearchUsers searches for users by full_name, nickname, or email
// Uses LIKE with wildcards for partial matching
func SearchUsers(db *sql.DB, query string, currentUserID int, limit int) ([]models.User, error) {
	searchTerm := "%" + query + "%"

	stmt := `
		SELECT id, uuid, email, full_name, date_of_birth, avatar_url, nickname, about_me, is_private, created_at
		FROM users
		WHERE id != ?
		AND (
			full_name LIKE ? COLLATE NOCASE
			OR nickname LIKE ? COLLATE NOCASE
			OR email LIKE ? COLLATE NOCASE
		)
		ORDER BY 
			CASE 
				WHEN full_name LIKE ? COLLATE NOCASE THEN 1
				WHEN nickname LIKE ? COLLATE NOCASE THEN 2
				ELSE 3
			END,
			full_name ASC
		LIMIT ?
	`

	// For exact match priority
	exactMatch := query + "%"

	rows, err := db.Query(stmt, currentUserID, searchTerm, searchTerm, searchTerm, exactMatch, exactMatch, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var u models.User
		var isPrivateInt int
		err := rows.Scan(
			&u.ID,
			&u.UUID,
			&u.Email,
			&u.FullName,
			&u.DateOfBirth,
			&u.AvatarURL,
			&u.Nickname,
			&u.AboutMe,
			&isPrivateInt,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		u.IsPrivate = isPrivateInt == 1
		users = append(users, u)
	}

	return users, rows.Err()
}