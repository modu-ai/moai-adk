//go:build !windows

package statusline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// SPEC-STATUSLINE-OFFICIAL-USAGE-001 AC-SOU-010 (absorbed sou-debt-1): the
// fallback read must abandon MID-READ when the render context ends — not only
// at entry. The fixture blocks the aggregator inside its record read (a FIFO
// standing in for a slow record file), the context is cancelled only AFTER the
// read has started, and the build must return promptly with n/a segments. An
// entry-check-only guard passes a pre-cancelled-context test while never
// honoring this shape: here a pre-launch check alone leaves Build blocked on
// the read, tripping the deadline assertion below.
//
// The file is POSIX-only (syscall.Mkfifo); the windows build skips it. The
// named REQ-SOU-006 residual — the abandoned os.ReadFile finishing in its own
// goroutine — is accepted by the requirement and irrelevant to the
// caller-visible contract asserted here.

// TestBuild_QuotaFallbackAbandonsMidRead pins the mid-read abandonment
// contract.
func TestBuild_QuotaFallbackAbandonsMidRead(t *testing.T) {
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	wt := filepath.Join(base, "wt-card")

	// The primary's context-usage directory holds a record that blocks the
	// reader: opening a FIFO for reading blocks until a writer appears, and no
	// writer ever does. The read has therefore STARTED (it is inside
	// ReadSessionTelemetry) for as long as the fixture exists.
	cuDir := filepath.Join(primary, ".moai", "state", contextUsageDirName)
	if err := os.MkdirAll(cuDir, 0o755); err != nil {
		t.Fatalf("create context-usage dir: %v", err)
	}
	fifo := filepath.Join(cuDir, "blocked.json")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Fatalf("mkfifo blocked record: %v", err)
	}

	seedWorktreeAnchor(t, primary, wt,
		"gitdir: "+filepath.Join(primary, ".git", "worktrees", "x"), true)

	builder := quotaTestBuilder()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type outcome struct {
		out string
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		out, err := builder.Build(ctx, strings.NewReader(glmFormStdin(wt)))
		done <- outcome{out, err}
	}()

	// Let the fallback read start: the aggregator goroutine is inside the
	// blocked FIFO open well before this elapses.
	time.Sleep(300 * time.Millisecond)
	cancel() // mid-read cancellation — the pin of AC-SOU-010

	select {
	case o := <-done:
		if o.err != nil {
			t.Fatalf("Build error after mid-read abandonment: %v", o.err)
		}
		seg := usageSegment(t, o.out, "5H:")
		if !strings.Contains(seg, "n/a") {
			t.Errorf("abandoned fallback should render n/a, got segment %q\noutput:\n%s", seg, o.out)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("fallback read did not abandon mid-read: Build did not return within 3s of context cancellation — an entry-check-only guard would read exactly like this")
	}
}
