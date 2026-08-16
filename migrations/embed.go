package migrations

import "embed"

// Files contains every database migration shipped with the executable.
//
//go:embed sqlite/*.sql duckdb/*.sql
var Files embed.FS
