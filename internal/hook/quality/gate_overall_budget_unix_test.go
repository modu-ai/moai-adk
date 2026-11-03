//go:build unix

package quality

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// A disabled gate never runs a step, so the budget pre-computation must not
// read project manifests: a package.json that is a FIFO blocks any reader.
func TestOverallBudgetDisabledGateDoesNotReadManifests(t *testing.T) {
	dir := t.TempDir()
	fifo := filepath.Join(dir, "package.json")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	t.Cleanup(func() {
		// Unblock a reader stuck in open(2) so a failing run does not leak it.
		// O_RDWR never blocks on a FIFO, even when no reader is waiting (a
		// plain O_WRONLY open would hang the cleanup of a passing run).
		if f, err := os.OpenFile(fifo, os.O_RDWR, 0); err == nil {
			_ = f.Close()
		}
	})

	cfg := budgetCfg(dir)
	cfg.Enabled = false
	done := make(chan time.Duration, 1)
	go func() {
		d, _ := NewQualityGate(cfg).OverallBudget()
		done <- d
	}()
	select {
	case d := <-done:
		if d <= 0 {
			t.Errorf("OverallBudget = %s, want a positive finite budget", d)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("OverallBudget blocked on a manifest while the gate is disabled")
	}
}
