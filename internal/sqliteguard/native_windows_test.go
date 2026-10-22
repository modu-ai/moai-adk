package sqliteguard

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"modernc.org/sqlite"
)

func nativePath(t *testing.T) (string, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "private-namespace")
	if err := EnsureDirectory(dir); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "atomic.db")
	p := filepath.ToSlash(path)
	if p[0] != '/' {
		p = "/" + p
	}
	return path, (&url.URL{Scheme: "file", Path: p, RawQuery: url.Values{"_pragma": {"journal_mode(WAL)", "busy_timeout(5000)"}}.Encode()}).String()
}

func nativeObjects(t *testing.T, path string, before map[string]os.FileInfo) map[string]os.FileInfo {
	t.Helper()
	out := map[string]os.FileInfo{}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		name := path + suffix
		if err := privatefile.Check(name); err != nil {
			t.Errorf("strict native policy suffix=%q: %v", suffix, err)
		}
		// Windows os.Stat can resolve its file ID lazily at SameFile time.
		// Capture from an open handle now, before any Close/reopen phase.
		f, err := os.Open(name)
		if err != nil {
			t.Errorf("object suffix=%q: %v", suffix, err)
			continue
		}
		info, statErr := f.Stat()
		closeErr := f.Close()
		if statErr != nil || closeErr != nil {
			t.Errorf("capture native object suffix=%q: stat=%v close=%v", suffix, statErr, closeErr)
			continue
		}
		if old := before[suffix]; old != nil && !os.SameFile(old, info) {
			t.Errorf("native object replaced suffix=%q", suffix)
		}
		out[suffix] = info
	}
	return out
}

func TestNativeAtomicPinsPhysicalReopenAndDirectClose(t *testing.T) {
	const closeBudget = 200 * time.Millisecond
	path, dsn := nativePath(t)
	var before map[string]os.FileInfo
	for phase := 1; phase <= 2; phase++ {
		db, err := Open(path, dsn)
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(1)
		t.Cleanup(func() {
			if err := db.Close(); err != nil {
				t.Errorf("cleanup DB: %v", err)
			}
		})
		if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS n(i INTEGER); INSERT INTO n VALUES(1)`); err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM n`).Scan(&n); err != nil || n != phase {
			t.Fatalf("phase=%d n=%d error=%v", phase, n, err)
		}
		before = nativeObjects(t, path, before)
		conn, err := db.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if err := conn.Raw(func(raw any) error {
			mode, err := raw.(sqlite.FileControl).FileControlPersistWAL("main", -1)
			if err == nil && mode != 1 {
				return errors.New("physical PERSIST_WAL readback mismatch")
			}
			return err
		}); err != nil {
			t.Fatal(err)
		}
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		err = db.Close()
		elapsed := time.Since(start)
		t.Logf("native production Close phase=%d elapsed=%v error=%v", phase, elapsed, err)
		if err != nil {
			t.Fatal(err)
		}
		if elapsed > closeBudget {
			t.Errorf("native Close=%v exceeds %v", elapsed, closeBudget)
		}
		nativeObjects(t, path, before)
	}
}

func TestNativeBadConnReappliesPrivacyAndPersistWAL(t *testing.T) {
	path, dsn := nativePath(t)
	db, err := Open(path, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Exec(`CREATE TABLE n(i INTEGER); INSERT INTO n VALUES(1)`); err != nil {
		t.Fatal(err)
	}
	before := nativeObjects(t, path, nil)
	conn, err := db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Raw(func(any) error { return driver.ErrBadConn }); !errors.Is(err, driver.ErrBadConn) {
		t.Fatalf("Raw=%v", err)
	}
	if err := conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM n`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("n=%d error=%v", n, err)
	}
	nativeObjects(t, path, before)
}

func TestNativeBindingRefusesDifferentFileAndModeRO(t *testing.T) {
	path, dsn := nativePath(t)
	if db, err := Open(path+"-other", dsn); err == nil {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		t.Fatal("different URI target admitted")
	}
	if db, err := Open(path, dsn+"&mode=ro"); err == nil {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		t.Fatal("read-only URI entered mutating guard path")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("binding refusal created database: %v", err)
	}
}
