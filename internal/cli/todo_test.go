// todo_test.go — SPEC-KANBAN-TODO-CLI-001 M2 acceptance tests (CLI verbs).
//
// Covers the CLI-process-level ACs the store-level M1 suite defers:
// AC-TODO-001 (concurrent add loses nothing), AC-TODO-003 (list lock-free
// + --json), AC-TODO-004 (done by id under lock), AC-TODO-005 (bare next
// read-only), AC-TODO-006 (next <n> --spec one locked write), AC-TODO-014
// (no-prompt static guard), plus the acceptance.md §C edge cases.
//
// The t71 block (pick-marking race hardening) covers: atomic `add --pick`
// (one cross-process-locked write appends AND picks), `unpick <n>` (the
// picked→queued recovery verb), and pick confirmations that carry the card
// text prefix (+ optional `--expect <prefix>` refusal guard).
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
	case "add-pick":
		// t71: run the FULL `add --pick` verb (not a direct store call) so
		// the concurrency test exercises the same atomic code path the
		// built binary serves.
		cmd := newTodoCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetArgs([]string{"add", os.Getenv("MOAI_TODO_HELPER_TEXT"), "--pick"})
		if err := cmd.Execute(); err != nil {
			t.Logf("helper add --pick: %v", err)
			os.Exit(3)
		}
		fmt.Print(out.String())
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

// --- t71: pick-marking race hardening ---
//
// Incident (2026-08-16, lead run tjv7iy): the lead ran `moai todo next t67`
// on a guessed id before reading the `add` output, while a concurrent
// session had just issued t67 to an unrelated card — so the wrong card was
// marked picked. Recovery needed done+re-add (id churn, added_at loss)
// because no unpick verb existed. Root causes: add→next was two separate
// writes; `next <n>` printed only "picked t67" so the mis-pick was not
// observable; no recovery verb.

// TestTodoAddPick_OneLockedWrite — `add --pick` appends AND marks picked in
// a single cross-process-locked write: no queued window exists between the
// add and the pick for a guessed id to address, and the issued id is printed
// (the add output the incident's `next` jumped ahead of).
func TestTodoAddPick_OneLockedWrite(t *testing.T) {
	_, store := todoFixture(t)

	out, _, err := runTodo(t, "add", "fix the widget race", "--pick")
	if err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	if got := strings.TrimSpace(out); got != "picked t1 fix the widget race" {
		t.Errorf("add --pick output = %q, want %q", got, "picked t1 fix the widget race")
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(rec.Items))
	}
	it := rec.Items[0]
	if it.ID != "t1" || it.State != kanban.BacklogStatePicked {
		t.Errorf("item = %s/%s, want t1/picked from the same write", it.ID, it.State)
	}

	// The high-water mark must advance inside the same locked write, so the
	// next plain add cannot re-mint t1.
	out, _, err = runTodo(t, "add", "follow-up card")
	if err != nil {
		t.Fatalf("post-pick add: %v", err)
	}
	if got := strings.TrimSpace(out); got != "t2 1" {
		t.Errorf("post-pick add output = %q, want %q (high-water mark must clear t1)", got, "t2 1")
	}
}

// TestTodoAddPick_ConcurrentProcesses — the incident's exact shape: a
// concurrent add from another session must not interleave between an add
// and its pick. 4 concurrent `add --pick` processes all land picked with
// distinct ids; nothing is left stranded queued mid-flight.
func TestTodoAddPick_ConcurrentProcesses(t *testing.T) {
	if runtimeIsWindowsForTodo() {
		t.Skip("helper re-exec exercised on unix; windows covered by GOOS=windows build/vet")
	}
	root, _ := todoFixture(t)

	const n = 4
	var wg sync.WaitGroup
	errs := make([]error, n)
	outputs := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run=TestTodoHelperProcess", "--")
			cmd.Env = append(os.Environ(),
				"MOAI_TODO_HELPER=add-pick",
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
			t.Fatalf("concurrent add --pick %d failed: %v: %s", i, errs[i], outputs[i])
		}
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
		t.Fatalf("items after %d concurrent add --pick = %d, want %d", n, len(rec.Items), n)
	}
	for _, it := range rec.Items {
		if it.State != kanban.BacklogStatePicked {
			t.Errorf("item %s state = %s, want picked (add+pick is one write; no queued window)", it.ID, it.State)
		}
	}
}

// TestTodoUnpick_RevertsPickedToQueued — the recovery verb the incident
// lacked: picked→queued in one locked write, preserving the id and added_at
// (no done+re-add churn) and clearing the spec_id attached at pick time so
// the card returns to the shape `add` issued.
func TestTodoUnpick_RevertsPickedToQueued(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "mis-picked card"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	before, err := store.Load()
	if err != nil {
		t.Fatalf("load before: %v", err)
	}
	addedAt := before.Items[0].AddedAt

	if _, _, err := runTodo(t, "next", "1", "--spec", "SPEC-X-001"); err != nil {
		t.Fatalf("seed pick: %v", err)
	}

	out, _, err := runTodo(t, "unpick", "1")
	if err != nil {
		t.Fatalf("unpick 1: %v", err)
	}
	if got := strings.TrimSpace(out); got != "unpicked t1 mis-picked card" {
		t.Errorf("unpick output = %q, want %q", got, "unpicked t1 mis-picked card")
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load after: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("items after unpick = %d, want 1 (no re-add churn)", len(rec.Items))
	}
	it := rec.Items[0]
	if it.ID != "t1" || it.State != kanban.BacklogStateQueued {
		t.Errorf("item = %s/%s, want t1/queued", it.ID, it.State)
	}
	if it.SpecID != nil {
		t.Errorf("spec_id = %v, want nil (queued cards carry no spec)", *it.SpecID)
	}
	if it.AddedAt != addedAt {
		t.Errorf("added_at = %s, want the original %s", it.AddedAt, addedAt)
	}
}

// TestTodoUnpick_RefusalsLeaveFileUntouched — unpick refuses a card that is
// not picked (nothing to revert) and a missing id, leaving the backlog
// byte-identical in both cases (the done-miss contract).
func TestTodoUnpick_RefusalsLeaveFileUntouched(t *testing.T) {
	root, _ := todoFixture(t)
	if _, _, err := runTodo(t, "add", "queued card"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	real := todoBacklogPath(root)
	before, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}

	if _, _, err := runTodo(t, "unpick", "1"); err == nil {
		t.Error("unpick on a queued (not picked) card must fail")
	}
	after, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read after refused unpick: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("file changed on a refused unpick; must be byte-identical")
	}

	if _, _, err := runTodo(t, "unpick", "t9"); err == nil {
		t.Error("unpick on a missing id must fail")
	}
	after2, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read after missed unpick: %v", err)
	}
	if !bytes.Equal(before, after2) {
		t.Error("file changed on a missed unpick; must be byte-identical")
	}
}

// TestTodoNextPick_ConfirmationShowsText — root cause 2 of the incident:
// `next <n>` printed only "picked t67", so a mis-pick was not immediately
// observable. The confirmation now carries a ~40-rune prefix of the card
// text, truncated on rune boundaries so multi-byte card text cannot be
// sliced mid-character.
func TestTodoNextPick_ConfirmationShowsText(t *testing.T) {
	_, _ = todoFixture(t)
	long := strings.Repeat("카드", 30) // 60 runes > the 40-rune window
	if _, _, err := runTodo(t, "add", long); err != nil {
		t.Fatalf("seed add: %v", err)
	}

	out, _, err := runTodo(t, "next", "1")
	if err != nil {
		t.Fatalf("next 1: %v", err)
	}
	got := strings.TrimSpace(out)
	want := "picked t1 " + strings.Repeat("카드", 20) + "..."
	if got != want {
		t.Errorf("pick confirmation = %q, want %q", got, want)
	}
	if strings.Contains(got, strings.Repeat("카드", 21)) {
		t.Error("confirmation leaked card text past the 40-rune prefix")
	}
}

// TestTodoNextPick_ExpectGuard — belt-and-braces for root cause 1: when the
// caller knows which card it expects, `--expect <prefix>` refuses the pick
// unless the addressed card's text starts with the prefix, leaving the file
// byte-identical on the refusal.
func TestTodoNextPick_ExpectGuard(t *testing.T) {
	root, _ := todoFixture(t)
	for _, text := range []string{"alpha widget", "beta widget"} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("seed add: %v", err)
		}
	}
	real := todoBacklogPath(root)
	before, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}

	if _, _, err := runTodo(t, "next", "2", "--expect", "alpha"); err == nil {
		t.Fatal("next --expect must refuse when the card text does not match")
	}
	after, err := os.ReadFile(real)
	if err != nil {
		t.Fatalf("read after refused pick: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("file changed on a refused --expect pick; must be byte-identical")
	}

	out, _, err := runTodo(t, "next", "2", "--expect", "beta")
	if err != nil {
		t.Fatalf("next 2 --expect beta: %v", err)
	}
	if got := strings.TrimSpace(out); got != "picked t2 beta widget" {
		t.Errorf("matching --expect pick output = %q, want %q", got, "picked t2 beta widget")
	}
}
