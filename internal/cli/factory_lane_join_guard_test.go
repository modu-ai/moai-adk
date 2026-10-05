// factory_lane_join_guard_test.go — card t1513: the serial slot is held by a
// recorded driver, not by a state alone, and the lane-join gate refuses a
// session whose working directory roots inside a linked worktree.
package cli

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactorySerialSlotHeldRequiresADriver — the slot predicate matrix. A row
// without an owner holds nothing (the t1453 wedge: picked, ownerless,
// lease-less, refusing every lease for a day); every driven row keeps the
// pre-existing behavior.
func TestFactorySerialSlotHeldRequiresADriver(t *testing.T) {
	now := time.Now()
	live := now.Add(time.Hour).Format(time.RFC3339Nano)
	expired := now.Add(-time.Minute).Format(time.RFC3339Nano)
	cases := []struct {
		name string
		card homestate.Card
		want bool
	}{
		{"picked without owner or lease holds nothing", homestate.Card{CardID: "t1453", State: homestate.CardPicked}, false},
		{"picked with an owner holds", homestate.Card{CardID: "t1", State: homestate.CardPicked, OwnerLabel: "lane-5"}, true},
		{"assigned with an owner holds", homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-2"}, true},
		{"leased with a live lease holds", homestate.Card{CardID: "t3", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live}, true},
		{"an expired lease frees", homestate.Card{CardID: "t4", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: expired}, false},
		{"merge-ready frees", homestate.Card{CardID: "t5", State: homestate.CardMergeReady, OwnerLabel: "lane-1"}, false},
	}
	for _, tc := range cases {
		if got := factorySerialSlotHeld(tc.card, now); got != tc.want {
			t.Errorf("%s: held=%v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestFactoryNextIgnoresDriverlessPickedRow — the end-to-end release: with a
// picked, ownerless, lease-less serial row sitting in the run record (the
// measured t1453 shape), a lane's `factory next --card` leases instead of
// being refused serial-slot.
func TestFactoryNextIgnoresDriverlessPickedRow(t *testing.T) {
	root, _ := flSerialPair(t)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root, homestate.Card{CardID: "t1453", State: homestate.CardPicked})
	nmLaneEnv(t, "lane-1", "")
	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if err != nil {
		t.Fatalf("factory next refused with a driverless picked row in the record: %v stderr=%q", err, stderr)
	}
	if st, owner := flRow(t, root, "t1"); st != homestate.CardLeased {
		t.Fatalf("t1 state=%s owner=%q, want leased", st, owner)
	}
}

// TestFactoryNextStillRefusedBehindADrivenRow — the control arm: with an
// owned serial row in flight (no lease, owner recorded), the slot still holds
// and `factory next` is refused serial-slot.
func TestFactoryNextStillRefusedBehindADrivenRow(t *testing.T) {
	root, _ := flSerialPair(t)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root, homestate.Card{CardID: "t1453", State: homestate.CardPicked, OwnerLabel: "lane-2"})
	nmLaneEnv(t, "lane-1", "")
	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if err == nil {
		t.Fatalf("factory next leased behind a driven serial row: stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "refused serial-slot") {
		t.Fatalf("refusal is not serial-slot: %v stderr=%q", err, stderr)
	}
}

// TestLaneJoinRefusesWorktreeRoot — the join gate: a primary-checkout cwd
// joins, a linked-worktree cwd is refused with the sentinel, and a cwd with
// no repository fails open.
func TestLaneJoinRefusesWorktreeRoot(t *testing.T) {
	root := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return string(out)
	}
	run("init", "-q")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "t")
	run("commit", "--allow-empty", "-qm", "seed")
	link := filepath.Join(t.TempDir(), "linked")
	if out, err := exec.Command("git", "-C", root, "worktree", "add", "-b", "wt", link).CombinedOutput(); err != nil {
		t.Fatalf("worktree add: %v: %s", err, out)
	}

	if err := refuseLaneJoinFromWorktreeAt(root); err != nil {
		t.Fatalf("a primary-checkout cwd must join: %v", err)
	}
	err := refuseLaneJoinFromWorktreeAt(link)
	if err == nil {
		t.Fatal("a linked-worktree cwd must be refused")
	}
	if !strings.HasPrefix(err.Error(), laneJoinWorktreeSentinel) {
		t.Fatalf("refusal %q lacks the %s sentinel", err, laneJoinWorktreeSentinel)
	}
	if err := refuseLaneJoinFromWorktreeAt(t.TempDir()); err != nil {
		t.Fatalf("a non-repository cwd must fail open: %v", err)
	}
}
