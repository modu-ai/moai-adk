package cli

import (
	"bytes"
	"strings"
	"sync"
	"time"
)

// codexErrorTailBytes bounds the stderr tail a receipt carries (t1788).
const codexErrorTailBytes = 1024

// codexStderrKeepBytes is how much stderr the capture holds before it trims back to
// its tail. The receipt reads only codexErrorTailBytes; the slack keeps a noisy child
// from forcing a trim on every write (t1788).
const codexStderrKeepBytes = 4 * codexErrorTailBytes

// codexStderrWaitDelay bounds how long close waits on the child's stderr pipe after the
// child has exited. A grandchild that inherited stderr can hold the pipe open for its
// whole life; without this bound a close would wait it out (t1788).
const codexStderrWaitDelay = 2 * time.Second

// codexStderrBuffer keeps what a codex child wrote to stderr. The exec copier
// writes it while the leg reads it, so access is locked.
type codexStderrBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

// Write appends p to the stream and trims it back to its tail once it grows past
// codexStderrKeepBytes. It never fails.
func (b *codexStderrBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n, err := b.buf.Write(p)
	if b.buf.Len() > codexStderrKeepBytes {
		tail := append([]byte(nil), b.buf.Bytes()[b.buf.Len()-codexErrorTailBytes:]...)
		b.buf.Reset()
		b.buf.Write(tail)
	}
	return n, err
}

// tail returns the last n bytes of the stream, trimmed of surrounding space.
func (b *codexStderrBuffer) tail(n int) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.buf.Bytes()
	if len(s) > n {
		s = s[len(s)-n:]
	}
	return strings.TrimSpace(string(s))
}

// codexErrorDetail is the codex_error a receipt records for a codex verdict: the
// cause of an inconclusive verdict, with the stderr tail when there is one. A pass
// or fail verdict carries no codex error (t1788).
func codexErrorDetail(verdict, summary, stderrTail string) string {
	if !strings.EqualFold(strings.TrimSpace(verdict), VerdictInconclusive) {
		return ""
	}
	detail := strings.TrimSpace(summary)
	if tail := strings.TrimSpace(stderrTail); tail != "" {
		if detail != "" {
			detail += " | "
		}
		detail += "stderr tail: " + tail
	}
	return detail
}

// codexReceiptErrorFromVerdicts returns the codex_error of an audit run from its
// codex per-backend verdict, or "" when the run had no codex verdict (t1788).
func codexReceiptErrorFromVerdicts(verdicts []PerBackendVerdict) string {
	for _, v := range verdicts {
		if v.Backend == BackendCodex {
			return codexErrorDetail(v.Verdict, v.Summary, v.StderrTail)
		}
	}
	return ""
}
