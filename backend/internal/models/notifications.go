package models

import "time"

type Notification struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Type       string    `json:"type"`
	ReferenceID *int     `json:"reference_id,omitempty"`
	FromUserID *int      `json:"from_user_id,omitempty"`
	Content    string    `json:"content,omitempty"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
	FromUserName string  `json:"from_user_name,omitempty"`
}

// Notification types
const (
	NotificationTypeFollowRequest = "follow_request"
	NotificationTypeFollowAccept  = "follow_accept"
	NotificationTypeNewMessage    = "new_message"
)