//go:build !windows

package hook

// POSIX-only FIFO fixture for the role-rules builder's regular-file guard
// (SPEC-ALWAYS-LOADED-BUDGET-001): a FIFO at the rule path must fail visible,
// not hang SessionStart. Split from role_rules_test.go because syscall.Mkfifo
// has no Windows symbol — the runtime.GOOS skip inside the test body cannot
// prevent the reference from breaking the Windows build.

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestSessionStartRoleRulesFifoFailVisible(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	root := t.TempDir()
	dispatch := roleRuleFiles[0]
	messaging := roleRuleFiles[1]
	writeRoleRuleFixture(t, root, dispatch, smallMarkedRule("dispatch"))
	writeRoleRuleFixture(t, root, messaging, smallMarkedRule("messaging"))

	path := filepath.Join(root, filepath.FromSlash(dispatch.Rel))
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove regular fixture: %v", err)
	}
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	// The builder stats before reading (regular files only), so the
	// injection returns the REQ-ALB-009 pair — and this test COMPLETING is
	// itself the no-hang proof: os.ReadFile on a FIFO would block until the
	// test deadline.
	inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
	assertFailVisible(t, inj, "fifo")
}

// TestReadRuleFileBytesFifoNonblocking is the TOCTOU-hardening seam test:
// the POSIX reader opens NON-BLOCKING and judges the opened handle, so a
// FIFO at the rule path is rejected immediately — no writer, no hang. This
// exercises the seam directly (a swap-race fixture would be flaky; the
// open-regular-file path is exercised by every other fixture in this file).
func TestReadRuleFileBytesFifoNonblocking(t *testing.T) {
	root := t.TempDir()
	fifo := filepath.Join(root, "role_rules_fifo_probe")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := readRuleFileBytes(fifo); err == nil {
			t.Errorf("FIFO read succeeded — the non-regular rejection did not fire")
		}
	}()
	select {
	case <-done:
		// rejected without a writer: the non-blocking open + handle fstat
		// path held
	case <-time.After(4 * time.Second):
		t.Fatal("readRuleFileBytes blocked 4s on a FIFO with no writer — the TOCTOU window is open")
	}
	<-done
}
