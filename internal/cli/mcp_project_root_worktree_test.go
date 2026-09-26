package cli

// SPEC-MCP-WORKTREE-UNTRACKED-001 — a linked worktree of a repository that
// keeps .moai/ untracked has no .moai of its own, yet it is still a MoAI tree:
// its primary checkout carries the configuration. These tests build that layout
// under t.TempDir() with a git configuration isolated from the machine.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// wtGitEnv returns an environment for fixture git commands: inherited GIT_*
// variables removed, and the global/system configuration pointed at empty files
// so user settings (worktree.useRelativePaths, init.defaultBranch, …) cannot
// change the fixture layout.
func wtGitEnv(t *testing.T) []string {
	t.Helper()
	empty := filepath.Join(t.TempDir(), "empty-gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	env := []string{}
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "GIT_") {
			continue
		}
		env = append(env, kv)
	}
	return append(env,
		"GIT_CONFIG_GLOBAL="+empty,
		"GIT_CONFIG_SYSTEM="+empty,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid",
	)
}

// wtGit runs git in dir with the isolated environment and fails on error.
func wtGit(t *testing.T, env []string, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// wtCanonTempDir returns a symlink-free temp dir (macOS /var → /private/var).
func wtCanonTempDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// wtWriteFile writes body to path, creating parent directories.
func wtWriteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// untrackedFixture is fixture F of the acceptance criteria.
type untrackedFixture struct {
	env []string
	P   string // primary checkout (canonical)
	W   string // linked worktree (canonical), no .moai
}

// newUntrackedFixture builds fixture F: a repository P with .moai/ gitignored
// but present (workflow.yaml plus one SPEC), one commit, and a linked worktree W
// made with `git worktree add`. workflowYAML is written verbatim to
// P/.moai/config/sections/workflow.yaml.
func newUntrackedFixture(t *testing.T, workflowYAML string) untrackedFixture {
	t.Helper()
	env := wtGitEnv(t)
	base := wtCanonTempDir(t)
	p := filepath.Join(base, "P")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	wtGit(t, env, p, "init", "-q", "-b", "main")
	wtWriteFile(t, filepath.Join(p, ".gitignore"), ".moai/\n")
	wtWriteFile(t, filepath.Join(p, "tracked.txt"), "base\n")
	wtWriteFile(t, filepath.Join(p, ".moai", "config", "sections", "workflow.yaml"), workflowYAML)
	wtWriteFile(t, filepath.Join(p, ".moai", "specs", "SPEC-WTFIX-001", "spec.md"),
		"---\nid: SPEC-WTFIX-001\ntitle: \"fixture\"\nversion: \"0.1.0\"\nstatus: draft\n"+
			"created: 2026-01-01\nupdated: 2026-01-01\nauthor: t\npriority: P3\n---\n\n# SPEC-WTFIX-001\n")
	wtGit(t, env, p, "add", ".gitignore", "tracked.txt")
	wtGit(t, env, p, "commit", "-q", "-m", "init")
	w := filepath.Join(base, "W")
	wtGit(t, env, p, "worktree", "add", "-q", "-b", "wt", w)
	return untrackedFixture{env: env, P: p, W: w}
}

const wtWorkflowNoGate = "workflow:\n  audit:\n    enabled: true\n"

// AC-MWU-001 (reproduction-first; REQ-MWU-002): a linked worktree of a
// repository that keeps .moai untracked is accepted, and the validator returns
// the worktree's canonical path.
func TestValidateProjectRoot_AcceptsLinkedWorktreeOfUntrackedMoai(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	if _, err := os.Stat(filepath.Join(fx.W, ".moai")); !os.IsNotExist(err) {
		t.Fatalf("fixture premise: W must have no .moai, stat err=%v", err)
	}
	got, err := validateProjectRoot(fx.W)
	if err != nil {
		t.Fatalf("validateProjectRoot(W) rejected a linked worktree: %v", err)
	}
	if got != fx.W {
		t.Fatalf("validateProjectRoot(W) = %q, want canonical %q", got, fx.W)
	}
}
