package hook

// worktree_registry_m5_test.go — SPEC-WORKTREE-HOME-ROOT-001 M5 RED (card
// t1640): AC-WHR-012a's re-key clause and AC-WHR-017b's state-file clauses
// (2), (4) and the file-unchanged halves of (6) and (7).
//
// Written before the re-key operation, the state-file lock and the atomic
// write exist — RekeyWorktreeEntry, RekeyWorktreeEntryInverse,
// StateLockLapseBudget, ErrStateLockWaitLapsed and the stateLockWaitBudget
// seam are undefined here, so this package fails to compile and that
// compile failure is the RED (plan §C test-compilation order; research
// §R13's M5 row). The pins: a re-key keys the entry to the new path and its
// inverse restores it; a fault between the temp write and the rename leaves
// the previous file intact; a state file holding invalid JSON fails every
// one of the four writers with the file byte-identical; two writers across
// processes lose neither entry; and a lock held live past the lowered wait
// bound lapses with a distinct error and the file unchanged.

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// m5Fixture writes a state file with the named entry paths and returns the
// project directory and the state file path.
func m5Fixture(t *testing.T, paths ...string) (string, string) {
	t.Helper()
	projectDir := t.TempDir()
	stateFile := worktreeStateFile(projectDir)
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	entries := make([]WorktreeEntry, 0, len(paths))
	now := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	for _, p := range paths {
		entries = append(entries, WorktreeEntry{Path: p, Branch: "main", AgentName: "m5", CreatedAt: now})
	}
	data, err := marshalWorktreeEntries(entries)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(stateFile, data, 0o644); err != nil {
		t.Fatalf("write state file: %v", err)
	}
	return projectDir, stateFile
}

// m5ReadStateFile returns the raw bytes of the state file.
func m5ReadStateFile(t *testing.T, stateFile string) string {
	t.Helper()
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}
	return string(data)
}

// TestRekeyWorktreeEntry_KeysTheEntryToTheNewPath pins AC-WHR-012a's re-key
// clause: the entry keyed to the old path is keyed to the new path, its
// other fields ride along, and the other entries are untouched.
func TestRekeyWorktreeEntry_KeysTheEntryToTheNewPath(t *testing.T) {
	pOld := "/repo/.moai/worktrees/t1"
	pNew := "/home/worktrees/projects/repo/t1"
	pOther := "/repo/.claude/worktrees/t2"
	projectDir, stateFile := m5Fixture(t, pOld, pOther)

	if err := RekeyWorktreeEntry(projectDir, pOld, pNew); err != nil {
		t.Fatalf("RekeyWorktreeEntry: %v", err)
	}
	entries := loadWorktreeEntries(stateFile)
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	seen := map[string]WorktreeEntry{}
	for _, e := range entries {
		seen[e.Path] = e
	}
	rekeyed, ok := seen[pNew]
	if !ok {
		t.Fatalf("no entry keyed to %s: %+v", pNew, entries)
	}
	if rekeyed.Branch != "main" || rekeyed.AgentName != "m5" {
		t.Errorf("re-keyed entry lost its fields: %+v", rekeyed)
	}
	if _, ok := seen[pOld]; ok {
		t.Errorf("the old key %s is still present", pOld)
	}
	if _, ok := seen[pOther]; !ok {
		t.Errorf("the untouched entry %s vanished", pOther)
	}
}

// TestRekeyWorktreeEntry_InverseRestoresTheOriginalPath pins the inverse
// operation: replacing the new path with the old one restores the original
// keying (the reverse re-key of REQ-WHR-012).
func TestRekeyWorktreeEntry_InverseRestoresTheOriginalPath(t *testing.T) {
	pOld := "/repo/.moai/worktrees/t1"
	pNew := "/home/worktrees/projects/repo/t1"
	projectDir, stateFile := m5Fixture(t, pOld)

	if err := RekeyWorktreeEntry(projectDir, pOld, pNew); err != nil {
		t.Fatalf("forward re-key: %v", err)
	}
	if err := RekeyWorktreeEntryInverse(projectDir, pNew, pOld); err != nil {
		t.Fatalf("inverse re-key: %v", err)
	}
	entries := loadWorktreeEntries(stateFile)
	if len(entries) != 1 || entries[0].Path != pOld {
		t.Fatalf("entries = %+v, want the single entry restored to %s", entries, pOld)
	}
}

// TestRekeyWorktreeEntry_MissingEntryFails pins that a re-key naming a path
// with no entry fails its step instead of writing anything.
func TestRekeyWorktreeEntry_MissingEntryFails(t *testing.T) {
	projectDir, stateFile := m5Fixture(t, "/repo/.moai/worktrees/t1")
	before := m5ReadStateFile(t, stateFile)

	if err := RekeyWorktreeEntry(projectDir, "/repo/.moai/worktrees/no-such", "/x"); err == nil {
		t.Fatal("a re-key of a missing entry was accepted")
	}
	if after := m5ReadStateFile(t, stateFile); after != before {
		t.Errorf("the failed re-key rewrote the file")
	}
}

// TestWorktreeRegistry_AtomicFaultKeepsPreviousFile pins the atomic-replace
// contract: a fault injected between the temp write and the rename fails the
// writer and leaves the previous file intact and parseable.
func TestWorktreeRegistry_AtomicFaultKeepsPreviousFile(t *testing.T) {
	pOld := "/repo/.moai/worktrees/t1"
	pNew := "/home/worktrees/projects/repo/t1"
	projectDir, stateFile := m5Fixture(t, pOld)
	before := m5ReadStateFile(t, stateFile)

	orig := worktreeAtomicFault
	worktreeAtomicFault = func() error {
		return errM5Injected
	}
	t.Cleanup(func() { worktreeAtomicFault = orig })

	if err := RekeyWorktreeEntry(projectDir, pOld, pNew); err == nil {
		t.Fatal("the faulted re-key was accepted")
	}
	if after := m5ReadStateFile(t, stateFile); after != before {
		t.Errorf("the faulted write changed the state file:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	// The previous file is still parseable.
	if entries := loadWorktreeEntries(stateFile); len(entries) != 1 || entries[0].Path != pOld {
		t.Errorf("the previous file is no longer parseable as its own content: %+v", entries)
	}
}

// TestWorktreeRegistry_InvalidJSONFailsEveryWriterByteIdentical pins
// AC-WHR-017b step (6)'s file-unchanged half: a state file holding invalid
// JSON fails each of the four writers (the re-key, registerWorktree,
// RegisterLauncherWorktree, unregisterWorktree) with the file
// byte-identical afterwards — never an empty rewrite.
func TestWorktreeRegistry_InvalidJSONFailsEveryWriterByteIdentical(t *testing.T) {
	projectDir, stateFile := m5Fixture(t)
	if err := os.WriteFile(stateFile, []byte("{ not json"), 0o644); err != nil {
		t.Fatalf("write invalid json: %v", err)
	}
	before := "{ not json"

	if err := RekeyWorktreeEntry(projectDir, "/a", "/b"); err == nil {
		t.Error("re-key accepted an unparseable state file")
	}
	if err := registerWorktree(projectDir, "/a", "b", "m5"); err == nil {
		t.Error("registerWorktree accepted an unparseable state file")
	}
	RegisterLauncherWorktree(projectDir, "/a", "b", "m5") // logs; must not rewrite
	if err := unregisterWorktree(projectDir, "/a"); err == nil {
		t.Error("unregisterWorktree accepted an unparseable state file")
	}
	if after := m5ReadStateFile(t, stateFile); after != before {
		t.Errorf("a writer rewrote the unparseable state file:\n%q", after)
	}
}

// TestWorktreeRegistry_TwoWritersAcrossProcessesLoseNeitherEntry pins
// AC-WHR-017b step (2)'s no-loss half: two writers in different processes,
// one registering a seventh tree and one re-keying a sixth, lose neither
// entry. The children are this test's own binary re-executed with the
// M5_HELPER env guard, bounded by the command's context timeout.
func TestWorktreeRegistry_TwoWritersAcrossProcessesLoseNeitherEntry(t *testing.T) {
	paths := []string{
		"/repo/.claude/worktrees/t6",
		"/repo/.claude/worktrees/t7",
	}
	projectDir, stateFile := m5Fixture(t, paths[0], paths[1])
	lockPath := stateFile + ".lock"

	run := func(helper string, entryPath string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestM5HelperProcess")
		cmd.Env = append(os.Environ(),
			"M5_HELPER="+helper,
			"M5_PROJECT_DIR="+projectDir,
			"M5_LOCK_PATH="+lockPath,
			"M5_ENTRY_PATH="+entryPath,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("helper %s: %v\n%s", helper, err, out)
		}
	}
	go run("m5-register", "/repo/.claude/worktrees/t7")
	run("m5-rekey", "/repo/.claude/worktrees/t6") // re-keys t6 inside the child

	time.Sleep(200 * time.Millisecond) // let both children finish their sections

	entries := loadWorktreeEntries(stateFile)
	seen := map[string]bool{}
	for _, e := range entries {
		seen[e.Path] = true
	}
	// No writer's update is lost: the re-key child's rename and the
	// register child's entry are both present afterwards.
	if !seen[paths[1]] || !seen[paths[0]+"-moved"] {
		t.Errorf("a writer's update was lost: %+v", entries)
	}
	if seen[paths[0]] {
		t.Errorf("the old key %s survived the re-key: %+v", paths[0], entries)
	}
}

// TestWorktreeRegistry_HeldLockPastLoweredBoundLapses pins AC-WHR-017b
// step (7)'s file-unchanged half: a live holder past the lowered wait bound
// makes the writer lapse with the distinct error, the file byte-identical,
// and neither an early return nor a hang past the bound.
func TestWorktreeRegistry_HeldLockPastLoweredBoundLapses(t *testing.T) {
	projectDir, stateFile := m5Fixture(t, "/repo/.claude/worktrees/t1")
	lockPath := stateFile + ".lock"

	// A live child the test started holds the state-file lock for 2s and
	// exits on its own (no kill needed). It signals readiness by creating
	// the ready file after the lock is held, so the parent's write cannot
	// race ahead of the acquisition.
	readyFile := lockPath + ".ready"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestM5HelperProcess")
	cmd.Env = append(os.Environ(),
		"M5_HELPER=hold-lock",
		"M5_LOCK_PATH="+lockPath,
		"M5_READY_FILE="+readyFile,
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start holder child: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Wait() })
	t.Cleanup(func() { _ = os.Remove(readyFile) })

	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(readyFile); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the holder child never signaled readiness")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Lower the wait bound through the seam for this run only.
	origBudget := stateLockWaitBudget
	stateLockWaitBudget = 100 * time.Millisecond
	t.Cleanup(func() { stateLockWaitBudget = origBudget })

	start := time.Now()
	err := RegisterLauncherWorktreeE(projectDir, "/repo/.claude/worktrees/t1-new", "b", "m5")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("the write succeeded while the lock was held past the bound")
	}
	if !errors.Is(err, ErrStateLockWaitLapsed) {
		t.Errorf("error %v is not the distinct lapse error", err)
	}
	if elapsed < 50*time.Millisecond {
		t.Errorf("the write returned before the bound lapsed (%v)", elapsed)
	}
	if elapsed > 10*time.Second {
		t.Errorf("the write hung past the bound (%v)", elapsed)
	}
	if after := m5ReadStateFile(t, stateFile); strings.Contains(after, "t1-new") {
		t.Errorf("the lapsed write rewrote the file:\n%s", after)
	}
	_ = ctx // the context bounds the child; the child exits on its own
	_ = cancel
}

// TestStateLockLapseBudget_IsOne pins the budget constant the M6 engine
// consumes: one lapse ends the run.
func TestStateLockLapseBudget_IsOne(t *testing.T) {
	if StateLockLapseBudget != 1 {
		t.Errorf("StateLockLapseBudget = %d, want 1", StateLockLapseBudget)
	}
	if ErrStateLockWaitLapsed.Error() == "" {
		t.Error("ErrStateLockWaitLapsed must carry a message")
	}
}

// TestM5HelperProcess is the re-execution target of the cross-process
// children; normal runs skip it on the missing env guard.
func TestM5HelperProcess(t *testing.T) {
	switch os.Getenv("M5_HELPER") {
	case "m5-register":
		projectDir := os.Getenv("M5_PROJECT_DIR")
		entryPath := os.Getenv("M5_ENTRY_PATH")
		RegisterLauncherWorktree(projectDir, entryPath, "worker-branch", "m5-child")
	case "m5-rekey":
		projectDir := os.Getenv("M5_PROJECT_DIR")
		oldPath := os.Getenv("M5_ENTRY_PATH")
		_ = RekeyWorktreeEntry(projectDir, oldPath, oldPath+"-moved")
	case "hold-lock":
		lockPath := os.Getenv("M5_LOCK_PATH")
		readyFile := os.Getenv("M5_READY_FILE")
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			os.Exit(3)
		}
		if err := lockfile.Lock(f); err != nil {
			os.Exit(4)
		}
		if readyFile != "" {
			_ = os.WriteFile(readyFile, []byte("held"), 0o644)
		}
		time.Sleep(2 * time.Second)
		_ = lockfile.Unlock(f)
		_ = f.Close()
	default:
		t.Skip("M5_HELPER not set")
	}
}
