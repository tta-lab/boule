-- name: SendMessage :one
INSERT INTO messages (id, sender, recipient, content)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetInbox :many
SELECT * FROM messages
WHERE recipient = ? AND read = 0
ORDER BY created_at DESC;

-- name: GetFeed :many
SELECT * FROM messages
WHERE (? = '' OR sender = ?)
  AND (? = '' OR recipient = ?)
  AND (? = 0 OR read = 0)
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: MarkRead :exec
UPDATE messages
SET read = 1
WHERE id = ?;

-- name: GetMessageByID :one
SELECT * FROM messages
WHERE id = ?;
