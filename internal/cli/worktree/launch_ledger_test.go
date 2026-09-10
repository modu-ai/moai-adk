package worktree

// Wiring tests for the launch-ledger reclamation (card t297): every moai
// worktree disposal path prunes dead ledger rows after a successful removal,
// nothing prunes when the removal did not happen (failed removal, --stale
// preview, --json inventory), and the --auto path reclaims quietly.
//
// The prune itself runs through the pruneLaunchLedgerFn seam, so no test here
// touches the real ~/.moai/claude-profiles/launch.yaml.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// ledgerSeam installs a counting fake for pruneLaunchLedgerFn and restores
// the original when the test ends.
func ledgerSeam(t *testing.T) *int {
	t.Helper()
	calls := 0
	orig := pruneLaunchLedgerFn
	pruneLaunchLedgerFn = func() ([]string, error) {
		calls++
		return []string{"/gone/row"}, nil
	}
	t.Cleanup(func() { pruneLaunchLedgerFn = orig })
	return &calls
}

func runSub(t *testing.T, name string, args ...string) (string, error) {
	t.Helper()
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() != name {
			continue
		}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		err := cmd.RunE(cmd, args)
		return buf.String(), err
	}
	t.Fatalf("subcommand %q not found", name)
	return "", nil
}

func TestPruneLaunchLedgerAfterDisposal_ReportsPrunedCount(t *testing.T) {
	ledgerSeam(t)
	out := new(bytes.Buffer)
	warn := new(bytes.Buffer)

	pruneLaunchLedgerAfterDisposal(out, warn)

	if !strings.Contains(out.String(), "Pruned 1 stale launch-ledger project row(s).") {
		t.Errorf("out = %q, want the pruned-count line", out.String())
	}
	if warn.Len() != 0 {
		t.Errorf("warn = %q, want empty", warn.String())
	}
}

func TestPruneLaunchLedgerAfterDisposal_SilentWhenNothingPruned(t *testing.T) {
	orig := pruneLaunchLedgerFn
	pruneLaunchLedgerFn = func() ([]string, error) { return nil, nil }
	t.Cleanup(func() { pruneLaunchLedgerFn = orig })

	out := new(bytes.Buffer)
	warn := new(bytes.Buffer)
	pruneLaunchLedgerAfterDisposal(out, warn)

	if out.Len() != 0 || warn.Len() != 0 {
		t.Errorf("out = %q warn = %q, want both empty when nothing was pruned", out.String(), warn.String())
	}
}

func TestPruneLaunchLedgerAfterDisposal_WarnsOnError(t *testing.T) {
	orig := pruneLaunchLedgerFn
	pruneLaunchLedgerFn = func() ([]string, error) { return nil, errTestPruneFailed }
	t.Cleanup(func() { pruneLaunchLedgerFn = orig })

	out := new(bytes.Buffer)
	warn := new(bytes.Buffer)
	pruneLaunchLedgerAfterDisposal(out, warn)

	if !strings.Contains(warn.String(), "Warning: could not prune the launch ledger") {
		t.Errorf("warn = %q, want the warning line", warn.String())
	}
	if out.Len() != 0 {
		t.Errorf("out = %q, want empty on failure", out.String())
	}
}

func TestPruneLaunchLedgerAfterDisposal_QuietFormSuppressesReportOnly(t *testing.T) {
	calls := ledgerSeam(t)
	warn := new(bytes.Buffer)

	pruneLaunchLedgerAfterDisposal(nil, warn) // nil out: --auto quiet form

	if *calls != 1 {
		t.Errorf("prune calls = %d, want 1 (quiet suppresses the report, not the reclamation)", *calls)
	}
}

var errTestPruneFailed = errFake("prune boom")

type errFake string

func (e errFake) Error() string { return string(e) }

// --- disposal-path wiring ---

func TestRunRemove_PrunesLaunchLedgerAfterSuccess(t *testing.T) {
	calls := ledgerSeam(t)
	origProvider := WorktreeProvider
	t.Cleanup(func() { WorktreeProvider = origProvider })
	WorktreeProvider = &mockWorktreeManager{}

	_, err := runSub(t, "remove", "/tmp/test-wt")
	if err != nil {
		t.Fatalf("runRemove: %v", err)
	}
	if *calls != 1 {
		t.Errorf("prune calls after successful remove = %d, want 1", *calls)
	}
}

func TestRunRemove_FailedRemovalSkipsPrune(t *testing.T) {
	calls := ledgerSeam(t)
	origProvider := WorktreeProvider
	t.Cleanup(func() { WorktreeProvider = origProvider })
	WorktreeProvider = &mockWorktreeManager{
		removeFunc: func(string, bool) error { return errTestPruneFailed },
	}

	if _, err := runSub(t, "remove", "/tmp/test-wt"); err == nil {
		t.Fatal("runRemove should fail when Remove fails")
	}
	if *calls != 0 {
		t.Errorf("prune calls after failed remove = %d, want 0", *calls)
	}
}

func TestRunDoneAuto_PrunesLaunchLedger(t *testing.T) {
	calls := ledgerSeam(t)
	origProvider := WorktreeProvider
	t.Cleanup(func() { WorktreeProvider = origProvider })
	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{{Path: "/repo/.claude/worktrees/wt", Branch: "feature/x"}}, nil
		},
	}

	success, err := runDoneWorktreeCleanup("feature/x", false, false)
	if err != nil || !success {
		t.Fatalf("runDoneWorktreeCleanup = (%v, %v), want (true, nil)", success, err)
	}
	if *calls != 1 {
		t.Errorf("prune calls after auto cleanup = %d, want 1", *calls)
	}
}

func TestRunDone_InteractivePrunesLaunchLedger(t *testing.T) {
	calls := ledgerSeam(t)
	origProvider := WorktreeProvider
	t.Cleanup(func() { WorktreeProvider = origProvider })
	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{{Path: "/repo/.claude/worktrees/wt", Branch: "feature/x"}}, nil
		},
	}

	if _, err := runSub(t, "done", "feature/x"); err != nil {
		t.Fatalf("runDone: %v", err)
	}
	if *calls != 1 {
		t.Errorf("prune calls after interactive done = %d, want 1", *calls)
	}
}

func TestRunClean_PlainPrunesLaunchLedger(t *testing.T) {
	calls := ledgerSeam(t)
	origProvider := WorktreeProvider
	t.Cleanup(func() { WorktreeProvider = origProvider })
	WorktreeProvider = &mockWorktreeManager{}

	if _, err := runSub(t, "clean"); err != nil {
		t.Fatalf("runClean: %v", err)
	}
	if *calls != 1 {
		t.Errorf("prune calls after plain clean = %d, want 1", *calls)
	}
}

func TestCleanStaleApply_PrunesLaunchLedger(t *testing.T) {
	env := staleTestEnv(t,
		[]git.Worktree{{Path: "/repo/wt-a", Branch: "feature/a"}},
		nil, // clean working trees
		map[string]bool{"feature/a": true},
	)
	calls := ledgerSeam(t)
	_ = env

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("clean --stale --yes: %v", err)
	}
	if *calls != 1 {
		t.Errorf("prune calls after stale apply = %d, want 1", *calls)
	}
	if !strings.Contains(out, "Pruned 1 stale launch-ledger project row(s).") {
		t.Errorf("out = %q, want the pruned-count line", out)
	}
}

func TestCleanStalePreview_DoesNotPrune(t *testing.T) {
	env := staleTestEnv(t,
		[]git.Worktree{{Path: "/repo/wt-a", Branch: "feature/a"}},
		nil,
		map[string]bool{"feature/a": true},
	)
	calls := ledgerSeam(t)
	_ = env

	if _, err := runStaleClean(t, map[string]string{"stale": "true"}); err != nil {
		t.Fatalf("clean --stale preview: %v", err)
	}
	if *calls != 0 {
		t.Errorf("prune calls after preview = %d, want 0 (a preview removes nothing)", *calls)
	}
}

func TestCleanStaleJSON_DoesNotPrune(t *testing.T) {
	env := staleTestEnv(t,
		[]git.Worktree{{Path: "/repo/wt-a", Branch: "feature/a"}},
		nil,
		map[string]bool{"feature/a": true},
	)
	calls := ledgerSeam(t)
	_ = env

	if _, err := runStaleClean(t, map[string]string{"stale": "true", "json": "true"}); err != nil {
		t.Fatalf("clean --stale --json: %v", err)
	}
	if *calls != 0 {
		t.Errorf("prune calls after --json inventory = %d, want 0 (the report removes nothing)", *calls)
	}
}

func TestCleanMergedOnly_PrunesLaunchLedger(t *testing.T) {
	env := staleTestEnv(t,
		[]git.Worktree{{Path: "/repo/wt-a", Branch: "feature/a"}},
		nil, // no ignored content
		map[string]bool{"feature/a": true},
	)
	calls := ledgerSeam(t)
	_ = env

	if _, err := runSub(t, "clean", "--merged-only"); err != nil {
		t.Fatalf("clean --merged-only: %v", err)
	}
	if *calls != 1 {
		t.Errorf("prune calls after merged-only sweep = %d, want 1", *calls)
	}
}
