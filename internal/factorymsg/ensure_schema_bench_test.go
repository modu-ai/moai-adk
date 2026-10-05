package factorymsg

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Measures the cost ensureSchema adds to OpenExistingWithDeadline on an
// already-migrated broker (the hook hot path, bounded by a 200ms deadline).
//
//	go test ./internal/factorymsg -run '^$' -bench 'BenchmarkEnsureSchema|BenchmarkOpenExisting' -benchtime=2000x -count=5

func benchMigratedStore(b *testing.B) (*Store, string) {
	b.Helper()
	root := b.TempDir()
	s, err := Open(root, "bench")
	if err != nil {
		b.Fatalf("open: %v", err)
	}
	b.Cleanup(func() { _ = s.Close() })
	return s, root
}

// BenchmarkEnsureSchemaMigrated isolates the catalogue probe: the read-only
// path ensureSchema takes when both the dispatches table and the sender_slot
// column already exist.
func BenchmarkEnsureSchemaMigrated(b *testing.B) {
	s, _ := benchMigratedStore(b)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ensureSchema(ctx, s.db); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPeersCountReference is the query OpenExistingWithDeadline already
// ran before ensureSchema was added, on the same handle, as a same-run scale.
func BenchmarkPeersCountReference(b *testing.B) {
	s, _ := benchMigratedStore(b)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var n int
		if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM peers`).Scan(&n); err != nil {
			b.Fatal(err)
		}
	}
}

// openExistingWithoutProbe replicates OpenExistingWithDeadline's connection
// setup, peers probe, and Close — everything except ensureSchema — so the
// pair of open benchmarks isolates what the catalogue probe adds.
func openExistingWithoutProbe(b *testing.B, root string, deadline time.Duration) {
	path, err := BrokerPath(root, "bench")
	if err != nil {
		b.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		b.Fatal(err)
	}
	v := url.Values{}
	v.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", deadline.Milliseconds()/10))
	v.Add("_txlock", "immediate")
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: v.Encode()}).String())
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM peers`).Scan(&count); err != nil {
		b.Fatal(err)
	}
	if err := db.Close(); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkOpenExistingWithoutProbeReplica(b *testing.B) {
	_, root := benchMigratedStore(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		openExistingWithoutProbe(b, root, 200*time.Millisecond)
	}
}

// BenchmarkOpenExistingMigrated is the whole hot-path open (sql.Open, peers
// probe, ensureSchema) plus Close, on an already-migrated broker.
func BenchmarkOpenExistingMigrated(b *testing.B) {
	_, root := benchMigratedStore(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, err := OpenExistingWithDeadline(root, "bench", 200*time.Millisecond)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
