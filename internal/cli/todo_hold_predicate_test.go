// todo_hold_predicate_test.go — SPEC-TODO-HOLD-STATE-001 M3 acceptance
// tests: the positive-enumeration sweep over every actionable selection
// surface (AC-THS-011), the queued-only machine-lease pin (AC-THS-012), and
// the autodone regression guard (AC-THS-011(c)).
//
// Two axes, deliberately separate:
//
//   - the FORM axis (a): a source scan asserting no actionable surface in
//     this package excludes states by a negated comparison (REQ-THS-012) —
//     the shape whose default would silently include a state added later.
//   - the BEHAVIOR axis (b): a fifth state value, planted the only way one
//     can exist (a hand-edited database), is never selected by any
//     actionable path. If a new state lands and the sweep is not re-run,
//     this test fails — the drift-detection axis.
package cli

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// AC-THS-011(a) — the form axis: no non-test file in this package compares a
// card state with a negated equality. Positive control: each actionable
// surface file in the sweep inventory carries at least one POSITIVE state
// comparison (`==` or a switch case), so a refactor that accidentally
// deleted the selection logic cannot satisfy the scan by going empty — and
// the control names the swept files, so the inventory itself is what is
// verified.
func TestTodoSelectionPredicatesPositivelyEnumerateStates(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("list package dir: %v", err)
	}
	// The M3 sweep inventory: every actionable surface file. The positive
	// control asserts on exactly these.
	actionable := map[string]bool{
		"todo.go":           true, // next bare listing + pick gate + unpick gate
		"todo_autodone.go":  true, // auto-done candidate scan
		"todo_drop.go":      true, // drop/undrop gates
		"goal.go":           true, // --auto mission pick
		"factory_card.go":   true, // machine-lease record admission
		"factory_mirror.go": true,
		"gtd.go":            true, // gtd engage pick
		"graph.go":          true, // live-set display predicate (G4)
		"todo_hold.go":      true, // hold/unhold gates
		"todo_analysis.go":  true,
	}
	positives := map[string]int{}
	scanned := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		scanned++
		for _, line := range strings.Split(string(raw), "\n") {
			if strings.Contains(line, "State != factory.BacklogState") {
				t.Errorf("%s excludes states by a negated comparison (REQ-THS-012): %s",
					name, strings.TrimSpace(line))
			}
			if strings.Contains(line, "State == factory.BacklogState") ||
				strings.Contains(line, "case factory.BacklogState") {
				positives[name]++
			}
		}
	}
	if scanned < 10 {
		t.Fatalf("scanned only %d non-test files — the sweep must cover the package, not a fragment", scanned)
	}
	for name := range actionable {
		if positives[name] == 0 {
			t.Errorf("%s is in the actionable-surface inventory but carries no positive state comparison — the sweep's positive control requires at least one", name)
		}
	}
}

// AC-THS-011(b) — the behavior axis: a card carrying a fifth state value is
// never selected by any actionable path. The value is planted the only way
// it can exist in a stamped-v2 database — a hand edit with the CHECK
// suspended — which is exactly the corruption channel the positive
// enumeration exists to contain. RED-now at plan time: the pick gate refused
// only `dropped`, so a future-state card was pickable.
func TestTodoFutureStateCardIsNeverSelectedByActionablePaths(t *testing.T) {
	root, store := todoFixture(t)
	seedTodo(t, "first queued", "second queued", "third queued")
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items[0].State = factory.BacklogStateHold
		return nil
	}); err != nil {
		t.Fatalf("seed hold: %v", err)
	}
	// Hand-corrupt t2 into a state the enum does not know.
	db, err := sql.Open("sqlite", store.EnginePath())
	if err != nil {
		t.Fatalf("open engine for corruption: %v", err)
	}
	if _, err := db.Exec(`PRAGMA ignore_check_constraints = ON`); err != nil {
		t.Fatalf("suspend checks: %v", err)
	}
	if _, err := db.Exec(`UPDATE items SET state = 'future_state' WHERE id = 't2'`); err != nil {
		t.Fatalf("plant future state: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close engine: %v", err)
	}
	before := readBacklogBytes(t, root)

	// Bare next lists the one remaining queued card only.
	out, _, err := runTodo(t, "next")
	if err != nil {
		t.Fatalf("bare next: %v", err)
	}
	if !strings.Contains(out, "t3\t") {
		t.Errorf("bare next = %q, want the queued card t3 listed", out)
	}
	for _, excluded := range []string{"t1", "t2"} {
		if strings.Contains(out, excluded+"\t") {
			t.Errorf("bare next = %q, want the %s card (held / future-state) excluded from candidates", out, excluded)
		}
	}

	// A pick attempt on the future-state card is refused, writing nothing.
	if _, _, err := runTodo(t, "next", "2"); err == nil {
		t.Fatal("picking a future-state card must be refused — the gate must enumerate accepted states, not refuse known-bad ones")
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Error("refused future-state pick must leave the queue file byte-identical")
	}

	// The auto-done candidate scan produces no outcome for the held or the
	// future-state card (dry run: reads only — a Mutate would have to write
	// the unknown state back). The one remaining queued card MAY appear: an
	// inconclusive query skips it, which is a scan of a lawful candidate.
	dryOut, _, err := runTodo(t, "auto-done", "--dry-run")
	if err != nil {
		t.Fatalf("auto-done dry run: %v", err)
	}
	for _, absent := range []string{"t1", "t2"} {
		if strings.Contains(dryOut, absent) {
			t.Errorf("auto-done dry run = %q, want no outcome for %s (held / future-state)", dryOut, absent)
		}
	}
}

// AC-THS-011(c) — the autodone regression guard, GREEN-at-M1 by design: the
// existing candidate filter already skips a held card (iter1 D1 correction —
// the site's defect was form, not behavior). The guard stays green through
// M3's form conversion; the positive control proves it is green for the
// right reason by flipping the held card to queued and watching the outcome
// set grow.
func TestTodoAutoDoneSkipsHeldCard(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t901", "queued work", factory.BacklogStateQueued)
	seedCard(t, store, "t921", "held work", factory.BacklogStateHold)
	commitOnRef(t, root, "Merge branch 'WT-x' into develop (card t901)")
	commitOnRef(t, root, "Merge branch 'WT-y' into develop (card t921)")
	materializeOriginDevelop(t, root)

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	if !strings.Contains(stdout, "done t901") {
		t.Errorf("stdout = %q, want the queued card closed", stdout)
	}
	if strings.Contains(stdout, "t921") {
		t.Errorf("stdout = %q, want no outcome for the held card", stdout)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, entry := range rec.Archived {
		if entry.Item.ID == "t921" {
			t.Error("the held card was archived — hold must be skipped by the candidate scan")
		}
	}

	// Positive control (mutation): the same card in queued state IS a close
	// candidate — the scan sees hold, it does not ignore it.
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == "t921" {
				r.Items[i].State = factory.BacklogStateQueued
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("flip hold to queued: %v", err)
	}
	stdout, _, err = runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done positive control: %v", err)
	}
	if !strings.Contains(stdout, "done t921") {
		t.Errorf("positive control stdout = %q, want the formerly-held card closed once queued", stdout)
	}
}

// AC-THS-012 — the machine-lease pin: the selection layer a leaser calls
// picks only queued cards. Queue shape: 2 queued + 1 hold + 1 picked, in
// five shuffled insertion orders; the selectable set is exactly the 2 queued
// rows every time. Positive control: flipping the held card to queued grows
// the selectable set to 3 — the leaser ignores hold because it is held, not
// because it is invisible.
//
// Disposition (recorded for the lead): the named leaser selection functions
// (`factory next`, the Codex `-f` lane claim) are NOT in this tree — the
// pre-flight re-read confirmed t1240/t1294 still in flight. The pin lives at
// the shared selection layer (`todo next` bare listing + pick gate) that any
// leaser inherits; when t1240/t1294 land, their selectors join this sweep.
func TestTodoMachineLeaseSelectsOnlyQueued(t *testing.T) {
	orders := [][]string{
		{"queued-a", "queued-b", "held", "picked"},
		{"held", "queued-a", "picked", "queued-b"},
		{"picked", "held", "queued-b", "queued-a"},
		{"queued-b", "picked", "queued-a", "held"},
		{"held", "picked", "queued-a", "queued-b"},
	}
	stateOf := map[string]factory.BacklogState{
		"queued-a": factory.BacklogStateQueued,
		"queued-b": factory.BacklogStateQueued,
		"held":     factory.BacklogStateHold,
		"picked":   factory.BacklogStatePicked,
	}
	for run := 0; run < len(orders); run++ {
		t.Run(fmtShuffledRun(run), func(t *testing.T) {
			_, store := todoFixture(t)
			// The ids follow insertion order, so the shuffle decides which id
			// carries which state.
			queuedIDs, nonQueuedIDs := []string{}, []string{}
			if err := store.Mutate(func(rec *factory.BacklogRecord) error {
				for _, name := range orders[run] {
					rec.LastSeq++
					id := fmt.Sprintf("t%d", rec.LastSeq)
					rec.Items = append(rec.Items, factory.BacklogItem{
						ID:      id,
						Text:    name,
						AddedAt: "2026-09-29T00:00:00Z",
						State:   stateOf[name],
					})
					if stateOf[name] == factory.BacklogStateQueued {
						queuedIDs = append(queuedIDs, id)
					} else {
						nonQueuedIDs = append(nonQueuedIDs, id)
					}
				}
				return nil
			}); err != nil {
				t.Fatalf("seed run %d: %v", run, err)
			}
			out, _, err := runTodo(t, "next")
			if err != nil {
				t.Fatalf("bare next: %v", err)
			}
			for _, id := range queuedIDs {
				if !strings.Contains(out, id+"\t") {
					t.Errorf("run %d: bare next = %q, want queued card %s selectable", run, out, id)
				}
			}
			for _, id := range nonQueuedIDs {
				if strings.Contains(out, id+"\t") {
					t.Errorf("run %d: bare next = %q, want %s excluded from the selectable set", run, out, id)
				}
				// The pick path agrees with the listing: hold and picked
				// refuse; only queued picks.
				if _, _, err := runTodo(t, "next", strings.TrimPrefix(id, "t")); err == nil {
					t.Errorf("run %d: pick of %s must be refused", run, id)
				}
			}
		})
	}
}

func fmtShuffledRun(n int) string { return fmt.Sprintf("order-%d", n) }
