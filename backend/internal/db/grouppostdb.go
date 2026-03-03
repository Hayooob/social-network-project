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
            u.full_name AS author_name,
            COUNT(c.id) AS comment_count
        FROM group_posts gp
        JOIN users u ON u.id = gp.user_id
        LEFT JOIN group_post_comments c ON c.post_id = gp.id
        WHERE gp.group_id = ?
        GROUP BY gp.id
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
        if err := rows.Scan(&p.ID, &p.GroupID, &p.UserID, &p.Content, &p.CreatedAt, &p.AuthorName, &p.CommentCount); err != nil {
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

// GetGroupPostComments retrieves all comments for a group post
func GetGroupPostComments(db *sql.DB, postID int) ([]models.GroupPostComment, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, u.full_name, c.content, COALESCE(c.image_path, ''), c.created_at
		FROM group_post_comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`
	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.GroupPostComment
	for rows.Next() {
		var c models.GroupPostComment
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.AuthorName, &c.Content, &c.ImagePath, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// AddGroupPostComment creates a new comment on a group post
func AddGroupPostComment(db *sql.DB, postID int, userID int, content string, imagePath string) (*models.GroupPostComment, error) {
	res, err := db.Exec(`
		INSERT INTO group_post_comments (post_id, user_id, content, image_path)
		VALUES (?, ?, ?, ?)
	`, postID, userID, content, imagePath)
	if err != nil {
		return nil, err
	}

	newID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Return the created comment
	query := `
		SELECT c.id, c.post_id, c.user_id, u.full_name, c.content, COALESCE(c.image_path, ''), c.created_at
		FROM group_post_comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = ?
	`
	var c models.GroupPostComment
	err = db.QueryRow(query, newID).Scan(&c.ID, &c.PostID, &c.UserID, &c.AuthorName, &c.Content, &c.ImagePath, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
