package models

import "time"

type GroupPost struct {
    ID           int       `json:"id"`
    GroupID      int       `json:"group_id"`
    UserID       int       `json:"user_id"`
    Content      string    `json:"content"`
    CreatedAt    time.Time `json:"created_at"`
    AuthorName   string    `json:"author_name,omitempty"`
    CommentCount int       `json:"comment_count"`
}