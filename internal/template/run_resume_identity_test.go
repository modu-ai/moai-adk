package template

import (
	"io/fs"
	"strings"
	"testing"
)

// Card t2008 / RUN-06. The run workflow's Resume Check reads the last completed
// phase from progress.md and skips every completed phase. It names no identity
// to re-verify, although the factory runtime already records the SPEC revision
// and git commit a card was assigned (internal/factory/factory_runtime.go
// factoryProvenance). A progress.md written for another tree, another HEAD, or
// an edited plan must not be trusted to skip work.

const contextLoadingPath = ".claude/skills/moai/workflows/run/context-loading.md"

// resumeIdentityViolations lists the identity checks a Resume Check section
// does not require before skipping a completed phase.
func resumeIdentityViolations(text string) []string {
	body, ok := section(text, "### Resume Check", "## Worktree Path Rules")
	if !ok {
		return []string{"Resume Check: section not found"}
	}
	var v []string
	for _, w := range []string{
		"repository root", "HEAD", "plan_artifact_hash", "card", "invalidate",
	} {
		if !strings.Contains(body, w) {
			v = append(v, "Resume Check: missing "+w)
		}
	}
	return v
}

// TestRunResumeIdentity_Deployed checks the deployed template body.
func TestRunResumeIdentity_Deployed(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fs.ReadFile(fsys, contextLoadingPath)
	if err != nil {
		t.Fatalf("read %s: %v", contextLoadingPath, err)
	}
	if v := resumeIdentityViolations(string(data)); len(v) > 0 {
		t.Errorf("%s skips completed phases without an identity check:\n  %s", contextLoadingPath, strings.Join(v, "\n  "))
	}
}

// TestRunResumeIdentity_Mutants proves the checker can fail: the current-style
// section is rejected, the positive control passes, and each required word is
// load-bearing.
func TestRunResumeIdentity_Mutants(t *testing.T) {
	t.Parallel()

	const stale = "### Resume Check\nIdentify last completed phase, skip all completed phases.\n## Worktree Path Rules\n"
	if v := resumeIdentityViolations(stale); len(v) != 5 {
		t.Fatalf("stale section must miss all 5 words, got %d: %v", len(v), v)
	}

	const good = "### Resume Check\nVerify repository root, card, HEAD and plan_artifact_hash; invalidate on a difference.\n## Worktree Path Rules\n"
	if v := resumeIdentityViolations(good); len(v) != 0 {
		t.Fatalf("positive control must pass, got: %v", v)
	}
	for _, drop := range []string{"repository root", "HEAD", "plan_artifact_hash", "card", "invalidate"} {
		if v := resumeIdentityViolations(strings.ReplaceAll(good, drop, "x")); len(v) == 0 {
			t.Errorf("dropping %q must be detected", drop)
		}
	}
}
