package db

import (
	"database/sql"

	"social-network/internal/models"
)

// InsertPost creates a new post for the given user (legacy signature).
// This version stores only the post row in `posts`.
// For images and private allow-lists, use InsertPostWithExtras.
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

	return GetPostByID(db, int(id))
}

// InsertPostWithExtras creates a post and optionally:
// - stores an image path in post_images
// - stores allowed viewers in post_allowed_viewers (for privacy='private')
func InsertPostWithExtras(db *sql.DB, userID int, content, privacy string, imagePath string, allowedViewerIDs []int) (*models.Post, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		INSERT INTO posts (user_id, content, privacy)
		VALUES (?, ?, ?)
	`, userID, content, privacy)
	if err != nil {
		return nil, err
	}

	postID64, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	postID := int(postID64)

	if imagePath != "" {
		if _, err := tx.Exec(`
			INSERT INTO post_images (post_id, image_path)
			VALUES (?, ?)
		`, postID, imagePath); err != nil {
			return nil, err
		}
	}

	if privacy == "private" && len(allowedViewerIDs) > 0 {
		stmt, err := tx.Prepare(`
			INSERT OR IGNORE INTO post_allowed_viewers (post_id, user_id)
			VALUES (?, ?)
		`)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		for _, vid := range allowedViewerIDs {
			if vid <= 0 || vid == userID {
				continue
			}
			if _, err := stmt.Exec(postID, vid); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return GetPostByID(db, postID)
}

// GetPostByID retrieves a single post by its ID
func GetPostByID(db *sql.DB, postID int) (*models.Post, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.privacy, p.created_at,
       u.full_name as author_name,
       COALESCE(pi.image_path, '') as image_path
FROM posts p
JOIN users u ON p.user_id = u.id
LEFT JOIN post_images pi ON pi.post_id = p.id
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
		&post.ImagePath,
	)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetPublicFeed retrieves recent posts for the global feed.
// Only shows posts from PUBLIC users with privacy='public'.
func GetPublicFeed(db *sql.DB, limit int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.privacy, p.created_at,
       u.full_name as author_name,
       COALESCE(pi.image_path, '') AS image_path,
       (SELECT COUNT(1) FROM post_likes pl WHERE pl.post_id = p.id) AS like_count,
       (SELECT COUNT(1) FROM post_comments pc WHERE pc.post_id = p.id) AS comment_count,
       0 AS liked_by_me
FROM posts p
JOIN users u ON p.user_id = u.id
LEFT JOIN post_images pi ON pi.post_id = p.id
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
// - public posts from public users
// - posts from accepted follows (public + almost-private)
// - private posts ONLY if viewer is in post_allowed_viewers
// - your own posts (all privacy values)
func GetPersonalizedFeed(db *sql.DB, userID int, limit int) ([]models.Post, error) {
	query := `
		SELECT DISTINCT
			p.id, p.user_id, p.content, p.privacy, p.created_at,
			u.full_name AS author_name,
			COALESCE(pi.image_path, '') AS image_path,
			(SELECT COUNT(1) FROM post_likes pl WHERE pl.post_id = p.id) AS like_count,
			(SELECT COUNT(1) FROM post_comments pc WHERE pc.post_id = p.id) AS comment_count,
			CASE WHEN EXISTS(
				SELECT 1 FROM post_likes pl2
				WHERE pl2.post_id = p.id AND pl2.user_id = ?
			) THEN 1 ELSE 0 END AS liked_by_me
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN post_images pi ON pi.post_id = p.id
		LEFT JOIN followers f
			ON f.following_id = p.user_id
			AND f.follower_id = ?
			AND f.status = 'accepted'
		WHERE
			(p.privacy = 'public' AND u.is_private = 0)
			OR (f.id IS NOT NULL AND p.privacy IN ('public', 'almost-private'))
			OR (p.privacy = 'private' AND EXISTS(
				SELECT 1 FROM post_allowed_viewers pav
				WHERE pav.post_id = p.id AND pav.user_id = ?
			))
			OR p.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT ?
	`
	rows, err := db.Query(query, userID, userID, userID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPosts(rows)
}

// GetPostsByUserID retrieves all posts by a specific user, ordered by newest first
func GetPostsByUserID(db *sql.DB, userID int) ([]models.Post, error) {
	query := `
		SELECT
			p.id, p.user_id, p.content, p.privacy, p.created_at,
			u.full_name AS author_name,
			COALESCE(pi.image_path, '') AS image_path,
			(SELECT COUNT(1) FROM post_likes pl WHERE pl.post_id = p.id) AS like_count,
			(SELECT COUNT(1) FROM post_comments pc WHERE pc.post_id = p.id) AS comment_count,
			0 AS liked_by_me
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN post_images pi ON pi.post_id = p.id
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
    &post.ImagePath,
    &post.LikeCount,
    &post.CommentCount,
    &post.LikedByMe,
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
// - It never includes privacy='private'. (Use GetPostsByUserVisibleForViewer for private allow-lists.)
func GetPostsByUserVisible(db *sql.DB, userID int, includeAlmost bool) ([]models.Post, error) {
query := `
	SELECT
		p.id, p.user_id, p.content, p.privacy, p.created_at,
		u.full_name AS author_name,
		COALESCE(pi.image_path, '') AS image_path,
		(SELECT COUNT(1) FROM post_likes pl WHERE pl.post_id = p.id) AS like_count,
		(SELECT COUNT(1) FROM post_comments pc WHERE pc.post_id = p.id) AS comment_count,
		0 AS liked_by_me
	FROM posts p
	JOIN users u ON p.user_id = u.id
	LEFT JOIN post_images pi ON pi.post_id = p.id
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

// GetPostsByUserVisibleForViewer retrieves posts for a profile page and supports private allow-lists.
// Rules:
// - If viewerID == userID (self): return all posts.
// - Otherwise: return public, plus almost-private if includeAlmost, plus private ONLY if viewer is allowed.
func GetPostsByUserVisibleForViewer(db *sql.DB, userID int, viewerID int, includeAlmost bool) ([]models.Post, error) {
	if viewerID == userID {
		return GetPostsByUserID(db, userID)
	}

	flag := 0
	if includeAlmost {
		flag = 1
	}

query := `
	SELECT
		p.id, p.user_id, p.content, p.privacy, p.created_at,
		u.full_name AS author_name,
		COALESCE(pi.image_path, '') AS image_path,
		(SELECT COUNT(1) FROM post_likes pl WHERE pl.post_id = p.id) AS like_count,
		(SELECT COUNT(1) FROM post_comments pc WHERE pc.post_id = p.id) AS comment_count,
		CASE WHEN EXISTS(
			SELECT 1 FROM post_likes pl2
			WHERE pl2.post_id = p.id AND pl2.user_id = ?
		) THEN 1 ELSE 0 END AS liked_by_me
	FROM posts p
	JOIN users u ON p.user_id = u.id
	LEFT JOIN post_images pi ON pi.post_id = p.id
	WHERE p.user_id = ?
	  AND (
		p.privacy = 'public'
		OR (? = 1 AND p.privacy = 'almost-private')
		OR (p.privacy = 'private' AND EXISTS(
			SELECT 1 FROM post_allowed_viewers pav
			WHERE pav.post_id = p.id AND pav.user_id = ?
		))
	  )
	ORDER BY p.created_at DESC
`

rows, err := db.Query(query, viewerID, userID, flag, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPosts(rows)
}
