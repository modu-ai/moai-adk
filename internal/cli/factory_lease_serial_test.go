// factory_lease_serial_test.go — SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458)
// AC-FAL-001 (two serial leases at once: exactly one wins, in one process and
// across two processes) and AC-FAL-005 (two lanes lease their own assigned
// serial cards: exactly one).
package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// flSerialPair builds the two-queued-serial-card queue of AC-FAL-001.
func flSerialPair(t *testing.T) (string, *kanban.BacklogStore) {
	t.Helper()
	root, store := nmBase(t, kanban.BacklogStateQueued, kanban.BacklogStateQueued)
	fcClassify(t, store, "t1", kanban.ClassPriorityNormal, false, kanban.ClassModeSerial)
	fcClassify(t, store, "t2", kanban.ClassPriorityNormal, false, kanban.ClassModeSerial)
	return root, store
}

// flSplitResults separates the lanes that leased from the ones that did not.
func flSplitResults(results []flLaneResult) (winners, losers []flLaneResult) {
	for _, r := range results {
		if r.err == nil {
			winners = append(winners, r)
		} else {
			losers = append(losers, r)
		}
	}
	return winners, losers
}

// TestFactoryLeaseSerialDistinctNomineesExactlyOne — AC-FAL-001 (a): two lanes
// nominate two different queued serial cards at once, held at the nomination
// seam by the tolerant gate. Exactly one card ends leased; the other lane is
// refused `serial-slot` (exit 4, one stderr line) and its card has no leased row.
func TestFactoryLeaseSerialDistinctNomineesExactlyOne(t *testing.T) {
	root, _ := flSerialPair(t)
	nmIsolatedWorktrees(t, "t1", "t2")
	g := newFLGate(2)
	nmSetSeam(t, func(cardID string) error {
		g.hold(cardID)
		return nil
	})
	results := flRaceLanes(t, g, []nmLaneRun{{"lane-1", "t1"}, {"lane-2", "t2"}})
	n, rows := flLeasedRows(t, root, "t1", "t2")
	t.Logf("serial cards leased=%d rows=%v", n, rows)
	if n != 1 {
		t.Fatalf("serial exclusivity: %d serial cards are leased (%v), want exactly 1", n, rows)
	}
	winners, losers := flSplitResults(results)
	if len(winners) != 1 || len(losers) != 1 {
		t.Fatalf("%d lanes leased and %d did not, want one of each (results %+v)", len(winners), len(losers), results)
	}
	flAssertOneRefusalLine(t, losers[0], 4, "factory next: refused serial-slot: ")
	if st, _ := flRow(t, root, losers[0].card); st == homestate.CardLeased {
		t.Errorf("the refused lane's card %s has a leased row", losers[0].card)
	}
}

// TestFactoryLeaseSerialBareLanesExactlyOne — AC-FAL-001 (b): two lanes run bare
// `factory next` in one test process, held at the pass entry (before either
// reads the record) by the tolerant gate. Exactly one serial card ends leased;
// the other lane exits 3 with `no card is available`.
func TestFactoryLeaseSerialBareLanesExactlyOne(t *testing.T) {
	root, _ := flSerialPair(t)
	nmIsolatedWorktrees(t, "t1", "t2")
	g := newFLGate(2)
	prev := factoryLeaseAtPassEntry
	factoryLeaseAtPassEntry = func() { g.hold(strconv.Itoa(flGoroutineID())) }
	t.Cleanup(func() { factoryLeaseAtPassEntry = prev })

	results := flRaceLanes(t, g, []nmLaneRun{{"lane-1", ""}, {"lane-2", ""}})
	n, rows := flLeasedRows(t, root, "t1", "t2")
	t.Logf("serial cards leased=%d rows=%v", n, rows)
	if n != 1 {
		t.Fatalf("serial exclusivity: %d serial cards are leased (%v), want exactly 1", n, rows)
	}
	winners, losers := flSplitResults(results)
	if len(winners) != 1 || len(losers) != 1 {
		t.Fatalf("%d lanes leased and %d did not, want one of each (results %+v)", len(winners), len(losers), results)
	}
	if code := nmExit(losers[0].err); code != 3 {
		t.Errorf("the losing lane exited %d (%v), want 3", code, losers[0].err)
	}
	if !strings.Contains(losers[0].out, "no card is available") {
		t.Errorf("the losing lane's stdout = %q, want `no card is available`", losers[0].out)
	}
}

const (
	flXProcLaneEnv  = "MOAI_T1458_XPROC_LANE"
	flXProcRootEnv  = "MOAI_T1458_XPROC_ROOT"
	flXProcRdvEnv   = "MOAI_T1458_XPROC_RDV"
	flXProcTreesEnv = "MOAI_T1458_XPROC_TREES"
	flXProcGrace    = 3 * time.Second
)

// TestFactoryLeaseSerialCrossProcessHelper is the child of the cross-process
// test (plan WM1, "cross-process lane helper"). It does nothing unless the
// parent set the lane variable. TestMain clears the factory ambient family and
// CLAUDE_PROJECT_DIR in the child, so the child re-establishes them here; the
// hold at the pass entry is a file rendezvous (a channel cannot cross a process
// boundary); the worktree creator stub is rebuilt from the paths the parent
// prepared; and the result goes back on stdout as one tagged line.
func TestFactoryLeaseSerialCrossProcessHelper(t *testing.T) {
	lane := os.Getenv(flXProcLaneEnv)
	if lane == "" {
		t.Skip("helper process only")
	}
	root := os.Getenv(flXProcRootEnv)
	rdv := os.Getenv(flXProcRdvEnv)
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv(config.EnvHome, "")
	t.Chdir(root)
	nmLaneEnv(t, lane, "")

	trees := map[string]string{}
	for _, kv := range strings.Split(os.Getenv(flXProcTreesEnv), ";") {
		if k, v, ok := strings.Cut(kv, "="); ok {
			trees[k] = v
		}
	}
	prevCreator := worktree.WorktreeCreator
	worktree.WorktreeCreator = func(name string, _ io.Writer) (string, error) {
		dir, ok := trees[name]
		if !ok {
			return "", fmt.Errorf("no isolated worktree prepared for %s", name)
		}
		return dir, nil
	}
	t.Cleanup(func() { worktree.WorktreeCreator = prevCreator })

	prevNow := factoryCardNow
	factoryCardNow = func() time.Time { return fcNow }
	t.Cleanup(func() { factoryCardNow = prevNow })

	held := false
	prevEntry := factoryLeaseAtPassEntry
	factoryLeaseAtPassEntry = func() {
		if held {
			return
		}
		held = true
		_ = os.WriteFile(filepath.Join(rdv, "arrived-"+lane), nil, 0o600)
		deadline := time.Now().Add(flXProcGrace)
		for time.Now().Before(deadline) {
			entries, _ := os.ReadDir(rdv)
			if len(entries) >= 2 {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	t.Cleanup(func() { factoryLeaseAtPassEntry = prevEntry })

	stdout, stderr, err := qasRunNext(t, "--run", fcRun)
	firstErr, _, _ := strings.Cut(strings.TrimSpace(stderr), "\n")
	fmt.Printf("XPROC-RESULT lane=%s held=%v exit=%d stdout=%q stderr-first-line=%q\n", lane, held, nmExit(err), stdout, firstErr)
}

// TestFactoryLeaseSerialCrossProcessExactlyOne — AC-FAL-001 (c): two separate
// helper PROCESSES run bare `factory next`, each held at the pass entry by a
// file rendezvous. Exactly one serial card ends leased; the other process exits
// 3 with `no card is available`. A process-local mutex cannot satisfy this test
// (mutant MU7): only a cross-process lock orders two processes.
func TestFactoryLeaseSerialCrossProcessExactlyOne(t *testing.T) {
	root, _ := flSerialPair(t)
	trees := make([]string, 0, 2)
	for _, id := range []string{"t1", "t2"} {
		dir := t.TempDir()
		initGitRepo(t, dir)
		trees = append(trees, id+"="+dir)
	}
	rdv := t.TempDir()

	lanes := []string{"lane-1", "lane-2"}
	outs := make([]bytes.Buffer, len(lanes))
	var wg sync.WaitGroup
	for i, lane := range lanes {
		cmd := exec.Command(os.Args[0], "-test.run=^TestFactoryLeaseSerialCrossProcessHelper$", "-test.timeout=3m")
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			factoryEnvPinnedEnv+"=1",
			flXProcLaneEnv+"="+lane,
			flXProcRootEnv+"="+root,
			flXProcRdvEnv+"="+rdv,
			flXProcTreesEnv+"="+strings.Join(trees, ";"),
		)
		cmd.Stdout = &outs[i]
		cmd.Stderr = &outs[i]
		if err := cmd.Start(); err != nil {
			t.Fatalf("start the helper process for %s: %v", lane, err)
		}
		wg.Add(1)
		go func(c *exec.Cmd) {
			defer wg.Done()
			_ = c.Wait()
		}(cmd)
	}
	wg.Wait()

	var results []string
	for i := range outs {
		for _, line := range strings.Split(outs[i].String(), "\n") {
			if strings.HasPrefix(line, "XPROC-RESULT") {
				results = append(results, line)
				t.Logf("%s", line)
			}
		}
	}
	n, rows := flLeasedRows(t, root, "t1", "t2")
	t.Logf("serial cards leased across two processes=%d rows=%v", n, rows)
	if len(results) != 2 {
		t.Fatalf("got %d helper result lines, want 2 (helper output: %q / %q)", len(results), outs[0].String(), outs[1].String())
	}
	if n != 1 {
		t.Fatalf("serial exclusivity across two processes: %d serial cards are leased (%v), want exactly 1", n, rows)
	}
	noCard := 0
	for _, line := range results {
		if strings.Contains(line, "exit=3") && strings.Contains(line, "no card is available") {
			noCard++
		}
	}
	if noCard != 1 {
		t.Errorf("%d helper processes exited 3 with `no card is available`, want exactly 1 (%v)", noCard, results)
	}
}

// TestFactoryLeaseOwnAssignedSerialSiblingsExactlyOne — AC-FAL-005: serial cards
// t1 and t2 are recorded `assigned` to lane-1 and lane-2, nothing in flight; both
// lanes run bare `factory next` held at their arm (a) claim by the tolerant
// gate. Exactly one card is leased; the other lane exits 3 and its card stays
// `assigned` with owner and version unchanged.
func TestFactoryLeaseOwnAssignedSerialSiblingsExactlyOne(t *testing.T) {
	root, store := nmBase(t, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
	fcClassify(t, store, "t1", kanban.ClassPriorityNormal, false, kanban.ClassModeSerial)
	fcClassify(t, store, "t2", kanban.ClassPriorityNormal, false, kanban.ClassModeSerial)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun},
		homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-2", Stage: homestate.CardRun})
	before := map[string]homestate.Card{"t1": fcCard(t, root, "t1"), "t2": fcCard(t, root, "t2")}
	nmIsolatedWorktrees(t, "t1", "t2")
	g := newFLGate(2)
	prev := factoryLeaseBeforeClaim
	factoryLeaseBeforeClaim = func(arm, cardID string) error {
		if arm == "a" {
			g.hold(cardID)
		}
		return nil
	}
	t.Cleanup(func() { factoryLeaseBeforeClaim = prev })

	results := flRaceLanes(t, g, []nmLaneRun{{"lane-1", ""}, {"lane-2", ""}})
	n, rows := flLeasedRows(t, root, "t1", "t2")
	t.Logf("serial cards leased=%d rows=%v", n, rows)
	if n != 1 {
		t.Fatalf("arm (a): %d serial cards are leased (%v), want exactly 1", n, rows)
	}
	_, losers := flSplitResults(results)
	if len(losers) != 1 {
		t.Fatalf("%d lanes did not lease, want exactly one (results %+v)", len(losers), results)
	}
	if code := nmExit(losers[0].err); code != 3 {
		t.Errorf("the losing lane exited %d (%v), want 3", code, losers[0].err)
	}
	for _, id := range []string{"t1", "t2"} {
		c := fcCard(t, root, id)
		if c.State == homestate.CardLeased {
			continue
		}
		b := before[id]
		if c.State != homestate.CardAssigned || c.OwnerLabel != b.OwnerLabel || c.Version != b.Version {
			t.Errorf("%s = %s owner=%q version=%d, want assigned owner=%q version=%d unchanged", id, c.State, c.OwnerLabel, c.Version, b.OwnerLabel, b.Version)
		}
	}
}
