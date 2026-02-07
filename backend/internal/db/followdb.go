package db

import (
	"database/sql"
	"errors"

	"social-network/internal/models"
)

func CreateFollow(db *sql.DB, followerID int, followingID int, status string) error {
	stmt := `
		INSERT INTO followers (follower_id, following_id, status)
		VALUES (?, ?, ?)
	`
	_, err := db.Exec(stmt, followerID, followingID, status)
	return err
}

func GetFollowStatus(db *sql.DB, followerID int, followingID int) (*models.Follow, error) {
	stmt := `
		SELECT
			id,
			follower_id,
			following_id,
			status,
			created_at
		FROM followers
		WHERE follower_id = ? AND following_id = ?
	`

	row := db.QueryRow(stmt, followerID, followingID)

	var f models.Follow

	err := row.Scan(
		&f.ID,
		&f.FollowerID,
		&f.FollowingID,
		&f.Status,
		&f.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &f, nil
}

func UpdateFollowStatus(db *sql.DB, followerID int, followingID int, newStatus string) error {
	stmt := `
		UPDATE followers
		SET status = ?
		WHERE follower_id = ? AND following_id = ?
	`
	_, err := db.Exec(stmt, newStatus, followerID, followingID)
	return err
}

func DeleteFollow(db *sql.DB, followerID int, followingID int) error {
	stmt := `
		DELETE FROM followers
		WHERE follower_id = ? AND following_id = ?
	`
	_, err := db.Exec(stmt, followerID, followingID)
	return err
}

func GetFollowers(db *sql.DB, userID int) ([]models.Follow, error) {
	stmt := `
		SELECT
			f.id,
			f.follower_id,
			f.following_id,
			f.status,
			f.created_at,
			u.full_name
		FROM followers f
		JOIN users u ON f.follower_id = u.id
		WHERE f.following_id = ? AND f.status = 'accepted'
		ORDER BY f.created_at DESC
	`

	rows, err := db.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var follows []models.Follow

	for rows.Next() {
		var f models.Follow
		err := rows.Scan(
			&f.ID,
			&f.FollowerID,
			&f.FollowingID,
			&f.Status,
			&f.CreatedAt,
			&f.FollowerName,
		)
		if err != nil {
			return nil, err
		}
		follows = append(follows, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return follows, nil
}

func GetFollowing(db *sql.DB, userID int) ([]models.Follow, error) {
	stmt := `
		SELECT
			f.id,
			f.follower_id,
			f.following_id,
			f.status,
			f.created_at,
			u.full_name
		FROM followers f
		JOIN users u ON f.following_id = u.id
		WHERE f.follower_id = ? AND f.status = 'accepted'
		ORDER BY f.created_at DESC
	`

	rows, err := db.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var follows []models.Follow

	for rows.Next() {
		var f models.Follow
		err := rows.Scan(
			&f.ID,
			&f.FollowerID,
			&f.FollowingID,
			&f.Status,
			&f.CreatedAt,
			&f.FollowingName,
		)
		if err != nil {
			return nil, err
		}
		follows = append(follows, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return follows, nil
}

func GetPendingFollowRequests(db *sql.DB, userID int) ([]models.Follow, error) {
	stmt := `
		SELECT
			f.id,
			f.follower_id,
			f.following_id,
			f.status,
			f.created_at,
			u.full_name
		FROM followers f
		JOIN users u ON f.follower_id = u.id
		WHERE f.following_id = ? AND f.status = 'pending'
		ORDER BY f.created_at DESC
	`

	rows, err := db.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var follows []models.Follow

	for rows.Next() {
		var f models.Follow
		err := rows.Scan(
			&f.ID,
			&f.FollowerID,
			&f.FollowingID,
			&f.Status,
			&f.CreatedAt,
			&f.FollowerName,
		)
		if err != nil {
			return nil, err
		}
		follows = append(follows, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return follows, nil
}

func GetFollowerCount(db *sql.DB, userID int) (int, error) {
	stmt := `
		SELECT COUNT(*)
		FROM followers
		WHERE following_id = ? AND status = 'accepted'
	`

	var count int
	err := db.QueryRow(stmt, userID).Scan(&count)
	return count, err
}

func GetFollowingCount(db *sql.DB, userID int) (int, error) {
	stmt := `
		SELECT COUNT(*)
		FROM followers
		WHERE follower_id = ? AND status = 'accepted'
	`

	var count int
	err := db.QueryRow(stmt, userID).Scan(&count)
	return count, err
}
// IsFollowing returns true if followerID follows followingID with status='accepted'.
func IsFollowing(db *sql.DB, followerID int, followingID int) (bool, error) {
    stmt := `
        SELECT 1
        FROM followers
        WHERE follower_id = ? AND following_id = ? AND status = 'accepted'
        LIMIT 1
    `

    var one int
    err := db.QueryRow(stmt, followerID, followingID).Scan(&one)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

// AreMutualFriends returns true if both users follow each other with status='accepted'
func AreMutualFriends(db *sql.DB, userID1 int, userID2 int) (bool, error) {
	stmt := `
		SELECT COUNT(*)
		FROM followers f1
		JOIN followers f2 ON f1.follower_id = f2.following_id AND f1.following_id = f2.follower_id
		WHERE f1.follower_id = ? AND f1.following_id = ?
		AND f1.status = 'accepted' AND f2.status = 'accepted'
	`

	var count int
	err := db.QueryRow(stmt, userID1, userID2).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMutualFriends returns list of users who mutually follow each other with the given user
func GetMutualFriends(db *sql.DB, userID int) ([]models.User, error) {
	stmt := `
		SELECT u.id, u.uuid, u.email, u.full_name, u.nickname, u.avatar_url, u.is_private
		FROM users u
		WHERE u.id IN (
			SELECT f1.following_id
			FROM followers f1
			JOIN followers f2 ON f1.follower_id = f2.following_id AND f1.following_id = f2.follower_id
			WHERE f1.follower_id = ?
			AND f1.status = 'accepted' AND f2.status = 'accepted'
		)
		ORDER BY u.full_name ASC
	`

	rows, err := db.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID,
			&u.UUID,
			&u.Email,
			&u.FullName,
			&u.Nickname,
			&u.AvatarURL,
			&u.IsPrivate,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}