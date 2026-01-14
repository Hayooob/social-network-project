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