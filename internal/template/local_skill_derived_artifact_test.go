package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Local skill copies are derived artifacts, and a divergence outside the
// redeploy glob is the case where a template-only fix loses work.
//
// # Why this guard exists
//
// Commit a64548a2a (#1273) swept 31 skills' worth of audit findings. 42 of its
// 44 files were under internal/template/templates/; exactly one local copy came
// with them. A later card was then dispatched against two of those findings as
// "live defects" because they still reproduced in .claude/skills/ — but the
// template copies had been correct since that commit, and only the local
// deployment was behind.
//
// The tempting reading is that the sweep had a scope bug and the fix is a guard
// requiring both copies to agree. Measured against this tree, that guard is not
// available: of the template skill .md files, well over a hundred already differ
// from their local twin. Divergence is the normal state, not the anomaly,
// because `moai update` deletes .claude/skills/moai* wholesale and redeploys it
// from the embedded templates (ManagedCleanTargets in
// internal/cli/update/deploy/deploy.go carries that path with IsGlob: true). A
// local copy under that glob is a deployment artifact whose content is whatever
// the last update wrote — so a template-only fix is complete, and the stale
// local copy is lag, not a missed edit.
//
// What is NOT safe is a local skill that diverges from a template twin while
// sitting OUTSIDE that glob. No update would ever overwrite it, so it is a real
// second source, and a template-only sweep silently leaves it behind. That is
// the case this guard names.
//
// The guard therefore does not require the copies to agree. It requires that
// every disagreement be one the redeploy glob covers.
//
// # How far this guard currently reaches
//
// Stated plainly, because a reader should not have to infer it: every skill
// directory shipped in the template tree today begins with "moai", so the
// failing branch below is NOT reachable from the steady tree. The sweep passes
// because there is nothing outside the glob to find, not because it searched
// and cleared something. Its value is prospective — it fires the day a skill
// ships under a name the glob does not take.
//
// That it fires at all was proven by mutation rather than assumed: planting a
// template skill "zzmutant/SKILL.md" with a diverging local twin produced
//
//	1 local skill file(s) diverge ... outside the `.claude/skills/moai*`
//	redeploy glob ...
//	  zzmutant/SKILL.md
//
// and removing the pair restored the pass. The count assertions above cover the
// other vacuity direction — the day the two trees converge, this says so
// instead of reporting a pass it did not earn.

// localSkillPathIsRedeployed reports whether a path relative to the skills root
// (for example "moai-domain-backend/SKILL.md") names a local skill copy that
// `moai update` would delete and rewrite from the embedded templates.
//
// The predicate mirrors the ManagedCleanTargets entry
// filepath.Join(ClaudeDir, SkillsSubdir, "moai*") with IsGlob: true. That glob
// matches at the first path segment only, which is why this tests the leading
// directory rather than the whole path.
func localSkillPathIsRedeployed(relFromSkillsRoot string) bool {
	rel := filepath.ToSlash(strings.TrimPrefix(filepath.ToSlash(relFromSkillsRoot), "./"))
	if rel == "" {
		return false
	}
	first := rel
	if i := strings.Index(rel, "/"); i >= 0 {
		first = rel[:i]
	}
	return strings.HasPrefix(first, "moai")
}

// TestLocalSkillPathIsRedeployed_Predicate exercises the predicate directly, so
// a change to the glob's shape fails here with a readable name rather than only
// as a tree-wide sweep. The false cases are the ones that matter: those are the
// paths a template-only fix would abandon.
func TestLocalSkillPathIsRedeployed_Predicate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		rel  string
		want bool
	}{
		{"moai/SKILL.md", true},
		{"moai/workflows/plan/spec-assembly.md", true},
		{"moai-domain-backend/SKILL.md", true},
		{"moai-foundation-cc/reference/sub-agents/index.md", true},
		// Names that do not start with the glob's literal prefix are never
		// redeployed, so a divergence there is a second source.
		{"hns-lsel-curator/SKILL.md", false},
		{"hns-workflow-ci-loop/SKILL.md", false},
		{"vendor-skill/SKILL.md", false},
		{"", false},
		// "moai" is a prefix match, not a segment match: the glob would also
		// take a name that merely starts with it. Pinning this stops a future
		// reader from "fixing" the predicate into an equality test and
		// silently narrowing what counts as redeployed.
		{"moaix/SKILL.md", true},
	}

	for _, c := range cases {
		if got := localSkillPathIsRedeployed(c.rel); got != c.want {
			t.Errorf("localSkillPathIsRedeployed(%q) = %v, want %v", c.rel, got, c.want)
		}
	}
}

// TestDivergingLocalSkillCopiesAreAllRedeployed walks the shipped skill tree and
// requires every local twin that differs to be one the redeploy glob covers.
//
// It deliberately does NOT require the copies to match. The count assertion
// below is what keeps it honest: if the two trees ever became identical the
// sweep would pass while observing nothing, so the test says so instead of
// reporting a pass it did not earn.
func TestDivergingLocalSkillCopiesAreAllRedeployed(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	templateSkills := filepath.Join(repoRoot, "internal", "template", "templates", ".claude", "skills")
	localSkills := filepath.Join(repoRoot, ".claude", "skills")

	if _, err := os.Stat(localSkills); err != nil {
		t.Skipf("no local skills tree to compare against: %v", err)
	}

	var compared, diverging int
	var offenders []string

	walkErr := filepath.Walk(templateSkills, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		rel, relErr := filepath.Rel(templateSkills, path)
		if relErr != nil {
			return relErr
		}
		localPath := filepath.Join(localSkills, rel)
		localBytes, localErr := os.ReadFile(localPath) //nolint:gosec // test-scoped path under the repo
		if localErr != nil {
			// No local twin at all is not this guard's concern: nothing
			// diverged, because nothing is there to diverge.
			return nil //nolint:nilerr // absence is not a finding here
		}
		templateBytes, tmplErr := os.ReadFile(path) //nolint:gosec // test-scoped path under the repo
		if tmplErr != nil {
			return tmplErr
		}
		compared++
		if string(localBytes) == string(templateBytes) {
			return nil
		}
		diverging++
		if !localSkillPathIsRedeployed(rel) {
			offenders = append(offenders, rel)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walking %s: %v", templateSkills, walkErr)
	}

	if compared == 0 {
		t.Fatal("compared no skill files — the template or local skills tree moved, so this guard observed nothing")
	}
	if diverging == 0 {
		t.Fatalf("every one of the %d compared skill files matches its local twin; "+
			"this guard's subject has disappeared and it now passes without observing anything", compared)
	}

	if len(offenders) > 0 {
		t.Errorf("%d local skill file(s) diverge from the shipped template while sitting outside the "+
			"`.claude/skills/moai*` redeploy glob, so no `moai update` will ever reconcile them — "+
			"a template-only fix silently leaves these behind:\n  %s",
			len(offenders), strings.Join(offenders, "\n  "))
	}
}
