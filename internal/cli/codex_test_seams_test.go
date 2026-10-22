package cli

// codex_test_seams_test.go — test seam helpers rehomed from the removed
// codex task/job test files (card t1909 step ⑤); the staying glm/mcp audit
// tests consume the same seams.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/session"
)

// withCodexProjectDir points the project-root seam at root for the duration of
// the test. codex_setup and the audit paths resolve their config through it.
func withCodexProjectDir(t *testing.T, root string) {
	t.Helper()
	prev := projectDirResolver
	projectDirResolver = func() string { return root }
	t.Cleanup(func() { projectDirResolver = prev })
}

// seamsForInit swaps the wizard seams the same way the autonomy-wiring tests
// do, so runInit runs non-interactively against the injected result.
func seamsForInit(t *testing.T) {
	t.Helper()
	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{AutonomyTier: config.AutonomyTierAutomatic}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })
}

// writeCodexWorkflowYAML writes .moai/config/sections/workflow.yaml with the
// given body. An empty body writes no file at all (the absent-file state).
func writeCodexWorkflowYAML(t *testing.T, root, body string) {
	t.Helper()
	if body == "" {
		return
	}
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// resultText concatenates the text content of a tool result. An IsError result
// carries its explanation there rather than in StructuredContent (see toolErr).
func resultText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

// writeFileForTest overwrites path with body.
func writeFileForTest(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// withCodexProjectRoot pins the project-root seam to root (AC-CL-002's cwd
// axis runs through it, never through the raw process cwd).
func withCodexProjectRoot(t *testing.T, root string) *int {
	t.Helper()
	calls := 0
	prev := findProjectRootFn
	findProjectRootFn = func() (string, error) { calls++; return root, nil }
	t.Cleanup(func() { findProjectRootFn = prev })
	return &calls
}

// anchorGit runs git with -C dir and fails the test on error.
func anchorGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git -C %s %v: %v\n%s", dir, args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// anchorRepo is a real repository with a project marker, its base branch, and
// the helpers that build trees under <root>/.claude/worktrees/.
type anchorRepo struct {
	root string
	base string
}

func newAnchorRepo(t *testing.T) *anchorRepo {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	root := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	anchorGit(t, root, "init", "-q")
	anchorGit(t, root, "config", "user.email", "anchor-test@example.com")
	anchorGit(t, root, "config", "user.name", "Anchor Test")
	anchorGit(t, root, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	anchorGit(t, root, "add", "README.md")
	anchorGit(t, root, "commit", "-q", "-m", "seed")
	r := &anchorRepo{root: root, base: anchorGit(t, root, "branch", "--show-current")}

	// Keep every registry the anchor decision consults inside the fixture.
	t.Setenv(config.EnvClaudeProjectDir, root)
	withCodexProjectRoot(t, root)

	t.Cleanup(func() {
		out, _ := exec.Command("git", "-C", root, "worktree", "list", "--porcelain").Output()
		for path := range session.ParseWorktreeLocks(string(out)) {
			_ = exec.Command("git", "-C", root, "worktree", "unlock", path).Run()
		}
	})
	return r
}

// tree creates an unlocked linked worktree <root>/.claude/worktrees/<name>.
func (r *anchorRepo) tree(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(r.root, ".claude", "worktrees", name)
	anchorGit(t, r.root, "worktree", "add", "-q", "-b", "WT-"+name, path, r.base)
	return path
}

func firstLineStr(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

// resumeHome isolates HOME and installs the shared init wizard seams.
func resumeHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	seamsForInit(t)
}

// isolatedInitHome points HOME at a fresh directory and installs the init
// wizard seams.
func isolatedInitHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	seamsForInit(t)
}
