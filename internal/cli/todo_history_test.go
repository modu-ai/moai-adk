// todo_history_test.go — SPEC-TODO-ARCHIVE-QUERY-001: the acceptance suite
// for `moai todo history`, the read surface over the archive.
//
// The suite follows acceptance.md §D: every AC names its test, and the
// FIXTURE convention holds — queues are seeded through the public CLI
// (`todo add`, `todo done`, …) except for the two storage surgeries the
// acceptance file names explicitly (a pre-archive `done` shape for AC-TAQ-004
// and a dropped-tables / legacy-JSON store for AC-TAQ-013).
//
// The goldens under testdata/golden/live-readers/ were captured at commit C
// (see progress.md §E.2) from the pre-verb tree; clause 3 of AC-TAQ-011
// compares this suite's replay of the same seed sequence against them.
package cli

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedHistoryFates builds the four-fates fixture: t1 queued, t2 picked,
// t3 dropped, t4 archived (done). All four fates of spec.md §A.3 in one
// queue, seeded entirely through the CLI. The fixture is isolated FIRST —
// todoFixture pins CLAUDE_PROJECT_DIR at a temp git repo so runTodo can
// never resolve the queue into the primary checkout.
func seedHistoryFates(t *testing.T) {
	t.Helper()
	_, _ = todoFixture(t)
	for _, text := range []string{
		"write the parser for the config file",
		"polish the docs landing page",
		"drop the legacy cache layer",
		"wire the banner into the shell",
	} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("add %q: %v", text, err)
		}
	}
	if _, _, err := runTodo(t, "next", "2", "--spec", "SPEC-X-001"); err != nil {
		t.Fatalf("next 2: %v", err)
	}
	if _, _, err := runTodo(t, "drop", "3", "superseded by the parser rewrite"); err != nil {
		t.Fatalf("drop 3: %v", err)
	}
	if _, _, err := runTodo(t, "done", "4"); err != nil {
		t.Fatalf("done 4: %v", err)
	}
}

// AC-TAQ-001 — a live card reports `live` and its own state, one line,
// exit 0. The three live states are proven distinct rather than collapsed.
func TestTodoHistoryReportsLiveCard(t *testing.T) {
	seedHistoryFates(t)

	cases := []struct {
		id   string
		want string
	}{
		{"t1", "t1\tlive\tqueued\twrite the parser for the config file\n"},
		{"t2", "t2\tlive\tpicked\tpolish the docs landing page\n"},
		{"t3", "t3\tlive\tdropped\t[DROPPED — superseded by the parser rewrite] drop the legacy cache layer\n"},
	}
	for _, tc := range cases {
		out, _, err := runTodo(t, "history", tc.id)
		if err != nil {
			t.Fatalf("history %s: %v", tc.id, err)
		}
		if out != tc.want {
			t.Errorf("history %s stdout = %q, want %q", tc.id, out, tc.want)
		}
	}
}

// AC-TAQ-002 — an archived card reports `archived` and the state it held at
// archive time.
func TestTodoHistoryReportsArchivedCard(t *testing.T) {
	seedHistoryFates(t)

	out, _, err := runTodo(t, "history", "t4")
	if err != nil {
		t.Fatalf("history t4: %v", err)
	}
	want := "t4\tarchived\tqueued\twire the banner into the shell\n"
	if out != want {
		t.Errorf("history t4 stdout = %q, want %q", out, want)
	}
}

// AC-TAQ-003 — an unknown id reports `absent`, exits 0, and is NOT
// byte-equal to what `why` prints for the same id (todo_why.go:34-37
// renders the same bytes for "no findings" and "never seen"; the conflation
// must stay closed).
func TestTodoHistoryReportsAbsentCard(t *testing.T) {
	seedHistoryFates(t)

	out, _, err := runTodo(t, "history", "t9999")
	if err != nil {
		t.Fatalf("history t9999: %v", err)
	}
	if !strings.Contains(out, "t9999") || !strings.Contains(out, "absent") {
		t.Errorf("history t9999 stdout = %q, want it to name t9999 and absent", out)
	}
	whyOut, _, err := runTodo(t, "why", "t9999")
	if err != nil {
		t.Fatalf("why t9999: %v", err)
	}
	if out == whyOut {
		t.Errorf("history and why render the same bytes for an unknown id (%q) — the conflation AC-TAQ-003 closes", out)
	}
}

// AC-TAQ-005 — a bare ordinal normalizes to the explicit id form.
func TestTodoHistoryNormalizesBareOrdinal(t *testing.T) {
	seedHistoryFates(t)

	bare, _, err := runTodo(t, "history", "4")
	if err != nil {
		t.Fatalf("history 4: %v", err)
	}
	explicit, _, err := runTodo(t, "history", "t4")
	if err != nil {
		t.Fatalf("history t4: %v", err)
	}
	if bare != explicit {
		t.Errorf("history 4 = %q, history t4 = %q — bare ordinal must normalize byte-identically", bare, explicit)
	}
}

// AC-TAQ-006 — the bare listing is most-recently-archived first.
func TestTodoHistoryListsNewestFirst(t *testing.T) {
	_, _ = todoFixture(t)
	for _, text := range []string{"alpha work", "beta work", "gamma work"} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("add %q: %v", text, err)
		}
	}
	for _, id := range []string{"t1", "t2", "t3"} {
		if _, _, err := runTodo(t, "done", id); err != nil {
			t.Fatalf("done %s: %v", id, err)
		}
	}

	out, _, err := runTodo(t, "history")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("history stdout carries %d lines (%q), want 3", len(lines), out)
	}
	wantIDs := []string{"t3", "t2", "t1"}
	for i, want := range wantIDs {
		if got := strings.SplitN(lines[i], "\t", 2)[0]; got != want {
			t.Errorf("line %d names %q, want %q", i, got, want)
		}
	}
}

// AC-TAQ-009 — an empty archive says so explicitly, exit 0.
func TestTodoHistoryEmptyArchiveIsExplicit(t *testing.T) {
	_, _ = todoFixture(t)
	if _, _, err := runTodo(t, "add", "alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}

	out, _, err := runTodo(t, "history")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty archive rendered zero bytes — the explicit line is the requirement")
	}
	if out != "archive is empty\n" {
		t.Errorf("empty-archive render = %q, want the explicit line", out)
	}
}

// seedFiveAndDeleteT3 builds the AC-TAQ-004 surgery: five cards issued
// through the CLI, then t3's row deleted from items WITHOUT lowering
// last_seq — the exact shape a pre-archive `done` left behind, which the
// current CLI (which archives) cannot produce. This is the first of the
// two storage surgeries acceptance.md names.
func seedFiveAndDeleteT3(t *testing.T) (root string, store *kanban.BacklogStore) {
	t.Helper()
	root, store = todoFixture(t)
	for _, text := range []string{"one", "two", "three", "four", "five"} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("add %q: %v", text, err)
		}
	}
	db, err := sql.Open("sqlite", store.EnginePath())
	if err != nil {
		t.Fatalf("open surgery db: %v", err)
	}
	defer func() { _ = db.Close() }()
	res, err := db.Exec(`DELETE FROM items WHERE id = 't3'`)
	if err != nil {
		t.Fatalf("delete t3 row: %v", err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		t.Fatalf("delete t3 row affected %d rows, want 1", n)
	}
	return root, store
}

// todoHistoryPreArchiveNote is the stderr qualifier REQ-TAQ-004 requires:
// an absent id at or below the issued-id mark may have been issued and
// destroyed, and absent does not establish never-issued.
const todoHistoryPreArchiveNote = "may have been issued"

// AC-TAQ-004 — `absent` is qualified for an id that was issued, and only
// for one. The second clause is the regression guard against keying the
// disclosure on archive emptiness: after one `done` the archive is NOT
// empty anymore, and the note must STILL fire because t3 stays destroyed.
func TestTodoHistoryDisclosesPreArchiveQueue(t *testing.T) {
	_, store := seedFiveAndDeleteT3(t)

	out, errOut, err := runTodo(t, "history", "t3")
	if err != nil {
		t.Fatalf("history t3: %v (stderr %q)", err, errOut)
	}
	if out != "t3\tabsent\n" {
		t.Errorf("history t3 stdout = %q, want exactly the single absent line", out)
	}
	if !strings.Contains(errOut, "t3") || !strings.Contains(errOut, todoHistoryPreArchiveNote) {
		t.Errorf("history t3 stderr = %q, want the at-or-below-the-mark qualifier naming t3", errOut)
	}

	// The same queue after one `done` on a different card: the archive now
	// holds an entry, and the note must survive that fact.
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done t1: %v", err)
	}
	out, errOut, err = runTodo(t, "history", "t3")
	if err != nil {
		t.Fatalf("history t3 (post-done): %v (stderr %q)", err, errOut)
	}
	if out != "t3\tabsent\n" {
		t.Errorf("history t3 stdout (post-done) = %q, want exactly the single absent line", out)
	}
	if !strings.Contains(errOut, todoHistoryPreArchiveNote) {
		t.Errorf("history t3 stderr (post-done) = %q — the qualifier went silent once the archive was non-empty; REQ-TAQ-004 is keyed on last_seq, not on emptiness", errOut)
	}

	// A never-issued id is not dressed up as a destroyed one: t9999 sits
	// above the mark, so stdout reports absent and stderr stays silent.
	out, errOut, err = runTodo(t, "history", "t9999")
	if err != nil {
		t.Fatalf("history t9999: %v", err)
	}
	if out != "t9999\tabsent\n" {
		t.Errorf("history t9999 stdout = %q, want exactly the single absent line", out)
	}
	if errOut != "" {
		t.Errorf("history t9999 stderr = %q, want empty — an id above the mark gets no destruction note", errOut)
	}
	_ = store
}

// todoHistoryDegradedStoreNote is what REQ-TAQ-013 requires a store that
// cannot vouch for an archive to say: WHICH store answered, and that no
// archive is available.
const todoHistoryDegradedStoreNote = "no archive is available"

// AC-TAQ-013 — a store that cannot vouch for an archive degrades with a
// disclosure rather than answering silently. Clause 1: a database whose
// archive tables were dropped. Clause 2: a legacy backlog.json with no
// backlog.db — the second storage surgery acceptance.md names.
func TestTodoHistoryDegradesWithoutArchiveTables(t *testing.T) {
	t.Run("dropped archive tables", func(t *testing.T) {
		_, store := todoFixture(t)
		if _, _, err := runTodo(t, "add", "alpha work"); err != nil {
			t.Fatalf("add: %v", err)
		}
		// Each scenario drops the tables FRESH: the first history read's
		// engine open runs the DDL and recreates them (the store's universal
		// open behavior — list would do the same), consuming the degraded
		// shape for every later invocation. One surgery per invocation.
		dropArchiveTables := func() {
			t.Helper()
			db, err := sql.Open("sqlite", store.EnginePath())
			if err != nil {
				t.Fatalf("open surgery db: %v", err)
			}
			for _, stmt := range []string{`DROP TABLE archived_items`, `DROP TABLE archived_findings`} {
				if _, err := db.Exec(stmt); err != nil {
					t.Fatalf("%s: %v", stmt, err)
				}
			}
			if err := db.Close(); err != nil {
				t.Fatalf("close surgery db: %v", err)
			}
		}

		dropArchiveTables()
		out, errOut, err := runTodo(t, "history", "t1")
		if err != nil {
			t.Fatalf("history t1: %v (stderr %q)", err, errOut)
		}
		if out != "t1\tlive\tqueued\talpha work\n" {
			t.Errorf("history t1 stdout = %q, want the live line — the degraded lookup must still answer", out)
		}
		if !strings.Contains(errOut, todoHistoryDegradedStoreNote) {
			t.Errorf("history t1 stderr = %q, want the store disclosure", errOut)
		}

		dropArchiveTables()
		_, errOut, err = runTodo(t, "history")
		if err != nil {
			t.Fatalf("history (listing): %v (stderr %q)", err, errOut)
		}
		if !strings.Contains(errOut, todoHistoryDegradedStoreNote) {
			t.Errorf("history (listing) stderr = %q, want the store disclosure", errOut)
		}
	})

	t.Run("legacy json only", func(t *testing.T) {
		root, store := todoFixture(t)
		legacy := `{"version":1,"last_seq":7,"items":[{"id":"t1","text":"alpha work","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"queued"}],"findings":[]}`
		if err := os.MkdirAll(filepath.Dir(store.Path()), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(store.Path(), []byte(legacy), 0o600); err != nil {
			t.Fatalf("write legacy json: %v", err)
		}

		out, errOut, err := runTodo(t, "history", "t9999")
		if err != nil {
			t.Fatalf("history t9999: %v (stderr %q)", err, errOut)
		}
		if !strings.Contains(out, "t9999") || !strings.Contains(out, "absent") {
			t.Errorf("history t9999 stdout = %q, want the absent answer", out)
		}
		if !strings.Contains(errOut, "legacy backlog.json") || !strings.Contains(errOut, todoHistoryDegradedStoreNote) {
			t.Errorf("history t9999 stderr = %q, want the legacy-store disclosure", errOut)
		}
		// The read-only posture on this path: serving the JSON must not have
		// migrated it into a database behind the verb's back.
		if _, statErr := os.Stat(store.EnginePath()); !os.IsNotExist(statErr) {
			t.Errorf("backlog.db exists after a history read on a JSON-only queue — the verb migrated the store (stat err %v)", statErr)
		}
		_ = root
	})
}

// seedArchivedCards archives n cards through the CLI and returns nothing —
// the queue is left with n archived entries and an empty live queue.
func seedArchivedCards(t *testing.T, n int) {
	t.Helper()
	_, _ = todoFixture(t)
	for i := 1; i <= n; i++ {
		text := fmt.Sprintf("work item %02d", i)
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("add %q: %v", text, err)
		}
		if _, _, err := runTodo(t, "done", fmt.Sprintf("t%d", i)); err != nil {
			t.Fatalf("done t%d: %v", i, err)
		}
	}
}

// countHistoryLines counts the listing lines on stdout.
func countHistoryLines(out string) int {
	trimmed := strings.TrimSuffix(out, "\n")
	if trimmed == "" {
		return 0
	}
	return len(strings.Split(trimmed, "\n"))
}

// AC-TAQ-007 — the listing is bounded by default, the bound is adjustable,
// and 0 means unbounded.
func TestTodoHistoryLimitBound(t *testing.T) {
	seedArchivedCards(t, 25)

	out, _, err := runTodo(t, "history")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if got := countHistoryLines(out); got != 20 {
		t.Errorf("default listing carries %d lines, want the default bound 20", got)
	}

	out, _, err = runTodo(t, "history", "--limit", "5")
	if err != nil {
		t.Fatalf("history --limit 5: %v", err)
	}
	if got := countHistoryLines(out); got != 5 {
		t.Errorf("--limit 5 listing carries %d lines, want 5", got)
	}

	out, _, err = runTodo(t, "history", "--limit", "0")
	if err != nil {
		t.Fatalf("history --limit 0: %v", err)
	}
	if got := countHistoryLines(out); got != 25 {
		t.Errorf("--limit 0 listing carries %d lines, want all 25 (0 = unbounded)", got)
	}
}

// AC-TAQ-008 — truncation states the withheld count on stderr, and stdout
// carries no such note (a machine reading stdout is unaffected).
func TestTodoHistoryStatesWithheldCount(t *testing.T) {
	seedArchivedCards(t, 25)

	out, errOut, err := runTodo(t, "history", "--limit", "5")
	if err != nil {
		t.Fatalf("history --limit 5: %v", err)
	}
	if !strings.Contains(errOut, "20") || !strings.Contains(errOut, "withheld") {
		t.Errorf("history --limit 5 stderr = %q, want it to state 20 entries withheld", errOut)
	}
	if strings.Contains(out, "withheld") {
		t.Errorf("history --limit 5 stdout = %q — the withheld note leaked onto stdout", out)
	}
}
