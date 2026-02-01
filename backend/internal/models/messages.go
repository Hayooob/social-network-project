package models

import "time"

type Message struct {
	ID           int       `json:"id"`
	SenderID     int       `json:"sender_id"`
	ReceiverID   int       `json:"receiver_id"`
	Content      string    `json:"content"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
	SenderName   string    `json:"sender_name,omitempty"`
	ReceiverName string    `json:"receiver_name,omitempty"`
}

type Conversation struct {
	UserID         int       `json:"user_id"`
	UserName       string    `json:"user_name"`
	LastMessage    string    `json:"last_message"`
	LastMessageAt  time.Time `json:"last_message_at"`
	UnreadCount    int       `json:"unread_count"`
}