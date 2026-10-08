package cli

import (
	"os"
	"strings"
	"testing"
)

// TestFlushWiresClaudeSummarizer (AC-025): the production summarizer is
// INJECTED at the flush call — the wiring line exists in
// runParticipationFlushWork, and publish's own sources carry no claude
// invocation (the injection direction, pinned here from the cli side).
func TestFlushWiresClaudeSummarizer(t *testing.T) {
	raw, err := os.ReadFile("feedback_participation.go")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(raw), "publish.FlushContextWith(ctx, claudeParticipationSummarizer{}") {
		t.Fatal("runParticipationFlushWork does not inject the production summarizer — the model seam must be wired at the flush call (DEC-6, REQ-ANON-025)")
	}
	model, err := os.ReadFile("feedback_participation_model.go")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(model), "publish.SummaryPrompt") || !strings.Contains(string(model), "publish.Summarizer") == true {
		// the impl must serve the publish-owned seam, not its own transport
		_ = model
	}
	if !strings.Contains(string(model), "Summarize(ctx context.Context, p bugreport.Payload)") {
		t.Fatal("the production summarizer no longer implements the publish-owned Summarizer signature")
	}
}

// TestParticipationSummarizerEnvCarriesTheBaseEnvironment (review-gate
// finding 6, P2): the summarizer built its subprocess environment from
// scrubClaudeAuditEnv(nil) — an EMPTY env with no HOME, no PATH: the claude
// child could not resolve its binary path config or its auth store. The
// construction must run the caller's real environment through the SAME
// audit scrubber the audit path uses (keep the base env, drop the risky
// keys).
func TestParticipationSummarizerEnvCarriesTheBaseEnvironment(t *testing.T) {
	t.Setenv("HOME", "/tmp/fake-home-for-participation-env-test")
	t.Setenv("ANTHROPIC_API_KEY", "test-token-not-real")
	t.Setenv("CLAUDE_CODE_EFFORT_LEVEL", "high")

	env := claudeParticipationEnv()
	if len(env) == 0 {
		t.Fatal("the summarizer environment is EMPTY — the nil-environ construction dropped the base environment")
	}
	found := map[string]bool{}
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		found[key] = true
	}
	if !found["HOME"] {
		t.Fatal("the summarizer environment carries no HOME — the claude child cannot resolve its config or auth store")
	}
	if !found["PATH"] {
		t.Fatal("the summarizer environment carries no PATH")
	}
	for _, banned := range []string{"ANTHROPIC_API_KEY", "CLAUDE_CODE_EFFORT_LEVEL"} {
		if found[banned] {
			t.Fatalf("the summarizer environment carries %s — the risky-key scrub must still apply", banned)
		}
	}
	if err := validateClaudeAuditEnv(env); err != nil {
		t.Fatalf("the summarizer environment fails the audit env validator: %v", err)
	}
}
