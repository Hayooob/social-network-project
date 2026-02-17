-- Stage 7 fix: explicit invitations

CREATE TABLE IF NOT EXISTS group_invitations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  group_id INTEGER NOT NULL,
  inviter_id INTEGER NOT NULL,
  invitee_id INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending', -- pending/accepted/declined
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
  FOREIGN KEY (inviter_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (invitee_id) REFERENCES users(id) ON DELETE CASCADE,

  UNIQUE (group_id, invitee_id)
);
