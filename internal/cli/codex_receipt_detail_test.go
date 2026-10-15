package cli

import (
	"strings"
	"testing"
)

// The detail names the cause of an inconclusive codex verdict and carries the
// tail of the codex child's stderr, so a receipt alone shows why the leg gave no
// verdict (t1788 brain ruling, condition 2).
func TestCodexErrorDetailNamesInconclusiveCause(t *testing.T) {
	if got := codexErrorDetail("pass", "Verdict: pass", "boom"); got != "" {
		t.Fatalf("a pass verdict must carry no codex error, got %q", got)
	}
	got := codexErrorDetail("inconclusive", "codex leg timed out after 20m0s", "approval request failed")
	if !strings.Contains(got, "codex leg timed out after 20m0s") {
		t.Fatalf("the cause is missing from the codex error: %q", got)
	}
	if !strings.Contains(got, "approval request failed") {
		t.Fatalf("the stderr tail is missing from the codex error: %q", got)
	}
}

// The tail is bounded and keeps the end of the stream, where the last messages are.
func TestCodexStderrBufferTailIsBounded(t *testing.T) {
	var b codexStderrBuffer
	if _, err := b.Write([]byte(strings.Repeat("x", 3000) + "END")); err != nil {
		t.Fatal(err)
	}
	tail := b.tail(1024)
	if len(tail) > 1024 {
		t.Fatalf("tail is %d bytes, want at most 1024", len(tail))
	}
	if !strings.HasSuffix(tail, "END") {
		t.Fatalf("tail must keep the end of the stream, got suffix %q", tail[len(tail)-10:])
	}
}

// The audit_multi receipt takes its codex detail from the codex per-backend verdict,
// and a run without a codex verdict carries none.
func TestCodexReceiptErrorFromVerdictsReadsCodexBackend(t *testing.T) {
	verdicts := []PerBackendVerdict{
		{Backend: BackendClaude, Verdict: "pass", Summary: "ok"},
		{
			Backend:    BackendCodex,
			Verdict:    VerdictInconclusive,
			Summary:    "codex stdout closed before response to id=1",
			StderrTail: "approval request failed",
		},
	}
	got := codexReceiptErrorFromVerdicts(verdicts)
	if !strings.Contains(got, "codex stdout closed before response to id=1") {
		t.Fatalf("the codex detail lacks the cause: %q", got)
	}
	if !strings.Contains(got, "approval request failed") {
		t.Fatalf("the codex detail lacks the stderr tail: %q", got)
	}
	none := codexReceiptErrorFromVerdicts([]PerBackendVerdict{{Backend: BackendClaude, Verdict: "pass"}})
	if none != "" {
		t.Fatalf("a run without a codex verdict must carry no codex error, got %q", none)
	}
}
