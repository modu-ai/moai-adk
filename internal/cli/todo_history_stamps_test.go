// todo_history_stamps_test.go — SPEC-TODO-TRANSITION-STAMPS-001 M4
// acceptance: the archive time axis on `todo history` (AC-TST-008) and the
// live WIP age on the queue surfaces (AC-TST-009).
//
// The line shape contract (todo_history.go's header) is extension-by-append:
// the pre-existing fields keep their order and content byte for byte, and
// the new fields ride BEFORE the card text, which stays last.
package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// historyFields splits one tab-separated history line.
func historyFields(line string) []string {
	return strings.Split(strings.TrimRight(line, "\n"), "\t")
}

// seedStampedArchive builds a queue with one archived card that carries the
// full time axis (picked → done) and one archived with none (add → done).
func seedStampedArchive(t *testing.T) (*kanban.BacklogStore, string) {
	t.Helper()
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "--pick", "stamped closed card"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done t1: %v", err)
	}
	if _, _, err := runTodo(t, "add", "unstampeded closed card"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "done", "t2"); err != nil {
		t.Fatalf("done t2: %v", err)
	}
	return store, ""
}

// TestHistoryArchivedRowExposesTimeAxis — AC-TST-008, lookup. The archived
// line appends picked_at, dropped_at, archived_at, and the verdict record as
// tab-separated fields before the card text; an unstampeded card renders `-`
// in the three stamp fields; the pre-existing fields are byte-for-byte
// unchanged.
func TestHistoryArchivedRowExposesTimeAxis(t *testing.T) {
	store, _ := seedStampedArchive(t)
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	stamped := rec.Archived[0]

	out, _, err := runTodo(t, "history", "t1")
	if err != nil {
		t.Fatalf("history t1: %v", err)
	}
	fields := historyFields(out)
	if len(fields) < 6 {
		t.Fatalf("history line carries %d fields, want the extended shape:\n%s", len(fields), out)
	}
	// Pre-existing fields byte-for-byte: id, fate, state-at-archive, landing
	// cell — the shape the history verb's contract pins.
	if fields[0] != "t1" || fields[1] != "archived" || fields[2] != string(stamped.Item.State) {
		t.Errorf("pre-existing fields drifted: %v", fields[:3])
	}
	if !strings.HasPrefix(fields[3], "landing=") {
		t.Errorf("field 4 = %q, want the landing cell", fields[3])
	}
	// The time axis: picked_at present, dropped_at absent, archived_at present.
	if fields[4] != stampOf(stamped.Item.PickedAt) || fields[4] == "" {
		t.Errorf("picked_at field = %q, want the stamp %q", fields[4], stampOf(stamped.Item.PickedAt))
	}
	if fields[5] != "-" {
		t.Errorf("dropped_at field = %q, want '-' (never picked-with-drop on this path)", fields[5])
	}
	if fields[6] != stampOf(stamped.ArchivedAt) || fields[6] == "" {
		t.Errorf("archived_at field = %q, want the archive stamp %q", fields[6], stampOf(stamped.ArchivedAt))
	}
	// The verdict record cell: no flag ran, so nothing was fabricated.
	if fields[7] != "verdict=-" {
		t.Errorf("verdict field = %q, want 'verdict=-' (no --require-landed ran)", fields[7])
	}
	// Card text stays LAST.
	if !strings.Contains(out, "stamped closed card\n") {
		t.Errorf("the card text is not last on the line:\n%s", out)
	}
}

// TestHistoryArchivedRowAbsentStampsRenderDash — AC-TST-008, the unstampeded
// archived card renders `-` in the picked_at / dropped_at fields. archived_at
// is NOT one of them: every done stamps the archive instant, so it is
// present on both cards by design.
func TestHistoryArchivedRowAbsentStampsRenderDash(t *testing.T) {
	_, _ = seedStampedArchive(t)
	out, _, err := runTodo(t, "history", "t2")
	if err != nil {
		t.Fatalf("history t2: %v", err)
	}
	fields := historyFields(out)
	if len(fields) < 8 {
		t.Fatalf("history line carries %d fields, want the extended shape:\n%s", len(fields), out)
	}
	if fields[4] != "-" || fields[5] != "-" {
		t.Errorf("picked_at/dropped_at fields = %v, want '-' for an unstampeded card", fields[4:6])
	}
	if fields[6] == "-" {
		t.Errorf("archived_at field = '-', want the archive instant every done stamps")
	}
}

// TestHistoryListingExposesTimeAxis — AC-TST-008, listing. Each archived row
// renders the appended fields, newest first, pre-existing fields unchanged.
func TestHistoryListingExposesTimeAxis(t *testing.T) {
	_, _ = seedStampedArchive(t)
	out, _, err := runTodo(t, "history")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("listing carried %d lines, want 2:\n%s", len(lines), out)
	}
	// Newest first.
	if !strings.HasPrefix(lines[0], "t2\t") || !strings.HasPrefix(lines[1], "t1\t") {
		t.Fatalf("listing order drifted: %v", lines)
	}
	for i, line := range lines {
		fields := historyFields(line)
		if len(fields) < 8 {
			t.Errorf("listing row %d carries %d fields, want the extended shape", i, len(fields))
		}
	}
}

// TestHistoryLiveRowExposesPickedAt — AC-TST-009 half 1. A picked live card's
// lookup line renders its picked_at.
func TestHistoryLiveRowExposesPickedAt(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "--pick", "live picked card"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	out, _, err := runTodo(t, "history", "t1")
	if err != nil {
		t.Fatalf("history t1: %v", err)
	}
	fields := historyFields(out)
	if len(fields) < 6 {
		t.Fatalf("live line carries %d fields, want the extended shape:\n%s", len(fields), out)
	}
	if fields[1] != "live" || fields[2] != string(kanban.BacklogStatePicked) {
		t.Errorf("pre-existing live fields drifted: %v", fields[:3])
	}
	if fields[4] != stampOf(rec.Items[0].PickedAt) || fields[4] == "" {
		t.Errorf("picked_at field = %q, want the live stamp %q", fields[4], stampOf(rec.Items[0].PickedAt))
	}
	if fields[5] != "-" {
		t.Errorf("dropped_at field = %q, want '-'", fields[5])
	}
}

// TestListJSONStampsOmitEmpty — AC-TST-009 half 2. A picked card's JSON
// object carries a non-empty picked_at; a never-picked card's object omits
// the key entirely.
func TestListJSONStampsOmitEmpty(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "--pick", "picked json card"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	if _, _, err := runTodo(t, "add", "never picked json card"); err != nil {
		t.Fatalf("add: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	out, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	var rendered struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rendered); err != nil {
		t.Fatalf("parse list --json: %v\n%s", err, out)
	}
	if len(rendered.Items) != len(rec.Items) {
		t.Fatalf("list --json rendered %d items, want %d", len(rendered.Items), len(rec.Items))
	}
	picked, never := rendered.Items[0], rendered.Items[1]
	if raw, ok := picked["picked_at"]; !ok {
		t.Errorf("the picked card's JSON object omits picked_at, want the key present")
	} else {
		var v string
		if err := json.Unmarshal(raw, &v); err != nil || v == "" {
			t.Errorf("picked_at = %s, want a non-empty timestamp", raw)
		}
	}
	if _, ok := never["picked_at"]; ok {
		t.Errorf("the never-picked card's JSON object carries picked_at, want it omitted (omitempty)")
	}
	if _, ok := never["dropped_at"]; ok {
		t.Errorf("the never-picked card's JSON object carries dropped_at, want it omitted (omitempty)")
	}
}
