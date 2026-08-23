// todo_analysis_test.go — acceptance tests for the queue's analysis layer:
// the mechanical duplicate analyser, the findings record, the agent-layer
// relation verbs, and the operator visibility surface.
//
// Every test runs on a t.TempDir() fixture. The operator's live queue is
// never a test subject: an earlier revision of this feature damaged a real
// card by reaching the project queue, which is why todoFixture (a git repo
// in a temp dir + CLAUDE_PROJECT_DIR) is the only seam these tests use.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedFindings writes findings into the fixture queue directly through the
// store, so a test can construct a Given state the CLI verbs would need
// several invocations to reach.
func seedFindings(t *testing.T, store *kanban.BacklogStore, findings ...kanban.BacklogFinding) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Findings = append(rec.Findings, findings...)
		return nil
	}); err != nil {
		t.Fatalf("seed findings: %v", err)
	}
}

// loadFindings returns the fixture queue's current findings.
func loadFindings(t *testing.T, store *kanban.BacklogStore) []kanban.BacklogFinding {
	t.Helper()
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return rec.Findings
}

// TestTodoDoneReclaimsFindings — AC-TA-008 (REQ-TA-007): removing a card
// removes every finding naming it, in either position, and nothing else.
// The length is asserted exactly, so under-reclaiming (a finding outliving
// its subject) and over-reclaiming (an unrelated finding swept away) both
// fail.
func TestTodoDoneReclaimsFindings(t *testing.T) {
	_, store := todoFixture(t)
	for _, text := range []string{"card one", "card two", "card three", "card four"} {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("seed add %q: %v", text, err)
		}
	}
	seedFindings(t, store,
		kanban.BacklogFinding{SubjectID: "t1", RelatedID: "t2", Relation: kanban.BacklogRelationContains, Source: kanban.BacklogSourceAgent},
		kanban.BacklogFinding{SubjectID: "t3", RelatedID: "t4", Relation: kanban.BacklogRelationContains, Source: kanban.BacklogSourceAgent},
	)

	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done t1: %v", err)
	}

	got := loadFindings(t, store)
	if len(got) != 1 {
		t.Fatalf("findings after done = %d, want exactly 1: %+v", len(got), got)
	}
	if got[0].SubjectID != "t3" || got[0].RelatedID != "t4" {
		t.Errorf("surviving finding = %s->%s, want t3->t4", got[0].SubjectID, got[0].RelatedID)
	}
	for _, f := range got {
		if f.SubjectID == "t1" || f.RelatedID == "t1" {
			t.Errorf("finding still names the removed card t1: %+v", f)
		}
	}
}

// TestTodoLegacyRecordRoundTrips — AC-TA-012 (REQ-TA-006): a queue file
// written before this feature loads unchanged, `findings` renders as an
// empty array rather than null or an omitted key, and the per-item contract
// stays exactly five fields.
//
// The item key set is counted DIRECTLY rather than inferred from a
// successful decode: encoding/json silently ignores unknown fields, so
// "decodes into the old struct without error" cannot detect a field that was
// ADDED to an item. The same reasoning rules out applying the check to the
// whole record — the record legitimately gained the additive `findings` key.
func TestTodoLegacyRecordRoundTrips(t *testing.T) {
	root, store := todoFixture(t)
	legacy := `{
  "version": 1,
  "last_seq": 2,
  "items": [
    {"id": "t1", "text": "legacy card one", "added_at": "2026-01-01T00:00:00Z", "spec_id": null, "state": "queued"},
    {"id": "t2", "text": "legacy card two", "added_at": "2026-01-02T00:00:00Z", "spec_id": "SPEC-X-001", "state": "picked"}
  ]
}
`
	writeQueueFile(t, root, legacy)

	out, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json over a legacy file: %v", err)
	}
	var rendered map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out), &rendered); err != nil {
		t.Fatalf("parse list --json: %v (out=%q)", err, out)
	}
	if string(rendered["findings"]) != "[]" {
		t.Errorf("findings rendered as %s, want an empty array", rendered["findings"])
	}
	var roundTripped []kanban.BacklogItem
	if err := json.Unmarshal(rendered["items"], &roundTripped); err != nil {
		t.Fatalf("parse items: %v", err)
	}
	if len(roundTripped) != 2 || roundTripped[0].Text != "legacy card one" ||
		roundTripped[1].SpecID == nil || *roundTripped[1].SpecID != "SPEC-X-001" {
		t.Errorf("legacy items did not round-trip: %+v", roundTripped)
	}

	if _, _, err := runTodo(t, "add", "new card"); err != nil {
		t.Fatalf("add over a legacy file: %v", err)
	}

	raw, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatalf("read queue file: %v", err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatalf("parse queue file: %v", err)
	}
	if _, ok := top["findings"]; !ok {
		t.Error("queue file lost the additive top-level findings key")
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(top["items"], &items); err != nil {
		t.Fatalf("parse items: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("items = %d, want 3", len(items))
	}
	want := map[string]bool{"id": true, "text": true, "added_at": true, "spec_id": true, "state": true}
	for i, it := range items {
		if len(it) != len(want) {
			t.Errorf("item %d has %d keys, want exactly %d: %v", i, len(it), len(want), todoJSONKeys(it))
		}
		for k := range it {
			if !want[k] {
				t.Errorf("item %d carries an out-of-contract key %q", i, k)
			}
		}
	}
}

// writeQueueFile plants raw JSON at the fixture queue's path.
func writeQueueFile(t *testing.T, root, content string) {
	t.Helper()
	path := todoBacklogPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir queue dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write queue file: %v", err)
	}
}

// todoJSONKeys renders a JSON object's key set for a failure message.
func todoJSONKeys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
