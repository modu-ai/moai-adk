package factorymsg

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestBrokerOpenWithSpentContextDoesNotInitialize(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	for _, canceled := range []bool{true, false} {
		t.Run(map[bool]string{true: "canceled", false: "expired"}[canceled], func(t *testing.T) {
			root := t.TempDir()
			var ctx context.Context
			var cancel context.CancelFunc
			want := context.DeadlineExceeded
			if canceled {
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
				want = context.Canceled
			} else {
				ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			}
			defer cancel()
			path, err := BrokerPath(root, "run-spent")
			if err != nil {
				t.Fatal(err)
			}
			store, err := OpenWithContext(ctx, root, "run-spent")
			if store != nil {
				_ = store.Close()
				t.Fatal("spent context returned a broker handle")
			}
			if !errors.Is(err, want) {
				t.Fatalf("error=%v, want %v", err, want)
			}
			if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("spent context initialized broker: %v", err)
			}
		})
	}
}

func TestBrokerOpenContextPreservesConnectionSettings(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	store, err := OpenWithContext(context.Background(), t.TempDir(), "run-settings")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	var mode string
	var busy int
	if err := store.db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if err := store.db.QueryRow("PRAGMA busy_timeout").Scan(&busy); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" || busy != 2500 {
		t.Fatalf("connection settings changed: journal=%s busy=%d", mode, busy)
	}
}
