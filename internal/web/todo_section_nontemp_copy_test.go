package web

// todo_section_nontemp_copy_test.go — SPEC-TODO-HOME-TEMP-GUARD-001 M2, the
// §C.1 C-row disposition for internal/web.
//
// TestTodoSectionReadsThroughToProjectLocalQueue keeps PASSING after the
// temporary-origin guard lands but stops exercising what it was written for:
// its root is a t.TempDir(), so the guard answers before the home-fallback
// branch is reached, and both of its load-bearing assertions survive for the
// wrong reason — the rendered rows come from the project-local queue either
// way, and "the fallback root was not created" is an ABSENCE the guard leaves
// true. A vacuously-passing test does not fail, so no non-destructive check
// can see this.
//
// The original stays where it is (still a valid regression guard for behaviour
// under the guard). This copy declares its root non-temporary through the
// REQ-THG-009 seam, which is exported precisely so a consuming package's tests
// can reach it, and adds the positive assertion that the seam was heard.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestTodoSectionReadsThroughToProjectLocalQueue_NonTemp — the C-row copy.
func TestTodoSectionReadsThroughToProjectLocalQueue_NonTemp(t *testing.T) {
	home := stubTodoHome(t)
	root := t.TempDir() // deliberately NOT a git repository

	origRoots := kanban.TempRootsFn
	isolated := filepath.Join(t.TempDir(), "a-root-that-contains-nothing")
	kanban.TempRootsFn = func() []string { return []string{isolated} }
	t.Cleanup(func() { kanban.TempRootsFn = origRoots })

	// The positive assertion: the injected root set was actually read, so this
	// fixture really is on the home-fallback path rather than on the guard's
	// refusal path. Remove the stub and this line goes RED.
	if reason, isTemp := kanban.TempOriginReason(root); isTemp {
		t.Fatalf("the injected temp-root set was not read: root %q still classifies temporary (reason %q)", root, reason)
	}

	local := writeBacklog(t, root, threeStateQueue)
	before, err := os.Stat(local)
	if err != nil {
		t.Fatalf("stat local queue: %v", err)
	}
	fallbackRoot := filepath.Join(home, ".moai", "todo", kanban.TodoQueueProjectKey(root))
	time.Sleep(10 * time.Millisecond)

	body := todoBodyFor(t, root)

	if n := strings.Count(body, "data-todo-row"); n != 3 {
		t.Fatalf("read-through renders %d rows, want 3", n)
	}
	after, err := os.Stat(local)
	if err != nil {
		t.Fatalf("the console moved or removed the local queue: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("the console changed the local queue's mtime: %v -> %v", before.ModTime(), after.ModTime())
	}
	if _, err := os.Stat(fallbackRoot); !os.IsNotExist(err) {
		t.Errorf("the console created the fallback root %q (stat err = %v)", fallbackRoot, err)
	}
}
