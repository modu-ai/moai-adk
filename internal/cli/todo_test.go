// todo_test.go — SPEC-KANBAN-TODO-CLI-001 M2 acceptance tests (CLI verbs).
//
// Covers the CLI-process-level ACs the store-level M1 suite defers:
// AC-TODO-001 (concurrent add loses nothing), AC-TODO-003 (list lock-free
// + --json), AC-TODO-004 (done by id under lock), AC-TODO-005 (bare next
// read-only), AC-TODO-006 (next <n> --spec one locked write), AC-TODO-014
// (no-prompt static guard), plus the acceptance.md §C edge cases.
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// runTodo executes the todo cobra command against args and returns
// (stdout, stderr, error). Exit-code semantics: a non-nil error is the
// command's failure (cobra maps it to exit 1).
func runTodo(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newTodoCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

// todoFixture prepares a project dir whose backlog the command resolves via
// CLAUDE_PROJECT_DIR, and returns the store for seeding/verification.
func todoFixture(t *testing.T) (root string, store *kanban.BacklogStore) {
	t.Helper()
	root = t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	store = kanban.NewBacklogStore(todoBacklogPath(root))
	return root, store
}

func TestTodoAdd_PrintsIDAndPosition(t *testing.T) {
	_, store := todoFixture(t)

	out, _, err := runTodo(t, "add", "first card")
	if err != nil {
		t.Fatalf("add first: %v", err)
	}
	if got := strings.TrimSpace(out); got != "t1 1" {
		t.Errorf("first add output = %q, want %q", got, "t1 1")
	}

	out, _, err = runTodo(t, "add", "second card")
	if err != nil {
		t.Fatalf("add second: %v", err)
	}
	if got := strings.TrimSpace(out); got != "t2 2" {
		t.Errorf("second add output = %q, want %q", got, "t2 2")
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rec.Version != 1 || len(rec.Items) != 2 {
		t.Errorf("record = version %d, %d items; want version 1, 2 items", rec.Version, len(rec.Items))
	}
}

func TestTodoAdd_EmptyText_RejectedNoWrite(t *testing.T) {
	root, _ := todoFixture(t)

	_, _, err := runTodo(t, "add", "")
	if err == nil {
		t.Fatal("empty-text add must fail")
	}
	if _, statErr := os.Stat(todoBacklogPath(root)); !os.IsNotExist(statErr) {
		t.Errorf("backlog file must not exist after rejected add (stat err = %v)", statErr)
	}
}

func TestTodoList_EmptyQueueIsNotAnError(t *testing.T) {
	_, _ = todoFixture(t)

	out, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list on missing file: %v", err)
	}
	if !strings.Contains(out, "empty") {
		t.Errorf("empty-queue render = %q, want it to report the empty queue", out)
	}
}

func TestTodoList_JSONStructured(t *testing.T) {
	_, _ = todoFixture(t)
	if _, _, err := runTodo(t, "add", "alpha"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	if _, _, err := runTodo(t, "add", "beta"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	if _, _, err := runTodo(t, "next", "1", "--spec", "SPEC-X-001"); err != nil {
		t.Fatalf("seed pick: %v", err)
	}

	out, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	if !json.Valid([]byte(out)) {
		t.Fatalf("list --json output is not valid JSON: %q", out)
	}
	var rec kanban.BacklogRecord
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(rec.Items) != 2 || rec.Items[0].State != kanban.BacklogStatePicked {
		t.Errorf("json items = %+v; want 2 items with t1 picked", rec.Items)
	}
}

// TestTodoList_LockFreeWhileForeignProcessHoldsLock — AC-TODO-003: list
// succeeds while backlog.lock is held by a separate OS process (the helper
// parks inside a Mutate, holding the lock for ~1.2s).
func TestTodoList_LockFreeWhileForeignProcessHoldsLock(t *testing.T) {
	if runtimeIsWindowsForTodo() {
		t.Skip("helper re-exec exercised on unix; windows covered by GOOS=windows build/vet")
	}
	root, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "held card"); err != nil {
		t.Fatalf("seed add: %v", err)
	}

	helper := exec.Command(os.Args[0], "-test.run=TestTodoHelperProcess", "--")
	helper.Env = append(os.Environ(),
		"MOAI_TODO_HELPER=hold-lock",
		"CLAUDE_PROJECT_DIR="+root,
	)
	if err := helper.Start(); err != nil {
		t.Fatalf("start lock-holder: %v", err)
	}
	defer func() { _ = helper.Wait() }()

	// Let the helper acquire the lock before listing.
	time.Sleep(400 * time.Millisecond)

	out, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list must succeed while a foreign process holds backlog.lock: %v", err)
	}
	if !strings.Contains(out, "held card") {
		t.Errorf("list output = %q, want the seeded card rendered", out)
	}
	if _, err := store.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}
}

func TestTodoDone_RemovesByID(t *testing.T) {
	_, store := todoFixture(t)
	for _, text := range []string{"one", "two", "three"} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("seed add: %v", err)
		}
	}

	if _, _, err := runTodo(t, "done", "3"); err != nil {
		t.Fatalf("done 3: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 2 {
		t.Fatalf("items after done = %d, want 2", len(rec.Items))
	}
	for _, it := range rec.Items {
		if it.ID == "t3" {
			t.Errorf("t3 still present after done")
		}
	}
}

func TestTodoDone_MissReportedFileUntouched(t *testing.T) {
	_, _ = todoFixture(t)
	if _, _, err := runTodo(t, "add", "only"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	root := os.Getenv("CLAUDE_PROJECT_DIR")
	real := todoBacklogPath(root)
	before, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}

	_, _, err = runTodo(t, "done", "t3")
	if err == nil {
		t.Fatal("done on a missing id must fail")
	}
	after, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("file changed on a missed done; must be byte-identical")
	}
}

func TestTodoNextBare_ReadOnlyOldestFirst(t *testing.T) {
	root, _ := todoFixture(t)
	for _, text := range []string{"oldest", "middle", "newest"} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("seed add: %v", err)
		}
	}
	real := todoBacklogPath(root)
	before, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}

	out, _, err := runTodo(t, "next")
	if err != nil {
		t.Fatalf("bare next: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("bare next printed %d lines, want 3: %q", len(lines), out)
	}
	if !strings.Contains(lines[0], "oldest") || !strings.Contains(lines[2], "newest") {
		t.Errorf("bare next order wrong: %q", lines)
	}

	after, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("bare next must leave the backlog byte-identical")
	}
}

func TestTodoNextPick_OneLockedWrite(t *testing.T) {
	_, store := todoFixture(t)
	for _, text := range []string{"one", "two", "three"} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("seed add: %v", err)
		}
	}

	if _, _, err := runTodo(t, "next", "2", "--spec", "SPEC-X-001"); err != nil {
		t.Fatalf("next 2 --spec: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	picked := 0
	for _, it := range rec.Items {
		if it.State == kanban.BacklogStatePicked {
			picked++
			if it.ID != "t2" {
				t.Errorf("picked item = %s, want t2", it.ID)
			}
			if it.SpecID == nil || *it.SpecID != "SPEC-X-001" {
				t.Errorf("picked spec_id = %v, want SPEC-X-001", it.SpecID)
			}
		} else if it.State != kanban.BacklogStateQueued {
			t.Errorf("item %s state = %s, want untouched queued", it.ID, it.State)
		}
	}
	if picked != 1 {
		t.Errorf("picked count = %d, want exactly 1", picked)
	}
}

func TestTodoNext_OutOfRangeFileUntouched(t *testing.T) {
	root, _ := todoFixture(t)
	if _, _, err := runTodo(t, "add", "only"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	real := todoBacklogPath(root)
	before, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}

	_, _, err = runTodo(t, "next", "99")
	if err == nil {
		t.Fatal("next on a missing id must fail")
	}
	after, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("file changed on a missed next; must be byte-identical")
	}
}

// TestTodoConcurrentAdd_8Processes — AC-TODO-001: 8 concurrent `todo add`
// processes against one backlog lose nothing and mint 8 distinct ids.
func TestTodoConcurrentAdd_8Processes(t *testing.T) {
	if runtimeIsWindowsForTodo() {
		t.Skip("helper re-exec exercised on unix; windows covered by GOOS=windows build/vet")
	}
	root, _ := todoFixture(t)

	const n = 8
	var wg sync.WaitGroup
	errs := make([]error, n)
	outputs := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run=TestTodoHelperProcess", "--")
			cmd.Env = append(os.Environ(),
				"MOAI_TODO_HELPER=add",
				"CLAUDE_PROJECT_DIR="+root,
				"MOAI_TODO_HELPER_TEXT=card "+string(rune('A'+i)),
			)
			out, err := cmd.CombinedOutput()
			errs[i] = err
			outputs[i] = string(out)
		}(i)
	}
	wg.Wait()

	seen := make(map[string]bool)
	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("concurrent add %d failed: %v: %s", i, errs[i], outputs[i])
		}
	}
	for i := 0; i < n; i++ {
		for _, field := range strings.Fields(outputs[i]) {
			if strings.HasPrefix(field, "t") && !seen[field] {
				seen[field] = true
			}
		}
	}
	if len(seen) != n {
		t.Errorf("distinct issued ids printed = %d, want %d (got %v)", len(seen), n, seen)
	}

	stdout, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	var rec kanban.BacklogRecord
	if err := json.Unmarshal([]byte(stdout), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(rec.Items) != n {
		t.Fatalf("items after 8 concurrent adds = %d, want %d", len(rec.Items), n)
	}
	texts := make(map[string]bool)
	for _, it := range rec.Items {
		texts[it.Text] = true
	}
	for i := 0; i < n; i++ {
		want := "card " + string(rune('A'+i))
		if !texts[want] {
			t.Errorf("text %q missing from the queue", want)
		}
	}
}

// TestTodoHelperProcess is the re-exec helper backing the cross-process
// tests above (same idiom as internal/kanban/board_lock_cross_test.go).
func TestTodoHelperProcess(t *testing.T) {
	mode := os.Getenv("MOAI_TODO_HELPER")
	if mode == "" {
		return // normal test run, not a helper invocation
	}
	root := os.Getenv("CLAUDE_PROJECT_DIR")
	if root == "" {
		os.Exit(4)
	}
	store := kanban.NewBacklogStore(todoBacklogPath(root))
	_ = store

	switch mode {
	case "add":
		// Run the FULL cobra verb (not store.Add directly) so the
		// concurrency test exercises the same code path the built binary
		// serves — AC-TODO-001's "concurrent `moai todo add` processes".
		cmd := newTodoCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetArgs([]string{"add", os.Getenv("MOAI_TODO_HELPER_TEXT")})
		if err := cmd.Execute(); err != nil {
			t.Logf("helper add: %v", err)
			os.Exit(3)
		}
		// Print to the process stdout directly — t.Logf output is dropped
		// when the helper runs without -v, and the parent parses this line.
		fmt.Print(out.String())
	case "hold-lock":
		err := store.Mutate(func(rec *kanban.BacklogRecord) error {
			time.Sleep(1200 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Logf("helper hold-lock: %v", err)
			os.Exit(3)
		}
	default:
		os.Exit(4)
	}
	os.Exit(0)
}

func runtimeIsWindowsForTodo() bool {
	return os.PathSeparator == '\\'
}

// TestTodoCmd_NoAskUserQuestion — AC-TODO-014: the todo command surface
// carries zero interactive-prompt references (static guard, the
// TestNew_NoAskUserQuestion idiom scoped to todo.go), with a negative
// control proving the guard detects an introduced reference.
func TestTodoCmd_NoAskUserQuestion(t *testing.T) {
	data, err := os.ReadFile("todo.go")
	if err != nil {
		t.Fatalf("read todo.go: %v", err)
	}
	if reason, bad := todoPromptGuard(string(data)); bad {
		t.Errorf("todo.go must stay prompt-free: %s", reason)
	}

	// Negative control: the guard must flag a synthetic violation.
	if _, bad := todoPromptGuard("x := AskUserQuestion()"); !bad {
		t.Error("guard must detect an AskUserQuestion reference (negative control)")
	}
}

// todoPromptGuard reports whether source contains an interactive-prompt
// reference. Split out so the negative control exercises the same predicate
// the guard uses.
func todoPromptGuard(source string) (reason string, bad bool) {
	for _, token := range []string{"AskUserQuestion", "mcp__askuser"} {
		if strings.Contains(source, token) {
			return "found " + token, true
		}
	}
	return "", false
}

// TestTodoBareInvocationLists pins the documented contract that a bare
// `moai todo` renders the queue. The skill surface (.claude/skills/moai)
// and workflows/todo.md both describe the bare form as the list surface;
// the command used to answer it with cobra's help text instead, so the
// documented entry point never reached the backlog it names.
//
// The subcommand `moai todo list` stays valid — this widens the surface
// rather than moving it.
func TestTodoBareInvocationLists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	if _, _, err := runTodo(t, "add", "first card"); err != nil {
		t.Fatalf("seed add: %v", err)
	}

	bare, _, err := runTodo(t)
	if err != nil {
		t.Fatalf("bare todo: %v", err)
	}
	if strings.Contains(bare, "USAGE") || strings.Contains(bare, "Usage:") {
		t.Errorf("bare todo printed help, want the queue:\n%s", bare)
	}
	if !strings.Contains(bare, "first card") {
		t.Errorf("bare todo did not render the seeded card:\n%s", bare)
	}

	listed, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("todo list: %v", err)
	}
	if bare != listed {
		t.Errorf("bare todo and `todo list` diverged:\nbare=%q\nlist=%q", bare, listed)
	}
}

// TestTodoUnknownSubcommandStillErrors guards the widening above: making
// the bare form do work must not turn a mistyped verb into a silent list.
func TestTodoUnknownSubcommandStillErrors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	if _, _, err := runTodo(t, "lsit"); err == nil {
		t.Error("mistyped subcommand was accepted, want an error")
	}
}
