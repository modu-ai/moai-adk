// todo_transition_stamps_test.go — SPEC-TODO-TRANSITION-STAMPS-001 M2
// acceptance: the transition stamps (AC-TST-002..004) and the edge cases
// acceptance.md §D.4 names. Every test drives the CLI verbs against an
// isolated fixture queue and reads the persisted record back through the
// store — the column surfaces, not the struct, are the contract.
package cli

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// stampOf returns the dereferenced stamp or "" when absent.
func stampOf(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// TestTransitionStamps_PickedAtStampsAndClears — AC-TST-002. add --pick
// stamps; unpick clears; a re-pick holds the SECOND episode's time, not the
// first's.
func TestTransitionStamps_PickedAtStampsAndClears(t *testing.T) {
	_, store := todoFixture(t)

	if _, _, err := runTodo(t, "add", "--pick", "picked once"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	first := rec.Items[0]
	if first.State != factory.BacklogStatePicked {
		t.Fatalf("state = %s, want picked", first.State)
	}
	if stampOf(first.PickedAt) == "" {
		t.Fatalf("picked_at is NULL after add --pick, want a timestamp")
	}

	if _, _, err := runTodo(t, "unpick", "t1"); err != nil {
		t.Fatalf("unpick: %v", err)
	}
	rec, err = store.Load()
	if err != nil {
		t.Fatalf("load after unpick: %v", err)
	}
	if got := stampOf(rec.Items[0].PickedAt); got != "" {
		t.Fatalf("picked_at = %q after unpick, want NULL", got)
	}

	// RFC3339 carries second precision, so the second episode must start in
	// a later second for "the SECOND transition's time, not the first's" to
	// be an observable distinction.
	time.Sleep(1100 * time.Millisecond)
	if _, _, err := runTodo(t, "next", "t1"); err != nil {
		t.Fatalf("next t1: %v", err)
	}
	rec, err = store.Load()
	if err != nil {
		t.Fatalf("load after re-pick: %v", err)
	}
	second := stampOf(rec.Items[0].PickedAt)
	if second == "" {
		t.Fatalf("picked_at is NULL after re-pick, want a timestamp")
	}
	if second == stampOf(first.PickedAt) {
		t.Fatalf("picked_at after re-pick = %q, want the SECOND episode's time (first was %q)", second, stampOf(first.PickedAt))
	}
}

// TestTransitionStamps_NextPickStamps — the `todo next <n>` pick path stamps
// the same way `add --pick` does (REQ-TST-004 names both transitions).
func TestTransitionStamps_NextPickStamps(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "queued then picked"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "next", "t1"); err != nil {
		t.Fatalf("next t1: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rec.Items[0].State != factory.BacklogStatePicked || stampOf(rec.Items[0].PickedAt) == "" {
		t.Fatalf("after next pick: state=%s picked_at=%q, want picked + a timestamp",
			rec.Items[0].State, stampOf(rec.Items[0].PickedAt))
	}
}

// TestTransitionStamps_DroppedAtStampsAndClears — AC-TST-003.
func TestTransitionStamps_DroppedAtStampsAndClears(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "to drop"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "drop", "t1", "no longer needed"); err != nil {
		t.Fatalf("drop: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if stampOf(rec.Items[0].DroppedAt) == "" {
		t.Fatalf("dropped_at is NULL after drop, want a timestamp")
	}
	if _, _, err := runTodo(t, "undrop", "t1"); err != nil {
		t.Fatalf("undrop: %v", err)
	}
	rec, err = store.Load()
	if err != nil {
		t.Fatalf("load after undrop: %v", err)
	}
	if got := stampOf(rec.Items[0].DroppedAt); got != "" {
		t.Fatalf("dropped_at = %q after undrop, want NULL", got)
	}
}

// TestTransitionStamps_ArchivePreservesStampsAndStampsArchivedAt — AC-TST-004.
func TestTransitionStamps_ArchivePreservesStampsAndStampsArchivedAt(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "--pick", "picked then done"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	pickedAt := stampOf(rec.Items[0].PickedAt)
	if pickedAt == "" {
		t.Fatalf("picked_at is NULL before done, want a timestamp")
	}
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	rec, err = store.Load()
	if err != nil {
		t.Fatalf("load after done: %v", err)
	}
	if len(rec.Items) != 0 {
		t.Fatalf("live table still holds %d rows after done, want 0", len(rec.Items))
	}
	if len(rec.Archived) != 1 {
		t.Fatalf("archive holds %d entries after done, want 1", len(rec.Archived))
	}
	entry := rec.Archived[0]
	if got := stampOf(entry.Item.PickedAt); got != pickedAt {
		t.Errorf("archived picked_at = %q, want the pre-archive stamp %q unchanged", got, pickedAt)
	}
	if stampOf(entry.ArchivedAt) == "" {
		t.Errorf("archived_at is NULL after done, want a timestamp")
	}
	if stampOf(entry.Item.DroppedAt) != "" {
		t.Errorf("archived dropped_at = %q, want NULL — the two stamps never coexist", stampOf(entry.Item.DroppedAt))
	}
}

// TestTransitionStamps_DropThenDoneCarriesDroppedStampOnly — acceptance.md
// §D.4: drop is queued-only, so the direct drop → done path archives a row
// carrying dropped_at with picked_at NULL.
func TestTransitionStamps_DropThenDoneCarriesDroppedStampOnly(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "dropped then done"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "drop", "t1", "changed mind, closed instead"); err != nil {
		t.Fatalf("drop: %v", err)
	}
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	entry := rec.Archived[0]
	if stampOf(entry.Item.DroppedAt) == "" {
		t.Errorf("archived dropped_at is NULL after drop→done, want a timestamp")
	}
	if stampOf(entry.Item.PickedAt) != "" {
		t.Errorf("archived picked_at = %q, want NULL — the two stamps never coexist", stampOf(entry.Item.PickedAt))
	}
}

// TestTransitionStamps_GtdEngagePickAndGoalMissionPick — sync-audit F1
// repair: REQ-TST-004's "or any other transition that sets state='picked'"
// clause covers two pick paths beyond the two todo verbs, and both stamped
// nothing before this repair:
//
//  1. `moai gtd engage --pick` — the published-card pick in the gtd engage
//     verb (internal/cli/gtd.go, the engage Mutate callback).
//  2. the goal auto-mission ActionPick owner-adapter pick (internal/cli/
//     goal.go, the mission supervise path).
//
// Each subtest is named after its entry point so a selector sweep cannot
// silently pass on one path alone. Both are asserted through the persisted
// record — the card ends picked AND carries a non-empty picked_at.
func TestTransitionStamps_GtdEngagePickAndGoalMissionPick(t *testing.T) {
	t.Run("gtd engage --pick", func(t *testing.T) {
		_, store := todoFixture(t)
		// capture → clarify → organize → engage --pick (no --dispatch: the
		// pick stamp is what is under test, and isolating it keeps the
		// dispatch machinery — lanes, leases — out of this test).
		out, _, err := runGTDCapture(t, "capture", "engaged pick stamp card", "--event", "fr-stamp-gtd", "--source", "user", "--sensitivity", "private", "--json")
		if err != nil {
			t.Fatalf("capture: %v", err)
		}
		var captured struct {
			ItemID string `json:"item_id"`
		}
		if err := json.Unmarshal([]byte(out), &captured); err != nil || captured.ItemID == "" {
			t.Fatalf("capture output=%q err=%v", out, err)
		}
		if _, _, err := runGTDCapture(t, "clarify", captured.ItemID, "--disposition", "action", "--outcome", "landed", "--evidence", "CI", "--authority", "queue,dispatch", "--trusted"); err != nil {
			t.Fatalf("clarify: %v", err)
		}
		if _, _, err := runGTDCapture(t, "organize", captured.ItemID, "--class", "action", "--context", "computer"); err != nil {
			t.Fatalf("organize: %v", err)
		}
		if _, stderr, err := runGTDCapture(t, "engage", captured.ItemID, "--approve", "--fresh", "--dependencies-ready", "--lane", "worker-4", "--resources", "--pick", "--json"); err != nil {
			t.Fatalf("engage: %v (stderr %s)", err, stderr)
		}
		rec, err := store.Load()
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		for i := range rec.Items {
			if rec.Items[i].ID == "t1" {
				if rec.Items[i].State != factory.BacklogStatePicked {
					t.Fatalf("state = %s, want picked after gtd engage --pick", rec.Items[i].State)
				}
				if stampOf(rec.Items[i].PickedAt) == "" {
					t.Fatalf("picked_at is NULL after `gtd engage --pick`, want a stamp (REQ-TST-004 covers this transition)")
				}
				return
			}
		}
		t.Fatalf("no live card t1 after gtd engage --pick: %d items", len(rec.Items))
	})

	t.Run("goal auto-mission ActionPick", func(t *testing.T) {
		root, store := todoFixture(t)
		// fcGoalDispatch walks publish → pick → dispatch through the mission
		// supervise path; the pick owner-adapter (goal.go) is the transition
		// under test, so the picked card must carry a picked_at stamp.
		_, cardID, _ := fcGoalDispatch(t, root, store, "goal mission pick stamp card", "fr-stamp-goal", "018f4f4a-7b7c-7a11-8f4d-f33333333333", "worker-4", "auto-run-stamp")
		rec, err := store.Load()
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		for i := range rec.Items {
			if rec.Items[i].ID == cardID {
				if rec.Items[i].State != factory.BacklogStatePicked {
					t.Fatalf("state = %s, want picked after the mission pick", rec.Items[i].State)
				}
				if stampOf(rec.Items[i].PickedAt) == "" {
					t.Fatalf("picked_at is NULL after the goal auto-mission ActionPick, want a stamp (REQ-TST-004 covers this transition)")
				}
				return
			}
		}
		t.Fatalf("no live card %s after the mission pick: %d items", cardID, len(rec.Items))
	})
}
