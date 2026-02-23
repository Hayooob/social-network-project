package db

import (
	"database/sql"
	"social-network/internal/models"
)

// SendGroupMessage saves a message to a group chat
func SendGroupMessage(db *sql.DB, groupID, senderID int, content string) (*models.GroupMessage, error) {
	stmt := `INSERT INTO group_messages (group_id, sender_id, content) VALUES (?, ?, ?)`
	result, err := db.Exec(stmt, groupID, senderID, content)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetGroupMessageByID(db, int(id))
}

// GetGroupMessageByID retrieves a single group message by ID
func GetGroupMessageByID(db *sql.DB, messageID int) (*models.GroupMessage, error) {
	stmt := `
		SELECT gm.id, gm.group_id, gm.sender_id, u.full_name, gm.content, gm.created_at
		FROM group_messages gm
		JOIN users u ON u.id = gm.sender_id
		WHERE gm.id = ?
	`
	var msg models.GroupMessage
	err := db.QueryRow(stmt, messageID).Scan(
		&msg.ID, &msg.GroupID, &msg.SenderID, &msg.SenderName, &msg.Content, &msg.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetGroupMessages retrieves messages for a group with pagination
func GetGroupMessages(db *sql.DB, groupID, limit, offset int) ([]models.GroupMessage, error) {
	stmt := `
		SELECT gm.id, gm.group_id, gm.sender_id, u.full_name, gm.content, gm.created_at
		FROM group_messages gm
		JOIN users u ON u.id = gm.sender_id
		WHERE gm.group_id = ?
		ORDER BY gm.created_at ASC
		LIMIT ? OFFSET ?
	`
	rows, err := db.Query(stmt, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.GroupMessage
	for rows.Next() {
		var msg models.GroupMessage
		if err := rows.Scan(
			&msg.ID, &msg.GroupID, &msg.SenderID, &msg.SenderName, &msg.Content, &msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetGroupMemberIDs returns all member IDs of a group
func GetGroupMemberIDs(db *sql.DB, groupID int) ([]int64, error) {
	rows, err := db.Query(`
		SELECT user_id FROM group_members 
		WHERE group_id = ? AND status = 'accepted'
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}