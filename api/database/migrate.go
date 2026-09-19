package database

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var createLinksSQL string

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, createLinksSQL); err != nil {
		return fmt.Errorf("failed to run db migrations: %w", err)
	}
	return nil
}
