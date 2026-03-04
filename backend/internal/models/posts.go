package models

import "time"

type Post struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Content    string    `json:"content"`
	Privacy    string    `json:"privacy"` // public, private, almost-private
	CreatedAt  time.Time `json:"created_at"`
	AuthorName string    `json:"author_name,omitempty"`
	// newly added avatar url of the author
	AuthorAvatarURL string `json:"author_avatar_url,omitempty"`
	ImagePath       string `json:"image_path,omitempty"`
	LikeCount       int    `json:"like_count"`
	CommentCount    int    `json:"comment_count"`
	LikedByMe       int    `json:"liked_by_me"`
}
