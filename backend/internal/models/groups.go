package models

import "time"

type Group struct {
	ID          int       `json:"id"`
	CreatorID   int       `json:"creator_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPrivate   int       `json:"is_private"` // 0/1 for sqlite
	CreatedAt   time.Time `json:"created_at"`

	// optional joined fields
	CreatorName string `json:"creator_name,omitempty"`
	MemberCount int    `json:"member_count,omitempty"`
	MyStatus    string `json:"my_status,omitempty"` // pending/accepted/none
	MyRole      string `json:"my_role,omitempty"`   // admin/member
}

type GroupMember struct {
	ID       int       `json:"id"`
	GroupID  int       `json:"group_id"`
	UserID   int       `json:"user_id"`
	Role     string    `json:"role"`   // admin/member
	Status   string    `json:"status"` // pending/accepted
	JoinedAt time.Time `json:"joined_at"`

	// optional joined fields
	UserName string `json:"user_name,omitempty"`
}
