//go:build !windows

package sqliteguard

import "database/sql"

// Open preserves the existing Unix SQLite connection behavior. Native Windows
// creation/sharing controls are applied only by the Windows implementation.
func Open(_ string, dsn string) (*sql.DB, error) { return sql.Open("sqlite", dsn) }
