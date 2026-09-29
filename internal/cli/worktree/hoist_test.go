package worktree

// Hoist tests — SPEC-REPORTS-LIFECYCLE-001 REQ-RLC-005 / REQ-RLC-006
// (AC-RLC-007 / AC-RLC-008).
//
// Coverage map (plan §F.1 M3):
//   - verb normal hoist: tree .moai/reports/ content lands under the main
//     root's .moai/reports/worktrees/<tree-name>/ with relative paths
//     preserved, count + bytes reported;
//   - conflict protection: an existing destination file with different
//     content is never overwritten and every unresolved path is reported;
//   - path refusal: targets outside the resolved project root are refused;
//   - done wiring: the L2 removal path invokes the same hoist routine
//     before removal (--no-hoist opts out).

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// runGitDir runs git in dir, failing the test on error.
func runGitDir(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s (in %s): %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return string(out)
}

// writeFileRel writes content to rel under base, creating parent dirs.
func writeFileRel(t *testing.T, base, rel, content string) {
	t.Helper()
	path := filepath.Join(base, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func readFileStrict(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// overrideMainRootFromTarget swaps gitMainRootFromTargetFunc for the test and
// restores it on cleanup.
func overrideMainRootFromTarget(t *testing.T, mainRoot string) {
	t.Helper()
	orig := gitMainRootFromTargetFunc
	gitMainRootFromTargetFunc = func(string) (string, error) { return mainRoot, nil }
	t.Cleanup(func() { gitMainRootFromTargetFunc = orig })
}

// TestHoistWorktreeReports_Core preserves the routine's contract directly:
// relative paths preserved, byte-identical copies, count and byte totals.
func TestHoistWorktreeReports_Core(t *testing.T) {
	main := t.TempDir()
	tree := filepath.Join(t.TempDir(), "t1320")
	writeFileRel(t, tree, ".moai/reports/evidence.md", "evidence body\n")
	writeFileRel(t, tree, ".moai/reports/sub/nested.md", "nested\n")

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	if res.Count != 2 {
		t.Errorf("Count = %d, want 2", res.Count)
	}
	if res.Bytes != int64(len("evidence body\n")+len("nested\n")) {
		t.Errorf("Bytes = %d, want %d", res.Bytes, len("evidence body\n")+len("nested\n"))
	}
	got := readFileStrict(t, filepath.Join(main, ".moai", "reports", "worktrees", "t1320", "evidence.md"))
	if got != "evidence body\n" {
		t.Errorf("hoisted evidence.md = %q", got)
	}
	got = readFileStrict(t, filepath.Join(main, ".moai", "reports", "worktrees", "t1320", "sub", "nested.md"))
	if got != "nested\n" {
		t.Errorf("hoisted nested.md = %q", got)
	}
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped = %v, want empty", res.Skipped)
	}
}

// TestHoistWorktreeReports_NoReportsDir is the nothing-to-hoist no-op.
func TestHoistWorktreeReports_NoReportsDir(t *testing.T) {
	main := t.TempDir()
	tree := filepath.Join(t.TempDir(), "t999")
	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports without .moai/reports: %v", err)
	}
	if res.Count != 0 || res.Bytes != 0 {
		t.Errorf("no-op hoist = (%d, %d), want (0, 0)", res.Count, res.Bytes)
	}
}

// TestHoistVerb_NormalPath runs the cobra verb end-to-end: evidence lands
// under the main root, output carries the count and byte total.
func TestHoistVerb_NormalPath(t *testing.T) {
	main := t.TempDir()
	overrideMainRootFromTarget(t, main)
	tree := filepath.Join(main, "wt-slug") // a tree under the project root, as real worktrees are
	writeFileRel(t, tree, ".moai/reports/evidence.md", "evidence body\n")

	cmd := newHoistCmd()
	cmd.SetOut(new(strings.Builder))
	errOut := new(strings.Builder)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{tree})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("hoist verb: %v", err)
	}
	output := cmd.OutOrStdout().(*strings.Builder).String() + errOut.String()
	// REQ-RLC-005 output clause: the count and byte total are load-bearing.
	if !strings.Contains(output, "Hoisted 1 file(s)") || !strings.Contains(output, "14 byte(s)") {
		t.Errorf("hoist output must report the count and byte total, got: %s", output)
	}
	want := filepath.Join(main, ".moai", "reports", "worktrees", "wt-slug", "evidence.md")
	if got := readFileStrict(t, want); got != "evidence body\n" {
		t.Errorf("hoisted file = %q", got)
	}
}

// TestHoistVerb_ConflictSkipAndReport (AC-RLC-008): a differing destination
// file is kept byte-identical and its path reported; the rest still hoists.
func TestHoistVerb_ConflictSkipAndReport(t *testing.T) {
	main := t.TempDir()
	overrideMainRootFromTarget(t, main)
	tree := filepath.Join(main, "wt-slug")
	writeFileRel(t, tree, ".moai/reports/evidence.md", "NEW evidence\n")
	writeFileRel(t, tree, ".moai/reports/other.md", "other\n")
	// Pre-existing destination with DIFFERENT content at the same rel path.
	dest := filepath.Join(main, ".moai", "reports", "worktrees", "wt-slug", "evidence.md")
	writeFileRel(t, main, filepath.Join(".moai", "reports", "worktrees", "wt-slug", "evidence.md"), "OLD evidence\n")

	cmd := newHoistCmd()
	cmd.SetOut(new(strings.Builder))
	cmd.SetErr(new(strings.Builder))
	cmd.SetArgs([]string{tree})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("hoist verb: %v", err)
	}
	output := cmd.OutOrStdout().(*strings.Builder).String() + cmd.ErrOrStderr().(*strings.Builder).String()
	if !strings.Contains(output, "evidence.md") || !strings.Contains(output, "Skipped") {
		t.Errorf("output must report the skipped path, got: %s", output)
	}
	if got := readFileStrict(t, dest); got != "OLD evidence\n" {
		t.Errorf("conflicting destination was overwritten: %q", got)
	}
	// Non-conflicting sibling still hoisted.
	if got := readFileStrict(t, filepath.Join(main, ".moai", "reports", "worktrees", "wt-slug", "other.md")); got != "other\n" {
		t.Errorf("non-conflicting file not hoisted: %q", got)
	}
}

// TestHoistVerb_RefusesPathOutsideProjectRoot (REQ-RLC-005 refusal clause).
func TestHoistVerb_RefusesPathOutsideProjectRoot(t *testing.T) {
	main := t.TempDir()
	overrideMainRootFromTarget(t, main)
	outside := filepath.Join(t.TempDir(), "elsewhere-tree")
	writeFileRel(t, outside, ".moai/reports/evidence.md", "should not hoist\n")

	cmd := newHoistCmd()
	cmd.SetOut(new(strings.Builder))
	cmd.SetErr(new(strings.Builder))
	cmd.SetArgs([]string{outside})
	if err := cmd.Execute(); err == nil {
		t.Fatal("hoist of a path outside the project root must be refused")
	}
	if _, err := os.Stat(filepath.Join(main, ".moai", "reports", "worktrees", "elsewhere-tree")); !os.IsNotExist(err) {
		t.Error("refused hoist must not create the destination")
	}
}

// TestDoneL2RemovalHoistsBeforeDisposal (AC-RLC-007 done wiring): a real
// git repo + L2-shaped worktree carrying evidence; the done removal core
// hoists the evidence before Remove, and --no-hoist opts out.
func TestDoneL2RemovalHoistsBeforeDisposal(t *testing.T) {
	repo := t.TempDir()
	runGitDir(t, repo, "init", "-q")
	runGitDir(t, repo, "-c", "user.email=t@test", "-c", "user.name=t", "commit", "--allow-empty", "-q", "-m", "root")
	tree := filepath.Join(repo, "wt-l2")
	runGitDir(t, repo, "worktree", "add", "-q", "-b", "feature/hoist-evidence", tree)
	writeFileRel(t, tree, ".moai/reports/evidence.md", "card evidence\n")

	t.Run("hoists before removal", func(t *testing.T) {
		origProvider := WorktreeProvider
		mock := &mockWorktreeProvider{worktrees: []git.Worktree{{Branch: "feature/hoist-evidence", Path: tree}}}
		WorktreeProvider = mock
		t.Cleanup(func() { WorktreeProvider = origProvider })

		success, err := runDoneWorktreeCleanupWithOptions("feature/hoist-evidence", false, false, true)
		if err != nil || !success {
			t.Fatalf("runDoneWorktreeCleanupWithOptions = (%v, %v)", success, err)
		}
		if !mock.removeCalled {
			t.Fatal("Remove was not called — hoist must not block a clean disposal")
		}
		hoisted := filepath.Join(repo, ".moai", "reports", "worktrees", "wt-l2", "evidence.md")
		if got := readFileStrict(t, hoisted); got != "card evidence\n" {
			t.Errorf("hoisted evidence = %q", got)
		}
	})

	t.Run("no-hoist opts out", func(t *testing.T) {
		origProvider := WorktreeProvider
		mock := &mockWorktreeProvider{worktrees: []git.Worktree{{Branch: "feature/hoist-evidence", Path: tree}}}
		WorktreeProvider = mock
		t.Cleanup(func() { WorktreeProvider = origProvider })

		// The previous subtest legitimately created the destination; clear it
		// so this subtest observes its own (non-)hoist in isolation.
		dest := filepath.Join(repo, ".moai", "reports", "worktrees", "wt-l2")
		if err := os.RemoveAll(dest); err != nil {
			t.Fatalf("clear destination: %v", err)
		}

		if _, err := runDoneWorktreeCleanupWithOptions("feature/hoist-evidence", false, false, false); err != nil {
			t.Fatalf("runDoneWorktreeCleanupWithOptions: %v", err)
		}
		if _, err := os.Stat(dest); !os.IsNotExist(err) {
			t.Error("--no-hoist must not create the hoist destination")
		}
	})
}
