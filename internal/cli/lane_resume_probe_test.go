package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestClaudeHelpSynopsisBoundedUnderLingeringChild (card-review r1, P2①) —
// the derivation probe's time bound must survive a PATH wrapper whose child
// process lingers holding the stdout pipe: exec.CommandContext's default
// WaitDelay of zero makes Output() wait for pipe EOF even after Cancel, so
// a wrapper that exits immediately but backgrounds a pipe-holding grandchild
// would block launcher entry until that grandchild exits. The probe must
// return within its bound (timeout + grace).
//
// This test drives a THROWAWAY fixture script — never the real claude —
// because the behavior under test is the probe's own process hygiene, which
// cannot be observed without a real process tree. It lives in this sibling
// file so the derivation-classification test files stay free of process
// references (the AC-SCV-012 no-spawn greps name those files, not this one).
func TestClaudeHelpSynopsisBoundedUnderLingeringChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the lingering-child repro needs a POSIX shell")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "claude-linger")
	// The wrapper prints a synopsis line and exits immediately, but the
	// backgrounded sleep inherits the stdout pipe and lingers — the exact
	// shape that breaks an unbounded pipe wait. The sleep duration is an
	// odd, greppable number so the cleanup kills only this test's child.
	body := "#!/bin/sh\nsleep 27 &\necho '  -p, --print   print response'\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatalf("write fixture script: %v", err)
	}
	t.Cleanup(func() { _ = exec.Command("pkill", "-f", "sleep 27").Run() })

	start := time.Now()
	_, _ = claudeHelpSynopsis(script) // err is expected — the point is the bounded return
	elapsed := time.Since(start)

	if elapsed > 10*time.Second {
		t.Fatalf("claudeHelpSynopsis returned after %v, want within the 3s timeout + grace bound", elapsed)
	}
}
