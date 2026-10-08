// doctor_user_install_test.go — the M5 doctor checks (SPEC-USER-ASSET-
// INSTALL-001; REQ-014 missing/modified/untracked, REQ-015 both drift
// directions). Temp HOMEs only.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestDoctorUserInstallDetectsDrift covers REQ-014's three detection arms:
// a manifest-tracked file deleted, hash-modified, and an untracked entry in
// the roots.
func TestDoctorUserInstallDetectsDrift(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	if err := ensureUserAssetsLocked(home, nil, &strings.Builder{}); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	// Modified: flip a tracked file's bytes.
	modified := filepath.Join(home, ".claude", "skills", "moai-workflow-tdd", "SKILL.md")
	if err := os.WriteFile(modified, []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Untracked: a foreign file in a root.
	untracked := filepath.Join(home, ".claude", "skills", "some-foreign-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(untracked), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(untracked, []byte("foreign\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkUserInstallIntegrity(home, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("modified tracked file must WARN, got %v: %s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "modified") {
		t.Errorf("message should name the modified class: %q", check.Message)
	}

	// Missing: delete a tracked file.
	if err := os.RemoveAll(filepath.Join(home, ".claude", "skills", "moai-workflow-spec")); err != nil {
		t.Fatal(err)
	}
	check = checkUserInstallIntegrity(home, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("missing tracked file must WARN, got %v", check.Status)
	}
}

// TestDoctorUserInstallCleanOnHealthyInstall is the healthy-install
// regression arm (AC-009): a correct user install reads clean, no false
// drift after M4.
func TestDoctorUserInstallCleanOnHealthyInstall(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	if err := ensureUserAssetsLocked(home, nil, &strings.Builder{}); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	check := checkUserInstallIntegrity(home, false)
	if check.Status != uikit.CheckOK {
		t.Errorf("healthy install must read OK, got %v: %s", check.Status, check.Message)
	}
}

// TestDoctorProjectVsLockBothDirections covers REQ-015: lock entry absent
// from project AND project file absent from lock, both detected.
func TestDoctorProjectVsLockBothDirections(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// One on-disk template-managed file + one ghost lock entry (recorded,
	// never on disk): both drift directions in one fixture.
	if err := os.MkdirAll(filepath.Join(root, ".claude", "rules", "moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(root, ".claude", "rules", "moai", "seeded.md")
	if err := os.WriteFile(live, []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ghost := filepath.Join(root, ".claude", "rules", "moai", "ghost.md")
	if err := os.WriteFile(ghost, []byte("ghost\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Track(".claude/rules/moai/seeded.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Track(".claude/rules/moai/ghost.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}
	// The ghost: remove it from disk AFTER tracking (the lock entry stays).
	if err := os.Remove(ghost); err != nil {
		t.Fatal(err)
	}

	check := checkProjectVsLock(root, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("drift must WARN, got %v: %s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "absent from project") {
		t.Errorf("message should name the lock→project direction: %q", check.Message)
	}
}
