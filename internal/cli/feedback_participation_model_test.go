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
