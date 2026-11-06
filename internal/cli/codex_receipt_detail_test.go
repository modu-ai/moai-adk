package cli

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
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

// The capture is bounded in memory too: a long leg that writes much to stderr keeps
// only its tail, not the whole stream (t1788).
func TestCodexStderrBufferRetainsBoundedMemory(t *testing.T) {
	var b codexStderrBuffer
	chunk := []byte(strings.Repeat("x", 4096))
	for i := 0; i < 256; i++ { // 1 MiB of stderr from one leg
		if _, err := b.Write(chunk); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := b.Write([]byte("END")); err != nil {
		t.Fatal(err)
	}
	if held := b.buf.Len(); held > 16*codexErrorTailBytes {
		t.Fatalf("the capture holds %d bytes, want at most %d", held, 16*codexErrorTailBytes)
	}
	if tail := b.tail(codexErrorTailBytes); !strings.HasSuffix(tail, "END") {
		t.Fatalf("the bounded capture lost the end of the stream: %q", tail)
	}
}

// A grandchild that inherits the child's stderr keeps the pipe open after the child
// has exited. close must not wait out that grandchild: a stalled close stalls the
// audit leg it closes (t1788).
func TestCodexConnCloseDoesNotWaitForOrphanedStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the orphaned-grandchild shape needs a POSIX shell")
	}
	conn, err := realCodexSessionRunner{}.start(context.Background(), "/bin/sh", []string{"-c", "sleep 8 & echo gc=$! >&2; echo ready >&2; read line"}, "")
	if err != nil {
		t.Fatal(err)
	}
	c := conn.(*realCodexConn)
	// The backgrounded sleep outlives the shell; stop it by the pid it reported, on
	// every path, so the test leaves no process behind.
	t.Cleanup(func() {
		tail := c.stderr.tail(codexErrorTailBytes)
		if _, after, ok := strings.Cut(tail, "gc="); ok && len(strings.Fields(after)) > 0 {
			if pid, convErr := strconv.Atoi(strings.Fields(after)[0]); convErr == nil {
				if p, findErr := os.FindProcess(pid); findErr == nil {
					_ = p.Kill()
				}
			}
		}
	})
	began := time.Now()
	_ = c.close()
	if took := time.Since(began); took > 5*time.Second {
		t.Fatalf("close waited %s on a grandchild that holds stderr, want at most 5s", took)
	}
	if tail := c.stderr.tail(codexErrorTailBytes); !strings.Contains(tail, "ready") {
		t.Fatalf("the capture lost the child's stderr: %q", tail)
	}
}

// A required codex gate turns an inconclusive verdict into fail before the receipt is
// written (applyGateUnmet). The receipt must still carry the no-verdict cause and the
// stderr tail: they are why the verdict was empty (t1788, review P2).
func TestCodexAudit_RequiredGateReceiptKeepsCauseAndStderrTail(t *testing.T) {
	root := newProbeProject(t, "SPEC-CAGRECEIPT-005")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	prevLook, prevRPC := codexLookPath, codexReviewRPC
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexReviewRPC = func(context.Context, string, string, map[string]any) (ReviewOutput, error) {
		out := inconclusiveReview("codex stdout closed before response to id=1")
		out.StderrTail = "approval request failed"
		return out, nil
	}
	t.Cleanup(func() { codexLookPath, codexReviewRPC = prevLook, prevRPC })

	got := structuredMap(t, callToolCodexAudit(t, map[string]any{"project_root": root}))
	id, _ := got["audit_receipt"].(string)
	if id == "" {
		t.Fatalf("a required-gate block must expose its receipt id, got result %v", got)
	}
	r, err := auditreceipt.ReadReceipt(root, id)
	if err != nil {
		t.Fatalf("ReadReceipt %s: %v", id, err)
	}
	if !strings.Contains(r.CodexError, "codex stdout closed before response to id=1") {
		t.Errorf("receipt codex_error = %q, want the no-verdict cause kept under the required gate", r.CodexError)
	}
	if !strings.Contains(r.CodexError, "approval request failed") {
		t.Errorf("receipt codex_error = %q, want the stderr tail kept under the required gate", r.CodexError)
	}
}

// A codex child that exits before the handshake answers leaves its reason on stderr.
// The failed-open path closes the child first, so the tail is read after the close and
// carried on the review output (t1788, review P2).
func TestCodexReviewStderrTailSurvivesFailedOpen(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the failing-child shape needs a POSIX shell")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Card t1632 CI repair: the REQ-MN-009 containment listing is a
	// precondition BEFORE the session open, and the real listing would exec
	// /bin/sh with `mcp list` and fail the spawn there — the subject here is
	// the failed-open CHILD's stderr, so the listing precondition is stubbed
	// empty (the withCodexSession rule) and the child itself still fails.
	stubCodexContainmentListing(t)
	// /bin/sh runs with the app-server subcommand as its script name, so it fails before
	// it can answer the handshake and names that subcommand on stderr.
	out, _ := runCodexReviewRPCResolved(ctx, "/bin/sh", codexMethodReviewStart, map[string]any{}, nil)
	if out.Verdict != VerdictInconclusive {
		t.Fatalf("verdict = %q, want inconclusive when the session never opens", out.Verdict)
	}
	if !strings.Contains(out.StderrTail, codexAppServerSubcmd) {
		t.Fatalf("stderr tail = %q, want the child's stderr to name %q", out.StderrTail, codexAppServerSubcmd)
	}
}
