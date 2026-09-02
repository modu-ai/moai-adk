// memory_drain_test.go — `moai memory drain` (SPEC-AGENT-MEMORY-DRAIN-001).
//
// The command's environment is fully seamed: primary resolution and worktree
// enumeration are function-variable seams pointing at t.TempDir() fixtures,
// so no real sibling tree is ever enumerated or written.
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// swapMemoryDrainSeams points the drain command's environment at fixture
// directories: primary is the primary checkout root, trees are the linked
// worktree roots to enumerate.
func swapMemoryDrainSeams(t *testing.T, primary string, trees []string) {
	t.Helper()
	prevRoot := memoryDrainPrimaryRoot
	prevList := memoryDrainListWorktrees
	memoryDrainPrimaryRoot = func(string) (string, bool, error) {
		return primary, false, nil
	}
	memoryDrainListWorktrees = func(string) ([]string, error) {
		return trees, nil
	}
	t.Cleanup(func() {
		memoryDrainPrimaryRoot = prevRoot
		memoryDrainListWorktrees = prevList
	})
}

// runMemoryDrain executes the drain command with args and returns stdout,
// stderr, and the execution error.
func runMemoryDrain(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newMemoryCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(append([]string{"drain"}, args...))
	execErr := cmd.Execute()
	return out.String(), errOut.String(), execErr
}

// TestMemoryDrainListedInHelp pins AC-AM-001a's surface half: the verb must
// be discoverable on `moai memory --help`.
func TestMemoryDrainListedInHelp(t *testing.T) {
	t.Parallel()
	cmd := newMemoryCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("memory --help: %v", err)
	}
	if !strings.Contains(out.String(), "drain") {
		t.Errorf("`drain` missing from `moai memory --help`:\n%s", out.String())
	}
}

// TestMemoryDrainPreviewDefaultWritesNothing covers AC-AM-001a: without the
// apply flag the command prints the per-tree copy set, exits 0, and writes
// nothing to the primary store.
func TestMemoryDrainPreviewDefaultWritesNothing(t *testing.T) {
	primary := t.TempDir()
	tree := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tree, ".claude", "agent-memory", "manager-spec"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tree, ".claude", "agent-memory", "manager-spec", "feedback_x.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	swapMemoryDrainSeams(t, primary, []string{tree})

	out, _, execErr := runMemoryDrain(t)
	if execErr != nil {
		t.Fatalf("drain (preview): %v", execErr)
	}
	if !strings.Contains(out, "feedback_x.md") {
		t.Errorf("preview does not name the would-be copy:\n%s", out)
	}
	if !strings.Contains(out, "preview") {
		t.Errorf("preview output does not announce preview mode:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(primary, ".claude", "agent-memory")); !os.IsNotExist(err) {
		t.Errorf("preview wrote into the primary store (err=%v)", err)
	}
}

// TestMemoryDrainJSONRecords covers AC-AM-001b: --json emits a single JSON
// array whose records carry path, agents, files, copied, collided, skipped.
func TestMemoryDrainJSONRecords(t *testing.T) {
	primary := t.TempDir()
	tree := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tree, ".claude", "agent-memory", "manager-spec"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tree, ".claude", "agent-memory", "manager-spec", "feedback_x.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	swapMemoryDrainSeams(t, primary, []string{tree})

	out, _, execErr := runMemoryDrain(t, "--json")
	if execErr != nil {
		t.Fatalf("drain --json: %v", execErr)
	}
	var records []hook.TreeDrainRecord
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &records); err != nil {
		t.Fatalf("--json output is not a JSON array: %v\n%s", err, out)
	}
	if len(records) != 1 {
		t.Fatalf("got %d record(s), want 1", len(records))
	}
	r := records[0]
	if r.Path != tree || r.Files != 1 || r.Copied != 1 || r.Collided != 0 || r.Skipped != 0 {
		t.Errorf("record fields wrong: %+v", r)
	}
	if len(r.Agents) != 1 || r.Agents[0].Name != "manager-spec" {
		t.Errorf("record agents wrong: %+v", r.Agents)
	}
}

// TestMemoryDrainApplyCopies covers the --yes path end-to-end through the
// command.
func TestMemoryDrainApplyCopies(t *testing.T) {
	primary := t.TempDir()
	tree := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tree, ".claude", "agent-memory", "manager-spec"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tree, ".claude", "agent-memory", "manager-spec", "feedback_x.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	swapMemoryDrainSeams(t, primary, []string{tree})

	out, _, execErr := runMemoryDrain(t, "--yes")
	if execErr != nil {
		t.Fatalf("drain --yes: %v", execErr)
	}
	if _, err := os.Stat(filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "feedback_x.md")); err != nil {
		t.Errorf("apply did not copy: %v", err)
	}
	if strings.Contains(out, "preview") {
		t.Errorf("apply output still announces preview:\n%s", out)
	}
}

// TestMemoryDrainSkipsThePrimaryItself pins the no-self-drain invariant: the
// enumerated set excludes the primary checkout.
func TestMemoryDrainSkipsThePrimaryItself(t *testing.T) {
	primary := t.TempDir()
	tree := t.TempDir()
	swapMemoryDrainSeams(t, primary, []string{primary, tree})

	out, _, execErr := runMemoryDrain(t, "--json")
	if execErr != nil {
		t.Fatalf("drain --json: %v", execErr)
	}
	var records []hook.TreeDrainRecord
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &records); err != nil {
		t.Fatalf("--json output is not a JSON array: %v\n%s", err, out)
	}
	for _, r := range records {
		if r.Path == primary {
			t.Errorf("the primary checkout was drained as a worktree: %+v", r)
		}
	}
}

// TestMemoryDrainRejectsUnknownFlags pins the verb-fallthrough lesson: an
// unknown flag must exit non-zero, never silently fall through to help.
func TestMemoryDrainRejectsUnknownFlags(t *testing.T) {
	primary := t.TempDir()
	swapMemoryDrainSeams(t, primary, nil)

	_, _, execErr := runMemoryDrain(t, "--definitely-not-a-flag")
	if execErr == nil {
		t.Error("unknown flag accepted (fell through to help?)")
	}
}

// gitRunIn runs a git command in dir, failing the test on error.
func gitRunIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git -C %s %v: %v\n%s", dir, args, err, out)
	}
}

// TestListWorktreesRealExcludesPrimary exercises the real git-backed
// enumeration against a throwaway repository: linked trees are listed, the
// main worktree is not.
func TestListWorktreesRealExcludesPrimary(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	primary := t.TempDir()
	// git reports /private/var for /var on macOS; compare in git's spelling.
	resolvedPrimary, evalErr := filepath.EvalSymlinks(primary)
	if evalErr != nil {
		t.Fatalf("eval symlinks: %v", evalErr)
	}
	primary = resolvedPrimary
	gitRunIn(t, primary, "init", "-q")
	gitRunIn(t, primary, "config", "user.email", "t@example.com")
	gitRunIn(t, primary, "config", "user.name", "t")
	gitRunIn(t, primary, "commit", "-q", "--allow-empty", "-m", "init")
	wtParent := t.TempDir()
	resolvedWtParent, evalErr := filepath.EvalSymlinks(wtParent)
	if evalErr != nil {
		t.Fatalf("eval symlinks (wt): %v", evalErr)
	}
	wt := filepath.Join(resolvedWtParent, "wt")
	gitRunIn(t, primary, "worktree", "add", "-q", "--detach", wt)

	trees, err := listWorktreesReal(primary)
	if err != nil {
		t.Fatalf("listWorktreesReal: %v", err)
	}
	found := false
	for _, p := range trees {
		if filepath.Clean(p) == filepath.Clean(primary) {
			t.Fatalf("primary listed as a linked worktree: %v", trees)
		}
		if filepath.Clean(p) == filepath.Clean(wt) {
			found = true
		}
	}
	if !found {
		t.Errorf("linked worktree missing from enumeration: %v", trees)
	}

	if _, err := listWorktreesReal(t.TempDir()); err == nil {
		t.Error("non-repo primary enumerated without error")
	}
}
