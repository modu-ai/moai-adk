package template

import (
	"io/fs"
	"regexp"
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

// resumeRequiredWords are the words the Resume Check section must carry. The
// "- Field:" entries are the identity lines the progress.md initial entry
// records (the Resume Check compares against recorded values, so a producer
// must exist), and "unrecorded identity field" is the rule for a field that was never written.
var resumeRequiredWords = []string{
	"repository root", "HEAD", "plan_artifact_hash", "card", "invalidate",
	"whether the file was just created or already existed", "- Phase {N} complete: HEAD",
	"unrecorded identity field", "- Repository root:", "- Worktree:", "- Lease owner:", "- HEAD:", "- plan_artifact_hash:",
}

// checkpointLineRE matches a phase-checkpoint template line anywhere in a run
// body. Every such line must carry the identity the Resume Check compares, or
// a run that follows the template resumes from the first phase every time.
var checkpointLineRE = regexp.MustCompile(`(?m)^- Phase (\d+) complete:.*$`)

// checkpointViolations lists checkpoint template lines that omit the HEAD or
// the plan_artifact_hash.
func checkpointViolations(text string) []string {
	var v []string
	for _, m := range checkpointLineRE.FindAllString(text, -1) {
		if !strings.Contains(m, "HEAD") || !strings.Contains(m, "plan_artifact_hash") {
			v = append(v, "checkpoint without HEAD and plan_artifact_hash: "+m)
		}
	}
	return v
}

// TestRunCheckpointTemplatesCarryIdentity checks every checkpoint template in
// the deployed run bodies.
func TestRunCheckpointTemplatesCarryIdentity(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	seen := 0
	for _, p := range []string{
		".claude/skills/moai/workflows/run.md",
		contextLoadingPath,
		".claude/skills/moai/workflows/run/phase-execution.md",
		".claude/skills/moai/workflows/run/task-decomposition.md",
	} {
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		seen += len(checkpointLineRE.FindAllString(string(b), -1))
		if v := checkpointViolations(string(b)); len(v) > 0 {
			t.Errorf("%s:\n  %s", p, strings.Join(v, "\n  "))
		}
	}
	if seen == 0 {
		t.Fatal("no checkpoint template lines found: the check would pass vacuously")
	}
}

// TestRunCheckpointTemplates_Mutants proves the checker can fail in each
// direction and that a complete line passes.
func TestRunCheckpointTemplates_Mutants(t *testing.T) {
	t.Parallel()

	if v := checkpointViolations("- Phase 2 complete: strategy=x\n"); len(v) != 1 {
		t.Errorf("bare checkpoint must be one violation, got %v", v)
	}
	if v := checkpointViolations("- Phase 2 complete: HEAD {sha}, strategy=x\n"); len(v) != 1 {
		t.Errorf("checkpoint without the hash must be one violation, got %v", v)
	}
	if v := checkpointViolations("- Phase 2 complete: plan_artifact_hash {h}, strategy=x\n"); len(v) != 1 {
		t.Errorf("checkpoint without HEAD must be one violation, got %v", v)
	}
	if v := checkpointViolations("- Phase 2 complete: HEAD {sha}, plan_artifact_hash {h}, strategy=x\n"); len(v) != 0 {
		t.Errorf("complete checkpoint must pass, got %v", v)
	}
}

// resumeIdentityViolations lists the identity checks a Resume Check section
// does not require before skipping a completed phase.
func resumeIdentityViolations(text string) []string {
	body, ok := section(text, "### Resume Check", "## Worktree Path Rules")
	if !ok {
		return []string{"Resume Check: section not found"}
	}
	var v []string
	for _, w := range resumeRequiredWords {
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
	if v := resumeIdentityViolations(stale); len(v) != len(resumeRequiredWords) {
		t.Fatalf("stale section must miss all %d words, got %d: %v", len(resumeRequiredWords), len(v), v)
	}

	const good = "### Resume Check\nVerify repository root, card, HEAD and plan_artifact_hash; invalidate on a difference or an unrecorded identity field; whether the file was just created or already existed, record the block.\n- Phase {N} complete: HEAD x\n- Repository root: x\n- Worktree: x\n- Lease owner: x\n- HEAD: x\n- plan_artifact_hash: x\n## Worktree Path Rules\n"
	if v := resumeIdentityViolations(good); len(v) != 0 {
		t.Fatalf("positive control must pass, got: %v", v)
	}
	for _, drop := range resumeRequiredWords {
		if v := resumeIdentityViolations(strings.ReplaceAll(good, drop, "x")); len(v) == 0 {
			t.Errorf("dropping %q must be detected", drop)
		}
	}
}
