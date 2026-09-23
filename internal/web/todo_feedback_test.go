package web

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

func watchSeededTodo(t *testing.T) (string, *kanban.BacklogStore, <-chan string) {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root))
	if _, _, err := store.Add("seed"); err != nil {
		t.Fatal(err)
	}
	control := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(control, 0700); err != nil {
		t.Fatal(err)
	}
	h := NewHub()
	ch := h.add()
	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() { done <- h.Watch(root, stop) }()
	t.Cleanup(func() {
		close(stop)
		if err := <-done; err != nil {
			t.Error(err)
		}
		h.remove(ch)
	})
	ticker := time.NewTicker(350 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(8 * time.Second)
	for {
		select {
		case <-ticker.C:
			if err := os.WriteFile(filepath.Join(control, "probe.yaml"), []byte("x: y"), 0600); err != nil {
				t.Fatal(err)
			}
		case ev := <-ch:
			if ev == "config" {
				return root, store, ch
			}
		case <-timeout:
			t.Fatal("existing-directory control did not deliver")
		}
	}
}

func TestTodoReadsDoNotRefreshThemselves(t *testing.T) {
	root, _, events := watchSeededTodo(t)
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	select {
	case ev := <-events:
		t.Fatalf("idle control emitted %s", ev)
	case <-time.After(600 * time.Millisecond):
	}
	for i := 0; i < 3; i++ {
		response := serveGet(t, a.routes(), "/todo")
		if response.Code != 200 {
			t.Fatalf("HTTP status=%d", response.Code)
		}
		select {
		case ev := <-events:
			t.Fatalf("read %d triggered %s without any queue write", i+1, ev)
		case <-time.After(600 * time.Millisecond):
		}
	}
}

func TestTodoCommitsStillRefresh(t *testing.T) {
	for _, held := range []bool{false, true} {
		name := "closed writer"
		if held {
			name = "held WAL writer"
		}
		t.Run(name, func(t *testing.T) {
			_, store, events := watchSeededTodo(t)
			if held {
				db, err := sql.Open("sqlite", store.EnginePath())
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if cerr := db.Close(); cerr != nil {
						t.Error(cerr)
					}
				}()
				db.SetMaxOpenConns(1)
				if _, err := db.Exec("PRAGMA wal_autocheckpoint=0"); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec("UPDATE items SET text='committed change' WHERE id='t1'"); err != nil {
					t.Fatal(err)
				}
				if info, err := os.Stat(store.EnginePath() + "-wal"); err != nil || info.Size() == 0 {
					t.Fatalf("held writer has no populated WAL: info=%v err=%v", info, err)
				}
			} else if _, _, err := store.Add("committed change"); err != nil {
				t.Fatal(err)
			}
			select {
			case ev := <-events:
				if ev != "kanban" {
					t.Fatalf("commit event=%s", ev)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("committed queue change did not refresh")
			}
		})
	}
}

// A WAL renamed into the watched directory produces Create and never Write,
// because its bytes were written before any watch on the file could exist.
// That is the event shape a held writer produces when it loses the race
// against the watch registration, made deterministic.
func TestTodoCreatedWALRefreshesOnlyWhenPopulated(t *testing.T) {
	for _, tc := range []struct {
		name string
		size int
		want bool
	}{
		{"populated WAL refreshes", 4096, true},
		{"empty reader WAL stays quiet", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, store, events := watchSeededTodo(t)
			wal := store.EnginePath() + "-wal"
			if _, err := os.Stat(wal); err == nil {
				t.Fatalf("precondition: WAL already exists at %s", wal)
			}
			staged := filepath.Join(t.TempDir(), "staged-wal")
			if err := os.WriteFile(staged, make([]byte, tc.size), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(staged, wal); err != nil {
				t.Fatal(err)
			}
			wait := 3 * time.Second
			if !tc.want {
				wait = 1200 * time.Millisecond
			}
			select {
			case ev := <-events:
				if !tc.want {
					t.Fatalf("empty WAL creation emitted %s", ev)
				}
				if ev != "kanban" {
					t.Fatalf("created WAL event=%s", ev)
				}
			case <-time.After(wait):
				if tc.want {
					t.Fatal("created populated WAL did not refresh")
				}
			}
		})
	}
}
