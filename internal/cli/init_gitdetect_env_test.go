package cli

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// scrubbedGit runs git in dir with the repository-scoping variables removed,
// so fixture setup cannot itself be redirected by the environment under test.
func scrubbedGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = gitenv.Env()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("git %v unavailable: %v: %s", args, err, out)
	}
}

// victimConfig queries a config file directly, bypassing any GIT_DIR the
// process carries. Exit 1 means "no match" and yields ""; any other failure is
// fatal, so an unreadable config can never pass as an untouched one.
func victimConfig(t *testing.T, configPath string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"config", "--file", configPath}, args...)...)
	cmd.Env = gitenv.Env() // keep an inherited GIT_DIR out of repository setup
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return ""
		}
		t.Fatalf("git config --file %s %v: %v", configPath, args, err)
	}
	return strings.TrimSpace(string(out))
}

// victimRemotes reads remote.* entries straight from a config file.
func victimRemotes(t *testing.T, configPath string) string {
	t.Helper()
	return victimConfig(t, configPath, "--get-regexp", `^remote\.`)
}

// TestDetectGitConfig_IgnoresInheritedGitDir reproduces card t1204: with
// GIT_DIR / GIT_WORK_TREE inherited (as under a git hook), `git -C <tempdir>`
// still obeys GIT_DIR. The fixture helpers then WRITE into the outer repository
// and the production detection READS the outer repository instead of dir.
func TestDetectGitConfig_IgnoresInheritedGitDir(t *testing.T) {
	victim := t.TempDir()
	scrubbedGit(t, victim, "init")
	victimConfigPath := filepath.Join(victim, ".git", "config")

	// Read shape: a repository built before the variables are set.
	prebuilt := t.TempDir()
	scrubbedGit(t, prebuilt, "init")
	scrubbedGit(t, prebuilt, "remote", "add", "origin", "https://gitlab.com/group/proj.git")

	t.Setenv("GIT_DIR", filepath.Join(victim, ".git"))
	t.Setenv("GIT_WORK_TREE", victim)

	if mode, provider := detectGitConfig(prebuilt); mode != "personal" || provider != "gitlab" {
		t.Errorf("read shape: detectGitConfig(prebuilt) = (%q, %q), want (personal, gitlab) — answered about the GIT_DIR repo", mode, provider)
	}

	// Write shape: the fixture helpers run with the variables set.
	dir := gitDetectInitRepo(t)
	gitAddRemote(t, dir, "upstream", "https://gitlab.com/group/proj.git")

	if got := victimRemotes(t, victimConfigPath); got != "" {
		t.Errorf("write shape: victim repo gained remotes:\n%s", got)
	}
	if mode, provider := detectGitConfig(dir); mode != "personal" || provider != "github" {
		t.Errorf("write shape: detectGitConfig(dir) = (%q, %q), want (personal, github)", mode, provider)
	}
}

// TestGitInitAt_IgnoresInheritedWorktreeGitDir covers the third shape from
// card t1204 (the TestInitGitDetectionFillsConfig fixture path): when the
// inherited GIT_DIR names a linked worktree's per-worktree gitdir, an unscrubbed
// `git -C <dir> init` re-initializes that gitdir and writes core.bare=true into
// the SHARED config, leaving the main repository unusable.
func TestGitInitAt_IgnoresInheritedWorktreeGitDir(t *testing.T) {
	mainRepo := t.TempDir()
	scrubbedGit(t, mainRepo, "init")
	scrubbedGit(t, mainRepo, "-c", "user.name=t1204", "-c", "user.email=t1204@example.invalid",
		"commit", "--allow-empty", "--no-verify", "-m", "init")
	linked := filepath.Join(t.TempDir(), "wt")
	scrubbedGit(t, mainRepo, "worktree", "add", "--detach", linked)

	sharedConfig := filepath.Join(mainRepo, ".git", "config")
	if got := victimConfig(t, sharedConfig, "--get", "core.bare"); got != "false" {
		t.Fatalf("premise: core.bare before = %q, want \"false\"", got)
	}

	t.Setenv("GIT_DIR", filepath.Join(mainRepo, ".git", "worktrees", "wt"))

	gitInitAt(t, t.TempDir())

	if got := victimConfig(t, sharedConfig, "--get", "core.bare"); got == "true" {
		t.Errorf("shared config gained core.bare=true — gitInitAt re-initialized the inherited worktree gitdir")
	}
	if got := victimRemotes(t, sharedConfig); got != "" {
		t.Errorf("shared config gained remotes:\n%s", got)
	}
}
