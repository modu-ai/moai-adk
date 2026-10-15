package cli

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
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
	conn, err := realCodexSessionRunner{}.start(context.Background(), "/bin/sh", []string{"-c", "sleep 8 & echo ready >&2; read line"})
	if err != nil {
		t.Fatal(err)
	}
	c := conn.(*realCodexConn)
	began := time.Now()
	_ = c.close()
	if took := time.Since(began); took > 5*time.Second {
		t.Fatalf("close waited %s on a grandchild that holds stderr, want at most 5s", took)
	}
	if tail := c.stderr.tail(codexErrorTailBytes); !strings.Contains(tail, "ready") {
		t.Fatalf("the capture lost the child's stderr: %q", tail)
	}
}
