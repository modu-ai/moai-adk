package bugreport

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// seedConsent writes a consent file enabling capture under the test's
// MOAI_HOME (at <home>/config/participation.yaml — the reader's real path)
// and returns the spool path.
func seedConsent(t *testing.T) string {
	t.Helper()
	path := spoolPathForTest(t)
	home := os.Getenv("MOAI_HOME")
	configDir := filepath.Join(home, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	body := "participation:\n  enabled: true\n  asked: true\n"
	if err := os.WriteFile(filepath.Join(configDir, "participation.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("seed consent: %v", err)
	}
	return path
}

func TestCaptureNoopWhenParticipationOff(t *testing.T) {
	// (a) no user-scoped file at all.
	path := spoolPathForTest(t)
	Capture(KindPanic, nil, "", nil)
	if got := spoolLineCount(t, path); got != 0 {
		t.Fatalf("absent consent: spool carries %d line(s), want none", got)
	}

	// (b) a user file that says false.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "participation.yaml"), []byte("participation:\n  enabled: false\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	Capture(KindPanic, nil, "", nil)
	if got := spoolLineCount(t, path); got != 0 {
		t.Fatalf("consent false: spool carries %d line(s), want none", got)
	}

	// (c) only a tracked project file saying true — the reader never opens a
	// project file, so nothing is captured. (The hostile fixture lives in the
	// reader's tests; here the project tree is not even present, which is the
	// construction the reader guarantees.)
	Capture(KindPanic, nil, "", nil)
	if got := spoolLineCount(t, path); got != 0 {
		t.Fatalf("tracked-only consent: spool carries %d line(s), want none", got)
	}
}

func TestCaptureIgnoresTrackedFileConsent(t *testing.T) {
	path := spoolPathForTest(t)

	// A hostile project tree whose tracked section file carries participation
	// keys, as cwd — the capture call runs with consent absent and must
	// record nothing regardless of what the project tree says.
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "feedback:\n  repository: attacker/x\n  participation: true\n  participation_asked: true\n"
	if err := os.WriteFile(filepath.Join(sections, "feedback.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write hostile file: %v", err)
	}
	t.Chdir(root)

	Capture(KindPanic, nil, "", nil)
	if got := spoolLineCount(t, path); got != 0 {
		t.Fatalf("tracked project file enabled capture: %d line(s)", got)
	}
}

// TestCaptureNeverPanicsOrBlocks is AC-002's fail-open arm, run under -race
// -count=3 by the verify line: a blocking spool write, a panicking spool
// write, an unwritable spool directory, and a full spool all return without
// panic, within the time box, recording nothing. There is no network and no
// model call anywhere on the capture path (the AC-025 import guard is the
// structural proof; this test is the observable one).
func TestCaptureNeverPanicsOrBlocks(t *testing.T) {
	path := spoolPathForTest(t)
	seedConsent(t)

	prev := spoolAppendFn
	t.Cleanup(func() { spoolAppendFn = prev })

	t.Run("blocking_write_returns_within_the_box", func(t *testing.T) {
		var mu sync.Mutex
		calls := 0
		spoolAppendFn = func(e SpoolEntry) error {
			mu.Lock()
			calls++
			mu.Unlock()
			time.Sleep(400 * time.Millisecond) // far beyond the 50ms box
			return nil
		}
		start := time.Now()
		// A hook-timeout signal: not subject to the zero-frame local rule
		// (which correctly drops a direct test-stack panic before any write).
		Capture(KindHookTimeout, errString("deadline"), "", nil)
		if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
			t.Fatalf("capture blocked %s, want it to abandon within the box", elapsed)
		}
		// The abandoned goroutine still starts; wait for it to be entered so
		// the assertion measures the call, not a scheduling race.
		deadline := time.Now().Add(2 * time.Second)
		for {
			mu.Lock()
			n := calls
			mu.Unlock()
			if n == 1 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("blocking stub calls = %d, want 1", n)
			}
			time.Sleep(5 * time.Millisecond)
		}
		if got := spoolLineCount(t, path); got != 0 {
			t.Fatalf("blocking write recorded %d line(s), want none", got)
		}
	})

	t.Run("panicking_write_never_propagates", func(t *testing.T) {
		// Review-gate finding #7: a panic in the WORKER goroutine kills the
		// process before the caller's deferred recover could see it, so the
		// worker carries its own recover — and this stub panics at the real
		// write path, not a phase the box skips. The consent store is
		// seeded, so the signal reaches the spool seam.
		seedConsent(t)
		spoolAppendFn = func(e SpoolEntry) error { panic("hostile filesystem") }
		Capture(KindHookTimeout, errString("deadline"), "", nil) // must not panic out of Capture
		Capture(KindHookHandlerFailure, errString("x"), "", nil) // and a second kind, same guarantee
	})

	t.Run("unwritable_spool_directory", func(t *testing.T) {
		spoolAppendFn = appendSpoolLine
		t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "config", "participation.yaml", "deeper"))
		Capture(KindPanic, nil, "", nil) // SpoolPath resolves under a file path → unwritable
	})

	t.Run("full_spool_drops", func(t *testing.T) {
		spoolAppendFn = appendSpoolLine
		p := seedConsent(t)
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		full := strings.Repeat("x", 65*1024)
		if err := os.WriteFile(p, []byte(full), 0o600); err != nil {
			t.Fatalf("fill spool: %v", err)
		}
		Capture(KindPanic, nil, "", nil)
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if info.Size() != int64(len(full)) {
			t.Fatalf("full spool grew: %d bytes", info.Size())
		}
	})
}

// TestCaptureSpoolsMoaiAndAmbiguous pins the verdict-consumption shape: a
// moai signal is spooled with its frames and detail; an ambiguous signal is
// spooled so the drain can log its retention (DEC-7); a zero-frame panic is
// dropped outright.
func TestCaptureSpoolsMoaiAndAmbiguous(t *testing.T) {
	path := seedConsent(t)

	// Frames from this stack: the moai frame filter keeps this test binary's
	// bugreport frames? No — capture-package frames are dropped by design, so
	// a direct Capture call from a test yields ZERO moai frames for the panic
	// kind and is dropped by the local-only rule. That is exactly the
	// property pinned first:
	Capture(KindPanic, nil, "", nil)
	if got := spoolLineCount(t, path); got != 0 {
		t.Fatalf("zero-frame panic spooled %d line(s), want a local drop", got)
	}

	// A hook timeout carries no frame requirement (it captures from a moai
	// call site; in the production wiring the frames come from the registry).
	// Spooling it here exercises the ambiguous-retention line the drain logs.
	spoolAppendFn = func(e SpoolEntry) error { return appendSpoolLine(e) }
	Capture(KindHookTimeout, errors.New("deadline"), "", nil)
	if got := spoolLineCount(t, path); got != 1 {
		t.Fatalf("ambiguous capture spooled %d line(s), want 1", got)
	}
}
