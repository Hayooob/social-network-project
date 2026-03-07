package models

// GroupPostComment represents a comment on a group post
type GroupPostComment struct {
	ID         int    `json:"id"`
	PostID     int    `json:"post_id"`
	UserID     int64  `json:"user_id"`
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
	ImagePath  string `json:"image_path,omitempty"`
	CreatedAt  string `json:"created_at"`
}
