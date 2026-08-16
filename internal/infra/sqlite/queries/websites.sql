-- name: CreateWebsite :one
INSERT INTO websites (id, name, domain, timezone, created_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetWebsite :one
SELECT * FROM websites WHERE id = ? LIMIT 1;

-- name: ListWebsites :many
SELECT * FROM websites ORDER BY created_at ASC;

