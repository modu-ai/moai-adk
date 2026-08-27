// todo_export_test.go — the downgrade route (SPEC-TODO-SQLITE-001
// AC-TOSQ-011, REQ-TOSQ-016, SC-3).
//
// The export is the ONLY way back to a pre-SQLite release, so "it wrote a
// file" is not the property under test. What matters is that the file a
// downgraded binary will read carries the live queue with full fidelity —
// including the cards added and the state changes made AFTER the cutover,
// which is where a naive export that simply copied the quarantine would
// silently lose work.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// AC-TOSQ-011 / SC-3: seed legacy → migrate → mutate → export → the exported
// file parses through the legacy record shape and matches the live reader
// field for field, findings order included.
func TestTodoExportJSON_RoundTripsThroughTheLegacyShape(t *testing.T) {
	root, store := todoFixture(t)

	// Build a queue that exercises every shape: all three states, a spec id,
	// a finding, and — critically — mutations made AFTER the cutover.
	if _, _, err := runTodo(t, "add", "first card"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "add", "second card"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "add", "third card"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "next", "2", "--spec", "SPEC-EXAMPLE-001"); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if _, _, err := runTodo(t, "drop", "t3", "superseded by the second card"); err != nil {
		t.Fatalf("drop: %v", err)
	}
	if _, _, err := runTodo(t, "relate", "t1", "t2", "--relation", "contains"); err != nil {
		t.Fatalf("relate: %v", err)
	}

	live, err := store.Load()
	if err != nil {
		t.Fatalf("Load live: %v", err)
	}

	out, _, err := runTodo(t, "export-json")
	if err != nil {
		t.Fatalf("export-json: %v", err)
	}
	if !strings.Contains(out, "exported") {
		t.Errorf("export-json stdout = %q, want it to report what it wrote", out)
	}

	// The file a downgraded binary reads is backlog.json at the queue root.
	target := kanban.BacklogPathForRoot(root)
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}

	// Parse it exactly as a legacy reader would: the record shape, nothing
	// engine-aware.
	var exported kanban.BacklogRecord
	if err := json.Unmarshal(raw, &exported); err != nil {
		t.Fatalf("exported file does not parse as a legacy record: %v\n%s", err, raw)
	}

	if exported.Version != live.Version {
		t.Errorf("version = %d, want %d", exported.Version, live.Version)
	}
	if exported.LastSeq != live.LastSeq {
		t.Errorf("last_seq = %d, want %d — a downgraded binary would reuse ids", exported.LastSeq, live.LastSeq)
	}
	if len(exported.Items) != len(live.Items) {
		t.Fatalf("items = %d, want %d", len(exported.Items), len(live.Items))
	}
	for i := range live.Items {
		w, g := live.Items[i], exported.Items[i]
		if w.ID != g.ID || w.Text != g.Text || w.AddedAt != g.AddedAt || w.State != g.State {
			t.Errorf("item %d: exported %+v != live %+v", i, g, w)
		}
		if (w.SpecID == nil) != (g.SpecID == nil) {
			t.Errorf("item %d (%s): spec_id null-shape differs", i, w.ID)
		} else if w.SpecID != nil && *w.SpecID != *g.SpecID {
			t.Errorf("item %d: spec_id %q != %q", i, *g.SpecID, *w.SpecID)
		}
	}
	if len(exported.Findings) != len(live.Findings) {
		t.Fatalf("findings = %d, want %d", len(exported.Findings), len(live.Findings))
	}
	for i := range live.Findings {
		if exported.Findings[i] != live.Findings[i] {
			t.Errorf("finding %d: exported %+v != live %+v", i, exported.Findings[i], live.Findings[i])
		}
	}

	// The mutations made after the cutover are present. An export that copied
	// the pre-migration quarantine would pass every count check above and fail
	// exactly here.
	var picked, dropped bool
	for _, it := range exported.Items {
		if it.State == kanban.BacklogStatePicked && it.SpecID != nil && *it.SpecID == "SPEC-EXAMPLE-001" {
			picked = true
		}
		if it.State == kanban.BacklogStateDropped {
			dropped = true
		}
	}
	if !picked || !dropped {
		t.Errorf("post-cutover state changes missing from the export (picked=%v dropped=%v)", picked, dropped)
	}
	if len(exported.Findings) == 0 {
		t.Error("the post-cutover finding is missing from the export")
	}
}

// REQ-TOSQ-016: the export is a COPY, not a migration back. It must not remove
// the database, must not change what this binary reads, and must leave the
// queue's own behavior untouched.
func TestTodoExportJSON_LeavesTheLiveStoreAuthoritative(t *testing.T) {
	root, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "a card"); err != nil {
		t.Fatalf("add: %v", err)
	}

	if _, _, err := runTodo(t, "export-json"); err != nil {
		t.Fatalf("export-json: %v", err)
	}

	if _, err := os.Stat(store.EnginePath()); err != nil {
		t.Fatalf("the database was removed by the export: %v", err)
	}

	// The next verb still reads and writes the database, not the export. If it
	// had switched to the exported file, this add would issue t1 again.
	item, _, err := store.Add("after export")
	if err != nil {
		t.Fatalf("Add after export: %v", err)
	}
	if item.ID != "t2" {
		t.Fatalf("id after export = %q, want t2 — the database must stay authoritative", item.ID)
	}

	// And the export is now stale by construction, which is correct: it is a
	// point-in-time copy. It still holds exactly one card.
	raw, err := os.ReadFile(kanban.BacklogPathForRoot(root))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	var exported kanban.BacklogRecord
	if err := json.Unmarshal(raw, &exported); err != nil {
		t.Fatalf("parse export: %v", err)
	}
	if len(exported.Items) != 1 {
		t.Errorf("export holds %d cards, want the 1 it captured", len(exported.Items))
	}
}

// The export leaves no temp residue on the success path — the same discipline
// the store's own writes hold.
func TestTodoExportJSON_NoTempResidue(t *testing.T) {
	root, _ := todoFixture(t)
	if _, _, err := runTodo(t, "add", "a card"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "export-json"); err != nil {
		t.Fatalf("export-json: %v", err)
	}

	entries, err := os.ReadDir(kanban.StateDirForRoot(root))
	if err != nil {
		t.Fatalf("read state dir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") || strings.HasPrefix(e.Name(), ".backlog-export-") {
			t.Errorf("export left residue %q", e.Name())
		}
	}
	if _, err := os.Stat(filepath.Join(kanban.StateDirForRoot(root), "backlog.json")); err != nil {
		t.Errorf("the export target is missing: %v", err)
	}
}

// REQ-TOSQ-013 / REQ-TOSQ-016: the export must SURVIVE subsequent verbs.
//
// A database beside a backlog.json is ambiguous — it is either an interrupted
// migration's leftover legacy file or a freshly written downgrade export — and
// getting that wrong in the quarantine direction silently deletes the only
// artifact a downgraded release can read. The operator would not find out until
// the older binary came up to an empty queue.
func TestTodoExportJSON_SurvivesSubsequentVerbs(t *testing.T) {
	root, _ := todoFixture(t)
	if _, _, err := runTodo(t, "add", "a card"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "export-json"); err != nil {
		t.Fatalf("export-json: %v", err)
	}
	target := kanban.BacklogPathForRoot(root)
	before, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}

	// Every verb shape: a read, a write, and another read.
	if _, _, err := runTodo(t, "list"); err != nil {
		t.Fatalf("list: %v", err)
	}
	if _, _, err := runTodo(t, "add", "another card"); err != nil {
		t.Fatalf("add after export: %v", err)
	}
	if _, _, err := runTodo(t, "list"); err != nil {
		t.Fatalf("list after add: %v", err)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("the export was renamed or removed by a later verb: %v", err)
	}
	if string(before) != string(after) {
		t.Error("the export's contents changed under a later verb; it is a point-in-time copy")
	}
	if _, err := os.Stat(target + ".migrated"); err == nil {
		t.Error("the export was quarantined as if it were pre-cutover legacy")
	}
}
