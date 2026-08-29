// todo_landing_test.go — SPEC-TODO-LANDING-STATE-001 AC-TLS-004, AC-TLS-005,
// AC-TLS-006, AC-TLS-007, AC-TLS-010 (M2, M3, M4).
//
// One thread runs through every criterion here: an answer nobody could obtain
// must not render as an answer somebody obtained. Before this SPEC, a `todo pr`
// row for a card whose landing query FAILED was byte-identical to the row for a
// card genuinely nobody had started, and `todo done --require-landed` printed
// `done <id>` and exited 0 whether the guard passed or never ran.
package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// prRow splits one rendered `todo pr` line into its tab-separated columns.
func prRow(t *testing.T, stdout, cardID string) []string {
	t.Helper()
	for _, line := range strings.Split(strings.TrimRight(stdout, "\n"), "\n") {
		cols := strings.Split(line, "\t")
		if len(cols) > 0 && cols[0] == cardID {
			return cols
		}
	}
	t.Fatalf("no row for %s in:\n%s", cardID, stdout)
	return nil
}

// AC-TLS-005 — an unanswerable landing query renders an outcome DISTINCT from
// no-link, and the degradation note names the ref that could not be reached.
//
// The control is the whole criterion: a second card in the same queue whose
// query is answerable renders `no-link`, so the assertion cannot be satisfied
// by uniformly renaming the outcome.
func TestTodoPR_UnanswerableRendersUnknownNotNoLink(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "unanswerable card", "genuinely unstarted card")
	installSpy(t, &spyRunner{
		prJSON:  `[]`,
		gitFail: map[string]error{ids[0]: fmt.Errorf("fatal: ambiguous argument 'origin/develop': unknown revision")},
	})

	stdout, stderr, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr: %v", err)
	}

	unanswerable := prRow(t, stdout, ids[0])
	answered := prRow(t, stdout, ids[1])
	if unanswerable[1] != string(kanban.PRLinkUnknown) {
		t.Errorf("outcome for the unanswerable card = %q, want %q", unanswerable[1], kanban.PRLinkUnknown)
	}
	if answered[1] != string(kanban.PRLinkNoLink) {
		t.Errorf("outcome for the answered card = %q, want %q", answered[1], kanban.PRLinkNoLink)
	}
	if unanswerable[1] == answered[1] {
		t.Error("the two rows render the same outcome — the false negative this SPEC closes")
	}
	// The note names the ref, so a reader learns WHICH question went
	// unanswered rather than only that one did.
	if !strings.Contains(stderr, todoLandedRef()) {
		t.Errorf("degradation note %q does not name the resolved ref %s", stderr, todoLandedRef())
	}
	// Exit code stays 0: the ruling is fail-open, and this SPEC repairs the
	// reporting axis, not the policy axis.
	if strings.Contains(stdout, "Error:") {
		t.Errorf("stdout %q reports an error; the render is fail-open", stdout)
	}
}

// AC-TLS-005 (JSON half) — the new kind reaches the machine-readable surface
// too, so a consumer of `todo pr --json` sees the same distinction the text
// render makes.
func TestTodoPR_UnknownReachesJSON(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "unanswerable card")
	installSpy(t, &spyRunner{
		prJSON:  `[]`,
		gitFail: map[string]error{ids[0]: fmt.Errorf("fatal: bad revision")},
	})

	stdout, _, err := runTodo(t, "pr", "--json")
	if err != nil {
		t.Fatalf("todo pr --json: %v", err)
	}
	if !strings.Contains(stdout, `"outcome":"`+string(kanban.PRLinkUnknown)+`"`) {
		t.Errorf("json %q does not carry the unknown outcome", stdout)
	}
}

// AC-TLS-010 — the row carries the queue state, so a `picked` card with no
// commits (the t338 shape) is distinguishable from a `queued`, never-started
// one. Both resolve to `no-link`; before this column they rendered alike.
//
// Stated limit, asserted as part of the criterion: this distinguishes *picked
// with no commits* from *never started*. It does NOT claim to detect work that
// produced no commit.
func TestTodoPR_RowCarriesQueueState(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "picked but no commits", "never started")
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == ids[0] {
				rec.Items[i].State = kanban.BacklogStatePicked
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("pick %s: %v", ids[0], err)
	}
	installSpy(t, &spyRunner{prJSON: `[]`})

	stdout, _, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr: %v", err)
	}
	picked := prRow(t, stdout, ids[0])
	queued := prRow(t, stdout, ids[1])

	// Six columns: CardID, Kind, PRs, Confidence, State, text.
	if len(picked) != 6 {
		t.Fatalf("row %v has %d columns, want 6 (the state column was added)", picked, len(picked))
	}
	if picked[1] != queued[1] {
		t.Fatalf("fixture premise broken: the two cards must share the %q outcome, got %q and %q",
			kanban.PRLinkNoLink, picked[1], queued[1])
	}
	if picked[4] != string(kanban.BacklogStatePicked) {
		t.Errorf("state column for the picked card = %q, want %q", picked[4], kanban.BacklogStatePicked)
	}
	if queued[4] != string(kanban.BacklogStateQueued) {
		t.Errorf("state column for the queued card = %q, want %q", queued[4], kanban.BacklogStateQueued)
	}
	// The card text stays the LAST field, so a consumer reading the tail
	// still reads the text after the column count changed.
	if picked[5] != "picked but no commits" {
		t.Errorf("last column = %q, want the card text", picked[5])
	}
}

// AC-TLS-006 — `done` prints exactly one landing verdict, and the three
// answers produce three DISTINCT tokens. This is the criterion whose RED is
// that "the guard passed" and "the guard did not run" were the same bytes.
func TestTodoDone_StdoutCarriesTheLandingVerdict(t *testing.T) {
	cases := []struct {
		name string
		args []string
		out  string
		err  error
		want kanban.LandingAnswer
	}{
		{"landed", []string{"done", "t1", "--require-landed"}, "abc1234 fix: something (t1)\n", nil, kanban.LandingLanded},
		{"unanswerable", []string{"done", "t1", "--require-landed"}, "", fmt.Errorf("fatal: bad revision"), kanban.LandingUnknown},
		{"no query at all", []string{"done", "t1"}, "", nil, kanban.LandingUnknown},
	}
	seen := map[string]string{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			todoFixture(t)
			seedTodo(t, "alpha work")
			stubLandingQuery(t, tc.out, tc.err)

			stdout, stderr, err := runTodo(t, tc.args...)
			if err != nil {
				t.Fatalf("todo %v: %v (stderr %q)", tc.args, err, stderr)
			}
			want := "landing=" + string(tc.want)
			if !strings.Contains(stdout, want) {
				t.Errorf("stdout = %q, want it to carry %q", stdout, want)
			}
			if n := strings.Count(stdout, "landing="); n != 1 {
				t.Errorf("stdout = %q carries %d landing verdicts, want exactly 1", stdout, n)
			}
			// The archive still goes through and the prefix is preserved —
			// the report changed, the act did not.
			if !strings.HasPrefix(stdout, "done t1") {
				t.Errorf("stdout = %q, want the archive confirmation to keep its prefix", stdout)
			}
			seen[tc.name] = strings.TrimSpace(stdout)
		})
	}
	// The distinctness half. Without it, three tokens that happen to be equal
	// would satisfy every assertion above.
	if seen["landed"] == seen["unanswerable"] {
		t.Errorf("a satisfied guard and an unanswerable one print the same bytes (%q) — the defect this SPEC closes", seen["landed"])
	}
}

// AC-TLS-007 (regression-guard) — the permissive policy is unchanged: an
// `unknown` answer still archives the card and still exits 0. This SPEC
// repairs the reporting axis, never the policy axis.
func TestTodoDone_UnknownStillProceeds(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "alpha work")
	stubLandingQuery(t, "", fmt.Errorf("fatal: bad revision 'origin/develop'"))

	stdout, stderr, err := runTodo(t, "done", "t1", "--require-landed")
	if err != nil {
		t.Fatalf("an unknown landing answer must proceed, got %v (stderr %q)", err, stderr)
	}
	if !strings.Contains(stdout, "landing="+string(kanban.LandingUnknown)) {
		t.Errorf("stdout = %q, want the unknown verdict named", stdout)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, it := range rec.Items {
		if it.ID == "t1" {
			t.Errorf("card t1 is still in the live queue; --require-landed must archive on unknown")
		}
	}
}
