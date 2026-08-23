// todo_analysis_add_test.go — acceptance tests for the mechanical analyser's
// two entry points, `add` and `analyze`.
//
// Every test runs on a t.TempDir() fixture; the operator's live queue is
// never a subject.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// queueDigest returns the SHA-256 of the fixture queue file. File-invariance
// claims are judged on this digest: "looks the same" is not evidence.
func queueDigest(t *testing.T, store *kanban.BacklogStore) string {
	t.Helper()
	raw, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatalf("read queue file: %v", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// seedItems writes cards straight through the store, bypassing `add` and
// therefore its analysis. A test that needs a queue holding a duplicate pair
// but NO findings yet cannot build that state through the CLI.
func seedItems(t *testing.T, store *kanban.BacklogStore, texts ...string) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for _, text := range texts {
			rec.LastSeq++
			rec.Items = append(rec.Items, kanban.BacklogItem{
				ID:      fmt.Sprintf("t%d", rec.LastSeq),
				Text:    text,
				AddedAt: "2026-01-01T00:00:00Z",
				State:   kanban.BacklogStateQueued,
			})
		}
		return nil
	}); err != nil {
		t.Fatalf("seed items: %v", err)
	}
}

// itemIDs returns the fixture queue's ids in file order.
func itemIDs(t *testing.T, store *kanban.BacklogStore) []string {
	t.Helper()
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	ids := make([]string, 0, len(rec.Items))
	for _, it := range rec.Items {
		ids = append(ids, it.ID)
	}
	return ids
}

// snapshotItems returns the fixture queue's items as JSON, for a later
// byte-comparison.
func snapshotItems(t *testing.T, store *kanban.BacklogStore) []byte {
	t.Helper()
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	raw, err := json.Marshal(rec.Items)
	if err != nil {
		t.Fatalf("marshal items: %v", err)
	}
	return raw
}

// assertItemsUnchanged compares the fixture queue's items against a snapshot
// taken earlier.
func assertItemsUnchanged(t *testing.T, store *kanban.BacklogStore, snapshot []byte, when string) {
	t.Helper()
	got := snapshotItems(t, store)
	if string(got) != string(snapshot) {
		t.Errorf("items changed %s:\n want %s\n  got %s", when, snapshot, got)
	}
}

// TestTodoAddRefusesExactDuplicate — AC-TA-001 (REQ-TA-001, REQ-TA-003): an
// exact duplicate is refused, the refusal names the colliding card, and the
// queue file is left byte-identical.
//
// The candidate differs from the seeded card in case and whitespace only, so
// an implementation that compares raw text classifies it as `none` and admits
// the card — the failure this test exists to catch.
func TestTodoAddRefusesExactDuplicate(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "Fix the flaky gate"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	before := queueDigest(t, store)

	out, errOut, err := runTodo(t, "add", "  fix   the FLAKY gate ")
	if err == nil {
		t.Fatalf("exact duplicate was admitted (stdout=%q), want a refusal", out)
	}
	if !strings.Contains(errOut, "t1") {
		t.Errorf("stderr does not name the colliding card t1:\n%s", errOut)
	}
	if !strings.Contains(errOut, "Fix the flaky gate") {
		t.Errorf("stderr does not carry the colliding card's text:\n%s", errOut)
	}
	if after := queueDigest(t, store); after != before {
		t.Errorf("queue file changed on a refused add: %s -> %s", before, after)
	}
}

// TestTodoAddForceAdmitsAndRecords — AC-TA-002 (REQ-TA-004): `--force`
// admits the card and leaves exactly one trace of the forced duplicate.
func TestTodoAddForceAdmitsAndRecords(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "Fix the flaky gate"); err != nil {
		t.Fatalf("seed add: %v", err)
	}

	if _, _, err := runTodo(t, "add", "fix the flaky gate", "--force"); err != nil {
		t.Fatalf("forced add: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(rec.Items))
	}
	forced := 0
	for _, f := range rec.Findings {
		if f.Relation == kanban.BacklogRelationDuplicateForced {
			forced++
			if f.RelatedID != "t1" || f.Source != kanban.BacklogSourceMechanical {
				t.Errorf("forced finding = %+v, want related t1 from the mechanical source", f)
			}
		}
	}
	if forced != 1 {
		t.Errorf("duplicate-forced findings = %d, want exactly 1: %+v", forced, rec.Findings)
	}
}

// TestTodoAddNearDuplicateRecordsOnly — AC-TA-003 (REQ-TA-001, REQ-TA-005):
// a near duplicate is admitted verbatim, the related card is untouched, one
// finding is recorded — and a card that resembles nothing records none.
//
// The negative half is what makes the positive half meaningful. A test
// asserting only "one finding for the near pair" would pass an
// implementation that records a finding for every pair; a test asserting
// only "no finding for the unrelated card" would pass an implementation with
// no analyser at all.
func TestTodoAddNearDuplicateRecordsOnly(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "Rework the auth middleware error paths"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	before, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	snapshot := before.Items[0]

	const nearText = "Rework auth middleware error paths"
	if _, _, err := runTodo(t, "add", nearText); err != nil {
		t.Fatalf("near-duplicate add: %v", err)
	}
	if _, _, err := runTodo(t, "add", "Add a Windows CI matrix job"); err != nil {
		t.Fatalf("unrelated add: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rec.Items[0] != snapshot {
		t.Errorf("the related card changed: %+v -> %+v", snapshot, rec.Items[0])
	}
	if rec.Items[1].Text != nearText {
		t.Errorf("the new card's text = %q, want the argument verbatim %q", rec.Items[1].Text, nearText)
	}

	near, unrelated := 0, 0
	for _, f := range rec.Findings {
		switch {
		case f.Names("t2") && f.Names("t1"):
			near++
			if f.Relation != kanban.BacklogRelationNearDuplicate {
				t.Errorf("t1/t2 finding relation = %q, want near-duplicate", f.Relation)
			}
			if f.Score < kanban.BacklogNearDuplicateThreshold {
				t.Errorf("t1/t2 score = %v, want >= %v", f.Score, kanban.BacklogNearDuplicateThreshold)
			}
		case f.Names("t3"):
			unrelated++
		}
	}
	if near != 1 {
		t.Errorf("near-duplicate findings for t1/t2 = %d, want exactly 1: %+v", near, rec.Findings)
	}
	if unrelated != 0 {
		t.Errorf("findings naming the unrelated card t3 = %d, want 0: %+v", unrelated, rec.Findings)
	}
}

// TestTodoAnalysisNeverReordersQueue — AC-TA-006 (REQ-TA-009): analysis
// leaves the queue's order alone. Order is the only signal the queue records
// about priority, and a reordered queue is indistinguishable, on the
// operator's ordinary screen, from one they ordered themselves — which is
// why reordering is forbidden outright rather than merely made reversible.
func TestTodoAnalysisNeverReordersQueue(t *testing.T) {
	_, store := todoFixture(t)
	for _, text := range []string{
		"Rework the auth middleware error paths",
		"Add a Windows CI matrix job",
		"Document the release checklist",
	} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("seed add %q: %v", text, err)
		}
	}

	adds := []struct {
		text  string
		force bool
	}{
		{"Rework auth middleware error paths", false}, // near of t1
		{"Add Windows CI matrix job", false},          // near of t2
		{"Document the release checklist", true},      // exact of t3, forced
		{"Ship the telemetry opt-out", false},
		{"Prune stale worktrees", false},
	}
	for _, a := range adds {
		args := []string{"add", a.text}
		if a.force {
			args = append(args, "--force")
		}
		if _, _, err := runTodo(t, args...); err != nil {
			t.Fatalf("add %q: %v", a.text, err)
		}
	}

	got := itemIDs(t, store)
	want := []string{"t1", "t2", "t3", "t4", "t5", "t6", "t7", "t8"}
	if len(got) != len(want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ids = %v, want %v — the seeded three keep positions 1-3 and the rest append in order",
				got, want)
		}
	}
}

// TestTodoExactRefusalWorksWithoutAgent — AC-TA-015 (REQ-TA-001,
// REQ-TA-003): the mechanical layer is a property of the CLI, not of any
// caller. A textual duplicate is refused; a card that means the same thing
// in different words is admitted with no finding, because judging THAT is
// beyond the mechanical layer's reach and a false positive there would
// discard a legitimate card.
func TestTodoExactRefusalWorksWithoutAgent(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "Fix the auth bug"); err != nil {
		t.Fatalf("seed add: %v", err)
	}

	if _, _, err := runTodo(t, "add", "fix the auth bug"); err == nil {
		t.Error("a textual duplicate was admitted with no agent in the loop, want a refusal")
	}
	if _, _, err := runTodo(t, "add", "Repair broken login"); err != nil {
		t.Fatalf("a semantically similar but textually distinct card was refused: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 2 {
		t.Errorf("items = %d, want 2", len(rec.Items))
	}
	if len(rec.Findings) != 0 {
		t.Errorf("findings = %d, want 0 — the mechanical layer must not guess at meaning: %+v",
			len(rec.Findings), rec.Findings)
	}
}

// TestTodoAnalyzeRerunIsIdempotent — AC-TA-016 (REQ-TA-002, REQ-TA-009):
// analyze records on the first run and records nothing new on the second,
// and it never touches a card.
//
// The "more than zero after the first run" assertion is the positive
// control: without it, an implementation that records nothing at all would
// satisfy "the length did not change" trivially.
func TestTodoAnalyzeRerunIsIdempotent(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store,
		"Rework the auth middleware error paths",
		"Rework auth middleware error paths",
		"Ship the telemetry opt-out",
	)
	before, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(before.Findings) != 0 {
		t.Fatalf("fixture already carries findings: %+v", before.Findings)
	}
	snapshot := snapshotItems(t, store)

	if _, _, err := runTodo(t, "analyze"); err != nil {
		t.Fatalf("first analyze: %v", err)
	}
	first, err := store.Load()
	if err != nil {
		t.Fatalf("load after the first analyze: %v", err)
	}
	if len(first.Findings) == 0 {
		t.Fatal("the first analyze recorded nothing, so the idempotence claim below would be vacuous")
	}
	assertItemsUnchanged(t, store, snapshot, "after the first analyze")

	if _, _, err := runTodo(t, "analyze"); err != nil {
		t.Fatalf("second analyze: %v", err)
	}
	second, err := store.Load()
	if err != nil {
		t.Fatalf("load after the second analyze: %v", err)
	}
	if len(second.Findings) != len(first.Findings) {
		t.Errorf("findings grew on re-analysis: %d -> %d", len(first.Findings), len(second.Findings))
	}
	assertItemsUnchanged(t, store, snapshot, "after the second analyze")

	seen := map[string]int{}
	for _, f := range second.Findings {
		seen[f.SubjectID+"|"+f.RelatedID+"|"+f.Relation+"|"+f.Source]++
	}
	for key, n := range seen {
		if n > 1 {
			t.Errorf("tuple %q recorded %d times, want at most 1", key, n)
		}
	}
}
