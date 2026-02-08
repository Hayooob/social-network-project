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

// GetPublicFeed retrieves recent posts for the global feed
// Only shows posts from PUBLIC users with privacy='public'
func GetPublicFeed(db *sql.DB, limit int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.privacy, p.created_at,
		       u.full_name as author_name
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.privacy = 'public'
		AND u.is_private = 0
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

// GetPersonalizedFeed returns posts for a logged-in user:
// - All public posts from public users
// - Posts from users they follow (even if private)
// - Their own posts
func GetPersonalizedFeed(db *sql.DB, userID int, limit int) ([]models.Post, error) {
	query := `
		SELECT DISTINCT p.id, p.user_id, p.content, p.privacy, p.created_at,
		       u.full_name as author_name
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN followers f ON f.following_id = p.user_id AND f.follower_id = ? AND f.status = 'accepted'
		WHERE 
			-- Public posts from public users
			(p.privacy = 'public' AND u.is_private = 0)
			-- OR posts from users I follow
			OR (f.id IS NOT NULL AND p.privacy IN ('public', 'almost-private'))
			-- OR my own posts
			OR p.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT ?
	`

	rows, err := db.Query(query, userID, userID, limit)
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

// GetPostsByUserVisible retrieves posts by a user filtered by viewer visibility.
// - If includeAlmost is true, it includes privacy='almost-private' in addition to public.
// - It never includes privacy='private' (stage 5 does not implement per-follower allow lists).
func GetPostsByUserVisible(db *sql.DB, userID int, includeAlmost bool) ([]models.Post, error) {
	query := `
        SELECT p.id, p.user_id, p.content, p.privacy, p.created_at,
               u.full_name as author_name
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.user_id = ?
          AND (
            p.privacy = 'public'
            OR (? = 1 AND p.privacy = 'almost-private')
          )
        ORDER BY p.created_at DESC
    `

	flag := 0
	if includeAlmost {
		flag = 1
	}

	rows, err := db.Query(query, userID, flag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPosts(rows)
}