package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/d3cie/dottie/migrations"
	_ "github.com/duckdb/duckdb-go/v2"
)

type Client struct {
	SQL *sql.DB
}

func Open(ctx context.Context, path string) (*Client, error) {
	database, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	database.SetMaxOpenConns(1)
	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping duckdb: %w", err)
	}
	if err := migrate(ctx, database); err != nil {
		_ = database.Close()
		return nil, err
	}
	return &Client{SQL: database}, nil
}

func (c *Client) Close() error {
	return c.SQL.Close()
}

func migrate(ctx context.Context, database *sql.DB) error {
	if _, err := database.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return fmt.Errorf("create duckdb migration table: %w", err)
	}
	entries, err := fs.Glob(migrations.Files, "duckdb/*.sql")
	if err != nil {
		return fmt.Errorf("list duckdb migrations: %w", err)
	}
	sort.Strings(entries)
	for _, name := range entries {
		version := strings.TrimSuffix(strings.TrimPrefix(name, "duckdb/"), ".sql")
		var count int
		if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&count); err != nil {
			return fmt.Errorf("inspect duckdb migration %s: %w", version, err)
		}
		if count > 0 {
			continue
		}
		contents, err := migrations.Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read duckdb migration %s: %w", version, err)
		}
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin duckdb migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply duckdb migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record duckdb migration %s: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit duckdb migration %s: %w", version, err)
		}
	}
	return nil
}
