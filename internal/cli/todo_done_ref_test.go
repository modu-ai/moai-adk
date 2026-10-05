// todo_done_ref_test.go — SPEC-TODO-LANDING-ATTRIBUTION-001 §C criteria
// AC-TLA-010, AC-TLA-011 (M2: the verdict names its ref, and the sub-config
// chain level is disclosed).
//
// The chain-level fixtures are REAL machine state: the level-2 test writes an
// actual refs/remotes/origin/HEAD symref into the fixture repository's git
// metadata, and the level-3 test builds a repository that genuinely lacks one
// — neither is stubbed in a code branch (acceptance.md AC-TLA-009).
package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// installOriginHEADSymref writes a real refs/remotes/origin/HEAD symref
// naming origin/<branch> into the fixture repository at root.
func installOriginHEADSymref(t *testing.T, root, branch string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(cmd.Environ(), "GIT_CONFIG_NOSYSTEM=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("update-ref", "refs/remotes/origin/"+branch, "HEAD")
	run("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/"+branch)
}

// configureWorktreeBase writes git_strategy.worktree_base_branch into the
// fixture's config, selecting chain level 1.
func configureWorktreeBase(t *testing.T, root, base string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "git_strategy:\n    worktree_base_branch: \"" + base + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
}

// AC-TLA-010 + AC-TLA-011, level 2 — the verdict line still begins with the
// `done <id> ` prefix every reader keys off AND names the ref that answered;
// a resolution that reached level 2 is disclosed on stderr with the level.
func TestTodoDone_VerdictNamesRefAndDisclosesLevel2(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha work")
	installOriginHEADSymref(t, root, "develop")
	stubLandingQuery(t, "abc1234 fix: something (t1)\n", nil)

	stdout, stderr, err := runTodo(t, "done", "t1", "--require-landed")
	if err != nil {
		t.Fatalf("todo done: %v (stderr %q)", err, stderr)
	}
	if !strings.HasPrefix(stdout, "done t1 ") {
		t.Errorf("stdout = %q, want the `done <id> ` prefix preserved (AC-TLA-010)", stdout)
	}
	if !strings.Contains(stdout, "ref=origin/develop") {
		t.Errorf("stdout = %q, want the answering ref named (AC-TLA-010)", stdout)
	}
	if !strings.Contains(stderr, "level 2") {
		t.Errorf("stderr = %q, want the answering chain level disclosed (AC-TLA-011)", stderr)
	}
}

// AC-TLA-011, level 3 — the compiled-in default is reached only through the
// absent symref, and that exceptional path is disclosed too.
func TestTodoDone_DisclosesLevel3(t *testing.T) {
	todoFixture(t) // git repo, no symref, no config
	seedTodo(t, "alpha work")
	stubLandingQuery(t, "abc1234 fix: something (t1)\n", nil)

	stdout, stderr, err := runTodo(t, "done", "t1", "--require-landed")
	if err != nil {
		t.Fatalf("todo done: %v (stderr %q)", err, stderr)
	}
	if !strings.Contains(stdout, "ref=origin/main") {
		t.Errorf("stdout = %q, want the default ref named as the answerer", stdout)
	}
	if !strings.Contains(stderr, "level 3") {
		t.Errorf("stderr = %q, want the level-3 fall-through disclosed", stderr)
	}
}

// AC-TLA-011, level 1 — a project that configures the key gets NO disclosure:
// the notice marks the exceptional path, not every path. C-4 keeps the
// configured behaviour exactly.
func TestTodoDone_NoDisclosureAtLevel1(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha work")
	configureWorktreeBase(t, root, "develop")
	stubLandingQuery(t, "abc1234 fix: something (t1)\n", nil)

	stdout, stderr, err := runTodo(t, "done", "t1", "--require-landed")
	if err != nil {
		t.Fatalf("todo done: %v (stderr %q)", err, stderr)
	}
	if !strings.Contains(stdout, "ref=origin/develop") {
		t.Errorf("stdout = %q, want the configured ref named", stdout)
	}
	if strings.Contains(stderr, "level") {
		t.Errorf("stderr = %q, want no chain-level disclosure at level 1", stderr)
	}
}

// Without --require-landed no query ran, so no ref answered: the verdict line
// must not name one. "The guard did not run" stays distinct from "the guard
// answered against ref X".
func TestTodoDone_WithoutFlagNoRefNamed(t *testing.T) {
	todoFixture(t)
	seedTodo(t, "alpha work")

	stdout, _, err := runTodo(t, "done", "t1")
	if err != nil {
		t.Fatalf("todo done: %v", err)
	}
	if !strings.HasPrefix(stdout, "done t1 ") || !strings.Contains(stdout, "landing=unknown") {
		t.Errorf("stdout = %q, want the unprefixed-unknown verdict preserved", stdout)
	}
	if strings.Contains(stdout, "ref=") {
		t.Errorf("stdout = %q, want no ref named when no landing query ran", stdout)
	}
}
