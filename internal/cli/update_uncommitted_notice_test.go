package cli

// t1729 — tests for the end-of-update uncommitted-template notice.
//
// The notice fires when the working tree carries uncommitted changes to the
// template files the sync just analyzed: worktrees branched from the
// committed base cannot see them until the operator commits (then syncs),
// which is exactly the stale-deployed-rule shape the SessionStart role-rule
// injection logs in factory sessions.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTrackedFile turns dir into a git repository with one tracked, committed
// file at rel (content v1). Returns the file's absolute path.
func initTrackedFile(t *testing.T, dir, rel, content string) string {
	t.Helper()
	return initTrackedFileUnder(t, dir, dir, rel, content)
}

// initTrackedFileUnder inits the git repository at repoRoot with one
// tracked, committed file at projectDir/rel — the project-in-subdirectory
// shape the path-basis test needs.
func initTrackedFileUnder(t *testing.T, repoRoot, projectDir, rel, content string) string {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q")
	path := filepath.Join(projectDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for seed file: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	run("add", ".")
	run("-c", "user.email=fixture@example.com", "-c", "user.name=fixture", "commit", "-q", "-m", "seed")
	return path
}

func TestUncommittedTemplateNotice(t *testing.T) {
	notice := func(t *testing.T, root string, paths ...string) string {
		t.Helper()
		return uncommittedTemplateNotice(root, paths)
	}

	t.Run("no_change_set_no_notice", func(t *testing.T) {
		dir := t.TempDir()
		initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		if got := notice(t, dir); got != "" {
			t.Errorf("empty change set must produce no notice, got: %s", got)
		}
	})

	t.Run("clean_tree_no_notice", func(t *testing.T) {
		dir := t.TempDir()
		initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		if got := notice(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md"); got != "" {
			t.Errorf("committed (clean) tree must produce no notice, got: %s", got)
		}
	})

	t.Run("uncommitted_template_change_fires", func(t *testing.T) {
		dir := t.TempDir()
		path := initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		// The update's uncommitted write: the working tree diverges from HEAD.
		if err := os.WriteFile(path, []byte("v2 refreshed by the update\n"), 0o644); err != nil {
			t.Fatalf("rewrite: %v", err)
		}
		got := notice(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md")
		if got == "" {
			t.Fatal("uncommitted template change must fire the notice")
		}
		if !strings.Contains(got, "git merge") {
			t.Errorf("notice does not name the working refresh (merge the base branch into the worktree): %s", got)
		}
		if strings.Contains(got, "commit on the base branch") {
			t.Errorf("notice still prescribes a base-branch commit that commit-protected repos refuse (audit round-4 P2): %s", got)
		}
		if !strings.Contains(got, "1 template file") {
			t.Errorf("notice does not carry the changed-file count: %s", got)
		}
	})

	t.Run("staged_only_change_still_fires", func(t *testing.T) {
		dir := t.TempDir()
		path := initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		if err := os.WriteFile(path, []byte("v2\n"), 0o644); err != nil {
			t.Fatalf("rewrite: %v", err)
		}
		cmd := exec.Command("git", "-C", dir, "add", ".")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git add: %v: %s", err, out)
		}
		if got := notice(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md"); got == "" {
			t.Fatal("a staged-but-uncommitted change is still uncommitted — the notice must fire")
		}
	})

	t.Run("unrelated_change_no_notice", func(t *testing.T) {
		dir := t.TempDir()
		initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		other := filepath.Join(dir, "user-notes.md")
		if err := os.WriteFile(other, []byte("the operator's own edit\n"), 0o644); err != nil {
			t.Fatalf("write unrelated: %v", err)
		}
		if got := notice(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md"); got != "" {
			t.Errorf("a user's unrelated edit must not trigger the notice, got: %s", got)
		}
	})

	t.Run("untracked_new_directory_fires", func(t *testing.T) {
		// Default porcelain collapses a wholly-untracked NEW directory to
		// "?? dir/" — the exact-path intersection would miss every file in
		// it and silently skip the notice (audit round-1 P2). The query
		// enumerates untracked files individually.
		dir := t.TempDir()
		initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		newPath := filepath.Join(dir, ".claude", "rules", "moai", "brand-new", "new-rule.md")
		if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
			t.Fatalf("mkdir new template dir: %v", err)
		}
		if err := os.WriteFile(newPath, []byte("new template file\n"), 0o644); err != nil {
			t.Fatalf("write new template file: %v", err)
		}
		got := notice(t, dir, ".claude/rules/moai/brand-new/new-rule.md")
		if got == "" {
			t.Fatal("an untracked NEW directory's template file must fire the notice")
		}
	})

	t.Run("project_subdir_of_repo_fires", func(t *testing.T) {
		// Porcelain paths are REPO-root-relative while templatePaths are
		// project-relative; with the project inside a subdirectory the raw
		// exact-match intersection misses every changed file (audit
		// round-2 P2). The bases must be normalized.
		repo := t.TempDir()
		proj := filepath.Join(repo, "apps", "service")
		path := initTrackedFileUnder(t, repo, proj, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		if err := os.WriteFile(path, []byte("v2 refreshed by the update\n"), 0o644); err != nil {
			t.Fatalf("rewrite: %v", err)
		}
		if got := uncommittedTemplateNotice(proj, []string{".claude/rules/moai/workflow/factory-dispatch.md"}); got == "" {
			t.Fatal("a changed template file under a project subdirectory must fire the notice")
		}
	})

	t.Run("inherited_git_env_cannot_redirect_the_query", func(t *testing.T) {
		// Discriminator shape (audit round-2 P2): the project is CLEAN while
		// the foreign repository (whose GIT_DIR the child git would inherit)
		// carries the same template path as an uncommitted change. Without
		// the env scrub the query reads the foreign repo and fires a FALSE
		// notice; with it, the clean project stays silent.
		dir := t.TempDir()
		initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1 committed, clean\n")
		other := t.TempDir()
		otherPath := initTrackedFile(t, other, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		if err := os.WriteFile(otherPath, []byte("v2 uncommitted in the FOREIGN repo\n"), 0o644); err != nil {
			t.Fatalf("rewrite foreign: %v", err)
		}
		t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
		if got := uncommittedTemplateNotice(dir, []string{".claude/rules/moai/workflow/factory-dispatch.md"}); got != "" {
			t.Fatalf("an inherited GIT_DIR leaked the foreign repo's state into the notice: %s", got)
		}
	})

	t.Run("configured_base_branch_named_in_notice", func(t *testing.T) {
		// A project running a non-main base gets advice naming ITS branch,
		// not a hardcoded main (audit round-3 P2).
		dir := t.TempDir()
		path := initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
			t.Fatalf("mkdir config: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".moai", "config", "sections", "git-strategy.yaml"), []byte("git_strategy:\n  worktree_base_branch: develop\n"), 0o644); err != nil {
			t.Fatalf("write git-strategy fixture: %v", err)
		}
		if err := os.WriteFile(path, []byte("v2\n"), 0o644); err != nil {
			t.Fatalf("rewrite: %v", err)
		}
		got := uncommittedTemplateNotice(dir, []string{".claude/rules/moai/workflow/factory-dispatch.md"})
		if got == "" {
			t.Fatal("uncommitted change must fire the notice")
		}
		if !strings.Contains(got, "git merge develop") {
			t.Errorf("notice does not name the configured base branch: %s", got)
		}
	})

	t.Run("deleted_template_file_fires", func(t *testing.T) {
		// A cleanup-only update DELETES stale managed files; the deletion is
		// uncommitted work the same as a modification (audit round-4 P2) —
		// porcelain " D path" must fire the notice.
		dir := t.TempDir()
		path := initTrackedFile(t, dir, ".claude/rules/moai/workflow/retired-rule.md", "v1\n")
		if err := os.Remove(path); err != nil {
			t.Fatalf("remove: %v", err)
		}
		if got := notice(t, dir, ".claude/rules/moai/workflow/retired-rule.md"); got == "" {
			t.Fatal("an uncommitted template DELETION must fire the notice")
		}
	})

	t.Run("not_a_repo_fails_open", func(t *testing.T) {
		dir := t.TempDir()
		if got := notice(t, dir, "any/path.md"); got != "" {
			t.Errorf("non-repository root must fail open (no notice), got: %s", got)
		}
	})
}
