package cli

import (
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

// victimRemotes reads remote.* entries straight from a config file, bypassing
// any GIT_DIR the process carries.
func victimRemotes(t *testing.T, configPath string) string {
	t.Helper()
	out, _ := exec.Command("git", "config", "--file", configPath, "--get-regexp", `^remote\.`).Output()
	return strings.TrimSpace(string(out))
}

// TestDetectGitConfig_IgnoresInheritedGitDir reproduces card t1204: with
// GIT_DIR / GIT_WORK_TREE inherited (as under a git hook), `git -C <tempdir>`
// still obeys GIT_DIR. The fixture helpers then WRITE into the outer repository
// and the production detection READS the outer repository instead of dir.
func TestDetectGitConfig_IgnoresInheritedGitDir(t *testing.T) {
	victim := t.TempDir()
	scrubbedGit(t, victim, "init")
	victimConfig := filepath.Join(victim, ".git", "config")

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

	if got := victimRemotes(t, victimConfig); got != "" {
		t.Errorf("write shape: victim repo gained remotes:\n%s", got)
	}
	if mode, provider := detectGitConfig(dir); mode != "personal" || provider != "github" {
		t.Errorf("write shape: detectGitConfig(dir) = (%q, %q), want (personal, github)", mode, provider)
	}
}
