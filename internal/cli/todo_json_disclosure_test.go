// todo_json_disclosure_test.go — SPEC-BACKLOG-JSON-DISCLOSURE-001 (card
// t395): a `backlog.json` at the canonical queue path is NOT the queue, and
// every `moai todo` verb that reads the queue and prints says so.
//
// THE READ SURFACE, ENUMERATED FROM THE CODE (plan.md M3 — the list is
// derived, never recalled). The 17 verbs registered at internal/cli/todo.go
// `cmd.AddCommand(...)` split on whether their RunE path reaches
// BacklogStore.Mutate (the locked write) or only Load/LoadPure (a
// lock-free read):
//
//	READ  list (and the bare `moai todo`, whose RunE is runTodoList)
//	READ  why      (newTodoWhyCmd → newTodoStore().Load)
//	READ  pr       (runTodoPR    → newTodoStore().Load)
//	READ  history  (runTodoHistory → store.LoadPure)
//	WRITE add, done, undone, next, unpick, edit, move, drop, undrop,
//	      analyze, relate, unrelate, export-json — each reaches Mutate
//
// `next` reads AND mutates; it is a write verb, and the disclosure does not
// ride it. Write verbs hold the queue lock and carry a different stdout
// contract, and the operator decision of 2026-09-02 scoped this to the read
// surface (spec.md REQ-BJD-002, plan.md §D.1).
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"
)

// disclosureMarker is the phrase the disclosure line is recognized by. It
// is written out here rather than referenced from the implementation so the
// test asserts observable output, not an internal constant.
const disclosureMarker = "is NOT the queue"

// staleBacklogJSON is a well-formed pre-SQLite record — the shape an export
// or an interrupted migration leaves at the canonical path.
const staleBacklogJSON = `{"version":1,"last_seq":7,"items":[{"id":"t1","text":"stale snapshot card","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"queued"}],"findings":[]}`

// todoReadVerb is one verb of the enumerated read surface.
type todoReadVerb struct {
	name string
	args []string
}

// todoReadSurface is the enumeration above, in invocable form.
var todoReadSurface = []todoReadVerb{
	{"bare", nil},
	{"list", []string{"list"}},
	{"why", []string{"why", "t1"}},
	{"pr", []string{"pr"}},
	{"history", []string{"history"}},
}

// seedDisclosureFixture builds an isolated queue with one card, seeded
// through the CLI so the database is created by the current binary and its
// archive tables are present — the Given AC-BJD-002 and AC-BJD-004 both
// carry, which keeps the existing REQ-TAQ-013 line from co-firing and
// making a count of "exactly one line" ambiguous.
func seedDisclosureFixture(t *testing.T) string {
	t.Helper()
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "write the parser for the config file"); err != nil {
		t.Fatalf("seed queue: %v", err)
	}
	// gh and git are stubbed for the whole fixture: `pr` is on the read
	// surface, and a live subprocess would make this suite depend on the
	// network and on the ambient repository.
	installSpy(t, &spyRunner{prJSON: "[]"})
	return store.Path()
}

// withStaleJSON writes the non-authoritative file beside the database.
func withStaleJSON(t *testing.T, jsonPath string) {
	t.Helper()
	if err := os.WriteFile(jsonPath, []byte(staleBacklogJSON), 0o600); err != nil {
		t.Fatalf("write stale backlog.json: %v", err)
	}
}

// countDisclosureLines returns how many lines of s carry the disclosure.
func countDisclosureLines(s string) int {
	n := 0
	for _, ln := range strings.Split(s, "\n") {
		if strings.Contains(ln, disclosureMarker) {
			n++
		}
	}
	return n
}

// TestTodoReadSurface_DisclosesNonAuthoritativeJSON — AC-BJD-002. Exactly
// one disclosure line per read verb, naming the answering SQLite store and
// naming backlog.json as not authoritative.
func TestTodoReadSurface_DisclosesNonAuthoritativeJSON(t *testing.T) {
	for _, verb := range todoReadSurface {
		t.Run(verb.name, func(t *testing.T) {
			jsonPath := seedDisclosureFixture(t)
			withStaleJSON(t, jsonPath)

			stdout, stderr, err := runTodo(t, verb.args...)
			if err != nil {
				t.Fatalf("todo %v: %v", verb.args, err)
			}
			if got := countDisclosureLines(stderr); got != 1 {
				t.Errorf("%d disclosure lines on stderr, want exactly 1\nstderr:\n%s", got, stderr)
			}
			if !strings.Contains(stderr, "SQLite backlog store") {
				t.Errorf("the disclosure does not name the store that answered\nstderr:\n%s", stderr)
			}
			if !strings.Contains(stderr, "backlog.json") {
				t.Errorf("the disclosure does not name backlog.json\nstderr:\n%s", stderr)
			}
			if strings.Contains(stdout, disclosureMarker) {
				t.Errorf("the disclosure reached stdout\nstdout:\n%s", stdout)
			}
		})
	}
}

// TestTodoReadSurface_StdoutUnpolluted — AC-BJD-003. stdout carries zero
// bytes of the disclosure AND is byte-identical to the same command's
// stdout against a layout with no backlog.json.
func TestTodoReadSurface_StdoutUnpolluted(t *testing.T) {
	for _, verb := range todoReadSurface {
		t.Run(verb.name, func(t *testing.T) {
			jsonPath := seedDisclosureFixture(t)
			clean, _, err := runTodo(t, verb.args...)
			if err != nil {
				t.Fatalf("todo %v (no json): %v", verb.args, err)
			}

			withStaleJSON(t, jsonPath)
			withJSON, stderr, err := runTodo(t, verb.args...)
			if err != nil {
				t.Fatalf("todo %v (with json): %v", verb.args, err)
			}
			if withJSON != clean {
				t.Errorf("stdout diverged when a backlog.json appeared\n--- without ---\n%q\n--- with ---\n%q", clean, withJSON)
			}
			if strings.Contains(withJSON, disclosureMarker) {
				t.Errorf("disclosure bytes on stdout:\n%s", withJSON)
			}
			if countDisclosureLines(stderr) != 1 {
				t.Errorf("the disclosure did not land on stderr\nstderr:\n%s", stderr)
			}
		})
	}
}

// TestTodoReadSurface_SilentWithoutJSON — AC-BJD-004. Nothing to disclose,
// nothing disclosed, on either stream. The fixture's archive tables are
// present, so the existing REQ-TAQ-013 line cannot fire either.
func TestTodoReadSurface_SilentWithoutJSON(t *testing.T) {
	for _, verb := range todoReadSurface {
		t.Run(verb.name, func(t *testing.T) {
			seedDisclosureFixture(t)
			stdout, stderr, err := runTodo(t, verb.args...)
			if err != nil {
				t.Fatalf("todo %v: %v", verb.args, err)
			}
			if countDisclosureLines(stdout)+countDisclosureLines(stderr) != 0 {
				t.Errorf("a disclosure line was emitted with no backlog.json present\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
			}
		})
	}
}

// TestTodoWriteVerbs_CarryNoDisclosure — the scope boundary of the operator
// decision, asserted rather than assumed: write verbs are out.
func TestTodoWriteVerbs_CarryNoDisclosure(t *testing.T) {
	jsonPath := seedDisclosureFixture(t)
	withStaleJSON(t, jsonPath)

	stdout, stderr, err := runTodo(t, "add", "a second card")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if countDisclosureLines(stdout)+countDisclosureLines(stderr) != 0 {
		t.Errorf("the disclosure rode a write verb — the decision scoped it to the read surface\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

// TestTodoDisclosure_LeavesBacklogJSONUntouched — AC-BJD-005. The file is
// evidence of a problem; the disclosure must not disturb it.
func TestTodoDisclosure_LeavesBacklogJSONUntouched(t *testing.T) {
	jsonPath := seedDisclosureFixture(t)
	withStaleJSON(t, jsonPath)
	// Backdate so an incidental rewrite that preserved the bytes would
	// still be caught by the mtime comparison.
	backdated := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(jsonPath, backdated, backdated); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	digest := func() (string, time.Time) {
		t.Helper()
		raw, err := os.ReadFile(jsonPath)
		if err != nil {
			t.Fatalf("read backlog.json: %v", err)
		}
		info, err := os.Stat(jsonPath)
		if err != nil {
			t.Fatalf("stat backlog.json: %v", err)
		}
		sum := sha256.Sum256(raw)
		return hex.EncodeToString(sum[:]), info.ModTime()
	}
	wantSum, wantMod := digest()

	for _, verb := range todoReadSurface {
		if _, _, err := runTodo(t, verb.args...); err != nil {
			t.Fatalf("todo %v: %v", verb.args, err)
		}
		gotSum, gotMod := digest()
		if gotSum != wantSum {
			t.Fatalf("todo %v changed backlog.json bytes: sha256 %s -> %s", verb.args, wantSum, gotSum)
		}
		if !gotMod.Equal(wantMod) {
			t.Fatalf("todo %v changed backlog.json mtime: %s -> %s", verb.args, wantMod, gotMod)
		}
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("backlog.json no longer exists after the read surface ran: %v", err)
	}
}
