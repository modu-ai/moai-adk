//go:build unix

package hook

// role_rules_fifo_unix_test.go — card t1619, codex review P2 (role rules version
// diagnostic): a FIFO at system.yaml must not hang the diagnostic. The non-blocking
// rule read refuses a non-regular file, and the diagnostic skips it without a detail.

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRoleRulesVersionSkewDiagnosticSkipsFIFO(t *testing.T) {
	root := t.TempDir()
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := syscall.Mkfifo(filepath.Join(cfgDir, "system.yaml"), 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	done := make(chan string, 1)
	go func() { done <- roleRuleSkewDetail(root) }()
	select {
	case detail := <-done:
		if detail != "" {
			t.Fatalf("a FIFO system.yaml must be skipped without a detail, got %q", detail)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("roleRuleSkewDetail blocked on a FIFO system.yaml: the version diagnostic must read non-blocking")
	}
}
