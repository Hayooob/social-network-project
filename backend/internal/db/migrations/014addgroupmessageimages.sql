-- Add image_path column to group_messages table.  Some environments may
-- already have this column (e.g. when the migration ran previously), so
-- the Go migration runner will quietly skip the error if it sees a
-- "duplicate column" message.  The extra CREATE/DROP statements from the
-- original version are no longer needed.
ALTER TABLE group_messages
ADD COLUMN image_path TEXT;