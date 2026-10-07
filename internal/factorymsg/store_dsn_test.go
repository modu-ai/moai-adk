package factorymsg

import (
	"database/sql"
	"net/url"
	"path/filepath"
	"testing"
	"time"
)

func TestBrokerDSNPathAndDriver(t *testing.T) {
	for _, path := range []string{"/tmp/with ?#%/broker.db", "C:/with ?#%/broker.db", "//server/share/broker.db"} {
		t.Run(path, func(t *testing.T) {
			query := url.Values{"mode": {"memory"}, "_pragma": {"busy_timeout(100)"}, "_txlock": {"immediate"}}
			dsn := brokerDSN(filepath.FromSlash(path), query)
			u, err := url.Parse(dsn)
			if err != nil {
				t.Fatal(err)
			}
			wantPath := path
			if wantPath[0] != '/' {
				wantPath = "/" + wantPath
			}
			if u.Host != "" || u.Path != wantPath {
				t.Fatalf("DSN lost native file path: %q host=%q path=%q want=%q", dsn, u.Host, u.Path, wantPath)
			}
			if u.Query().Encode() != query.Encode() {
				t.Fatalf("query settings changed: %q", u.RawQuery)
			}
			db, err := sql.Open("sqlite", dsn)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			if err := db.Ping(); err != nil {
				t.Fatalf("SQLite rejected DSN %q: %v", dsn, err)
			}
		})
	}
}

func TestBrokerOpenAndReopenEncodedHome(t *testing.T) {
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home with #percent%"))
	root := t.TempDir()
	store, err := OpenWithDeadline(root, "run", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	var mode string
	if err := store.db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		_ = store.Close()
		t.Fatal(err)
	}
	if mode != "wal" {
		_ = store.Close()
		t.Fatalf("journal mode = %q", mode)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	existing, err := OpenExistingWithDeadline(root, "run", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = existing.Close() })
	var timeout int
	if err := existing.db.QueryRow("PRAGMA busy_timeout").Scan(&timeout); err != nil {
		t.Fatal(err)
	}
	if timeout != 500 {
		t.Fatalf("existing broker busy_timeout=%d, want500", timeout)
	}
}
