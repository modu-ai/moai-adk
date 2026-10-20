package cli

// m2_launcher_short_name_test.go — SPEC-WORKTREE-HOME-ROOT-001 M2 RED (card
// t1640): AC-WHR-006, the launcher short name under both PM-3 branches.
//
// The PM-3 measurement recorded FAIL (operator ruling bd72: claude 2.1.303
// rejects the absolute home-root path at argument validation, "Invalid
// worktree name", exit 1), so D-5 selected branch Y: a short name with no
// existing tree passes through unchanged and the launcher creates nothing;
// a short name that names an existing tree is entered — a home-root tree by
// handing the backend its absolute path, a legacy-root tree natively by the
// pass-through claude already resolves. The branch is selected from the
// recorded PM-3 line, and a missing or GAP line fails the test (the AC
// requires the selector, never a default).
//
// The backend is stubbed (launchClaudeFunc) and records its argv; the
// assertions are on what the launcher hands it and on what the launcher
// created — the backend's own acceptance of absolute paths is the entry
// measurement's domain (PM-3, REQ-WHR-025), not this criterion's.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// pm3BranchOf returns the recorded PM-3 classification (PASS or FAIL) from
// the tracked measurement record, failing the test on an absent, GAP or
// unrecognized line (AC-WHR-006: the test selects its branch from the
// recorded PM-3 line and never defaults).
func pm3BranchOf(t *testing.T) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", "pm3-record-t1640.md"))
	if err != nil {
		t.Fatalf("PM-3 record unreadable: %v", err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		rest, ok := strings.CutPrefix(line, "RESULT PM-3: ")
		if !ok {
			continue
		}
		branch, _, found := strings.Cut(rest, " ;;")
		if !found {
			t.Fatalf("PM-3 record line is malformed: %q", line)
		}
		switch branch {
		case "PASS", "FAIL":
			return branch
		default:
			t.Fatalf("PM-3 classification %q is neither PASS nor FAIL", branch)
		}
	}
	t.Fatalf("no RESULT PM-3 line in the tracked measurement record")
	return ""
}

// m2LaunchFixture overrides the project root and the backend launch seam and
// returns the recorded argv of the stubbed backend.
func m2LaunchFixture(t *testing.T, repo string) *[]string {
	t.Helper()
	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return repo, nil }
	t.Cleanup(func() { findProjectRootFn = origFn })

	var gotArgs []string
	origLaunch := launchClaudeFunc
	launchClaudeFunc = func(profile string, args []string) error {
		gotArgs = args
		return nil
	}
	t.Cleanup(func() { launchClaudeFunc = origLaunch })
	return &gotArgs
}

// m2AssertNoTreeCreated asserts the launcher created no tree of name under
// either the legacy root or the home root.
func m2AssertNoTreeCreated(t *testing.T, repo, home, name string) {
	t.Helper()
	for _, p := range []string{
		filepath.Join(repo, ".claude", "worktrees", name),
		filepath.Join(home, "worktrees", "projects", filepath.Base(repo), name),
	} {
		if _, err := os.Lstat(p); !os.IsNotExist(err) {
			t.Errorf("the launcher created %s (lstat err=%v); branch Y creates nothing", p, err)
		}
	}
}

// TestLauncherShortName_NoExistingTree_PassesThrough is the branch-Y arm of
// AC-WHR-006: with no tree of the name in any root, the backend receives the
// short name unchanged and the launcher creates nothing.
func TestLauncherShortName_NoExistingTree_PassesThrough(t *testing.T) {
	if pm3BranchOf(t) != "FAIL" {
		t.Skip("PM-3 recorded PASS; branch X is the selected arm and this branch-Y assertion does not apply")
	}
	repo := m2Repo(t)
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	gotArgs := m2LaunchFixture(t, repo)

	if err := ccCmd.RunE(ccCmd, []string{"-w", "newname"}); err != nil {
		t.Fatalf("moai cc -w newname: %v", err)
	}
	joined := strings.Join(*gotArgs, " ")
	if !strings.Contains(joined, "newname") {
		t.Errorf("backend argv %v does not carry the short name newname unchanged", *gotArgs)
	}
	if strings.Contains(joined, home) {
		t.Errorf("backend argv %q names an absolute home-root path; branch Y passes the short name through", joined)
	}
	m2AssertNoTreeCreated(t, repo, home, "newname")
}

// TestLauncherShortName_HomeRootTree_IsEntered pins the both-branches clause
// for the home root: a short name that names an existing home-root tree is
// ENTERED — the launcher starts the backend inside the tree (the process
// working directory) and hands it NO worktree selection. The argv-rewrite
// form this test used to pin is the one PM-3 measured failing: claude 2.1.303
// exits 1 on absolute --worktree values (card t1640 review P1-2; M2's own
// Gaps (b) recorded the real-backend entry strategy as the open domain).
// Nothing is created under the legacy root. The fixture registers a REAL
// worktree: the M3 writer precheck reads the resolved tree's lock state,
// which needs a registered tree, not a bare directory.
func TestLauncherShortName_HomeRootTree_IsEntered(t *testing.T) {
	repo := m2Repo(t)
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	gotArgs := m2LaunchFixture(t, repo)

	tree := filepath.Join(home, "worktrees", "projects", filepath.Base(repo), "t1")
	if err := os.MkdirAll(filepath.Dir(tree), 0o755); err != nil {
		t.Fatalf("mkdir home projects: %v", err)
	}
	m3Gitf(t, repo, "worktree", "add", "-q", "-b", "WT-t1", "--", tree)

	wdBefore, werr := os.Getwd()
	if werr != nil {
		t.Fatalf("getwd: %v", werr)
	}
	t.Cleanup(func() { _ = os.Chdir(wdBefore) })

	if err := ccCmd.RunE(ccCmd, []string{"-w", "t1"}); err != nil {
		t.Fatalf("moai cc -w t1: %v", err)
	}
	joined := strings.Join(*gotArgs, " ")
	if strings.Contains(joined, tree) {
		t.Errorf("backend argv %q names the absolute tree path; PM-3 measured claude rejecting it (exit 1 at argument validation)", joined)
	}
	if strings.Contains(joined, "-w") {
		t.Errorf("backend argv %q still carries a worktree selection; the entry-by-directory form hands the backend none", joined)
	}
	if wdAfter, _ := os.Getwd(); wdAfter != tree {
		t.Errorf("the launcher did not enter the tree: cwd %q, want %q", wdAfter, tree)
	}
	if _, err := os.Lstat(filepath.Join(repo, ".claude", "worktrees", "t1")); !os.IsNotExist(err) {
		t.Errorf("a duplicate was created under .claude/worktrees (lstat err=%v)", err)
	}
}

// TestLauncherShortName_LegacyTree_PassesThrough pins the both-branches
// clause for the legacy root: a short name that names an existing
// .claude/worktrees tree passes through unchanged — claude re-enters it
// natively — and the launcher creates nothing.
func TestLauncherShortName_LegacyTree_PassesThrough(t *testing.T) {
	repo := m2Repo(t)
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	gotArgs := m2LaunchFixture(t, repo)

	if err := os.MkdirAll(filepath.Join(repo, ".claude", "worktrees", "t7c"), 0o755); err != nil {
		t.Fatalf("mkdir legacy tree: %v", err)
	}

	if err := ccCmd.RunE(ccCmd, []string{"-w", "t7c"}); err != nil {
		t.Fatalf("moai cc -w t7c: %v", err)
	}
	joined := strings.Join(*gotArgs, " ")
	if !strings.Contains(joined, "t7c") {
		t.Errorf("backend argv %v does not carry the short name t7c", *gotArgs)
	}
	if strings.Contains(joined, home) {
		t.Errorf("backend argv %q names an absolute path; the legacy-root tree is entered by the pass-through", joined)
	}
	// The fixture tree is the thing being entered, so only a HOME-root
	// duplicate would count as the launcher creating something.
	if _, err := os.Lstat(filepath.Join(home, "worktrees", "projects", filepath.Base(repo), "t7c")); !os.IsNotExist(err) {
		t.Errorf("the launcher created a home-root duplicate of t7c (lstat err=%v)", err)
	}
}

// TestLauncherShortName_Glm_NoExistingTree_PassesThrough repeats the
// branch-Y arm on the glm launch path (AC-WHR-006 names both backends).
func TestLauncherShortName_Glm_NoExistingTree_PassesThrough(t *testing.T) {
	if pm3BranchOf(t) != "FAIL" {
		t.Skip("PM-3 recorded PASS; branch X is the selected arm and this branch-Y assertion does not apply")
	}
	repo := m2Repo(t)
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	t.Setenv(config.DefaultGLMEnvVar, "m2-test-only-key")
	gotArgs := m2LaunchFixture(t, repo)

	if err := glmCmd.RunE(glmCmd, []string{"-w", "newname"}); err != nil {
		t.Fatalf("moai glm -w newname: %v", err)
	}
	joined := strings.Join(*gotArgs, " ")
	if !strings.Contains(joined, "newname") {
		t.Errorf("backend argv %v does not carry the short name newname unchanged", *gotArgs)
	}
	if strings.Contains(joined, home) {
		t.Errorf("backend argv %q names an absolute home-root path; branch Y passes the short name through", joined)
	}
	m2AssertNoTreeCreated(t, repo, home, "newname")
}
