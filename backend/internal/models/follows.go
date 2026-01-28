package models

import "time"

type Follow struct {
	ID            int       `json:"id"`
	FollowerID    int       `json:"follower_id"`
	FollowingID   int       `json:"following_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	FollowerName  string    `json:"follower_name,omitempty"`
	FollowingName string    `json:"following_name,omitempty"`
}