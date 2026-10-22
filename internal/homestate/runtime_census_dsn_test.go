package homestate

import (
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeCensusFactoryDSNUsesSQLiteFileURI(t *testing.T) {
	for _, path := range []string{"/tmp/with ?#%/factory.db", "C:/with ?#%/factory.db", "//server/share/factory.db"} {
		t.Run(path, func(t *testing.T) {
			dsn := runtimeCensusFactoryDSN(filepath.FromSlash(path))
			u, err := url.Parse(dsn)
			if err != nil {
				t.Fatal(err)
			}
			if u.Query().Get("mode") != "ro" {
				t.Fatalf("census must keep a read-only connection: %q", dsn)
			}
			// Memory mode exercises the actual driver's URI parser on every
			// host, without creating a foreign-platform filesystem path.
			query := u.Query()
			query.Set("mode", "memory")
			u.RawQuery = query.Encode()
			db, err := sql.Open("sqlite", u.String())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			if err := db.Ping(); err != nil {
				t.Fatalf("SQLite rejected census URI %q: %v", dsn, err)
			}
			want := path
			if !strings.HasPrefix(want, "/") {
				want = "/" + want
			}
			if u.Host != "" || u.Path != want {
				t.Fatalf("census URI changed filesystem path: %q, want path %q without authority", dsn, want)
			}
		})
	}
}

func TestRuntimeCensusReadsNativeFactoryPathWithReservedURIBytes(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home #%"))
	store, err := OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES('live',?,'now','now')`, os.Getpid()); err != nil {
		_ = store.Close()
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	census, err := ReadRuntimeCensus(root)
	if err != nil || census.ActiveFactoryLanes != 1 {
		t.Fatalf("native factory census = %+v, err=%v", census, err)
	}
}
