package cli

import (
	"os"
	"testing"
)

// /dev/null is a character device but not a terminal. Detection that checks
// os.ModeCharDevice treats `moai ... </dev/null` as interactive (t1262).
func TestDefaultCodexPromptCapableRejectsDevNull(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer func() { _ = f.Close() }()
	prev := os.Stdin
	os.Stdin = f
	t.Cleanup(func() { os.Stdin = prev })
	if defaultCodexPromptCapable() {
		t.Fatalf("defaultCodexPromptCapable() with stdin=%s = true, want false", os.DevNull)
	}
}

func TestWriterIsTerminalRejectsDevNull(t *testing.T) {
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer func() { _ = f.Close() }()
	if writerIsTerminal(f) {
		t.Fatalf("writerIsTerminal(%s) = true, want false", os.DevNull)
	}
}
