-- Canonical source: internal/store/migrations/001_create_messages.sql (embedded via go:embed)
-- This copy exists because sqlc reads schema from this directory.
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    content TEXT NOT NULL,
    read INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_recipient_read ON messages(recipient, read);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
