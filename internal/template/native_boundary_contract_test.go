package template

import (
	"io/fs"
	"strings"
	"testing"
)

// TestNativeReviewBoundaryMatchesOfficialCommandsPage pins two statements of the
// native /code-review boundary to the official commands reference (re-read for
// card t2011): the argument syntax carries the --max-findings flag (v2.1.288) and
// a pr#|branch|path target, and /code-review is described as a correctness-bug
// reviewer whose cleanup findings depend on model and effort, not as a reviewer
// that always covers reuse, simplification and efficiency.
func TestNativeReviewBoundaryMatchesOfficialCommandsPage(t *testing.T) {
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	renderer := NewRenderer(embedded)
	render := func(path string) string {
		t.Helper()
		body, err := renderer.Render(path, nil)
		if err != nil {
			t.Fatalf("render %s: %v", path, err)
		}
		return string(body)
	}

	matrix := render(".claude/rules/moai/workflow/native-invocation-model.md")
	for _, required := range []string{"--max-findings", "[pr#"} {
		if !strings.Contains(matrix, required) {
			t.Errorf("native-invocation-model.md /code-review row lacks %q", required)
		}
	}

	// Official code-review-local: "Claude can start /code-review on its own"
	// (flag gate removed in v2.1.246); skills.md lists /init and /security-review
	// as the built-in commands exposed through the Skill tool, not /review.
	for _, stale := range []string{
		"| `/code-review` | HUMAN-ONLY",
		"| `/review` | HUMAN-ONLY",
		"including `/init`, `/review`, and `/security-review`",
	} {
		if strings.Contains(matrix, stale) {
			t.Errorf("native-invocation-model.md still carries stale classification %q", stale)
		}
	}
	if !strings.Contains(matrix, "user-invocable-only") {
		t.Error("native-invocation-model.md does not name the skillOverrides user-invocable-only switch")
	}

	goal := render(".claude/skills/moai/workflows/goal.md")
	if strings.Contains(goal, "the model cannot set it") {
		t.Error("goal.md states an unobserved generalization about native /goal")
	}

	// sync-auditor must not say it carries no Write while also mandating that it
	// writes the verdict file: the digest and contract files belong to the
	// orchestrator, the verdict export is the one file the auditor writes.
	// Agent definitions carry shell tokens the renderer refuses, so read them raw.
	raw := func(path string) string {
		t.Helper()
		body, err := fs.ReadFile(embedded, path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(body)
	}
	auditor := raw(".claude/agents/moai/sync-auditor.md")
	for _, conflict := range []string{"carries no `Write` or `Edit`", "has no Write tool"} {
		if strings.Contains(auditor, conflict) {
			t.Errorf("sync-auditor.md still carries the conflicting storage statement %q", conflict)
		}
	}

	// plan-auditor MP-5 / MP-6: an unexecutable verification verb is a Gap, so the
	// text must not tell the auditor that the N/A "auto-passes" for those two.
	planAuditor := raw(".claude/agents/moai/plan-auditor.md")
	for _, stale := range []string{
		"If the D7 verification verb is not executable (e.g., target files absent), mark N/A following the MP-4 precedent (N/A auto-passes)",
		"If the D8 verification verb is not executable, mark N/A following the MP-4 precedent (N/A auto-passes)",
	} {
		if strings.Contains(planAuditor, stale) {
			t.Errorf("plan-auditor.md still auto-passes an unexecuted verb: %q", stale)
		}
	}

	review := render(".claude/skills/moai/workflows/review.md")
	if strings.Contains(review, "covering the correctness-bug + reuse/simplification/efficiency portion") {
		t.Error("review.md still claims /code-review always covers reuse/simplification/efficiency")
	}
	if !strings.Contains(review, "ahead of its upstream plus uncommitted changes") {
		t.Error("review.md does not state that the no-flag HEAD~1 scope differs from native /code-review's default")
	}
	if !strings.Contains(review, "effort-dependent") {
		t.Error("review.md compose section does not say the cleanup findings are effort-dependent")
	}
}
