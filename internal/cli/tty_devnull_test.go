package cli

import (
	"os"
	"testing"
)

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
