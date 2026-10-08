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
