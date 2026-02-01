package db

import (
	"database/sql"

	"social-network/internal/models"
)

func SendMessage(db *sql.DB, senderID int, receiverID int, content string) (*models.Message, error) {
	stmt := `
		INSERT INTO messages (sender_id, receiver_id, content)
		VALUES (?, ?, ?)
	`

	result, err := db.Exec(stmt, senderID, receiverID, content)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetMessageByID(db, int(id))
}

func GetMessageByID(db *sql.DB, messageID int) (*models.Message, error) {
	stmt := `
		SELECT 
			m.id, m.sender_id, m.receiver_id, m.content, m.is_read, m.created_at,
			s.full_name as sender_name, r.full_name as receiver_name
		FROM messages m
		JOIN users s ON m.sender_id = s.id
		JOIN users r ON m.receiver_id = r.id
		WHERE m.id = ?
	`

	row := db.QueryRow(stmt, messageID)

	var m models.Message
	var isReadInt int

	err := row.Scan(
		&m.ID,
		&m.SenderID,
		&m.ReceiverID,
		&m.Content,
		&isReadInt,
		&m.CreatedAt,
		&m.SenderName,
		&m.ReceiverName,
	)
	if err != nil {
		return nil, err
	}

	m.IsRead = isReadInt == 1
	return &m, nil
}

func GetConversation(db *sql.DB, userID int, otherUserID int, limit int, offset int) ([]models.Message, error) {
	stmt := `
		SELECT 
			m.id, m.sender_id, m.receiver_id, m.content, m.is_read, m.created_at,
			s.full_name as sender_name, r.full_name as receiver_name
		FROM messages m
		JOIN users s ON m.sender_id = s.id
		JOIN users r ON m.receiver_id = r.id
		WHERE (m.sender_id = ? AND m.receiver_id = ?)
		   OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at ASC
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(stmt, userID, otherUserID, otherUserID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message

	for rows.Next() {
		var m models.Message
		var isReadInt int

		err := rows.Scan(
			&m.ID,
			&m.SenderID,
			&m.ReceiverID,
			&m.Content,
			&isReadInt,
			&m.CreatedAt,
			&m.SenderName,
			&m.ReceiverName,
		)
		if err != nil {
			return nil, err
		}

		m.IsRead = isReadInt == 1
		messages = append(messages, m)
	}

	return messages, rows.Err()
}

func GetConversationList(db *sql.DB, userID int) ([]models.Conversation, error) {
	stmt := `
		SELECT 
			other_user_id,
			u.full_name,
			last_message,
			last_message_at,
			unread_count
		FROM (
			SELECT 
				CASE 
					WHEN sender_id = ? THEN receiver_id 
					ELSE sender_id 
				END as other_user_id,
				content as last_message,
				created_at as last_message_at,
				(SELECT COUNT(*) FROM messages m2 
				 WHERE m2.sender_id = CASE WHEN messages.sender_id = ? THEN messages.receiver_id ELSE messages.sender_id END
				 AND m2.receiver_id = ?
				 AND m2.is_read = 0) as unread_count,
				ROW_NUMBER() OVER (
					PARTITION BY CASE 
						WHEN sender_id = ? THEN receiver_id 
						ELSE sender_id 
					END 
					ORDER BY created_at DESC
				) as rn
			FROM messages
			WHERE sender_id = ? OR receiver_id = ?
		) sub
		JOIN users u ON sub.other_user_id = u.id
		WHERE rn = 1
		ORDER BY last_message_at DESC
	`

	rows, err := db.Query(stmt, userID, userID, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.Conversation

	for rows.Next() {
		var c models.Conversation

		err := rows.Scan(
			&c.UserID,
			&c.UserName,
			&c.LastMessage,
			&c.LastMessageAt,
			&c.UnreadCount,
		)
		if err != nil {
			return nil, err
		}

		conversations = append(conversations, c)
	}

	return conversations, rows.Err()
}

func MarkMessagesAsRead(db *sql.DB, receiverID int, senderID int) error {
	stmt := `
		UPDATE messages
		SET is_read = 1
		WHERE receiver_id = ? AND sender_id = ? AND is_read = 0
	`

	_, err := db.Exec(stmt, receiverID, senderID)
	return err
}

func GetUnreadMessageCount(db *sql.DB, userID int) (int, error) {
	stmt := `
		SELECT COUNT(*)
		FROM messages
		WHERE receiver_id = ? AND is_read = 0
	`

	var count int
	err := db.QueryRow(stmt, userID).Scan(&count)
	return count, err
}