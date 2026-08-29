// prlink_landedref_test.go — SPEC-TODO-LANDING-STATE-001 AC-TLS-001, AC-TLS-002
// (M1: the ref the landed question is asked about is RESOLVED, not constant).
//
// The defect these criteria close is silent in exactly the way the engine-flag
// defect was: a project that integrates on a branch other than the default one
// asks the landed question about a ref its work never reaches, and every card
// that shipped reads as not-landed. No error, no empty-output warning — just a
// wrong answer that looks like a right one.
package kanban

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// configuredProject writes a git-strategy.yaml carrying base as the card
// worktree base branch, and returns the project root.
func configuredProject(t *testing.T, base string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "git_strategy:\n"
	if base != "" {
		body += "    worktree_base_branch: \"" + base + "\"\n"
	} else {
		body += "    mode: personal\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
	return root
}

// AC-TLS-001 / AC-TLS-002 — the resolver reads the configured integration
// branch, and an unconfigured project keeps the historical default.
func TestLandedRefFor(t *testing.T) {
	cases := []struct {
		name string
		base string
		want string
	}{
		{"configured develop", "develop", "origin/develop"},
		{"configured with surrounding space", "  develop  ", "origin/develop"},
		{"configured main is still main", "main", "origin/main"},
		{"empty configuration keeps the default", "", DefaultLandedRef},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := LandedRefFor(configuredProject(t, tc.base)); got != tc.want {
				t.Errorf("LandedRefFor = %q, want %q", got, tc.want)
			}
		})
	}
	// A root that is not a project at all resolves to the default rather than
	// to an error: an unreadable project behaves as an unconfigured one.
	if got := LandedRefFor(t.TempDir()); got != DefaultLandedRef {
		t.Errorf("LandedRefFor(no project) = %q, want %q", got, DefaultLandedRef)
	}
	if DefaultLandedRef != "origin/main" {
		t.Errorf("DefaultLandedRef = %q, want origin/main — AC-TLS-002 pins the unconfigured behaviour", DefaultLandedRef)
	}
}

// AC-TLS-001 — the resolved ref reaches the argv the check actually runs. The
// ref is an INPUT to the builder, so a caller cannot forget to thread it.
func TestLandedGrepArgs_CarriesTheResolvedRef(t *testing.T) {
	args, err := LandedGrepArgs("origin/develop", "t293")
	if err != nil {
		t.Fatalf("argv: %v", err)
	}
	if !slices.Contains(args, "origin/develop") {
		t.Errorf("argv %v does not name the resolved ref", args)
	}
	if slices.Contains(args, "origin/main") {
		t.Errorf("argv %v still names the hardcoded default", args)
	}
	// The engine flag survives the signature change — the silent-empty guard
	// this SPEC inherits is not weakened by threading a ref through.
	if !slices.Contains(args, LandedRegexpEngineFlag) {
		t.Errorf("argv %v lost %s", args, LandedRegexpEngineFlag)
	}
	// An empty ref falls back rather than emitting an empty argv element,
	// which git would read as the working tree.
	back, backErr := LandedGrepArgs("", "t293")
	if backErr != nil {
		t.Fatalf("argv (empty ref): %v", backErr)
	}
	if !slices.Contains(back, DefaultLandedRef) {
		t.Errorf("argv %v with an empty ref does not fall back to %s", back, DefaultLandedRef)
	}
}

// AC-TLS-001 — a t293-shaped fixture (named by the integration branch and NOT
// by origin/main) answers landed for a project configured on develop, and
// not-landed for one that is not. This is the measured misjudgement input from
// acceptance.md §A.4 rendered as a fixture.
func TestLandedQuerier_AnswersAboutTheConfiguredRef(t *testing.T) {
	dir := developLandedRepo(t)
	run := gitIn(dir)

	onDevelop, devErr := (GitLandedQuerier{Run: run, Ref: "origin/develop"}).Landed("t293")
	if devErr != nil {
		t.Fatalf("landed on origin/develop: %v", devErr)
	}
	if onDevelop != LandingLanded {
		t.Errorf("t293 on origin/develop = %q, want %q — the card shipped on the integration branch", onDevelop, LandingLanded)
	}

	// The control that makes the above non-vacuous: the SAME card, asked
	// about origin/main, is genuinely not there. Without this half the test
	// would pass against a querier that answers "landed" unconditionally.
	onMain, mainErr := (GitLandedQuerier{Run: run, Ref: "origin/main"}).Landed("t293")
	if mainErr != nil {
		t.Fatalf("landed on origin/main: %v", mainErr)
	}
	if onMain != LandingNotLanded {
		t.Errorf("t293 on origin/main = %q, want %q — this is the false negative the SPEC closes", onMain, LandingNotLanded)
	}

	// And the default-ref querier (Ref unset) behaves exactly as origin/main
	// did before this change — AC-TLS-002's byte-identity, at the seam.
	onDefault, defErr := (GitLandedQuerier{Run: run}).Landed("t293")
	if defErr != nil {
		t.Fatalf("landed on the default ref: %v", defErr)
	}
	if onDefault != LandingNotLanded {
		t.Errorf("t293 on the default ref = %q, want %q", onDefault, LandingNotLanded)
	}
}

// developLandedRepo builds a repository whose origin/develop names t293 and
// whose origin/main does not — the shape measured on this project's own
// history (origin/main: 0 commits naming t293; origin/develop: 9).
func developLandedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitCfg := filepath.Join(dir, "gitconfig")
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(cmd.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
			"GIT_CONFIG_GLOBAL="+gitCfg, "GIT_CONFIG_NOSYSTEM=1",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q", "-b", "main")
	commit := func(file, msg string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, file), []byte(msg+"\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
		run("add", file)
		run("commit", "-q", "-m", msg)
	}
	commit("a.txt", "chore: seed the tree")
	// origin/main stops here — it does NOT name t293.
	run("update-ref", "refs/remotes/origin/main", "HEAD")
	commit("b.txt", "fix(kanban): card landing state (t293)")
	run("update-ref", "refs/remotes/origin/develop", "HEAD")
	return dir
}
