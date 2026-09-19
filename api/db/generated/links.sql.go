package generated

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createLink = `-- name: CreateLink :one
INSERT INTO links (
    id,
    short_code,
    original_url,
    expires_at
)
VALUES ($1, $2, $3, $4)
RETURNING id, short_code, original_url, expires_at, created_at, click_count
`

type CreateLinkParams struct {
	ID          uuid.UUID `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalUrl string    `json:"original_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (q *Queries) CreateLink(ctx context.Context, arg CreateLinkParams) (Link, error) {
	row := q.db.QueryRow(ctx, createLink,
		arg.ID,
		arg.ShortCode,
		arg.OriginalUrl,
		arg.ExpiresAt,
	)
	var i Link
	err := row.Scan(
		&i.ID,
		&i.ShortCode,
		&i.OriginalUrl,
		&i.ExpiresAt,
		&i.CreatedAt,
		&i.ClickCount,
	)
	return i, err
}

const deleteLink = `-- name: DeleteLink :exec
DELETE FROM links
WHERE short_code = $1
`

func (q *Queries) DeleteLink(ctx context.Context, shortCode string) error {
	_, err := q.db.Exec(ctx, deleteLink, shortCode)
	return err
}

const getLinkByShortCode = `-- name: GetLinkByShortCode :one
SELECT id, short_code, original_url, expires_at, created_at, click_count
FROM links
WHERE short_code = $1
  AND expires_at > NOW()
`

func (q *Queries) GetLinkByShortCode(ctx context.Context, shortCode string) (Link, error) {
	row := q.db.QueryRow(ctx, getLinkByShortCode, shortCode)
	var i Link
	err := row.Scan(
		&i.ID,
		&i.ShortCode,
		&i.OriginalUrl,
		&i.ExpiresAt,
		&i.CreatedAt,
		&i.ClickCount,
	)
	return i, err
}

const getLinkDetails = `-- name: GetLinkDetails :one
SELECT id, short_code, original_url, expires_at, created_at, click_count
FROM links
WHERE short_code = $1
`

func (q *Queries) GetLinkDetails(ctx context.Context, shortCode string) (Link, error) {
	row := q.db.QueryRow(ctx, getLinkDetails, shortCode)
	var i Link
	err := row.Scan(
		&i.ID,
		&i.ShortCode,
		&i.OriginalUrl,
		&i.ExpiresAt,
		&i.CreatedAt,
		&i.ClickCount,
	)
	return i, err
}

const incrementClickCount = `-- name: IncrementClickCount :exec
UPDATE links
SET click_count = click_count + 1
WHERE short_code = $1
`

func (q *Queries) IncrementClickCount(ctx context.Context, shortCode string) error {
	_, err := q.db.Exec(ctx, incrementClickCount, shortCode)
	return err
}

const listLinks = `-- name: ListLinks :many
SELECT id, short_code, original_url, expires_at, created_at, click_count
FROM links
WHERE expires_at > NOW()
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`

type ListLinksParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListLinks(ctx context.Context, arg ListLinksParams) ([]Link, error) {
	rows, err := q.db.Query(ctx, listLinks, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Link
	for rows.Next() {
		var i Link
		if err := rows.Scan(
			&i.ID,
			&i.ShortCode,
			&i.OriginalUrl,
			&i.ExpiresAt,
			&i.CreatedAt,
			&i.ClickCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
