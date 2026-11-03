package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
)

// OpenReadOnly opens an existing ledger for reading only. It never creates the file or its directory, never runs the
// schema DDL, never migrates, and refuses a file whose schema version is not SchemaVersion (older or newer): a reader
// that cannot understand the file must not rewrite it. The connection is mode=ro with query_only, so a write through
// the returned Store fails at the engine.
func OpenReadOnly(ctx context.Context, path string) (*Store, error) {
	q := url.Values{}
	q.Set("mode", "ro")
	q.Add("_pragma", "query_only(ON)")
	q.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", BusyTimeout.Milliseconds()))
	p := filepath.ToSlash(path)
	if len(p) == 0 || p[0] != '/' {
		p = "/" + p
	}
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: p, RawQuery: q.Encode()}).String())
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	var have int
	if err := db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&have); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ledger: %w", err)
	}
	if have != SchemaVersion {
		_ = db.Close()
		return nil, fmt.Errorf("ledger: the file has schema %d, this reader knows %d", have, SchemaVersion)
	}
	return &Store{db: db}, nil
}
