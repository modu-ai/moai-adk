// doctor_user_install_test.go — the M5 doctor checks (SPEC-USER-ASSET-
// INSTALL-001; REQ-014 missing/modified/untracked, REQ-015 both drift
// directions). Temp HOMEs only.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/userassets"
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

// TestDoctorUserInstallWarnsOnUnreadableProfileStore pins the round-4 doctor
// decision (card t1702): the User Install check surfaces a profile-store
// enumeration failure as a WARN naming the store — never the disguise of
// "no profiles exist", which would silently skip every profile-root file in
// the comparison. Skips on Windows and under effective root (0o000
// unobservable).
func TestDoctorUserInstallWarnsOnUnreadableProfileStore(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not enforceable on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("effective root reads through 0o000 modes; the permission arm is unobservable")
	}
	home := t.TempDir()
	if err := ensureUserAssetsLocked(home, nil, &strings.Builder{}); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	// A profile store that exists but cannot be read.
	store := filepath.Join(home, ".moai", "claude-profiles")
	if err := os.MkdirAll(filepath.Join(store, "work"), 0o755); err != nil {
		t.Fatalf("create profile dir: %v", err)
	}
	if err := os.Chmod(store, 0o000); err != nil {
		t.Fatalf("chmod store: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(store, 0o755) })

	check := checkUserInstallIntegrity(home, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("unreadable profile store must WARN, got %v: %s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "cannot enumerate") || !strings.Contains(check.Message, store) {
		t.Errorf("warn should name the enumeration failure and the store path: %q", check.Message)
	}
}

// TestDoctorUserInstallUnenumeratedRootNotCountedVerified pins F3 (card
// t1702 audit iter4): manifest records under a root that is not currently
// enumerated (a deleted profile) are reported as unverifiable — never
// counted inside the "file(s) verified" total.
func TestDoctorUserInstallUnenumeratedRootNotCountedVerified(t *testing.T) {
	home := t.TempDir()
	if err := ensureUserAssetsLocked(home, nil, &strings.Builder{}); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	// A record under a profile slug that no longer enumerates ("ghost" has
	// no directory under the profile store).
	m, err := userassets.Load(userassets.ManifestPath(home))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	installed := len(m.Files)
	m.Files["claude-skills@ghost/moai/SKILL.md"] = userassets.FileEntry{SHA256: "deadbeef", Bundle: "core"}
	if err := m.Save(userassets.ManifestPath(home)); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	check := checkUserInstallIntegrity(home, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("a record under an unenumerated root must WARN, got %v: %s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "not currently enumerated") || !strings.Contains(check.Message, "1 record") {
		t.Errorf("warn should name the unverifiable record separately: %q", check.Message)
	}
	if !strings.Contains(check.Message, fmt.Sprintf("%d file(s) verified", installed)) {
		t.Errorf("verified count must cover only the enumerated records (%d): %q", installed, check.Message)
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
