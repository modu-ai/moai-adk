package browser

import (
	"context"
	"testing"
)

// ExecProbe is exercised with plain shell utilities only; no browser command
// is ever started from a test.
func TestExecProbe(t *testing.T) {
	p := ExecProbe{}
	ctx := context.Background()

	if err := p.LookPath("sh"); err != nil {
		t.Fatalf("LookPath(sh): %v", err)
	}
	if err := p.LookPath("moai-no-such-binary"); err == nil {
		t.Fatalf("LookPath of a missing binary returned nil")
	}

	code, out, err := p.Run(ctx, "sh", "-c", "echo hi")
	if err != nil || code != 0 || out != "hi\n" {
		t.Fatalf("Run ok = (%d, %q, %v)", code, out, err)
	}
	code, _, err = p.Run(ctx, "sh", "-c", "exit 3")
	if err != nil || code != 3 {
		t.Fatalf("Run exit 3 = (%d, %v)", code, err)
	}
	if _, _, err = p.Run(ctx, "moai-no-such-binary"); err == nil {
		t.Fatalf("Run of a missing binary returned nil error")
	}
}
