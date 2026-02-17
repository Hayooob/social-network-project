package db

import (
	"database/sql"
	"time"

	"social-network/internal/models"
)

// CreateGroup inserts a new group and adds the creator as an accepted admin member.
func CreateGroup(db *sql.DB, creatorID int, name, description string, isPrivate bool) (*models.Group, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	priv := 0
	if isPrivate {
		priv = 1
	}

	res, err := tx.Exec(`
		INSERT INTO groups (creator_id, name, description, is_private)
		VALUES (?, ?, ?, ?)
	`, creatorID, name, description, priv)
	if err != nil {
		return nil, err
	}

	groupID64, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	groupID := int(groupID64)

	// Creator becomes admin + accepted member
	_, err = tx.Exec(`
		INSERT INTO group_members (group_id, user_id, role, status, joined_at)
		VALUES (?, ?, 'admin', 'accepted', CURRENT_TIMESTAMP)
	`, groupID, creatorID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return GetGroupByID(db, groupID, creatorID)
}

// GetGroupByID returns group details along with creator name, member count, and viewer membership status.
func GetGroupByID(db *sql.DB, groupID int, viewerID int) (*models.Group, error) {
	query := `
		SELECT
			g.id, g.creator_id, g.name, g.description, g.is_private, g.created_at,
			u.full_name AS creator_name,
			(SELECT COUNT(1) FROM group_members gm WHERE gm.group_id = g.id AND gm.status = 'accepted') AS member_count,
			COALESCE((SELECT gm2.status FROM group_members gm2 WHERE gm2.group_id = g.id AND gm2.user_id = ?), '') AS my_status,
			COALESCE((SELECT gm3.role   FROM group_members gm3 WHERE gm3.group_id = g.id AND gm3.user_id = ?), '') AS my_role
		FROM groups g
		JOIN users u ON u.id = g.creator_id
		WHERE g.id = ?
	`

	var grp models.Group
	var myStatus, myRole string
	if err := db.QueryRow(query, viewerID, viewerID, groupID).Scan(
		&grp.ID,
		&grp.CreatorID,
		&grp.Name,
		&grp.Description,
		&grp.IsPrivate,
		&grp.CreatedAt,
		&grp.CreatorName,
		&grp.MemberCount,
		&myStatus,
		&myRole,
	); err != nil {
		return nil, err
	}
	if myStatus != "" {
		grp.MyStatus = myStatus
	}
	if myRole != "" {
		grp.MyRole = myRole
	}
	return &grp, nil
}

// GetUserGroups lists groups where the user is a member (pending or accepted).
func GetUserGroups(db *sql.DB, userID int) ([]models.Group, error) {
	query := `
		SELECT
			g.id, g.creator_id, g.name, g.description, g.is_private, g.created_at,
			u.full_name AS creator_name,
			(SELECT COUNT(1) FROM group_members gm WHERE gm.group_id = g.id AND gm.status = 'accepted') AS member_count,
			gm.status AS my_status,
			gm.role AS my_role
		FROM group_members gm
		JOIN groups g ON g.id = gm.group_id
		JOIN users u ON u.id = g.creator_id
		WHERE gm.user_id = ?
		ORDER BY g.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		if err := rows.Scan(
			&g.ID,
			&g.CreatorID,
			&g.Name,
			&g.Description,
			&g.IsPrivate,
			&g.CreatedAt,
			&g.CreatorName,
			&g.MemberCount,
			&g.MyStatus,
			&g.MyRole,
		); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// AddGroupMember adds a user to a group.
// If the group is private, the member is added as 'pending'.
func AddGroupMember(db *sql.DB, groupID int, userID int) (*models.GroupMember, error) {
	var isPrivate int
	if err := db.QueryRow(`SELECT is_private FROM groups WHERE id = ?`, groupID).Scan(&isPrivate); err != nil {
		return nil, err
	}

	status := "accepted"
	if isPrivate == 1 {
		status = "pending"
	}

	// idempotent insert
	_, err := db.Exec(`
		INSERT OR IGNORE INTO group_members (group_id, user_id, role, status, joined_at)
		VALUES (?, ?, 'member', ?, CURRENT_TIMESTAMP)
	`, groupID, userID, status)
	if err != nil {
		return nil, err
	}

	return getGroupMember(db, groupID, userID)
}

// UpdateMemberStatus updates status to accepted/declined.
// If accepted, sets joined_at to current time.
func UpdateMemberStatus(db *sql.DB, groupID int, userID int, status string) error {
	if status == "accepted" {
		_, err := db.Exec(`
			UPDATE group_members
			SET status = 'accepted', joined_at = CURRENT_TIMESTAMP
			WHERE group_id = ? AND user_id = ?
		`, groupID, userID)
		return err
	}
	// decline -> remove membership row
	_, err := db.Exec(`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, userID)
	return err
}

// RemoveGroupMember removes a user from group (leave or kick).
func RemoveGroupMember(db *sql.DB, groupID int, userID int) error {
	_, err := db.Exec(`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, userID)
	return err
}

// GetGroupMembers returns members (pending + accepted), newest first.
func GetGroupMembers(db *sql.DB, groupID int) ([]models.GroupMember, error) {
	query := `
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.joined_at,
			u.full_name AS user_name
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ?
		ORDER BY gm.joined_at DESC
	`

	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.GroupMember
	for rows.Next() {
		var m models.GroupMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt, &m.UserName); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// IsGroupMember checks whether user is an accepted member of the group.
func IsGroupMember(db *sql.DB, groupID int, userID int) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(1) FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsGroupAdmin checks whether user is an accepted admin of the group.
func IsGroupAdmin(db *sql.DB, groupID int, userID int) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(1) FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'accepted' AND role = 'admin'
	`, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func getGroupMember(db *sql.DB, groupID int, userID int) (*models.GroupMember, error) {
	query := `
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.joined_at,
			u.full_name AS user_name
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ? AND gm.user_id = ?
	`

	var m models.GroupMember
	var joinedAtStr string
	if err := db.QueryRow(query, groupID, userID).Scan(&m.ID, &m.GroupID, &m.UserID, &m.Role, &m.Status, &joinedAtStr, &m.UserName); err != nil {
		return nil, err
	}
	// sqlite may scan DATETIME into string depending on driver settings
	if t, err := time.Parse("2006-01-02 15:04:05", joinedAtStr); err == nil {
		m.JoinedAt = t
	}
	return &m, nil
}
// GetVisibleGroups returns groups visible to the user:
// - groups created by users they follow (accepted follows)
// - groups they created
// - groups they are already related to (member or invited or requested)
func GetVisibleGroups(dbConn *sql.DB, userID int) ([]models.Group, error) {
	query := `
		SELECT
			g.id, g.creator_id, g.name, g.description, g.is_private, g.created_at,
			u.full_name AS creator_name,
			(SELECT COUNT(1) FROM group_members gm WHERE gm.group_id = g.id AND gm.status = 'accepted') AS member_count,
			COALESCE(gm.status, '') AS my_status,
			COALESCE(gm.role, '') AS my_role,
			COALESCE(gi.id, 0) AS invitation_id,
			COALESCE(gi.status, '') AS invitation_status
		FROM groups g
		JOIN users u ON u.id = g.creator_id
		LEFT JOIN group_members gm
			ON gm.group_id = g.id AND gm.user_id = ?
		LEFT JOIN group_invitations gi
			ON gi.group_id = g.id AND gi.invitee_id = ? AND gi.status = 'pending'
		WHERE
			g.creator_id = ?
			OR gm.user_id IS NOT NULL
			OR gi.id IS NOT NULL
			OR g.creator_id IN (
				SELECT following_id FROM followers
				WHERE follower_id = ? AND status = 'accepted'
			)
		ORDER BY g.created_at DESC
	`

	rows, err := dbConn.Query(query, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		if err := rows.Scan(
			&g.ID,
			&g.CreatorID,
			&g.Name,
			&g.Description,
			&g.IsPrivate,
			&g.CreatedAt,
			&g.CreatorName,
			&g.MemberCount,
			&g.MyStatus,
			&g.MyRole,
			&g.InvitationID,
			&g.InvitationStatus,
		); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}
// GetAcceptedGroupMemberIDs returns user IDs of accepted members.
func GetAcceptedGroupMemberIDs(dbConn *sql.DB, groupID int) ([]int, error) {
	rows, err := dbConn.Query(`
		SELECT user_id FROM group_members
		WHERE group_id = ? AND status = 'accepted'
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
