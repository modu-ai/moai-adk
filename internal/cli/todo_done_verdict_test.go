// todo_done_verdict_test.go — SPEC-TODO-TRANSITION-STAMPS-001 M3 acceptance:
// the done-time landing verdict record (AC-TST-005/006/007) and its
// coexistence rule (REQ-TST-009). Every scenario drives the real `todo done`
// verb against a fixture queue whose git repository carries a controlled
// landed ref, and reads the archived row back through the store.
package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// seedLandedRepo turns the fixture repository into one whose recorded
// integration ref answers `landed` for the addressed card: a branch `landed`
// carrying a commit whose subject attributes t1 through the axis-F
// trailing-parenthetical form, with refs/remotes/origin/landed tracking it
// and refs/remotes/origin/HEAD pointed there — the level-2 link of the
// landed-ref resolution chain, so the verb and the test resolve the SAME
// ref (`origin/landed`).
func seedLandedRepo(t *testing.T, root string) {
	t.Helper()
	runGitIn(t, root, "commit", "--allow-empty", "-q", "-m", "feat: work the card (t1)")
	runGitIn(t, root, "branch", "landed")
	runGitIn(t, root, "update-ref", "refs/remotes/origin/landed", "landed")
	runGitIn(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/landed")
}

// landedRefName returns the ref the resolution chain answers for the fixture
// root — the same resolution the verb performs (level 2 here).
func landedRefName(t *testing.T, root string) string {
	t.Helper()
	ref, level := factory.LandedRefForWithLevel(root)
	if level != factory.LandedRefOriginHEAD || ref != "origin/landed" {
		t.Fatalf("fixture ref resolution = (%q, level %d), want (origin/landed, level 2) — the seeded control is broken", ref, level)
	}
	return ref
}

// TestDoneVerdict_PersistedWithRefAndTime — AC-TST-005. The answering path
// persists verdict + ref + time, readable from the archived row without
// re-running any git command, and the record carries NO SHA.
func TestDoneVerdict_PersistedWithRefAndTime(t *testing.T) {
	root, store := todoFixture(t)
	seedLandedRepo(t, root)
	ref := landedRefName(t, root)

	if _, _, err := runTodo(t, "add", "--pick", "card with a landed commit"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	stdout, _, err := runTodo(t, "done", "t1", "--require-landed")
	if err != nil {
		t.Fatalf("done --require-landed: %v", err)
	}
	if !strings.Contains(stdout, "landing=landed") || !strings.Contains(stdout, "ref="+ref) {
		t.Fatalf("done line = %q, want landing=landed and ref=%s", strings.TrimSpace(stdout), ref)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Archived) != 1 {
		t.Fatalf("archive holds %d entries, want 1", len(rec.Archived))
	}
	v := rec.Archived[0].LandingVerdict
	if v == nil {
		t.Fatalf("landing_verdict is NULL after a --require-landed done, want a record")
	}
	if v.Verdict != factory.LandingLanded {
		t.Errorf("verdict = %q, want %q", v.Verdict, factory.LandingLanded)
	}
	if v.Ref != ref {
		t.Errorf("ref = %q, want the answering ref %q", v.Ref, ref)
	}
	if v.At == "" {
		t.Errorf("verdict time is empty, want an instant")
	}
	// NO SHA, structurally: the record's own wire keys are exactly the three
	// the requirement names — asked of the implementation's own key names,
	// not a transcription.
	raw, err := json.Marshal(*v)
	if err != nil {
		t.Fatalf("marshal verdict: %v", err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatalf("unmarshal verdict: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("verdict record carries %d keys (%v), want exactly 3", len(keys), keys)
	}
	for _, key := range []string{factory.LandingVerdictKeyVerdict, factory.LandingVerdictKeyRef, factory.LandingVerdictKeyAt} {
		if _, ok := keys[key]; !ok {
			t.Errorf("verdict record is missing key %q", key)
		}
	}
	for key := range keys {
		if strings.Contains(key, "sha") {
			t.Errorf("verdict record carries a SHA-named key %q — query-derived SHAs are outside the evidence store's write authority", key)
		}
	}
}

// TestDoneVerdict_NotPersistedWithoutFlag — AC-TST-006. Without the flag no
// query runs and the archived row's landing record is NULL, even though the
// printed line honestly reads landing=unknown.
func TestDoneVerdict_NotPersistedWithoutFlag(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "plain close"); err != nil {
		t.Fatalf("add: %v", err)
	}
	stdout, _, err := runTodo(t, "done", "t1")
	if err != nil {
		t.Fatalf("done: %v", err)
	}
	if !strings.Contains(stdout, "landing=unknown") {
		t.Fatalf("done line = %q, want landing=unknown (no query ran)", strings.TrimSpace(stdout))
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := rec.Archived[0].LandingVerdict; got != nil {
		t.Fatalf("landing_verdict = %+v without --require-landed, want NULL (no query ran, so nothing is fabricated)", *got)
	}
}

// TestDoneVerdict_RefusalArchivesNothing — acceptance.md §D.4: the
// not-landed refusal path archives nothing; the live row survives with its
// stamps intact and no archive entry exists to carry a refusal verdict.
func TestDoneVerdict_RefusalArchivesNothing(t *testing.T) {
	root, store := todoFixture(t)
	// The landed branch points at the init commit alone: its subject stream
	// attributes nothing, so --require-landed refuses on positive evidence.
	runGitIn(t, root, "branch", "landed")
	runGitIn(t, root, "update-ref", "refs/remotes/origin/landed", "landed")
	runGitIn(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/landed")
	_ = landedRefName(t, root)

	if _, _, err := runTodo(t, "add", "--pick", "card nothing attributes"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	if _, _, err := runTodo(t, "done", "t1", "--require-landed"); err == nil {
		t.Fatalf("done --require-landed on an unattributed card succeeded, want the refusal")
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 1 || rec.Items[0].State != factory.BacklogStatePicked {
		t.Fatalf("live row did not survive the refusal: %d items, state %s", len(rec.Items), rec.Items[0].State)
	}
	if stampOf(rec.Items[0].PickedAt) == "" {
		t.Fatalf("the live row's picked_at did not survive the refusal")
	}
	if len(rec.Archived) != 0 {
		t.Fatalf("the refusal archived %d entries, want 0", len(rec.Archived))
	}
}

// TestDoneVerdict_RefusedWithoutRefThroughStoreAPI — AC-TST-007. The store
// offers no write path that persists a query verdict with an empty ref:
// attempting it through the store API is refused and the record stands
// unchanged.
func TestDoneVerdict_RefusedWithoutRefThroughStoreAPI(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "card to archive"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Archived[0].LandingVerdict = &factory.LandingVerdict{
			Verdict: factory.LandingLanded,
			Ref:     "",
			At:      "2026-09-29T00:00:00Z",
		}
		return nil
	})
	if err == nil {
		t.Fatalf("the store accepted a verdict without its answering ref — REQ-TST-010's refusal is missing")
	}
	if !strings.Contains(err.Error(), "ref") {
		t.Errorf("the refusal does not name the missing ref: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rec.Archived[0].LandingVerdict != nil {
		t.Fatalf("the refused verdict was persisted anyway: %+v", *rec.Archived[0].LandingVerdict)
	}
}

// TestDoneVerdict_CoexistsWithOperatorEvidence — REQ-TST-009. An
// operator-authored landing evidence recorded earlier and a query verdict
// from a flag-run done meet on the same archived row: both survive, the
// verdict never overwrites the operator's evidence.
func TestDoneVerdict_CoexistsWithOperatorEvidence(t *testing.T) {
	root, store := todoFixture(t)
	seedLandedRepo(t, root)
	ref := landedRefName(t, root)

	if _, _, err := runTodo(t, "add", "--pick", "operator evidence then query"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	// Record operator evidence the way `todo landed` writes it — through the
	// evidence store's own discipline.
	ev := factory.LandingEvidence{
		Ref:        ref,
		RefHead:    "abc1234567890abcdef1234567890abcdef1234",
		ObservedAt: "2026-09-28T10:00:00Z",
		SHA:        "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		SHASource:  factory.LandingSHASourceOperator,
	}
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items[0].Landing = &ev
		return nil
	}); err != nil {
		t.Fatalf("record evidence: %v", err)
	}
	if _, _, err := runTodo(t, "done", "t1", "--require-landed"); err != nil {
		t.Fatalf("done --require-landed: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	entry := rec.Archived[0]
	if entry.Item.Landing == nil {
		t.Fatalf("the operator's evidence did not survive the done — REQ-TST-009's preservation is broken")
	}
	if entry.Item.Landing.SHA != ev.SHA || entry.Item.Landing.SHASource != factory.LandingSHASourceOperator {
		t.Errorf("operator evidence mutated: %+v", *entry.Item.Landing)
	}
	if entry.LandingVerdict == nil || entry.LandingVerdict.Verdict != factory.LandingLanded {
		t.Fatalf("the query verdict is missing beside the operator evidence: %+v", entry.LandingVerdict)
	}
}
