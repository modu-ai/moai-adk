package template

import (
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

	review := render(".claude/skills/moai/workflows/review.md")
	if strings.Contains(review, "covering the correctness-bug + reuse/simplification/efficiency portion") {
		t.Error("review.md still claims /code-review always covers reuse/simplification/efficiency")
	}
	if !strings.Contains(review, "effort-dependent") {
		t.Error("review.md compose section does not say the cleanup findings are effort-dependent")
	}
}
