// prlink_landedref_chain_test.go — SPEC-TODO-LANDING-ATTRIBUTION-001 §C
// criteria (M2: the ref chain, REQ-TLA-007..009).
//
// AC-TLA-009's fixture requirement is machine state, not a code branch: the
// level-3 branch is reached by building a real repository whose git metadata
// genuinely lacks refs/remotes/origin/HEAD, so the criterion is red until the
// chain actually falls through.
package kanban

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// chainRepo builds a real git repository with no written configuration (an
// unreadable config reads as empty), plus — when withHead is set — a
// refs/remotes/origin/HEAD symref naming target.
func chainRepo(t *testing.T, withHead bool, target string) string {
	t.Helper()
	dir := t.TempDir()
	gitCfg := dir + "/gitconfig"
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
	// The symref must point at a ref that exists: seed one commit.
	run("-c", "user.name=t", "-c", "user.email=t@example.com",
		"commit", "-q", "--allow-empty", "-m", "chore: seed the tree")
	if withHead {
		run("update-ref", "refs/remotes/origin/"+target, "HEAD")
		run("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/"+target)
	}
	return dir
}

// AC-TLA-008, second direction — with the configured key empty, the resolver
// consults refs/remotes/origin/HEAD and answers what it names. The current
// two-level resolver (MUT-CONFIG-ONLY) answers DefaultLandedRef here.
func TestLandedRefFor_ChainLevel2_OriginHEAD(t *testing.T) {
	dir := chainRepo(t, true, "develop")
	if got := LandedRefFor(dir); got != "origin/develop" {
		t.Errorf("LandedRefFor = %q, want %q — level 2 reads the repository's own recorded default (MUT-CONFIG-ONLY)", got, "origin/develop")
	}
}

// AC-TLA-009, clause 1 — a repository whose symref is genuinely ABSENT from
// its git metadata falls through to DefaultLandedRef without failing, and the
// answering level is reported as level 3.
func TestLandedRefFor_ChainLevel3_SymrefAbsent(t *testing.T) {
	dir := chainRepo(t, false, "")
	if got := LandedRefFor(dir); got != DefaultLandedRef {
		t.Errorf("LandedRefFor = %q, want %q — a missing symref falls through, it does not fail", got, DefaultLandedRef)
	}
}

// refRunRecorder swaps the chain's git seam for a recorder that reports
// unresolvable (forcing the level-3 fall-through) and restores it at cleanup.
// It returns the pointer the assertions read.
func refRunRecorder(t *testing.T) *[][]string {
	t.Helper()
	original := landedRefGitRun
	t.Cleanup(func() { landedRefGitRun = original })
	calls := &[][]string{}
	landedRefGitRun = func(name string, args ...string) (string, error) {
		*calls = append(*calls, append([]string{name}, args...))
		return "", errRefUnresolvable
	}
	return calls
}

// errRefUnresolvable stands in for a symref the recorder cannot answer.
var errRefUnresolvable = errors.New("fatal: ref refs/remotes/origin/HEAD is not a symbolic ref")

// AC-TLA-008 + AC-TLA-009 — LandedRefForWithLevel reports WHICH level
// answered, and the level is provenance, not a value inequality: a repository
// whose symref legitimately names main answers origin/main AT LEVEL 2, and
// the level is the only thing distinguishing that from the level-3 default.
func TestLandedRefForWithLevel_ReportsTheAnsweringLevel(t *testing.T) {
	dir := chainRepo(t, true, "release/v9")
	ref, level := LandedRefForWithLevel(dir)
	if ref != "origin/release/v9" || level != LandedRefOriginHEAD {
		t.Errorf("ref=%q level=%d, want origin/release/v9 at level %d", ref, level, LandedRefOriginHEAD)
	}

	dir = chainRepo(t, false, "")
	ref, level = LandedRefForWithLevel(dir)
	if ref != DefaultLandedRef || level != LandedRefDefault {
		t.Errorf("ref=%q level=%d, want %s at level %d", ref, level, DefaultLandedRef, LandedRefDefault)
	}

	root := configuredProject(t, "develop")
	ref, level = LandedRefForWithLevel(root)
	if ref != "origin/develop" || level != LandedRefConfigured {
		t.Errorf("ref=%q level=%d, want origin/develop at level %d", ref, level, LandedRefConfigured)
	}
}

// AC-TLA-008, first direction — the configured key answers first and the
// symref is NOT consulted. The census below makes the not-consulted half
// observable rather than asserted.
func TestLandedRefForWithLevel_Level1SkipsLevel2(t *testing.T) {
	root := configuredProject(t, "develop")
	calls := refRunRecorder(t)
	ref, level := LandedRefForWithLevel(root)
	if ref != "origin/develop" || level != LandedRefConfigured {
		t.Errorf("ref=%q level=%d, want the configured answer at level 1", ref, level)
	}
	if n := len(*calls); n != 0 {
		t.Errorf("level 2 consulted %d time(s) at level 1 — C-4 violated: %v", n, *calls)
	}
}

// censusViolation is AC-TLA-012's write-detector over the recorded
// invocations: it names the first ref-writing git command it sees, or "".
func censusViolation(calls [][]string) string {
	for _, c := range calls {
		if len(c) < 2 || c[0] != "git" {
			continue
		}
		joined := strings.Join(c[1:], " ")
		at := indexOf(c, "symbolic-ref")
		switch {
		case strings.Contains(joined, "set-head"), strings.Contains(joined, "update-ref"):
			return joined
		case at >= 0 && len(c)-at > 2:
			// the WRITE form: a second operand after the ref being read
			return joined
		}
	}
	return ""
}

func indexOf(hay []string, needle string) int {
	for i, s := range hay {
		if s == needle {
			return i
		}
	}
	return -1
}

// AC-TLA-012 — the resolution path runs exactly one git invocation and it is
// the READ form of symbolic-ref: no ref is ever written (REQ-TLA-012).
func TestLandedRefForWithLevel_ResolutionWritesNothing(t *testing.T) {
	dir := chainRepo(t, false, "")
	calls := refRunRecorder(t)
	ref, level := LandedRefForWithLevel(dir)
	if ref != DefaultLandedRef || level != LandedRefDefault {
		t.Fatalf("ref=%q level=%d, want the default at level 3 under the recorder", ref, level)
	}
	if n := len(*calls); n != 1 {
		t.Fatalf("resolution made %d git invocations, want exactly 1: %v", n, *calls)
	}
	call := (*calls)[0]
	if v := censusViolation(*calls); v != "" {
		t.Errorf("the resolution path issued a ref-writing command: %s", v)
	}
	if indexOf(call, "symbolic-ref") < 0 {
		t.Errorf("resolution called %v, want the symbolic-ref read", call)
	}
	// The read form carries exactly one operand after the subcommand.
	at := indexOf(call, "symbolic-ref")
	if operands := len(call) - at - 1; operands != 1 {
		t.Errorf("symbolic-ref invoked with %d operand(s) %v, want exactly 1 (the read form)", operands, call[at+1:])
	}

	// The census observes rather than passes vacuously: a planted write
	// command in the recorded path MUST be flagged.
	planted := append([][]string{}, *calls...)
	planted = append(planted, []string{"git", "-C", dir, "symbolic-ref",
		"refs/remotes/origin/HEAD", "refs/heads/evil"})
	if v := censusViolation(planted); v == "" {
		t.Error("the census failed to flag a planted symbolic-ref WRITE — it is vacuous")
	}
	planted = append(planted, []string{"git", "remote", "set-head", "origin", "develop"})
	if v := censusViolation(planted); v == "" {
		t.Error("the census failed to flag a planted remote set-head — it is vacuous")
	}
}
