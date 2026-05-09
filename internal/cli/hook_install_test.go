package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestInstallPrePushHook_FreshRepo verifies that installing into a fresh git repo
// creates an executable pre-push hook containing the ci-local make target.
func TestInstallPrePushHook_FreshRepo(t *testing.T) {
	dir := t.TempDir()
	// Initialize a git repo in the temp dir.
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}

	installer := NewPrePushInstaller(dir)
	if err := installer.InstallPrePushHook(false); err != nil {
		t.Fatalf("InstallPrePushHook(false) error: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-push")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("hook file not created at %s: %v", hookPath, err)
	}

	// Verify executable bit is set (mode & 0o111).
	// Windows does not support Unix executable bits; skip this check on Windows.
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		t.Errorf("hook file is not executable: mode %v", info.Mode())
	}

	// Verify content contains the MoAI marker and ci-local invocation.
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook file: %v", err)
	}

	wantStrings := []string{
		moaiPrePushMarker,
		"ci-local",
	}
	for _, want := range wantStrings {
		if !strings.Contains(string(content), want) {
			t.Errorf("hook content missing %q\ncontent:\n%s", want, content)
		}
	}
}

// TestInstallPrePushHook_SkipFlag verifies that skip=true results in no hook file.
func TestInstallPrePushHook_SkipFlag(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}

	installer := NewPrePushInstaller(dir)
	if err := installer.InstallPrePushHook(true); err != nil {
		t.Fatalf("InstallPrePushHook(true) error: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-push")
	if _, err := os.Stat(hookPath); err == nil {
		t.Error("hook file should NOT be created when skip=true")
	}
}

// TestInstallPrePushHook_PreservesUserHook verifies that a pre-existing hook
// without the MoAI marker is preserved and ErrUserHookExists is returned.
func TestInstallPrePushHook_PreservesUserHook(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}

	// Create a pre-existing hook WITHOUT the MoAI marker.
	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	userHookContent := "#!/bin/sh\n# User hook (no MoAI marker)\necho 'my hook'\n"
	hookPath := filepath.Join(hooksDir, "pre-push")
	if err := os.WriteFile(hookPath, []byte(userHookContent), 0o755); err != nil {
		t.Fatalf("write user hook: %v", err)
	}

	installer := NewPrePushInstaller(dir)
	err := installer.InstallPrePushHook(false)

	if !errors.Is(err, ErrUserHookExists) {
		t.Errorf("expected ErrUserHookExists, got: %v", err)
	}

	// Verify user hook content is preserved.
	got, readErr := os.ReadFile(hookPath)
	if readErr != nil {
		t.Fatalf("read hook: %v", readErr)
	}
	if string(got) != userHookContent {
		t.Errorf("user hook was modified; expected original content preserved")
	}
}

// TestInstallPrePushHook_OverwritesMoaiHook verifies that a pre-existing hook
// WITH the MoAI marker is safely overwritten.
func TestInstallPrePushHook_OverwritesMoaiHook(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}

	// Create a pre-existing hook WITH the MoAI marker (old version).
	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	oldContent := "#!/bin/sh\n" + moaiPrePushMarker + "\n# old version\n"
	hookPath := filepath.Join(hooksDir, "pre-push")
	if err := os.WriteFile(hookPath, []byte(oldContent), 0o755); err != nil {
		t.Fatalf("write old moai hook: %v", err)
	}

	installer := NewPrePushInstaller(dir)
	if err := installer.InstallPrePushHook(false); err != nil {
		t.Fatalf("InstallPrePushHook() error: %v", err)
	}

	got, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	// New content should contain ci-local, not the "old version" placeholder.
	if strings.Contains(string(got), "# old version") {
		t.Error("hook should be overwritten with new content, found old marker")
	}
	if !strings.Contains(string(got), "ci-local") {
		t.Error("new hook should contain ci-local")
	}
}
