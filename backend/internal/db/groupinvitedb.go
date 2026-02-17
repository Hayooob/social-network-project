package db

import (
	"database/sql"
	"errors"
)

type GroupInvitationRow struct {
	ID        int    `json:"id"`
	GroupID   int    `json:"group_id"`
	InviterID int    `json:"inviter_id"`
	InviteeID int    `json:"invitee_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`

	GroupName   string `json:"group_name,omitempty"`
	InviterName string `json:"inviter_name,omitempty"`
}

func CreateGroupInvitation(db *sql.DB, groupID, inviterID, inviteeID int) (int, error) {
	// If already member, block
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(1) FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`, groupID, inviteeID).Scan(&count); err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("user already a member")
	}

	// Insert or reset to pending
	res, err := db.Exec(`
		INSERT INTO group_invitations (group_id, inviter_id, invitee_id, status)
		VALUES (?, ?, ?, 'pending')
		ON CONFLICT(group_id, invitee_id) DO UPDATE SET
			inviter_id = excluded.inviter_id,
			status = 'pending',
			created_at = CURRENT_TIMESTAMP
	`, groupID, inviterID, inviteeID)
	if err != nil {
		return 0, err
	}

	id64, err := res.LastInsertId()
	if err == nil && id64 > 0 {
		return int(id64), nil
	}

	// If conflict updated, fetch id
	var id int
	err = db.QueryRow(`
		SELECT id FROM group_invitations
		WHERE group_id = ? AND invitee_id = ?
	`, groupID, inviteeID).Scan(&id)
	return id, err
}

func GetMyPendingInvitations(db *sql.DB, userID int) ([]GroupInvitationRow, error) {
	rows, err := db.Query(`
		SELECT gi.id, gi.group_id, gi.inviter_id, gi.invitee_id, gi.status, gi.created_at,
		       g.name as group_name, u.full_name as inviter_name
		FROM group_invitations gi
		JOIN groups g ON g.id = gi.group_id
		JOIN users u ON u.id = gi.inviter_id
		WHERE gi.invitee_id = ? AND gi.status = 'pending'
		ORDER BY gi.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []GroupInvitationRow
	for rows.Next() {
		var r GroupInvitationRow
		if err := rows.Scan(&r.ID, &r.GroupID, &r.InviterID, &r.InviteeID, &r.Status, &r.CreatedAt, &r.GroupName, &r.InviterName); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func GetInvitationByID(db *sql.DB, inviteID int) (*GroupInvitationRow, error) {
	var r GroupInvitationRow
	err := db.QueryRow(`
		SELECT id, group_id, inviter_id, invitee_id, status, created_at
		FROM group_invitations
		WHERE id = ?
	`, inviteID).Scan(&r.ID, &r.GroupID, &r.InviterID, &r.InviteeID, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func DeleteInvitation(db *sql.DB, inviteID int) error {
	_, err := db.Exec(`DELETE FROM group_invitations WHERE id = ?`, inviteID)
	return err
}
