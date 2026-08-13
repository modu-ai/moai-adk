// agent_model_matcher_test.go — settings.json.tmpl wiring guards for the
// PreToolUse agent-model observation hook.
//
// Two invariants are pinned here:
//
//  1. A DEDICATED PreToolUse matcher block observes Agent/Task. It must be a
//     separate block, never a widening of the existing "Write|Edit|Bash"
//     matcher — widening would run the Bash-only logic (branch guard, commit
//     quality gate) on every Agent spawn and spread any regression's blast
//     radius across Write/Edit as well.
//  2. The new block follows the house hook convention: shell wrapper,
//     quoted/braced $CLAUDE_PROJECT_DIR, timeout 10.
package template_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// amProjectRoot walks up to the directory containing go.mod.
func amProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}

// amSettingsBody returns the raw settings.json.tmpl body.
func amSettingsBody(t *testing.T) string {
	t.Helper()
	path := filepath.Join(amProjectRoot(t), "internal", "template", "templates", ".claude", "settings.json.tmpl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("template settings.json.tmpl not found at %s (consumer checkout?)", path)
	}
	return string(data)
}

// hookBlock is one entry of a PreToolUse array.
type hookBlock struct {
	Matcher string `json:"matcher"`
	Hooks   []struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
		Timeout int      `json:"timeout"`
		Type    string   `json:"type"`
	} `json:"hooks"`
}

// extractJSONArray returns the JSON array that follows `"<key>": [` in body,
// matched by bracket depth. The PreToolUse section carries no Go template
// directives, so the extracted span parses as plain JSON.
func extractJSONArray(t *testing.T, body, key string) string {
	t.Helper()
	marker := `"` + key + `": [`
	start := strings.Index(body, marker)
	if start < 0 {
		t.Fatalf("key %q not found in settings template", key)
	}
	open := start + len(marker) - 1 // position of '['
	depth := 0
	for i := open; i < len(body); i++ {
		switch body[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return body[open : i+1]
			}
		}
	}
	t.Fatalf("unterminated array for key %q", key)
	return ""
}

func preToolUseBlocks(t *testing.T, body string) []hookBlock {
	t.Helper()
	raw := extractJSONArray(t, body, "PreToolUse")
	var blocks []hookBlock
	if err := json.Unmarshal([]byte(raw), &blocks); err != nil {
		t.Fatalf("PreToolUse array does not parse as JSON: %v\n%s", err, raw)
	}
	return blocks
}

// matchesAgent reports whether a matcher string selects the Agent or Task tool.
func matchesAgent(matcher string) bool {
	for _, alt := range strings.Split(matcher, "|") {
		if alt == "Agent" || alt == "Task" {
			return true
		}
	}
	return false
}

// TestSettingsPreToolUseAgentMatcher pins AC-AME-010.
func TestSettingsPreToolUseAgentMatcher(t *testing.T) {
	t.Parallel()
	blocks := preToolUseBlocks(t, amSettingsBody(t))

	// (a) exactly one block observes Agent/Task.
	agentBlocks := 0
	for _, b := range blocks {
		if matchesAgent(b.Matcher) {
			agentBlocks++
		}
	}
	if agentBlocks != 1 {
		t.Errorf("PreToolUse blocks matching Agent/Task: got %d, want exactly 1", agentBlocks)
	}

	// (b) the pre-existing Write|Edit|Bash block survives byte-for-byte.
	found := false
	for _, b := range blocks {
		if b.Matcher == "Write|Edit|Bash" {
			found = true
		}
	}
	if !found {
		t.Errorf(`the existing PreToolUse matcher "Write|Edit|Bash" is absent — it must not be renamed or widened`)
	}
}

// TestSettingsPreToolUseAgentMatcherRedFixture is the (c) RED fixture: a
// variant that widens the existing matcher instead of adding a block must fail
// the (b) assertion. This proves the (b) check actually discriminates rather
// than passing vacuously.
func TestSettingsPreToolUseAgentMatcherRedFixture(t *testing.T) {
	t.Parallel()
	widened := strings.Replace(amSettingsBody(t), `"matcher": "Write|Edit|Bash"`, `"matcher": "Write|Edit|Bash|Agent"`, 1)
	blocks := preToolUseBlocks(t, widened)

	for _, b := range blocks {
		if b.Matcher == "Write|Edit|Bash" {
			t.Fatalf("RED fixture did not take effect: the unwidened matcher is still present")
		}
	}
	// The widened variant would pass a naive "some block matches Agent" check,
	// which is exactly why the (b) byte-equality assertion above is required.
	sawAgent := false
	for _, b := range blocks {
		if matchesAgent(b.Matcher) {
			sawAgent = true
		}
	}
	if !sawAgent {
		t.Errorf("RED fixture sanity: the widened matcher should still select Agent")
	}
}

// TestHookWrapperConventions pins AC-AME-052: the new PreToolUse entry follows
// the house hook wiring convention.
func TestHookWrapperConventions(t *testing.T) {
	t.Parallel()
	blocks := preToolUseBlocks(t, amSettingsBody(t))

	for _, b := range blocks {
		if !matchesAgent(b.Matcher) {
			continue
		}
		if len(b.Hooks) != 1 {
			t.Fatalf("agent matcher block: got %d hooks, want 1", len(b.Hooks))
		}
		h := b.Hooks[0]

		if h.Type != "command" || h.Command != "bash" {
			t.Errorf("hook invocation: got type=%q command=%q, want type=command command=bash", h.Type, h.Command)
		}
		// Guarded exec form: args = ["-c", "<missing-script guard>", "<wrapper>"].
		// The wrapper path is the final args element.
		if len(h.Args) != 3 || h.Args[0] != "-c" {
			t.Fatalf("hook args: got %v, want guarded exec form [\"-c\", guard, path]", h.Args)
		}
		arg := h.Args[2]
		// (a) points at a shell wrapper under the moai hook directory.
		if !strings.HasSuffix(arg, ".sh") || !strings.Contains(arg, "/.claude/hooks/moai/") {
			t.Errorf("hook arg %q does not point at a .claude/hooks/moai/*.sh wrapper", arg)
		}
		// (b) $CLAUDE_PROJECT_DIR is brace-delimited, never bare.
		if !strings.Contains(arg, "${CLAUDE_PROJECT_DIR}") {
			t.Errorf("hook arg %q does not use ${CLAUDE_PROJECT_DIR}", arg)
		}
		// (c) the MoAI 10s timeout budget (#1473 widened PreToolUse 5→10).
		if h.Timeout != 10 {
			t.Errorf("hook timeout: got %d, want 10", h.Timeout)
		}
	}
}
