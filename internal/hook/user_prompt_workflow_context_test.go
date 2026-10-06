package hook

// user_prompt_workflow_context_test.go — card t1499 M2.
//
// detectWorkflowContext used strings.Contains, so "running", "planning" and
// "prune" fired the injection, and it fired on every matching prompt: 3,043 of
// 3,218 injections in a 14-day sample were this one line. The rule chosen here
// for bare words is the word boundary; the slash-command forms (/moai run,
// /moai plan, /moai loop) are covered by the same boundary because the
// keyword is a whole word there.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestDetectWorkflowContext_WordBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		prompt string
		want   string // expected keyword, "" for no injection
	}{
		{"running is not the word run", "running tests now", ""},
		{"planning is not the word plan", "planning the release", ""},
		{"prune does not contain a whole-word keyword", "prune the old branches", ""},
		{"runtime is not the word run", "check the runtime config", ""},
		{"looped is not the word loop", "it looped forever", ""},
		{"underscore identifier is not the word run", "call run_tests helper", ""},
		{"slash command plan fires", "/moai plan x", "plan"},
		{"slash command run fires", "/moai run SPEC-001", "run"},
		{"slash command loop fires", "/moai loop fix errors", "loop"},
		{"bare word plan fires", "please plan the work", "plan"},
		{"bare word run fires", "Help me run this task", "run"},
		{"upper case bare word fires", "LOOP until fixed", "loop"},
		{"keyword before punctuation fires", "ok, plan.", "plan"},
		{"keyword followed by Korean particle fires", "plan을 세워줘", "plan"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := detectWorkflowContext(tt.prompt)
			if tt.want == "" {
				if got != "" {
					t.Fatalf("detectWorkflowContext(%q) = %q, want empty", tt.prompt, got)
				}
				return
			}
			wantMsg := "workflow keyword '" + tt.want + "' detected — MoAI workflow context may be active"
			if got != wantMsg {
				t.Fatalf("detectWorkflowContext(%q) = %q, want %q", tt.prompt, got, wantMsg)
			}
		})
	}
}

// promptContext runs one UserPromptSubmit through the real handler and returns
// the additionalContext it produced.
func promptContext(t *testing.T, projectDir, sessionID, prompt string) string {
	t.Helper()
	h := NewUserPromptSubmitHandler(&mockConfigProvider{cfg: newTestConfig()})
	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:     sessionID,
		CWD:           projectDir,
		HookEventName: "UserPromptSubmit",
		Prompt:        prompt,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil || out.HookSpecificOutput == nil {
		return ""
	}
	return out.HookSpecificOutput.AdditionalContext
}

func TestUserPromptSubmit_WorkflowContextOncePerSession(t *testing.T) {
	projectDir := newMoaiProjectRoot(t)
	t.Setenv(config.EnvClaudeProjectDir, projectDir)

	const marker = "workflow keyword"

	if got := promptContext(t, projectDir, "sess-a", "/moai plan x"); !strings.Contains(got, marker) {
		t.Fatalf("first matching prompt: additionalContext = %q, want the workflow line", got)
	}
	if got := promptContext(t, projectDir, "sess-a", "now run it"); strings.Contains(got, marker) {
		t.Errorf("second matching prompt in the same session repeated the line: %q", got)
	}
	if got := promptContext(t, projectDir, "sess-b", "/moai run SPEC-001"); !strings.Contains(got, marker) {
		t.Errorf("a different session must fire once: additionalContext = %q", got)
	}
	if got := promptContext(t, projectDir, "sess-b", "plan again"); strings.Contains(got, marker) {
		t.Errorf("second prompt in the other session repeated the line: %q", got)
	}
}

// A prompt that does not match must not consume the session's one injection.
func TestUserPromptSubmit_NonMatchingPromptDoesNotSpendTheInjection(t *testing.T) {
	projectDir := newMoaiProjectRoot(t)
	t.Setenv(config.EnvClaudeProjectDir, projectDir)

	if got := promptContext(t, projectDir, "sess-c", "running tests"); strings.Contains(got, "workflow keyword") {
		t.Fatalf("non-matching prompt injected: %q", got)
	}
	if got := promptContext(t, projectDir, "sess-c", "plan the work"); !strings.Contains(got, "workflow keyword") {
		t.Errorf("first real match was suppressed: %q", got)
	}
}
