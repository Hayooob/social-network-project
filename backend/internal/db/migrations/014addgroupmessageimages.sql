-- Add image_path column to group_messages table
-- Wrapped in a transaction to handle if already exists
CREATE TABLE IF NOT EXISTS group_messages_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);
DROP TABLE IF EXISTS group_messages_new;

ALTER TABLE group_messages ADD COLUMN image_path TEXT;