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
				defer db.Close()
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
