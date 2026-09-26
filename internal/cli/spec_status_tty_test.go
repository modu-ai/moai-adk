package cli

import (
	"os"
	"testing"
)

// A character device that is not a terminal (/dev/null) must not pass the
// non-TTY guard; otherwise `moai spec status --sync-git </dev/null` prompts
// and silently skips instead of reporting the non-TTY error (t1254).
func TestIsTerminalFileRejectsDevNull(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer func() { _ = f.Close() }()
	if isTerminalFile(f) {
		t.Fatalf("isTerminalFile(%s) = true, want false", os.DevNull)
	}
}

func TestIsTerminalFileRejectsPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer func() { _ = r.Close() }()
	defer func() { _ = w.Close() }()
	if isTerminalFile(r) {
		t.Fatal("isTerminalFile(pipe) = true, want false")
	}
}
