package models

import "time"

type Group struct {
	ID          int       `json:"id"`
	CreatorID   int       `json:"creator_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPrivate   int       `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`

	// joined
	CreatorName string `json:"creator_name,omitempty"`
	MemberCount int    `json:"member_count,omitempty"`

	// viewer state
	MyStatus string `json:"my_status,omitempty"` // accepted/pending/""
	MyRole   string `json:"my_role,omitempty"`   // admin/member/""

	// invitation state (for UI: Accept/Decline)
	InvitationID     int    `json:"invitation_id,omitempty"`     // 0 if none
	InvitationStatus string `json:"invitation_status,omitempty"` // pending/""
}

type GroupMember struct {
	ID       int       `json:"id"`
	GroupID  int       `json:"group_id"`
	UserID   int       `json:"user_id"`
	Role     string    `json:"role"`
	Status   string    `json:"status"`
	JoinedAt time.Time `json:"joined_at"`

	UserName string `json:"user_name,omitempty"`
}
