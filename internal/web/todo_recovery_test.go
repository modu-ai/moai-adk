package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestTodoUnreadableDoesNotClaimEmpty(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	writeBacklog(t, root, `{"version":1,"items":[{"id":`)
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	for _, route := range []string{"/todo", "/"} {
		body := serveGet(t, a.routes(), route).Body.String()
		if !strings.Contains(body, "data-todo-unavailable") {
			t.Errorf("%s lacks unreadable diagnostic", route)
		}
		if strings.Contains(body, `data-i18n="todo.empty"`) || strings.Contains(body, "todo-summary__metric") {
			t.Errorf("%s claims empty/counts for unreadable queue", route)
		}
		if strings.Contains(body, "unexpected end of JSON") {
			t.Errorf("%s leaks raw error", route)
		}
	}
}

func TestTodoWatcherRegistersLateDirectories(t *testing.T) {
	for _, layout := range []string{"project-local", "home"} {
		t.Run(layout, func(t *testing.T) {
			t.Setenv("MOAI_HOME", t.TempDir())
			root := t.TempDir()
			control := filepath.Join(root, ".moai", "config", "sections")
			if err := os.MkdirAll(control, 0700); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(root, ".moai", "state", "todo")
			if layout == "home" {
				var err error
				target, err = homestate.TodoDir(root)
				if err != nil {
					t.Fatal(err)
				}
			}
			h := NewHub()
			ch := h.add()
			defer h.remove(ch)
			stop := make(chan struct{})
			done := make(chan error, 1)
			go func() { done <- h.Watch(root, stop) }()
			defer func() {
				close(stop)
				if err := <-done; err != nil {
					t.Error(err)
				}
			}()
			// Confirm a real watcher is running before testing the absent directory.
			ticker := time.NewTicker(350 * time.Millisecond)
			defer ticker.Stop()
			deadline := time.After(8 * time.Second)
		ready:
			for {
				select {
				case <-ticker.C:
					if err := os.WriteFile(filepath.Join(control, "probe.yaml"), []byte("x: y"), 0600); err != nil {
						t.Fatal(err)
					}
				case event := <-ch:
					if event == "config" {
						break ready
					}
				case <-deadline:
					t.Fatal("existing-directory control did not deliver")
				}
			}
			if err := os.MkdirAll(target, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(target, "backlog.json"), []byte(`{"version":1,"items":[]}`), 0600); err != nil {
				t.Fatal(err)
			}
			deadline = time.After(4 * time.Second)
			for {
				select {
				case event := <-ch:
					if event == "kanban" {
						return
					}
				case <-deadline:
					t.Fatal("late todo directory never refreshes")
				}
			}
		})
	}
}
