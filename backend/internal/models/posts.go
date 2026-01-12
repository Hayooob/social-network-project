package models

import "time"

// Post represents a user's post in the social network
type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	Privacy   string    `json:"privacy"` // public, private, almost-private
	CreatedAt time.Time `json:"created_at"`
	
	// Optional fields for joined data
	AuthorName string `json:"author_name,omitempty"`
}