// todo_list_limit_test.go — SPEC-TODO-SURFACE-POLISH-001 M1 (card t1349): the
// list render's limit contract and the render-completeness invariant.
//
// AC-TSP-020 raises the default bound (20 → 100) while keeping the withheld
// notice; AC-TSP-021 pins the surviving contract (--limit 0 unbounded, --json
// ignores the limit, --limit <0 refused, --dropped view); AC-TSP-030 fixes
// the invariant "the rendered id set is EXACTLY the filter-satisfying row
// set" as a property test, so a row the store holds can never silently
// vanish from the render again (the t1338 observation's structural block).
package cli

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoListCardIDShaped matches the id form the queue issues, so the
// completeness assertions read only render rows and skip count lines.
var todoListCardIDShaped = regexp.MustCompile(`^t\d+$`)

// todoListRenderedIDs extracts the card ids a list render put on stdout.
func todoListRenderedIDs(stdout string) map[string]bool {
	ids := map[string]bool{}
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) > 0 && todoListCardIDShaped.MatchString(fields[0]) {
			ids[fields[0]] = true
		}
	}
	return ids
}

// seedTodoCards appends n cards directly through the store — the seeder the
// limit fixtures use (CLI adds would run the analyser n times for no
// additional contract).
func seedTodoCards(t *testing.T, store *kanban.BacklogStore, n int) {
	t.Helper()
	for i := 1; i <= n; i++ {
		if _, _, err := store.Add(fmt.Sprintf("limit fixture card %d", i)); err != nil {
			t.Fatalf("seed card %d: %v", i, err)
		}
	}
}

// AC-TSP-020 — the default render shows a 60-row live queue in full with no
// withheld notice, and a 120-row queue renders exactly 100 rows with the
// withheld line naming the remainder on stderr.
func TestTodoListDefaultLimit(t *testing.T) {
	_, store := todoFixture(t)
	seedTodoCards(t, store, 60)

	out, errOut, err := runTodo(t) // bare todo renders with the default limit
	if err != nil {
		t.Fatalf("bare todo: %v (stderr %q)", err, errOut)
	}
	if ids := todoListRenderedIDs(out); len(ids) != 60 {
		t.Errorf("bare todo rendered %d ids, want all 60 (stdout %q...)", len(ids), truncateForTest(out, 200))
	}
	if strings.Contains(errOut, "withheld") {
		t.Errorf("bare todo stderr = %q, want no withheld notice at 60 rows under the 100 default", errOut)
	}

	_, store2 := todoFixture(t)
	seedTodoCards(t, store2, 120)

	out, errOut, err = runTodo(t, "list")
	if err != nil {
		t.Fatalf("list: %v (stderr %q)", err, errOut)
	}
	if ids := todoListRenderedIDs(out); len(ids) != 100 {
		t.Errorf("list rendered %d ids, want exactly the 100-row bound", len(ids))
	}
	want := "list: 20 rows withheld — showing 100 of 120 (--limit 0 lists all)"
	if !strings.Contains(errOut, want) {
		t.Errorf("list stderr = %q, want the withheld line %q", errOut, want)
	}
}

// AC-TSP-021 — the surviving limit contract, each clause judged exactly as
// before the default moved.
func TestTodoListLimitZero(t *testing.T) {
	_, store := todoFixture(t)
	seedTodoCards(t, store, 30)

	out, errOut, err := runTodo(t, "list", "--limit", "0")
	if err != nil {
		t.Fatalf("list --limit 0: %v", err)
	}
	if ids := todoListRenderedIDs(out); len(ids) != 30 {
		t.Errorf("list --limit 0 rendered %d ids, want all 30", len(ids))
	}
	if strings.Contains(errOut, "withheld") {
		t.Errorf("list --limit 0 stderr = %q, want no withheld notice", errOut)
	}
}

func TestTodoListJSONIgnoresLimit(t *testing.T) {
	_, store := todoFixture(t)
	seedTodoCards(t, store, 25)

	out, _, err := runTodo(t, "list", "--json", "--limit", "1")
	if err != nil {
		t.Fatalf("list --json --limit 1: %v", err)
	}
	var rec kanban.BacklogRecord
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatalf("list --json output is not a BacklogRecord: %v (stdout %q...)", err, truncateForTest(out, 120))
	}
	if len(rec.Items) != 25 {
		t.Errorf("list --json returned %d items, want all 25 (--json ignores the limit)", len(rec.Items))
	}
}

func TestTodoListNegativeLimit(t *testing.T) {
	_, _ = todoFixture(t)

	_, _, err := runTodo(t, "list", "--limit", "-1")
	if err == nil {
		t.Fatalf("list --limit -1: want the refusal, got nil")
	}
	if !strings.Contains(err.Error(), "--limit must be >= 0 (got -1)") {
		t.Errorf("list --limit -1 error = %v, want the exact todo.go:753 wording", err)
	}
}

func TestTodoListDroppedOnly(t *testing.T) {
	_, store := todoFixture(t)
	seedTodoCards(t, store, 3)
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Items[2].State = kanban.BacklogStateDropped
		return nil
	}); err != nil {
		t.Fatalf("drop surgery: %v", err)
	}

	out, _, err := runTodo(t, "list", "--dropped")
	if err != nil {
		t.Fatalf("list --dropped: %v", err)
	}
	if ids := todoListRenderedIDs(out); len(ids) != 1 || !ids["t3"] {
		t.Errorf("list --dropped rendered %v, want exactly {t3}", ids)
	}

	out, _, err = runTodo(t, "list")
	if err != nil {
		t.Fatalf("default list: %v", err)
	}
	if ids := todoListRenderedIDs(out); len(ids) != 2 || ids["t3"] {
		t.Errorf("default list rendered %v, want {t1,t2} with t3 behind the count line", ids)
	}
	if !strings.Contains(out, "1 dropped (hidden") {
		t.Errorf("default list stdout = %q, want the dropped count line", out)
	}
}

// AC-TSP-030 — the render-completeness invariant: the rendered id set IS the
// filter-satisfying row set, for the default (all live states, hold
// included) and --dropped views alike. A row present in the store but
// missing from the render is a defect regardless of its state.
func TestTodoRenderCompleteness(t *testing.T) {
	_, store := todoFixture(t)
	seedTodoCards(t, store, 5)
	// Spread the five rows across every state the DDL admits.
	states := map[int]kanban.BacklogState{
		1: kanban.BacklogStateQueued,
		2: kanban.BacklogStatePicked,
		3: kanban.BacklogStateDropped,
		4: kanban.BacklogStateHold,
		5: kanban.BacklogStateQueued,
	}
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			n := i + 1
			if st, ok := states[n]; ok {
				rec.Items[i].State = st
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("state surgery: %v", err)
	}

	liveWant := map[string]bool{"t1": true, "t2": true, "t4": true, "t5": true}
	droppedWant := map[string]bool{"t3": true}

	out, errOut, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("default list: %v (stderr %q)", err, errOut)
	}
	got := todoListRenderedIDs(out)
	if len(got) != len(liveWant) {
		t.Errorf("default render id set = %v, want exactly %v (render must equal the filter-satisfying set)", got, liveWant)
	}
	for id := range liveWant {
		if !got[id] {
			t.Errorf("default render is missing live card %s — a store row the filter admits vanished from the render", id)
		}
	}
	if !got["t4"] {
		t.Errorf("default render omits the hold card t4 — hold is a live state and belongs in the default view")
	}

	out, _, err = runTodo(t, "list", "--dropped")
	if err != nil {
		t.Fatalf("list --dropped: %v", err)
	}
	got = todoListRenderedIDs(out)
	if len(got) != len(droppedWant) || !got["t3"] {
		t.Errorf("--dropped render id set = %v, want exactly %v", got, droppedWant)
	}
}

// truncateForTest bounds a stdout sample inside a failure message.
func truncateForTest(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
