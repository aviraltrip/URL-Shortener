-- name: CreateLink :one
INSERT INTO links (
    id,
    short_code,
    original_url,
    expires_at
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLinkByShortCode :one
SELECT *
FROM links
WHERE short_code = $1
  AND expires_at > NOW();

-- name: GetLinkDetails :one
SELECT *
FROM links
WHERE short_code = $1;

-- name: IncrementClickCount :exec
UPDATE links
SET click_count = click_count + 1
WHERE short_code = $1;

-- name: DeleteLink :exec
DELETE FROM links
WHERE short_code = $1;

-- name: ListLinks :many
SELECT *
FROM links
WHERE expires_at > NOW()
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
