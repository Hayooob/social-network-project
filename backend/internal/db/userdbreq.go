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

// inserts a new user into the users table
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
			return nil, nil // not found
		}
		return nil, err
	}

	u.IsPrivate = isPrivateInt == 1
	return &u, nil
}

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
