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
