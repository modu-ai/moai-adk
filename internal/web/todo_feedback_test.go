package web

import (
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

func watchSeededTodo(t *testing.T) (string, *factory.BacklogStore, <-chan string) {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	store := factory.NewBacklogStore(factory.BacklogPathForRoot(root))
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
	// Single-flight probe: write one probe, then wait for its outcome — the
	// "config" event, or one debounce window plus margin with nothing — before
	// the next write. A second write in flight could fire after the first event
	// and leak into the caller's quiet assertions. After the first event, drain
	// until one debounce window passes with nothing, so the tree is handed over
	// quiet.
	timeout := time.After(8 * time.Second)
	settle := debounce + 250*time.Millisecond
	for n := 0; ; n++ {
		if err := os.WriteFile(filepath.Join(control, "probe.yaml"), []byte("x: "+strconv.Itoa(n)), 0600); err != nil {
			t.Fatal(err)
		}
		attempt := time.After(settle)
	wait:
		for {
			select {
			case ev := <-ch:
				if ev != "config" {
					continue
				}
				for drained := false; !drained; {
					select {
					case <-ch:
					case <-time.After(settle):
						drained = true
					case <-timeout:
						t.Fatal("existing-directory control did not go quiet")
					}
				}
				return root, store, ch
			case <-attempt:
				break wait
			case <-timeout:
				t.Fatal("existing-directory control did not deliver")
			}
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
				if ev != "factory" {
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
				if ev != "factory" {
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

// TestWatchSeededTodoReturnsQuiet: a helper that hands over a watched tree must
// hand it over quiet. The probe writes and the 250ms debounce overlap, so a probe
// written just before the first "config" event can fire a second one right after
// the helper returned — and the next assertion ("idle control emitted nothing")
// reads it as a spurious refresh. The helper drains until one debounce window
// passes with nothing.
func TestWatchSeededTodoReturnsQuiet(t *testing.T) {
	_, _, events := watchSeededTodo(t)
	select {
	case ev := <-events:
		t.Fatalf("the helper handed over a tree that was not quiet: %s", ev)
	case <-time.After(debounce + 150*time.Millisecond):
	}
}
