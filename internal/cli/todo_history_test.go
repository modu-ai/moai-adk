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
	"strings"
	"testing"
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
