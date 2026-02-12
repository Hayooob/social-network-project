package db

import (
	"database/sql"

	"social-network/internal/models"
)

func CreateGroupPost(db *sql.DB, groupID int, userID int, content string) (*models.GroupPost, error) {
	res, err := db.Exec(`
		INSERT INTO group_posts (group_id, user_id, content)
		VALUES (?, ?, ?)
	`, groupID, userID, content)
	if err != nil {
		return nil, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetGroupPostByID(db, int(id64))
}

func GetGroupPostByID(db *sql.DB, postID int) (*models.GroupPost, error) {
	query := `
		SELECT gp.id, gp.group_id, gp.user_id, gp.content, gp.created_at,
			u.full_name AS author_name
		FROM group_posts gp
		JOIN users u ON u.id = gp.user_id
		WHERE gp.id = ?
	`

	var p models.GroupPost
	if err := db.QueryRow(query, postID).Scan(&p.ID, &p.GroupID, &p.UserID, &p.Content, &p.CreatedAt, &p.AuthorName); err != nil {
		return nil, err
	}
	return &p, nil
}

func GetGroupPosts(db *sql.DB, groupID int, limit int) ([]models.GroupPost, error) {
	query := `
		SELECT gp.id, gp.group_id, gp.user_id, gp.content, gp.created_at,
			u.full_name AS author_name
		FROM group_posts gp
		JOIN users u ON u.id = gp.user_id
		WHERE gp.group_id = ?
		ORDER BY gp.created_at DESC
		LIMIT ?
	`
	rows, err := db.Query(query, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.GroupPost
	for rows.Next() {
		var p models.GroupPost
		if err := rows.Scan(&p.ID, &p.GroupID, &p.UserID, &p.Content, &p.CreatedAt, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func DeleteGroupPost(db *sql.DB, postID int) error {
	_, err := db.Exec(`DELETE FROM group_posts WHERE id = ?`, postID)
	return err
}
