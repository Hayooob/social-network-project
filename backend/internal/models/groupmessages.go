package models

// GroupMessage represents a message in a group chat
type GroupMessage struct {
	ID         int    `json:"id"`
	GroupID    int    `json:"group_id"`
	SenderID   int    `json:"sender_id"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}