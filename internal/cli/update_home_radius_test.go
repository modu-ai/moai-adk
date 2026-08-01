package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEnsureGlobalSettingsEnv_HooksRemovalRadius pins the deletion radius of
// ensureGlobalSettingsEnv's global-hooks cleanup to <HOME>/.claude/hooks/moai.
// A widened radius (e.g. <HOME>/.claude/hooks) would destroy user-owned hook
// files that live as siblings of the moai subdirectory.
//
// The test redirects HOME through the package-level userHomeDirFn seam rather
// than mutating the HOME environment variable: a process-wide env mutation is
// visible to every parallel test in this package. Because it reassigns a
// package-level variable, this test MUST NOT call t.Parallel().
func TestEnsureGlobalSettingsEnv_HooksRemovalRadius(t *testing.T) {
	tmp := t.TempDir()

	origFn := userHomeDirFn
	userHomeDirFn = func() (string, error) { return tmp, nil }
	t.Cleanup(func() { userHomeDirFn = origFn })

	// The removal target comes from the named symbol under test, not from an
	// independently re-derived path (REQ-UDS-013).
	moaiHooksDir := globalMoaiHooksDir(tmp)
	if err := os.MkdirAll(moaiHooksDir, 0o755); err != nil {
		t.Fatalf("create moai hooks dir: %v", err)
	}
	moaiHookFile := filepath.Join(moaiHooksDir, "handle-session-start.sh")
	if err := os.WriteFile(moaiHookFile, []byte("#!/bin/bash\n"), 0o755); err != nil {
		t.Fatalf("write moai hook file: %v", err)
	}

	// A user-owned sibling of the moai subdirectory that MUST survive.
	siblingHook := filepath.Join(tmp, ".claude", "hooks", "user-hook.sh")
	if err := os.WriteFile(siblingHook, []byte("#!/bin/bash\n# user-owned\n"), 0o755); err != nil {
		t.Fatalf("write sibling hook: %v", err)
	}

	if err := ensureGlobalSettingsEnv(); err != nil {
		t.Fatalf("ensureGlobalSettingsEnv failed: %v", err)
	}

	if _, err := os.Stat(moaiHooksDir); !os.IsNotExist(err) {
		t.Errorf("removal target %s should be gone, stat err = %v", moaiHooksDir, err)
	}

	if _, err := os.Stat(siblingHook); err != nil {
		t.Errorf("user-owned sibling %s must survive, stat err = %v", siblingHook, err)
	}
}
