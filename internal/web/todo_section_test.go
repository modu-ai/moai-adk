package web

// todo_section_test.go — SPEC-WEB-TODO-QUEUE-001 M3: the read-only backlog
// section the /todo route serves.
//
// Resolved decision G-5: the section is the AUDIT view — all three states, each
// with a state badge — because a queued-only list cannot answer "where did card
// X go". The badge is half of that decision, so it is asserted rather than
// assumed.
//
// Every test here stubs kanban.HomeDirFn, which is process-global, so none run
// in parallel.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// stubTodoHome points the queue resolution's home seam at a throwaway
// directory, so no test reads or resolves against the developer's real
// ~/.moai/todo.
func stubTodoHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	orig := kanban.HomeDirFn
	kanban.HomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { kanban.HomeDirFn = orig })
	return home
}

// writeBacklog writes raw bytes to root's backlog file, creating the directory.
func writeBacklog(t *testing.T, root, body string) string {
	t.Helper()
	path := kanban.BacklogPathForRoot(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir queue dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write backlog: %v", err)
	}
	return path
}

// threeStateQueue is the AC-WTQ-003 fixture: one queued item with no SPEC id,
// one picked item with a SPEC id, one dropped item.
const threeStateQueue = `{"version":1,"last_seq":3,"items":[` +
	`{"id":"t1","text":"queued card","added_at":"2026-08-20T00:00:00Z","spec_id":null,"state":"queued"},` +
	`{"id":"t2","text":"picked card","added_at":"2026-08-20T00:01:00Z","spec_id":"SPEC-EXAMPLE-001","state":"picked"},` +
	`{"id":"t3","text":"dropped card","added_at":"2026-08-20T00:02:00Z","spec_id":null,"state":"dropped"}]}`

// todoBodyFor renders GET /todo for a console served from projectRoot.
func todoBodyFor(t *testing.T, projectRoot string) string {
	t.Helper()
	a := newApp(Config{ProjectRoot: projectRoot, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	req := httptest.NewRequest(http.MethodGet, "/todo", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /todo status = %d, want 200\nbody:\n%s", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// TestTodoSectionListsAllThreeStates — AC-WTQ-003: no item is filtered out,
// each row carries its identifier, its text and a state badge whose text is
// that item's state, and the picked row carries its SPEC id.
func TestTodoSectionListsAllThreeStates(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	writeBacklog(t, root, threeStateQueue)

	body := todoBodyFor(t, root)

	for _, want := range []string{
		"t1", "queued card", "t2", "picked card", "t3", "dropped card", "SPEC-EXAMPLE-001",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered section is missing %q", want)
		}
	}
	for _, state := range []string{"queued", "picked", "dropped"} {
		if !strings.Contains(body, `data-todo-state="`+state+`"`) {
			t.Errorf("no state badge for %q", state)
		}
		if !strings.Contains(body, `data-todo-state="`+state+`">`+state+`<`) {
			t.Errorf("the %q badge does not render its state as its text", state)
		}
	}
	if n := strings.Count(body, `data-todo-row`); n != 3 {
		t.Fatalf("section renders %d rows, want 3 (none filtered out)", n)
	}
	if strings.Contains(body, "data-todo-empty") {
		t.Errorf("a populated queue rendered the empty state")
	}
}

// TestTodoSectionEmptyStates — AC-WTQ-009: an absent, empty or malformed queue
// file renders the empty state at 200, never an error response.
func TestTodoSectionEmptyStates(t *testing.T) {
	cases := []struct {
		name string
		body *string
	}{
		{"absent file", nil},
		{"empty file", strPtr("")},
		{"malformed JSON", strPtr(`{"version":1,"items":[{"id":`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubTodoHome(t)
			root := t.TempDir()
			if tc.body != nil {
				writeBacklog(t, root, *tc.body)
			}

			body := todoBodyFor(t, root)

			if !strings.Contains(body, "data-todo-empty") {
				t.Errorf("no empty-state marker for %s", tc.name)
			}
			if strings.Contains(body, "data-todo-row") {
				t.Errorf("%s rendered rows", tc.name)
			}
		})
	}
}

func strPtr(s string) *string { return &s }

// TestTodoSectionCarriesExistingKanbanMarker — AC-WTQ-010 first half: the
// section sits inside an element carrying the EXISTING data-live="kanban"
// attribute that refresh() keys on. No new event name is introduced.
func TestTodoSectionCarriesExistingKanbanMarker(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	writeBacklog(t, root, threeStateQueue)

	body := todoBodyFor(t, root)

	marker := strings.Index(body, `data-live="kanban"`)
	if marker < 0 {
		t.Fatal("the todo section carries no data-live=\"kanban\" marker")
	}
	if row := strings.Index(body, "data-todo-row"); row < marker {
		t.Fatalf("a todo row (at %d) sits outside the data-live marker (at %d)", row, marker)
	}
}

// TestTodoSectionFromWorktreeReadsPrimaryQueue — AC-WTQ-005: a console served
// from a linked worktree lists the PRIMARY checkout's cards, not zero. This is
// the divergence the shared resolution exists to prevent.
func TestTodoSectionFromWorktreeReadsPrimaryQueue(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	stubTodoHome(t)
	primary := t.TempDir()
	gitInit(t, primary)
	wt := filepath.Join(t.TempDir(), "wt")
	runGit(t, primary, "worktree", "add", wt)
	writeBacklog(t, primary, threeStateQueue)

	body := todoBodyFor(t, wt)

	if n := strings.Count(body, "data-todo-row"); n != 3 {
		t.Fatalf("console served from the worktree renders %d rows, want the primary's 3", n)
	}
}

// TestTodoSectionReadsThroughToProjectLocalQueue — AC-WTQ-007: under the exact
// preconditions AC-WTQ-006 describes, the section lists the project-local
// queue's N items rather than an empty queue, and the disk stays untouched.
//
// Both halves are asserted together: AC-WTQ-006 pins that nothing was written,
// this pins what is rendered WHILE nothing was written. Asserting only the
// first leaves the divergence shippable.
func TestTodoSectionReadsThroughToProjectLocalQueue(t *testing.T) {
	home := stubTodoHome(t)
	root := t.TempDir() // deliberately NOT a git repository
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

// TestConsoleRoutesLeaveBacklogUntouched — AC-WTQ-001 runtime half: with a
// backlog file and its lock file present, exercising every console route in one
// run leaves the backlog's bytes identical and the lock file's modification
// time unchanged.
func TestConsoleRoutesLeaveBacklogUntouched(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	path := writeBacklog(t, root, threeStateQueue)
	lock := filepath.Join(filepath.Dir(path), "backlog.lock")
	if err := os.WriteFile(lock, []byte("held"), 0o600); err != nil {
		t.Fatalf("write lock file: %v", err)
	}
	beforeBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read backlog: %v", err)
	}
	beforeLock, err := os.Stat(lock)
	if err != nil {
		t.Fatalf("stat lock: %v", err)
	}

	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	h := a.routes()
	time.Sleep(10 * time.Millisecond)
	for _, p := range []string{"/", "/kanban", "/specs", "/monitor", "/settings", "/todo"} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}

	afterBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("backlog file gone after exercising the routes: %v", err)
	}
	if string(afterBytes) != string(beforeBytes) {
		t.Errorf("the console changed the backlog file's bytes")
	}
	afterLock, err := os.Stat(lock)
	if err != nil {
		t.Fatalf("lock file gone after exercising the routes: %v", err)
	}
	if !afterLock.ModTime().Equal(beforeLock.ModTime()) {
		t.Errorf("the console touched the lock file: %v -> %v", beforeLock.ModTime(), afterLock.ModTime())
	}
}

// gitInit turns dir into a committed repository. User- and system-level git
// config are pointed at /dev/null so a host's global config cannot leak in.
func gitInit(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "t")
	runGit(t, dir, "commit", "--allow-empty", "-q", "-m", "init")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", full, err, out)
	}
}
