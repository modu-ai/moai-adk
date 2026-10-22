package sqliteguard

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/url"
	"path/filepath"
	"reflect"
	"testing"

	"modernc.org/sqlite"
)

func realPool(t *testing.T) (*sql.DB, *manager) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "guard.db")
	p := filepath.ToSlash(path)
	if p[0] != '/' {
		p = "/" + p
	}
	dsn := (&url.URL{Scheme: "file", Path: p, RawQuery: url.Values{"_pragma": {"journal_mode(WAL)"}}.Encode()}).String()
	base, err := sqlite.NewConnector(dsn)
	if err != nil {
		t.Fatal(err)
	}
	m := newManager(4, func(Snapshot) {})
	db := sql.OpenDB(&connector{path: path, base: base, owners: m, prepare: func(_ string, acquired func(io.Closer)) error { acquired(&fakePin{}); return nil }})
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil && !errors.Is(err, ErrRetained) {
			t.Error(err)
		}
		// Only test-owned synthetic raw resources are explicitly finalized by
		// each test before this cleanup. Production exposes no force release.
		for _, o := range m.live {
			if o.conn != nil {
				if err := o.conn.Close(); err != nil {
					t.Error(err)
				}
			}
		}
	})
	return db, m
}

func TestSQLitePublicAPIAndPoolReconnection(t *testing.T) {
	db, m := realPool(t)
	if _, err := db.Exec(`CREATE TABLE t(n INTEGER); INSERT INTO t VALUES(1)`); err != nil {
		t.Fatal(err)
	}
	conn, err := db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Raw(func(raw any) error {
		api, ok := raw.(connectionAPI)
		if !ok {
			return errors.New("lost connection methods")
		}
		if _, ok := raw.(driver.NamedValueChecker); ok {
			return errors.New("added NamedValueChecker")
		}
		if _, err := api.ColumnInfo(`SELECT n FROM t`); err != nil {
			return err
		}
		if mode, err := api.FileControlPersistWAL("main", -1); err != nil || mode != 1 {
			return errors.New("physical persist setting missing")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	db.SetMaxIdleConns(0)
	for range 2 {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM t`).Scan(&n); err != nil || n != 1 {
			t.Fatalf("rows=%d error=%v", n, err)
		}
		if len(m.snapshots()) != 0 {
			t.Fatalf("idle physical lifetime not released: %+v", m.snapshots())
		}
	}
}

func TestRawRowsAndPreparedStatementAccounting(t *testing.T) {
	db, m := realPool(t)
	c, err := db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Raw(func(raw any) error {
		api := raw.(connectionAPI)
		s, err := api.Prepare(`SELECT 7 AS n`)
		if err != nil {
			return err
		}
		if _, ok := s.(statementAPI); !ok {
			return errors.New("lost statement context methods")
		}
		r, err := s.Query(nil)
		if err != nil {
			return err
		}
		if _, ok := r.(rowsAPI); !ok {
			return errors.New("lost rows metadata methods")
		}
		if got := reflect.TypeOf(r).NumMethod(); got != 8 {
			return errors.New("unexpected rows interface shape")
		}
		values := make([]driver.Value, 1)
		if err := r.Next(values); err != nil {
			return err
		}
		if values[0] != int64(7) {
			return errors.New("query value changed")
		}
		if err := r.Next(values); !errors.Is(err, io.EOF) {
			return errors.New("expected end of rows")
		}
		if m.snapshots()[0].Outstanding != 2 {
			return errors.New("EOF incorrectly released raw resource")
		}
		if err := r.Close(); err != nil {
			return err
		}
		if err := r.Close(); err != nil {
			return err
		}
		if m.snapshots()[0].Outstanding != 1 {
			return errors.New("prepared statement ownership lost")
		}
		return s.Close()
	}); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if len(m.snapshots()) != 0 {
		t.Fatalf("owners=%+v", m.snapshots())
	}
}
