package discovery

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestReadOnlySQLiteURIRetainsDriveAndEscapedPath(t *testing.T) {
	for _, path := range []string{"C:/Users/worker/broker.db", "C:/Users/space user/hash#percent%question?.db", "/tmp/space user/hash#percent%question?.db"} {
		t.Run(path, func(t *testing.T) {
			u, err := url.Parse(readOnlySQLiteURI(path))
			if err != nil {
				t.Fatal(err)
			}
			want := path
			if want[0] != '/' {
				want = "/" + want
			}
			if u.Scheme != "file" || u.Host != "" || u.Path != want {
				t.Fatalf("read-only SQLite URI changes the local path: scheme=%q host=%q path=%q want=%q", u.Scheme, u.Host, u.Path, want)
			}
			if u.Query().Get("mode") != "ro" || u.Query().Get("_pragma") != "busy_timeout(100)" || len(u.Query()) != 2 {
				t.Fatalf("read-only enumeration query policy changed: %v", u.Query())
			}
		})
	}
}

func TestOpenSQLiteReadOnlyNeverCreatesMissingBroker(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent broker.db")
	db, err := openSQLiteReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close read-only broker: %v", err)
		}
	})
	if err := db.PingContext(context.Background()); err == nil {
		t.Fatal("a missing broker was opened successfully")
	}
	for _, suffix := range []string{"", "-wal", "-shm", "-journal"} {
		if _, err := os.Lstat(path + suffix); !os.IsNotExist(err) {
			t.Errorf("read-only enumeration created a broker artifact %q: %v", path+suffix, err)
		}
	}
}
