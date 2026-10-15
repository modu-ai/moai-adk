package cli

// codex_lane_prompt_test.go — card t1776: the Codex `-l` lane's initial
// directive must be IMPERATIVE. The operator-measured failure: the previous
// descriptive prompt ("the cycle ranks…, dispatches…, judges…") sent the
// session reading todo_auto*.go and factory-dispatch*.md in full before it
// established the command's existence, burning 72% of the dispatch context.
// The prompt must command the run, forbid the source reading, pin the
// output-line contract, and stop on a refusal.

import (
	"strings"
	"testing"
)

func TestCodexLaneAutoPromptImperative(t *testing.T) {
	p := codexLaneAutoPrompt
	if !strings.HasPrefix(p, "Run `moai todo --auto` now.") {
		t.Errorf("prompt must open with the immediate command, got: %s", p)
	}
	if !strings.Contains(p, "Do not read") {
		t.Errorf("prompt must forbid the engine/rule source reading: %s", p)
	}
	for _, token := range []string{"selection:", "accept", "dispatch"} {
		if !strings.Contains(p, token) {
			t.Errorf("prompt does not pin the output-line contract token %q: %s", token, p)
		}
	}
	if !strings.Contains(p, "report that one line and stop") {
		t.Errorf("prompt lacks the refusal-stop directive: %s", p)
	}
	if strings.Contains(p, "the cycle ranks") {
		t.Errorf("prompt still carries the descriptive mechanism explainer that invited source reading: %s", p)
	}
}
