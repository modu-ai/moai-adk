package worktree

// migrate_busy_tree_t2094_test.go — card t2094, F3: a forward move that fails
// with the tree still at its source is usually a process holding the tree as
// its working directory (the Windows rename refusal). The row must then say
// what to do about it instead of a bare git error, a degraded probe must keep
// the git error, and a run whose only failures are such trees — nothing moved
// — must not advertise --rollback.

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// t2094BusyTreeWorld is the two-tree world with the forward move of
// w.movedTree refused before any rename, and the probe answering the plan
// with no occupant and the post-failure re-probe with the reprobe function's
// answer for the busy source: the occupant arrived (or became observable)
// between the plan and the move.
func t2094BusyTreeWorld(t *testing.T, reprobe func(source string) ([]string, error)) *m6World {
	t.Helper()
	w := t2053TwoTreeWorld(t)
	origMove := forwardGitMove
	forwardGitMove = func(root, from, to string) error {
		if from == w.movedTree {
			return fmt.Errorf("fatal: rename refused")
		}
		return origMove(root, from, to)
	}
	t.Cleanup(func() { forwardGitMove = origMove })
	origProbe := activeProcessCWDProbe
	calls := 0
	activeProcessCWDProbe = func() ([]string, error) {
		calls++
		if calls == 1 {
			return nil, nil // the plan sees no occupant
		}
		return reprobe(w.movedTree)
	}
	t.Cleanup(func() { activeProcessCWDProbe = origProbe })
	return w
}

func TestApply_BusyTreeNamesTheCauseAndTheWayOut(t *testing.T) {
	t2094BusyTreeWorld(t, func(source string) ([]string, error) { return []string{source}, nil })

	out, err := m4RunMigrate(t, "--apply")
	if err == nil {
		t.Fatalf("apply with a failed tree exited clean:\n%s", out)
	}
	if !strings.Contains(out, "a process is inside this worktree") ||
		!strings.Contains(out, "close it or move its working directory elsewhere") {
		t.Errorf("the failed row does not carry the busy-tree guidance:\n%s", out)
	}
	// The underlying git error stays for the record.
	if !strings.Contains(out, "rename refused") {
		t.Errorf("the busy-tree reason lost the underlying git error:\n%s", out)
	}
}

func TestApply_BusyTreeProbeFailureKeepsTheGitError(t *testing.T) {
	t2094BusyTreeWorld(t, func(string) ([]string, error) { return nil, errors.New("scan unavailable") })

	out, err := m4RunMigrate(t, "--apply")
	if err == nil {
		t.Fatalf("apply with a failed tree exited clean:\n%s", out)
	}
	if !strings.Contains(out, "git worktree move: fatal: rename refused") {
		t.Errorf("a degraded probe must fall back to the git error:\n%s", out)
	}
	if strings.Contains(out, "a process is inside this worktree") {
		t.Errorf("a degraded probe must not claim a busy tree:\n%s", out)
	}
}

func TestApply_AllFailuresNeverMoved_NoRollbackHint(t *testing.T) {
	t2053TwoTreeWorld(t)
	origMove := forwardGitMove
	forwardGitMove = func(root, from, to string) error {
		return fmt.Errorf("fatal: rename refused")
	}
	t.Cleanup(func() { forwardGitMove = origMove })

	out, err := m4RunMigrate(t, "--apply")
	if err == nil {
		t.Fatalf("apply with failed trees exited clean:\n%s", out)
	}
	if strings.Contains(out, "undo with") {
		t.Errorf("a run whose failures never moved anything advertised --rollback:\n%s", out)
	}
}

// TestApply_StrandedTreeKeepsTheRollbackHint — card t2094, audit F3: a run
// whose forward moves reach the destination and whose reverse moves then fail
// leaves the trees failed with a surviving pending line — the journal holds
// moves --rollback must converge, so the hint prints even though nothing
// counts as moved.
func TestApply_StrandedTreeKeepsTheRollbackHint(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	origMove := forwardGitMove
	forwardGitMove = func(root, from, to string) error {
		if err := origMove(root, from, to); err != nil {
			return err
		}
		// The directory rename happened; git then reports the move as failed —
		// the half-completed state the failure handling converges from.
		return fmt.Errorf("fatal: injected failure after the rename")
	}
	t.Cleanup(func() { forwardGitMove = origMove })
	origReverse := reverseGitMove
	reverseGitMove = func(root, from, to string) error { return fmt.Errorf("fatal: reverse refused") }
	t.Cleanup(func() { reverseGitMove = origReverse })

	out, err := m4RunMigrate(t, "--apply")
	if err == nil {
		t.Fatalf("apply with stranded trees exited clean:\n%s", out)
	}
	if !strings.Contains(out, "undo with") {
		t.Errorf("a run with destination-stranded trees must keep the rollback hint:\n%s", out)
	}
	if !strings.Contains(out, w.movedTree) {
		t.Errorf("the run does not name the stranded tree %s:\n%s", w.movedTree, out)
	}
}

// TestMigrateSummary_RollbackHintOnlyWhenTheJournalHoldsMoves — the summary
// gate itself: the hint prints for a run that moved trees or inherited an
// interrupted migration, and stays silent for a run of busy-tree failures.
func TestMigrateSummary_RollbackHintOnlyWhenTheJournalHoldsMoves(t *testing.T) {
	row := func(outcome string) migrationRow {
		return migrationRow{Class: "moai", Source: "s", Destination: "d", Outcome: outcome}
	}
	cases := []struct {
		name   string
		report *migrationReport
		want   bool
	}{
		{"busy failure only", &migrationReport{rows: []migrationRow{row("failed")}, failed: 1, journal: "j.json"}, false},
		{"a move to undo", &migrationReport{rows: []migrationRow{row("moved"), row("failed")}, moved: 1, failed: 1, journal: "j.json"}, true},
		{"interrupted", &migrationReport{rows: []migrationRow{row("interrupted")}, interrupted: 1, journal: "j.json"}, true},
		{"stranded failure", &migrationReport{rows: []migrationRow{row("failed")}, failed: 1, stranded: 1, journal: "j.json"}, true},
		{"no journal at all", &migrationReport{rows: []migrationRow{row("moved")}, moved: 1}, false},
	}
	for _, tc := range cases {
		buf := &strings.Builder{}
		cmd := newMigrateCmd()
		cmd.SetOut(buf)
		if err := printMigrationReport(cmd, tc.report, false); err != nil {
			t.Fatalf("%s: print: %v", tc.name, err)
		}
		if got := strings.Contains(buf.String(), "undo with"); got != tc.want {
			t.Errorf("%s: hint printed = %v, want %v:\n%s", tc.name, got, tc.want, buf.String())
		}
	}
}
