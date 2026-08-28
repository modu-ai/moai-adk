// todo_undone_test.go — SPEC-TODO-DESTRUCTIVE-GUARD-001 (card t330): the
// acceptance suite for `done`'s reversal, its guards, and the visibility
// split between live-queue readers and the downgrade export.
//
// The load-bearing test is the round trip: `done` followed by `undone` must
// leave the queue BYTE-IDENTICAL, findings included. A restore that recovers
// the card row but loses the relations recorded about it is the silent half
// of the failure this SPEC exists to close (spec.md §A.2).
package cli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// stubLandingQuery pins the landing query's answer for the duration of a test.
//
// It replaces the process seam rather than seeding a ref in the fixture repo,
// because the query runs `git` in the PROCESS working directory — the real
// repository the suite runs in — not in the fixture's queue root. A fixture
// ref would therefore be ignored and the assertion would rest on whatever the
// ambient repository's origin/main happens to contain: green today, red on a
// machine whose remote names the card, and never actually measuring the guard.
func stubLandingQuery(t *testing.T, out string, err error) *[][]string {
	t.Helper()
	original := todoRunCommand
	t.Cleanup(func() { todoRunCommand = original })
	calls := &[][]string{}
	todoRunCommand = func(name string, args ...string) (string, error) {
		*calls = append(*calls, append([]string{name}, args...))
		return out, err
	}
	return calls
}

// exportedRecord runs `export-json` and returns the parsed artifact plus the
// two streams, captured SEPARATELY: `internal/cli/todo.go:20-22` contracts
// one structured stdout line, so a disclosure on stdout would pollute the
// surface agents parse (AC-TDG-015).
func exportedRecord(t *testing.T, root string) (parsed map[string]any, stdout, stderr string) {
	t.Helper()
	stdout, stderr, err := runTodo(t, "export-json")
	if err != nil {
		t.Fatalf("export-json: %v (stderr %q)", err, stderr)
	}
	raw, err := os.ReadFile(filepath.Join(root, ".moai", "state", "todo", "backlog.json"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse export: %v", err)
	}
	return parsed, stdout, stderr
}

// archivedIDs reads the ids the export's archive field carries.
func archivedIDs(t *testing.T, parsed map[string]any) []string {
	t.Helper()
	entries, ok := parsed["archived"].([]any)
	if !ok {
		t.Fatalf("export carries no `archived` array: keys %v", exportKeysOf(parsed))
	}
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		entry, ok := e.(map[string]any)
		if !ok {
			t.Fatalf("archive entry is not an object: %T", e)
		}
		item, ok := entry["item"].(map[string]any)
		if !ok {
			t.Fatalf("archive entry carries no `item` object: %v", entry)
		}
		id, _ := item["id"].(string)
		ids = append(ids, id)
	}
	return ids
}

// exportKeysOf renders an export's top-level key set for a failure message.
func exportKeysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// liveIDs reads the ids the export's live `items` array carries.
func liveIDs(t *testing.T, parsed map[string]any) []string {
	t.Helper()
	items, ok := parsed["items"].([]any)
	if !ok {
		t.Fatalf("export carries no `items` array: keys %v", exportKeysOf(parsed))
	}
	ids := make([]string, 0, len(items))
	for _, e := range items {
		item, _ := e.(map[string]any)
		id, _ := item["id"].(string)
		ids = append(ids, id)
	}
	return ids
}

// AC-TDG-001 — `undone` restores the card.
func TestTodoUndone_RestoresTheCard(t *testing.T) {
	todoFixture(t)
	seedTodo(t, "alpha work")

	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	out, _, err := runTodo(t, "undone", "t1")
	if err != nil {
		t.Fatalf("undone: %v", err)
	}
	if !strings.HasPrefix(out, "undone t1 ") {
		t.Errorf("undone output = %q, want it to start with %q", out, "undone t1 ")
	}
	list, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(list, "t1") {
		t.Errorf("list = %q, want it to name the restored card t1", list)
	}
}

// AC-TDG-002 — the round trip is byte-identical, findings included.
func TestTodoDone_UndoneRoundTripIsByteIdentical(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha work", "beta work")
	if _, _, err := runTodo(t, "relate", "t1", "t2", "--relation", "contains"); err != nil {
		t.Fatalf("relate: %v", err)
	}
	before := readBacklogBytes(t, root)

	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	if same := readBacklogBytes(t, root); string(same) == string(before) {
		t.Fatal("done must actually change the record")
	}
	if _, _, err := runTodo(t, "undone", "t1"); err != nil {
		t.Fatalf("undone: %v", err)
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Errorf("done+undone is not an exact reversal.\n got: %s\nwant: %s", got, before)
	}
}

// AC-TDG-003 — `done` retains rather than discards, verifiable WITHOUT undone.
func TestTodoDone_ArchivesRatherThanDiscards(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha work")
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	parsed, _, _ := exportedRecord(t, root)
	if ids := archivedIDs(t, parsed); len(ids) != 1 || ids[0] != "t1" {
		t.Errorf("archive ids = %v, want [t1]", ids)
	}
}

// AC-TDG-006 — the archive survives migration from a JSON-only queue.
func TestTodoUndone_SurvivesMigrationFromLegacyJSON(t *testing.T) {
	root, store := todoFixture(t)
	queue := filepath.Dir(store.Path())
	if err := os.MkdirAll(queue, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	legacy := `{"version":1,"last_seq":2,"items":[` +
		`{"id":"t1","text":"alpha work","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"queued"},` +
		`{"id":"t2","text":"beta work","added_at":"2026-01-01T00:00:01Z","spec_id":null,"state":"queued"}],` +
		`"findings":[]}` + "\n"
	if err := os.WriteFile(store.Path(), []byte(legacy), 0o644); err != nil {
		t.Fatalf("seed legacy json: %v", err)
	}
	if _, err := os.Stat(store.EnginePath()); err == nil {
		t.Fatalf("fixture must start with no database at %s", store.EnginePath())
	}

	before := readBacklogBytes(t, root)
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	if _, err := os.Stat(store.EnginePath()); err != nil {
		t.Fatalf("done must have migrated the queue into %s: %v", store.EnginePath(), err)
	}
	if _, _, err := runTodo(t, "undone", "t1"); err != nil {
		t.Fatalf("undone: %v", err)
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Errorf("legacy-origin round trip is not byte-identical.\n got: %s\nwant: %s", got, before)
	}
}

// AC-TDG-007 — archived rows are invisible to live-queue readers.
func TestTodoDone_ArchivedRowsInvisibleToLiveReaders(t *testing.T) {
	todoFixture(t)
	seedTodo(t, "alpha work", "beta work")
	if _, _, err := runTodo(t, "relate", "t1", "t2", "--relation", "contains"); err != nil {
		t.Fatalf("relate: %v", err)
	}
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"list", []string{"list"}},
		{"next", []string{"next"}},
		{"analyze", []string{"analyze"}},
	} {
		out, _, err := runTodo(t, tc.args...)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if strings.Contains(out, "t1") {
			t.Errorf("%s names the archived card t1:\n%s", tc.name, out)
		}
	}

	// `why` echoes the addressed id back on the empty path, so a grep for
	// `t1` matches even when it is reporting nothing. Assert the exact line.
	out, _, err := runTodo(t, "why", "t1")
	if err != nil {
		t.Fatalf("why: %v", err)
	}
	if got := strings.TrimSpace(out); got != "t1: no findings" {
		t.Errorf("why t1 = %q, want %q", got, "t1: no findings")
	}

	// Re-adding the archived card's exact text is not a duplicate of it.
	if _, stderr, err := runTodo(t, "add", "alpha work"); err != nil {
		t.Fatalf("re-add: %v (stderr %q)", err, stderr)
	}
}

// AC-TDG-008 — `--expect` refuses a mismatch and writes nothing.
func TestTodoDone_ExpectRefusesMismatch(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha work")
	before := readBacklogBytes(t, root)

	_, stderr, err := runTodo(t, "done", "t1", "--expect", "beta")
	if err == nil {
		t.Fatal("done --expect beta must refuse a card reading \"alpha work\"")
	}
	if !strings.Contains(stderr, "alpha work") {
		t.Errorf("refusal %q must name the observed prefix", stderr)
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Errorf("refused done wrote to the record.\n got: %s\nwant: %s", got, before)
	}
}

// AC-TDG-009 — `--require-landed` refuses on positive evidence of not-landed.
func TestTodoDone_RequireLandedRefusesWhenNotLanded(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha work")
	// An empty commit set is the ref answering "nothing here names this card".
	stubLandingQuery(t, "", nil)
	before := readBacklogBytes(t, root)

	_, stderr, err := runTodo(t, "done", "t1", "--require-landed")
	if err == nil {
		t.Fatal("done --require-landed must refuse a card no commit names")
	}
	if !strings.Contains(stderr, kanban.LandedRef) {
		t.Errorf("refusal %q must name the ref the answer is about (%s)", stderr, kanban.LandedRef)
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Errorf("refused done wrote to the record.\n got: %s\nwant: %s", got, before)
	}
}

// AC-TDG-009 (proceed half) — an inconclusive answer proceeds rather than
// refusing. The fixture has no origin/main, so the query errors.
func TestTodoDone_RequireLandedProceedsWhenInconclusive(t *testing.T) {
	todoFixture(t)
	seedTodo(t, "alpha work")
	// A failing query — no git, no such ref, a broken remote — is an
	// unanswerable question, not evidence of not-landed.
	stubLandingQuery(t, "", errors.New("fatal: bad revision 'origin/main'"))
	stdout, stderr, err := runTodo(t, "done", "t1", "--require-landed")
	if err != nil {
		t.Fatalf("an inconclusive landing answer must proceed, got %v (stderr %q)", err, stderr)
	}
	if !strings.HasPrefix(stdout, "done t1") {
		t.Errorf("stdout = %q, want the archive to have gone through", stdout)
	}
	if !strings.Contains(stderr, "could not answer") {
		t.Errorf("stderr = %q, want the inconclusive answer surfaced rather than swallowed", stderr)
	}
}

// AC-TDG-010 — the flag is opt-in and costs nothing when absent.
func TestTodoDone_NoLandingQueryWithoutTheFlag(t *testing.T) {
	todoFixture(t)
	seedTodo(t, "alpha work", "beta work")

	original := todoRunCommand
	t.Cleanup(func() { todoRunCommand = original })
	var calls [][]string
	todoRunCommand = func(name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		return "abc1234 landed t2\n", nil
	}

	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	if len(calls) != 0 {
		t.Errorf("done without --require-landed spawned %d subprocess(es): %v", len(calls), calls)
	}

	// Paired with the positive control: WITH the flag, exactly one query runs.
	if _, _, err := runTodo(t, "done", "t2", "--require-landed"); err != nil {
		t.Fatalf("done --require-landed (stub reports landed): %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("done --require-landed spawned %d subprocess(es), want 1: %v", len(calls), calls)
	}
	if calls[0][0] != "git" {
		t.Errorf("landing query ran %q, want git", calls[0][0])
	}
}

// AC-TDG-011 — every refusal path writes nothing.
func TestTodoDoneUndone_RefusalsWriteNothing(t *testing.T) {
	root, store := todoFixture(t)
	seedTodo(t, "alpha work")
	stubLandingQuery(t, "", nil)
	// Archive t1, then reissue the id to a different live card so the
	// collision path is reachable.
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("seed done: %v", err)
	}
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Items = append(rec.Items, kanban.BacklogItem{
			ID: "t1", Text: "a different card", AddedAt: "2026-01-01T00:00:00Z",
			State: kanban.BacklogStateQueued,
		})
		return nil
	}); err != nil {
		t.Fatalf("seed reissue: %v", err)
	}
	before := readBacklogBytes(t, root)

	refusals := []struct {
		name string
		args []string
	}{
		{"absent id", []string{"done", "t99"}},
		{"expect mismatch", []string{"done", "t1", "--expect", "zzz"}},
		{"require-landed", []string{"done", "t1", "--require-landed"}},
		{"undone of an id never archived", []string{"undone", "t99"}},
		{"undone into a reissued id", []string{"undone", "t1"}},
	}
	for _, r := range refusals {
		_, stderr, err := runTodo(t, r.args...)
		if err == nil {
			t.Errorf("%s: must refuse (args %v)", r.name, r.args)
			continue
		}
		if got := readBacklogBytes(t, root); string(got) != string(before) {
			t.Errorf("%s: refusal wrote to the record (stderr %q).\n got: %s\nwant: %s",
				r.name, stderr, got, before)
		}
	}
}

// AC-TDG-012 — nothing prompts, on any path, with stdin closed. The exit code
// is determinate either way; what the criterion asserts is that every call
// RETURNS rather than blocking on a terminal read.
func TestTodoDoneUndone_NeverPrompt(t *testing.T) {
	todoFixture(t)
	seedTodo(t, "alpha work")
	stubLandingQuery(t, "", nil)

	paths := [][]string{
		{"done", "t1", "--expect", "zzz"},  // refusal
		{"done", "t99"},                    // error
		{"done", "t1", "--require-landed"}, // guard refusal
		{"done", "t1"},                     // success
		{"undone", "t1"},                   // success
		{"undone", "t1"},                   // refusal (already restored)
	}
	for _, args := range paths {
		err := runTodoWithClosedStdin(t, args...)
		t.Logf("%v terminated with %v", args, err)
	}
}

// AC-TDG-013 — a reissued id refuses restore and does not overwrite the live card.
func TestTodoUndone_ReissuedIDRefuses(t *testing.T) {
	root, store := todoFixture(t)
	seedTodo(t, "alpha work")
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Items = append(rec.Items, kanban.BacklogItem{
			ID: "t1", Text: "a different card", AddedAt: "2026-01-01T00:00:00Z",
			State: kanban.BacklogStateQueued,
		})
		return nil
	}); err != nil {
		t.Fatalf("seed reissue: %v", err)
	}
	before := readBacklogBytes(t, root)

	_, stderr, err := runTodo(t, "undone", "t1")
	if err == nil {
		t.Fatal("undone must refuse when the id has been reissued")
	}
	// "names the collision" means naming the reissue, not merely echoing the
	// argument back: the parent command's mistyped-verb guard already emits an
	// error carrying `t1` at the base tree, so an id-only assertion passes for
	// the wrong reason.
	if !strings.Contains(stderr, "t1") || !strings.Contains(stderr, "reissued") {
		t.Errorf("refusal %q must name the collision (the id AND the reissue)", stderr)
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Errorf("refused undone wrote to the record.\n got: %s\nwant: %s", got, before)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, it := range rec.Items {
		if it.ID == "t1" && it.Text != "a different card" {
			t.Errorf("the live t1 was overwritten: %q", it.Text)
		}
	}
}

// AC-TDG-015 — the export carries the archive AND discloses the downgrade cost.
func TestTodoExportJSON_CarriesArchiveAndDisclosesDowngrade(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha work", "beta work")
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}

	parsed, stdout, stderr := exportedRecord(t, root)
	if ids := archivedIDs(t, parsed); len(ids) != 1 || ids[0] != "t1" {
		t.Errorf("archive ids = %v, want [t1]", ids)
	}
	if ids := liveIDs(t, parsed); len(ids) != 1 || ids[0] != "t2" {
		t.Errorf("live ids = %v, want [t2]", ids)
	}
	if !strings.HasPrefix(stdout, "exported ") {
		t.Errorf("stdout = %q, want the structured export line", stdout)
	}
	if strings.Contains(stdout, "discard") {
		t.Errorf("the disclosure must not pollute stdout: %q", stdout)
	}
	for _, want := range []string{"archived", "discard"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr disclosure %q must mention %q", stderr, want)
		}
	}
}

// AC-TDG-015 (negative half) — no archive, no disclosure.
func TestTodoExportJSON_NoDisclosureWithoutArchivedRows(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha work")
	_, _, stderr := exportedRecord(t, root)
	if strings.Contains(stderr, "discard") {
		t.Errorf("an archive-free export must not disclose anything: %q", stderr)
	}
}

// AC-TDG-016 — restore empties the archive entry.
func TestTodoUndone_EmptiesTheArchiveEntry(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha work")
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	if _, _, err := runTodo(t, "undone", "t1"); err != nil {
		t.Fatalf("undone: %v", err)
	}
	parsed, _, _ := exportedRecord(t, root)
	if ids := archivedIDs(t, parsed); len(ids) != 0 {
		t.Errorf("archive ids after restore = %v, want empty", ids)
	}

	before := readBacklogBytes(t, root)
	if _, _, err := runTodo(t, "undone", "t1"); err == nil {
		t.Fatal("a second undone must refuse")
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Errorf("the second undone wrote to the record.\n got: %s\nwant: %s", got, before)
	}
}
