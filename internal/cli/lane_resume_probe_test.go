package cli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

// TestClaudeHelpSynopsisBoundedUnderLingeringChild (card-review r1, P2①;
// cleanup reworked per r2, P2) — the derivation probe's time bound must
// survive a PATH wrapper whose child process lingers holding the stdout
// pipe: exec.CommandContext's default WaitDelay of zero makes Output() wait
// for pipe EOF even after Cancel, so a wrapper that exits immediately but
// backgrounds a pipe-holding grandchild would block launcher entry until
// that grandchild exits. The probe must return inside its actual bound
// (3s timeout + 1s grace) with the bounded failure shape, and the
// degradation path must leave the active model — the snapshot — untouched.
//
// The wrapper is a THROWAWAY fixture script — never the real claude —
// because the behavior under test is the probe's own process hygiene, which
// cannot be observed without a real process tree. It lives in this sibling
// file so the derivation-classification test files stay free of process
// references (the AC-SCV-012 no-spawn greps name those files, not this
// one). Cleanup kills ONLY the grandchild this test started: the wrapper
// records its backgrounded job's pid in a file inside t.TempDir, and the
// cleanup reads that exact pid — no pattern matching over any process
// table, ever (card-review r2, P2).
func TestClaudeHelpSynopsisBoundedUnderLingeringChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the lingering-child repro needs a POSIX shell")
	}
	saveActiveClaudeOptionModel(t)
	activeClaudeOptionModel = claudeOptionModelSnapshot // pin: the degradation asserts against this

	dir := t.TempDir()
	pidFile := filepath.Join(dir, "linger.pid")
	script := filepath.Join(dir, "claude-linger")
	// The wrapper prints a synopsis line and exits immediately; the
	// backgrounded sleep inherits the stdout pipe and lingers — the exact
	// shape that breaks an unbounded pipe wait — and its pid is recorded
	// for an exact, registered cleanup. The sleep is long enough to hold
	// the pipe past the probe's whole bound and short enough that a runner
	// whose sandbox denies signaling (kill attempts are observed, never
	// discarded) sees the stray expire on its own moments later.
	body := "#!/bin/sh\nsleep 12 &\necho $! > " + pidFile + "\necho '  -p, --print   print response'\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatalf("write fixture script: %v", err)
	}
	t.Cleanup(func() {
		raw, err := os.ReadFile(pidFile)
		if err != nil {
			return // the wrapper never got to record it — nothing of ours to kill
		}
		pid := 0
		for _, b := range raw {
			if b < '0' || b > '9' {
				break
			}
			pid = pid*10 + int(b-'0')
		}
		if pid <= 1 {
			return
		}
		// Exactly the grandchild this test started — never a pattern sweep.
		// A sandboxed runner may deny the signal; that outcome is surfaced,
		// and the fixture's own 12s expiry bounds the stray regardless.
		if p, err := os.FindProcess(pid); err != nil {
			t.Logf("linger cleanup: find pid %d: %v", pid, err)
		} else if err := p.Kill(); err != nil {
			t.Logf("linger cleanup: kill pid %d denied (%v) — the fixture expires on its own", pid, err)
		}
	})

	start := time.Now()
	text, err := claudeHelpSynopsis(script)
	elapsed := time.Since(start)

	// The bound is the 3s context plus the 1s grace; a compliant probe
	// returns well inside twice that.
	if elapsed > 5*time.Second {
		t.Fatalf("claudeHelpSynopsis returned after %v, want well inside the 3s+1s bound", elapsed)
	}
	// The bounded return is a FAILURE shape: the wrapper exited
	// successfully while the lingering grandchild held the pipe, so the
	// wait-delay grace expired — the probe reports the bounded error (or a
	// context/exec failure) rather than a truncated success.
	if err == nil {
		t.Fatal("claudeHelpSynopsis returned nil error for a pipe-holding wrapper; want the bounded failure shape")
	}
	if !errors.Is(err, exec.ErrWaitDelay) &&
		!errors.Is(err, context.DeadlineExceeded) &&
		!errors.As(err, new(*exec.ExitError)) {
		t.Fatalf("unexpected bounded error shape: %v", err)
	}
	_ = text // whatever partial output arrived is discarded by the degradation

	// The degradation shape: applyClaudeOptionModel over the failed probe
	// must leave the active model — the pinned snapshot — untouched.
	applyClaudeOptionModel(script)
	if !reflect.DeepEqual(activeClaudeOptionModel, claudeOptionModelSnapshot) {
		t.Fatalf("active model swapped off the snapshot after a bounded probe failure: %v", activeClaudeOptionModel)
	}
}
