package db

import (
	"database/sql"

	"social-network/internal/models"
)

func CreateNotification(db *sql.DB, userID int, notifType string, referenceID *int, fromUserID *int, content string) (*models.Notification, error) {
	stmt := `
		INSERT INTO notifications (user_id, type, reference_id, from_user_id, content)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := db.Exec(stmt, userID, notifType, referenceID, fromUserID, content)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetNotificationByID(db, int(id))
}

func GetNotificationByID(db *sql.DB, notifID int) (*models.Notification, error) {
	stmt := `
		SELECT 
			n.id, n.user_id, n.type, n.reference_id, n.from_user_id, n.content, n.is_read, n.created_at,
			COALESCE(u.full_name, '') as from_user_name
		FROM notifications n
		LEFT JOIN users u ON n.from_user_id = u.id
		WHERE n.id = ?
	`

	row := db.QueryRow(stmt, notifID)

	var n models.Notification
	var isReadInt int

	err := row.Scan(
		&n.ID,
		&n.UserID,
		&n.Type,
		&n.ReferenceID,
		&n.FromUserID,
		&n.Content,
		&isReadInt,
		&n.CreatedAt,
		&n.FromUserName,
	)
	if err != nil {
		return nil, err
	}

	n.IsRead = isReadInt == 1
	return &n, nil
}

func GetNotifications(db *sql.DB, userID int, limit int, offset int) ([]models.Notification, error) {
	stmt := `
		SELECT 
			n.id, n.user_id, n.type, n.reference_id, n.from_user_id, n.content, n.is_read, n.created_at,
			COALESCE(u.full_name, '') as from_user_name
		FROM notifications n
		LEFT JOIN users u ON n.from_user_id = u.id
		WHERE n.user_id = ?
		ORDER BY n.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(stmt, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification

	for rows.Next() {
		var n models.Notification
		var isReadInt int

		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.ReferenceID,
			&n.FromUserID,
			&n.Content,
			&isReadInt,
			&n.CreatedAt,
			&n.FromUserName,
		)
		if err != nil {
			return nil, err
		}

		n.IsRead = isReadInt == 1
		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

func MarkNotificationAsRead(db *sql.DB, notifID int, userID int) error {
	stmt := `
		UPDATE notifications
		SET is_read = 1
		WHERE id = ? AND user_id = ?
	`

	_, err := db.Exec(stmt, notifID, userID)
	return err
}

func MarkAllNotificationsAsRead(db *sql.DB, userID int) error {
	stmt := `
		UPDATE notifications
		SET is_read = 1
		WHERE user_id = ? AND is_read = 0
	`

	_, err := db.Exec(stmt, userID)
	return err
}

func GetUnreadNotificationCount(db *sql.DB, userID int) (int, error) {
	stmt := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ? AND is_read = 0
	`

	var count int
	err := db.QueryRow(stmt, userID).Scan(&count)
	return count, err
}

func DeleteNotification(db *sql.DB, notifID int, userID int) error {
	stmt := `
		DELETE FROM notifications
		WHERE id = ? AND user_id = ?
	`

	_, err := db.Exec(stmt, notifID, userID)
	return err
}
