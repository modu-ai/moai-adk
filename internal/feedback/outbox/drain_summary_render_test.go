package outbox

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// TestRenderWithSummaryIdentity: the summary slots AHEAD of the
// deterministic template text, and an empty summary renders BYTE-IDENTICAL
// to RenderReport — the template path never drifts from the one render
// (design section 8; D38's M6 half).
func TestRenderWithSummaryIdentity(t *testing.T) {
	p := fixedRenderPayload()
	title, plain := RenderReport(p)

	if withTitle, withPlain := RenderReportWithSummary(p, ""); withTitle != title || withPlain != plain {
		t.Fatal("an empty summary did not render byte-identically to RenderReport")
	}

	summary := "A validated model summary."
	title2, withSummary := RenderReportWithSummary(p, summary)
	if title2 != title {
		t.Fatal("the summary changed the title")
	}
	markerEnd := strings.Index(withSummary, "-->")
	if markerEnd < 0 {
		t.Fatal("no marker block")
	}
	rest := withSummary[markerEnd+3:]
	if !strings.HasPrefix(rest, "\n\n"+summary+"\n\n") {
		t.Fatalf("the summary is not slotted ahead of the template text: %q", rest[:60])
	}
	if !strings.Contains(withSummary, "A moai-adk user opted into automatic improvement participation") {
		t.Fatal("the template text disappeared")
	}
	if plain == withSummary {
		t.Fatal("a non-empty summary changed nothing")
	}
}

// fixedRenderPayload is a platform-independent payload for the render
// tests (the golden's shape).
func fixedRenderPayload() bugreport.Payload {
	return bugreport.Payload{
		Schema:      "v1",
		Kind:        bugreport.KindPanic,
		Fingerprint: "0123456789abcdef",
		Version:     "v3.2.0",
		Commit:      "abcdef1234567",
		OS:          "linux",
		Arch:        "amd64",
		Frames:      []string{"internal/cli.Execute"},
	}
}
