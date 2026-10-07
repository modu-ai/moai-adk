package homestate

import (
	"database/sql"
	"net/url"
	"path/filepath"
	"testing"
)

func TestProfileLeaseDSNPathAndDriver(t *testing.T) {
	for _, path := range []string{"/tmp/with ?#%/profile.db", "C:/with ?#%/profile.db", "//server/share/profile.db"} {
		t.Run(path, func(t *testing.T) {
			query := url.Values{"mode": {"memory"}, "_pragma": {"busy_timeout(5000)"}, "_txlock": {"immediate"}}
			dsn := profileLeaseDSN(filepath.FromSlash(path), query)
			db, err := sql.Open("sqlite", dsn)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			if err := db.Ping(); err != nil {
				t.Fatalf("SQLite rejected DSN %q: %v", dsn, err)
			}
			u, err := url.Parse(dsn)
			if err != nil {
				t.Fatal(err)
			}
			wantPath := path
			if wantPath[0] != '/' {
				wantPath = "/" + wantPath
			}
			if u.Host != "" || u.Path != wantPath || u.Query().Encode() != query.Encode() {
				t.Fatalf("DSN changed file path or query: %q", dsn)
			}
		})
	}
}
