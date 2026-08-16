package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/d3cie/dottie/internal/infra/sqlite/db"
	"github.com/d3cie/dottie/migrations"
	_ "modernc.org/sqlite"
)

type Client struct {
	SQL     *sql.DB
	Queries *db.Queries
}

func Open(ctx context.Context, path string) (*Client, error) {
	database, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	database.SetMaxOpenConns(1)
	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := migrate(ctx, database); err != nil {
		_ = database.Close()
		return nil, err
	}
	return &Client{SQL: database, Queries: db.New(database)}, nil
}

func (c *Client) Close() error {
	return c.SQL.Close()
}

func migrate(ctx context.Context, database *sql.DB) error {
	if _, err := database.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return fmt.Errorf("create sqlite migration table: %w", err)
	}
	entries, err := fs.Glob(migrations.Files, "sqlite/*.sql")
	if err != nil {
		return fmt.Errorf("list sqlite migrations: %w", err)
	}
	sort.Strings(entries)
	for _, name := range entries {
		version := strings.TrimSuffix(strings.TrimPrefix(name, "sqlite/"), ".sql")
		var count int
		if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&count); err != nil {
			return fmt.Errorf("inspect sqlite migration %s: %w", version, err)
		}
		if count > 0 {
			continue
		}
		contents, err := migrations.Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read sqlite migration %s: %w", version, err)
		}
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin sqlite migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply sqlite migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record sqlite migration %s: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit sqlite migration %s: %w", version, err)
		}
	}
	return nil
}
