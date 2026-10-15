package cli

import (
	"os"
	"path/filepath"
	"runtime"
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

// TestUpdateSettingsEnvHooksSymlinkRefused pins F1 (card t1702 audit
// iter4): the global-hooks cleanup never FOLLOWS a symlink — a config
// root whose hooks (or hooks/moai) is a symbolic link into an external
// directory keeps the external subtree intact; the removal is refused and
// even the link itself is left in place. Skips on windows (os.Symlink
// needs privileges there).
func TestUpdateSettingsEnvHooksSymlinkRefused(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Symlink requires privileges on windows")
	}
	tmp := t.TempDir()

	// Shape 1: <root>/hooks is a symlink to an external dir holding moai/.
	root1 := filepath.Join(tmp, "prof1")
	if err := os.MkdirAll(root1, 0o755); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(tmp, "external")
	keep := filepath.Join(external, "moai", "deep", "keep.txt")
	if err := os.MkdirAll(filepath.Dir(keep), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(root1, "hooks")); err != nil {
		t.Fatal(err)
	}

	if err := ensureSettingsEnvInConfigRoot(root1); err != nil {
		t.Fatalf("ensureSettingsEnvInConfigRoot(root1): %v", err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("external content deleted through the symlinked hooks dir: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(root1, "hooks")); err != nil {
		t.Fatalf("the symlink itself must be left in place (refused, not removed): %v", err)
	}

	// Shape 2: hooks is a real dir, hooks/moai is a symlink to an external dir.
	root2 := filepath.Join(tmp, "prof2")
	hooks2 := filepath.Join(root2, "hooks")
	if err := os.MkdirAll(hooks2, 0o755); err != nil {
		t.Fatal(err)
	}
	external2 := filepath.Join(tmp, "external2")
	keep2 := filepath.Join(external2, "deep", "keep.txt")
	if err := os.MkdirAll(filepath.Dir(keep2), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keep2, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external2, filepath.Join(hooks2, "moai")); err != nil {
		t.Fatal(err)
	}

	if err := ensureSettingsEnvInConfigRoot(root2); err != nil {
		t.Fatalf("ensureSettingsEnvInConfigRoot(root2): %v", err)
	}
	if _, err := os.Stat(keep2); err != nil {
		t.Fatalf("external content deleted through the symlinked hooks/moai: %v", err)
	}
}
