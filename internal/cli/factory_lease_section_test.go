// factory_lease_section_test.go — SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458)
// AC-FAL-009 clauses (i) and (ii): no queue lock is held during worktree
// creation or the card-worktree record write, and per lease path the section
// issues exactly the table's promotions, restores and record writes, none of
// them before the queue lock is held.
package cli

import (
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// flProbeWindow is AC-FAL-009's lock-probe window (100 ms, the window of L11).
const flProbeWindow = 100 * time.Millisecond

// TestFactoryLeaseSectionExcludesWorktreeStep — AC-FAL-009 (i): a public Mutate
// issued from inside the worktree creator stub completes within 500 ms, and so
// does one issued at the card-worktree record write — the section is over
// before the worktree step begins.
func TestFactoryLeaseSectionExcludesWorktreeStep(t *testing.T) {
	for _, form := range []string{"nominated", "bare"} {
		t.Run(form, func(t *testing.T) {
			_, store := nmBase(t, factory.BacklogStateQueued)
			nmLaneEnv(t, "lane-1", "")
			dir := t.TempDir()
			initGitRepo(t, dir)
			var mu sync.Mutex
			var inCreator, atRecord *flOp
			creatorDone, recordDone := false, false
			prevCreator := worktree.WorktreeCreator
			worktree.WorktreeCreator = func(string, io.Writer) (string, error) {
				op := flStartNoop(store)
				ok := op.within(flMargin)
				mu.Lock()
				inCreator, creatorDone = op, ok
				mu.Unlock()
				return dir, nil
			}
			t.Cleanup(func() { worktree.WorktreeCreator = prevCreator })
			prevNow := factoryCardNow
			factoryCardNow = func() time.Time {
				if flStackHas("factoryEnsureCardWorktree") {
					mu.Lock()
					first := atRecord == nil
					mu.Unlock()
					if first {
						op := flStartNoop(store)
						ok := op.within(flMargin)
						mu.Lock()
						atRecord, recordDone = op, ok
						mu.Unlock()
					}
				}
				return prevNow()
			}
			t.Cleanup(func() { factoryCardNow = prevNow })

			args := []string{"--run", fcRun}
			if form == "nominated" {
				args = append(args, "--card", "t1")
			}
			_, stderr, err := qasRunNext(t, args...)
			factoryCardNow = prevNow
			if err != nil {
				t.Fatalf("next: %v (stderr %q)", err, stderr)
			}
			mu.Lock()
			defer mu.Unlock()
			if inCreator == nil || atRecord == nil {
				t.Fatalf("the creator stub (reached=%v) or the card-worktree record write (reached=%v) was not reached", inCreator != nil, atRecord != nil)
			}
			flJoin(t, []*flOp{inCreator, atRecord})
			if !creatorDone {
				t.Errorf("a Mutate issued from inside the creator stub did not complete within %s (a queue lock is held during worktree creation)", flMargin)
			}
			if !recordDone {
				t.Errorf("a Mutate issued at the card-worktree record write did not complete within %s", flMargin)
			}
		})
	}
}

// flArmCase is one success row of AC-FAL-009's table.
type flArmCase struct {
	name                       string
	itemState                  factory.BacklogState
	row                        *homestate.Card
	nominated                  bool
	promotions, restores, wrts int
}

func flArmCases() []flArmCase {
	assigned := func() *homestate.Card {
		return &homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun}
	}
	picked := func() *homestate.Card { return &homestate.Card{CardID: "t1", State: homestate.CardPicked} }
	return []flArmCase{
		{name: "arm-a", itemState: factory.BacklogStatePicked, row: assigned(), wrts: 1},
		{name: "arm-b", itemState: factory.BacklogStatePicked, row: picked(), wrts: 2},
		{name: "arm-b2", itemState: factory.BacklogStatePicked, wrts: 3},
		{name: "arm-c", itemState: factory.BacklogStateQueued, promotions: 1, wrts: 3},
		{name: "nominated-queued", itemState: factory.BacklogStateQueued, nominated: true, promotions: 1, wrts: 3},
		{name: "nominated-queued-assigned", itemState: factory.BacklogStateQueued, row: assigned(), nominated: true, promotions: 1, wrts: 1},
		{name: "nominated-picked", itemState: factory.BacklogStatePicked, nominated: true, wrts: 3},
		{name: "nominated-picked-assigned", itemState: factory.BacklogStatePicked, row: assigned(), nominated: true, wrts: 1},
	}
}

// TestFactoryLeaseSectionRecordWritesPerArm — AC-FAL-009 (ii): for every success
// row of the table, the lease issues exactly the row's queue promotions, queue
// restores and factory-record write transactions, and at every record write the
// queue's lock is held (an operator write started there stays pending for
// 100 ms): writes_before_lock is 0.
func TestFactoryLeaseSectionRecordWritesPerArm(t *testing.T) {
	for _, c := range flArmCases() {
		t.Run(c.name, func(t *testing.T) {
			root, store := nmBase(t, c.itemState)
			if c.row != nil {
				fcPlace(t, root, *c.row)
			}
			nmLaneEnv(t, "lane-1", "")
			nmIsolatedWorktrees(t, "t1")

			var mu sync.Mutex
			writes, beforeLock := 0, 0
			var stateAtFirstWrite factory.BacklogState
			var probes []*flOp
			prevNow := factoryCardNow
			factoryCardNow = func() time.Time {
				if flAtClaimWrite() {
					rec, err := store.LoadPure()
					op := flStartNoop(store)
					pending := !op.within(flProbeWindow)
					mu.Lock()
					writes++
					if writes == 1 && err == nil {
						for _, it := range rec.Items {
							if it.ID == "t1" {
								stateAtFirstWrite = it.State
							}
						}
					}
					if !pending {
						beforeLock++
					}
					probes = append(probes, op)
					mu.Unlock()
				}
				return prevNow()
			}
			t.Cleanup(func() { factoryCardNow = prevNow })

			args := []string{"--run", fcRun}
			if c.nominated {
				args = append(args, "--card", "t1")
			}
			_, stderr, err := qasRunNext(t, args...)
			factoryCardNow = prevNow
			mu.Lock()
			ops := append([]*flOp(nil), probes...)
			gotWrites, gotBefore, first := writes, beforeLock, stateAtFirstWrite
			mu.Unlock()
			flJoin(t, ops)
			if err != nil {
				t.Fatalf("next: %v (stderr %q)", err, stderr)
			}
			nmAssertLeased(t, root, "t1", "lane-1")
			final := nmQueueState(t, store, "t1")
			promotions, restores := 0, 0
			if c.itemState == factory.BacklogStateQueued && (first == factory.BacklogStatePicked || final == factory.BacklogStatePicked) {
				promotions = 1
			}
			if promotions == 1 && final == factory.BacklogStateQueued {
				restores = 1
			}
			t.Logf("%s: promotions=%d restores=%d record_writes=%d writes_before_lock=%d", c.name, promotions, restores, gotWrites, gotBefore)
			if promotions != c.promotions || restores != c.restores || gotWrites != c.wrts {
				t.Errorf("%s: promotions=%d restores=%d record writes=%d, want %d/%d/%d", c.name, promotions, restores, gotWrites, c.promotions, c.restores, c.wrts)
			}
			if gotBefore != 0 {
				t.Errorf("%s: writes_before_lock=%d, want 0 (at %s of %d record writes an operator write completed, so the queue lock was not held)", c.name, gotBefore, fmt.Sprint(gotBefore), gotWrites)
			}
		})
	}
}
