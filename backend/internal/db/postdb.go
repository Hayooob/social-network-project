package db

import (
	"database/sql"
	"social-network/internal/models"
)

// InsertPost creates a new post for the given user
func InsertPost(db *sql.DB, userID int, content, privacy string) (*models.Post, error) {
	query := `
		INSERT INTO posts (user_id, content, privacy)
		VALUES (?, ?, ?)
	`
	
	result, err := db.Exec(query, userID, content, privacy)
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	// Fetch the created post
	return GetPostByID(db, int(id))
}

// GetPostByID retrieves a single post by its ID
func GetPostByID(db *sql.DB, postID int) (*models.Post, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.privacy, p.created_at,
		       u.full_name as author_name
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`
	
	var post models.Post
	err := db.QueryRow(query, postID).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&post.Privacy,
		&post.CreatedAt,
		&post.AuthorName,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &post, nil
}

// GetPublicFeed retrieves recent public posts for the global feed
func GetPublicFeed(db *sql.DB, limit int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.privacy, p.created_at,
		       u.full_name as author_name
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.privacy = 'public'
		ORDER BY p.created_at DESC
		LIMIT ?
	`
	
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return scanPosts(rows)
}

// GetPostsByUserID retrieves all posts by a specific user, ordered by newest first
func GetPostsByUserID(db *sql.DB, userID int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.privacy, p.created_at,
		       u.full_name as author_name
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
	`
	
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return scanPosts(rows)
}

// scanPosts is a helper function to scan multiple post rows
func scanPosts(rows *sql.Rows) ([]models.Post, error) {
	var posts []models.Post
	
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&post.Privacy,
			&post.CreatedAt,
			&post.AuthorName,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return posts, nil
}