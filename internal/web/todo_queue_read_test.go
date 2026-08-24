package web

// todo_queue_read_test.go — the guard for the console's single backlog read
// seam. The backlog's storage is under review (JSON file today), so every call
// into the store is kept in ONE function; a migration then changes that
// function and nothing else. This test is what makes "one place" checkable
// rather than a claim in a comment.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// backlogStoreSymbols are the store-facing symbols the console must reach only
// through readTodoQueue.
var backlogStoreSymbols = []string{
	"kanban.ResolveTodoQueueRoot",
	"kanban.NewBacklogStore",
	"kanban.BacklogPathForRoot",
	"BacklogItem",
	"BacklogRecord",
}

// TestBacklogReadSeamIsSingleFile asserts that no production file in
// internal/web other than todo_queue_read.go names a backlog-store symbol.
func TestBacklogReadSeamIsSingleFile(t *testing.T) {
	const seam = "todo_queue_read.go"

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	seamHits := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, sym := range backlogStoreSymbols {
			if !strings.Contains(string(src), sym) {
				continue
			}
			if name == seam {
				seamHits++
				continue
			}
			t.Errorf("%s names the backlog-store symbol %q; the console reaches the store only through %s", name, sym, seam)
		}
	}
	if seamHits == 0 {
		t.Fatalf("%s names no backlog-store symbol — the read seam is not where this test says it is", seam)
	}
}
