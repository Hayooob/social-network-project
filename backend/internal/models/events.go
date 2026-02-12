package models

import "time"

type Event struct {
	ID          int       `json:"id"`
	GroupID     *int      `json:"group_id,omitempty"` // nullable
	CreatorID   int       `json:"creator_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	EventDate   time.Time `json:"event_date"`
	CreatedAt   time.Time `json:"created_at"`

	// optional joined fields
	CreatorName   string `json:"creator_name,omitempty"`
	GroupName     string `json:"group_name,omitempty"`
	GoingCount    int    `json:"going_count,omitempty"`
	MaybeCount    int    `json:"maybe_count,omitempty"`
	NotGoingCount int    `json:"not_going_count,omitempty"`
	MyResponse    string `json:"my_response,omitempty"`
}

type EventResponse struct {
	ID        int       `json:"id"`
	EventID   int       `json:"event_id"`
	UserID    int       `json:"user_id"`
	Response  string    `json:"response"` // going, not_going, maybe
	CreatedAt time.Time `json:"created_at"`

	// optional joined fields
	UserName string `json:"user_name,omitempty"`
}
