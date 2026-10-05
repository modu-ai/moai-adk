package hook

// env_literal_diff_test.go — AC-SRL-009's verdict instrument (REQ-SRL-009):
// new env-name references use the internal/config/envkeys.go constants, so
// the card's implementation diff must introduce zero MOAI_(FACTORY|KANBAN)
// literals in the hook, factorymsg, and cli packages.
//
// The extraction is diff-scoped, not whole-tree: the audit's iteration-1
// measurement (plan-audit-r1.md §3, tree d194083fb) showed the whole-tree
// sweep returns 28 literals on a tree with zero new ones — report-not-verdict
// (verification-completeness.md §1.1). The RED cell for this AC is that
// pinned baseline, not a run of this test. The swept added-line count is
// always logged so an empty sweep is visible, never silent.
//
// The working-tree form of the diff (git diff <base>) is used rather than
// the committed base...HEAD form: the acceptance contract's mutate probe
// requires a working-tree edit to flip the extraction, which only the
// working-tree form can observe; on a clean tree the two forms are the same
// diff.

import (
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestNoNewEnvLiteralsInDiff(t *testing.T) { // AC-SRL-009
	base := srlDiffBase(t)
	root := srlRepoRoot(t)
	out, err := exec.Command("git", "-C", root, "diff", base, "--",
		"internal/hook", "internal/factorymsg", "internal/cli",
		":!*envkeys.go", ":!*_test.go").Output()
	if err != nil {
		t.Fatalf("git diff failed: %v", err)
	}
	re := regexp.MustCompile(`MOAI_(FACTORY|KANBAN)[A-Z_]*`)
	literals := map[string]bool{}
	added := 0
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}
		added++
		for _, lit := range re.FindAllString(line, -1) {
			literals[lit] = true
		}
	}
	// The swept count rides the log so an empty sweep is checkable, never
	// silent; the recorded verdict cites this log line.
	t.Logf("env-literal sweep: %d added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=%s, distinct literals=%d", added, base, len(literals))
	if len(literals) > 0 {
		names := make([]string, 0, len(literals))
		for name := range literals {
			names = append(names, name)
		}
		sort.Strings(names)
		t.Fatalf("diff introduces env-name literals (REQ-SRL-009 — use internal/config/envkeys.go constants): %v", names)
	}
}

func srlRepoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// srlDiffBase resolves the card's diff base as the merge-base with the
// integration branch. With no resolvable base ref (a shallow CI checkout)
// it returns HEAD — the sweep is then empty and the log line says so, which
// the verdict records rather than reading as a silent pass.
func srlDiffBase(t *testing.T) string {
	t.Helper()
	root := srlRepoRoot(t)
	for _, ref := range []string{"develop", "origin/develop"} {
		if out, err := exec.Command("git", "-C", root, "merge-base", "HEAD", ref).Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return "HEAD"
}
