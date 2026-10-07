package cli

// factory_merge_test.go — the lane-direct merge verbs (AC-FLA-009/010/011).
//
// The fixtures build throwaway git repositories under t.TempDir: every merge
// the tests perform happens inside those repositories, never into any real
// develop branch, and the integration window is exercised through the
// package's own record APIs against a throwaway lock root
// (CLAUDE_PROJECT_DIR), never the developer's state.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/spf13/cobra"
)

// mergeReadyFixture builds the environment one `factory merge ready` run
// resolves: a throwaway repository whose card branch has absorbed develop
// (so the triple's merge probes read clean), a SPEC directory whose sync
// record carries the given sync_status value, and a throwaway lock root.
// The process cwd moves into the repository; the lock root is exported as
// CLAUDE_PROJECT_DIR.
func mergeReadyFixture(t *testing.T, syncStatus string, lane, sessionID string) (repo, lockRoot, specID string) {
	t.Helper()
	repo = t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(rel, content string) {
		if err := os.WriteFile(filepath.Join(repo, rel), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	git("init", "-b", "develop")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")
	write("base.txt", "base\n")
	git("add", ".")
	git("commit", "-m", "base")
	git("checkout", "-b", "WT-card")
	write("card.txt", "card\n")
	git("add", ".")
	git("commit", "-m", "card work")
	git("checkout", "develop")
	write("dev.txt", "dev\n")
	git("add", ".")
	git("commit", "-m", "develop work")
	git("checkout", "WT-card")
	git("merge", "--no-ff", "-m", "absorb develop", "develop")

	specID = "SPEC-MERGE-READY-FIXTURE-001"
	section := ""
	if syncStatus != "" {
		section = "## §E.4 Sync-phase Audit-Ready Signal\n\nsync_status: " + syncStatus + "\n"
	}
	specDir := filepath.Join(repo, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"),
		[]byte("## §E.2 Run-phase Evidence\n\n...\n\n"+section), 0o644); err != nil {
		t.Fatal(err)
	}

	lockRoot = t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", lockRoot)
	t.Setenv(config.EnvMoaiFactoryWorker, lane)
	// REQ-MWQ-021 (card t1479): merge-readiness runs the FOURTH condition —
	// a valid re-measure record keyed to the candidate tree. Seed one for
	// the WT-card tree; a test that needs the condition to FAIL uses a
	// syncStatus that fails earlier (sync-audit) or clears the store.
	tree := strings.TrimSpace(func() string {
		cmd := exec.Command("git", "rev-parse", "WT-card^{tree}")
		cmd.Dir = repo
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("read candidate tree: %v", err)
		}
		return string(out)
	}())
	if err := factory.WriteRemeasureRecord(lockRoot, strings.TrimSpace(tree), factory.RemeasureRecord{
		Base:          "seeded-by-mergeReadyFixture",
		Command:       "true",
		ExitCode:      0,
		BuildIdentity: "moai merge-ready fixture",
		RecordedAt:    time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	return repo, lockRoot, specID
}

// holdWindowFor takes the window on behalf of another lane through the
// package's own acquire API — the forged-ledger path, never a real merge.
func holdWindowFor(t *testing.T, lockRoot, sessionID, laneName string) {
	t.Helper()
	_, err := factory.AcquireIntegrationLock(lockRoot, factory.IntegrationLock{
		SessionID:   sessionID,
		SessionName: laneName,
		PID:         os.Getpid(), // this test process: a live holder
		Branch:      "develop",
		AcquiredAt:  time.Now().UTC().Format(time.RFC3339),
	}, false)
	if err != nil {
		t.Fatalf("forge window hold: %v", err)
	}
}

func runFactoryMerge(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newFactoryCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	walk(cmd, func(c *cobra.Command) { c.SilenceUsage = true; c.SilenceErrors = true })
	err := cmd.Execute()
	return out.String(), err
}

// readLockRoot reads the lock record from the fixture's lock root.
func readLockRoot(t *testing.T, lockRoot string) factory.IntegrationLock {
	t.Helper()
	lock, err := factory.ReadIntegrationLock(lockRoot)
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	return *lock
}

// The full lane-direct sequence on the absorbed shape: the triple is
// evaluated and recorded, the window is taken, and the recorded checks
// predate the window's acquire stamp (AC-FLA-009's second half).
func TestFactoryMergeReady_ClearsAndTakesWindow(t *testing.T) {
	_, lockRoot, specID := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")

	out, err := runFactoryMerge(t, "merge", "ready", "--card", "t9001", "--spec", specID, "--branch", "WT-card", "--develop", "develop", "--session", "sess-lane-9", "--json")
	if err != nil {
		t.Fatalf("merge ready: %v", err)
	}
	if !strings.Contains(out, `"verdict":"cleared"`) {
		t.Fatalf("ready did not clear on the absorbed shape: %s", out)
	}

	// The window the path took is recorded with this lane as holder.
	lock := readLockRoot(t, lockRoot)
	if !lock.Held() || lock.SessionName != "lane-9" || lock.Card != "t9001" {
		t.Fatalf("window not recorded for lane-9/t9001: %+v", lock)
	}

	// AC-FLA-009 second half, read back from the records: the recorded run
	// passed and predates the acquire stamp. The stamp carries RFC3339
	// second precision, so the comparison runs at the stamp's own
	// resolution — exactly what the verb's cleared verdict is built on.
	acquireAt, err := time.Parse(time.RFC3339, lock.AcquiredAt)
	if err != nil {
		t.Fatalf("acquired_at %q: %v", lock.AcquiredAt, err)
	}
	store := factorylane.NewStore(".", nil)
	run, err := store.LatestMergeCheckRun("lane-9", "t9001")
	if err != nil || run == nil {
		t.Fatalf("recorded run missing: run=%+v err=%v", run, err)
	}
	if !run.AllPassed {
		t.Fatalf("recorded run did not pass: %+v", run)
	}
	if !run.CheckedAt.Before(acquireAt.Add(time.Second)) {
		t.Fatalf("recorded at %s is not within or before the acquire stamp's second %s",
			run.CheckedAt, acquireAt)
	}
}

// A failing triple refuses BEFORE any window is taken: the failing condition
// is named, the run is recorded, and the lock root stays empty.
func TestFactoryMergeReady_RefusesFailingConditionBeforeWindow(t *testing.T) {
	_, lockRoot, specID := mergeReadyFixture(t, "audit-ready", "lane-9", "sess-lane-9")

	out, err := runFactoryMerge(t, "merge", "ready", "--card", "t9001", "--spec", specID, "--branch", "WT-card", "--develop", "develop", "--session", "sess-lane-9", "--json")
	if err != nil {
		t.Fatalf("merge ready errored on a refusal (a refusal is a verdict, exit 0): %v", err)
	}
	if !strings.Contains(out, "sync-audit") {
		t.Fatalf("refusal does not name the failing condition: %s", out)
	}
	if lock := readLockRoot(t, lockRoot); lock.Held() {
		t.Fatalf("window taken despite a failing triple: %+v", lock)
	}
	store := factorylane.NewStore(".", nil)
	run, err := store.LatestMergeCheckRun("lane-9", "t9001")
	if err != nil || run == nil {
		t.Fatalf("failing run not recorded: run=%+v err=%v", run, err)
	}
	if run.AllPassed || run.FailedCondition != "sync-audit" {
		t.Errorf("recorded run shape wrong: %+v", run)
	}
}

// AC-FLA-010: acquire against a held window is refused with the holder
// named, the window stays with its holder, and the lane's checks are
// already recorded — the verdict the lane waits on.
func TestFactoryMergeReady_HeldWindowRefusedWithHolderNamed(t *testing.T) {
	_, lockRoot, specID := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")
	holdWindowFor(t, lockRoot, "sess-other", "lane-7")

	out, err := runFactoryMerge(t, "merge", "ready", "--card", "t9001", "--spec", specID, "--branch", "WT-card", "--develop", "develop", "--session", "sess-lane-9", "--json")
	if err != nil {
		t.Fatalf("merge ready errored on contention (exit 0 with the holder named): %v", err)
	}
	if !strings.Contains(out, "lane-7") {
		t.Fatalf("refusal does not name the holder lane-7: %s", out)
	}
	if !strings.Contains(out, `"verdict":"waiting"`) {
		t.Fatalf("contention verdict not reported as waiting: %s", out)
	}
	lock := readLockRoot(t, lockRoot)
	if lock.SessionID != "sess-other" || lock.SessionName != "lane-7" {
		t.Fatalf("window moved off its holder: %+v", lock)
	}
}

// AC-FLA-011, negative case at the gate: with no window record the merge
// gate refuses, naming the absence — a merge without the window is not
// available on this path.
func TestFactoryMergeGate_NoRecordRefuses(t *testing.T) {
	mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")

	out, err := runFactoryMerge(t, "merge", "gate", "--card", "t9001")
	if err != nil {
		t.Fatalf("merge gate errored on a refusal (a refusal is a verdict, exit 0): %v", err)
	}
	if !strings.Contains(out, "no integration acquire record") {
		t.Fatalf("gate refusal does not name the absent record: %s", out)
	}
	if !strings.Contains(out, "REFUSED") {
		t.Fatalf("gate refusal not marked: %s", out)
	}
}

// AC-FLA-011, positive case: with a live acquire record for this lane the
// gate proceeds — the window covers the moment.
func TestFactoryMergeGate_LiveHoldProceeds(t *testing.T) {
	_, lockRoot, _ := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")
	holdWindowFor(t, lockRoot, "sess-lane-9", "lane-9")

	out, err := runFactoryMerge(t, "merge", "gate", "--card", "t9001")
	if err != nil {
		t.Fatalf("merge gate errored: %v", err)
	}
	if !strings.Contains(out, "PROCEED") {
		t.Fatalf("gate did not proceed under a live hold: %s", out)
	}

	// A foreign lane's hold does not cover this lane's merge.
	other := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", other)
	holdWindowFor(t, other, "sess-x", "lane-2")
	out, err = runFactoryMerge(t, "merge", "gate", "--card", "t9001")
	if err != nil {
		t.Fatalf("merge gate errored: %v", err)
	}
	if !strings.Contains(out, "lane-2") || strings.Contains(out, "PROCEED") {
		t.Fatalf("foreign holder did not refuse: %s", out)
	}
}
