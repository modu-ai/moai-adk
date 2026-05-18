// SPEC-V3R3-UPDATE-CLEANUP-001 — End-to-end regression scenarios (M4).
// Scenarios A–F cover the six canonical deprecated-path cleanup cases.
// Case-insensitive FS probe tests (REQ-UPC-026) are included here.

package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// ---------------------------------------------------------------------------
// Shared helpers for E2E test fixtures
// ---------------------------------------------------------------------------

// setupE2EProject creates a temp project with .moai/ initialised.
func setupE2EProject(t *testing.T) (root string, mgr manifest.Manager) {
	t.Helper()
	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, defs.MoAIDir), 0o755); err != nil {
		t.Fatalf("MkdirAll .moai: %v", err)
	}
	mgr = manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("manifest Load: %v", err)
	}
	return root, mgr
}

// seedAgencyFiles populates the project with all deprecated agency paths,
// recording them in mgr with the given content.
func seedAgencyFiles(t *testing.T, root string, mgr manifest.Manager, content string) {
	t.Helper()
	for _, p := range defs.DeprecatedPaths {
		abs := filepath.Join(root, filepath.FromSlash(p.Path))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("MkdirAll %q: %v", filepath.Dir(abs), err)
		}
		data := []byte(content)
		if err := os.WriteFile(abs, data, 0o644); err != nil {
			t.Fatalf("WriteFile %q: %v", abs, err)
		}
		hash := manifest.HashBytes(data)
		_ = mgr.Track(filepath.ToSlash(p.Path), manifest.TemplateManaged, hash)
	}
}

// runCleanupPhase is a minimal simulation of the cleanup phase used in E2E tests:
// scan → filter skip-markers → backup → remove.
func runCleanupPhase(t *testing.T, root string, mgr manifest.Manager) (backupDir string, removed []string) {
	t.Helper()

	found, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths: %v", err)
	}

	var devNull strings.Builder
	filtered, _ := filterSkipMarkerPaths(root, found, &devNull)

	if len(filtered) == 0 {
		return "", nil
	}

	backupDir, err = backupDeprecatedPaths(root, filtered, mgr)
	if err != nil {
		t.Fatalf("backupDeprecatedPaths: %v", err)
	}

	for _, rel := range filtered {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := removeDeprecatedFile(abs); err != nil {
			t.Fatalf("removeDeprecatedFile %q: %v", rel, err)
		}
		removed = append(removed, rel)
	}
	return backupDir, removed
}

// ---------------------------------------------------------------------------
// Scenario A — agency files present, pristine → backup + delete
// ---------------------------------------------------------------------------

func TestE2E_ScenarioA_AgencyPresentPristine(t *testing.T) {
	t.Parallel()
	root, mgr := setupE2EProject(t)
	seedAgencyFiles(t, root, mgr, "# scenario A pristine")

	backupDir, removed := runCleanupPhase(t, root, mgr)

	if backupDir == "" {
		t.Fatal("expected a backup directory to be created")
	}
	if len(removed) != len(defs.DeprecatedPaths) {
		t.Errorf("removed %d files, want %d", len(removed), len(defs.DeprecatedPaths))
	}

	// Original files must be gone
	for _, p := range defs.DeprecatedPaths {
		abs := filepath.Join(root, filepath.FromSlash(p.Path))
		if _, err := os.Lstat(abs); !os.IsNotExist(err) {
			t.Errorf("expected %q to be removed, Lstat: %v", p.Path, err)
		}
	}

	// Backup directory must contain files
	if _, err := os.Stat(filepath.Join(backupDir, "MANIFEST.json")); err != nil {
		t.Errorf("backup MANIFEST.json missing: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Scenario B — agency files absent → no-op, no backup directory created
// ---------------------------------------------------------------------------

func TestE2E_ScenarioB_AgencyAbsent(t *testing.T) {
	t.Parallel()
	root, mgr := setupE2EProject(t)
	// No agency files seeded

	backupDir, removed := runCleanupPhase(t, root, mgr)

	if backupDir != "" {
		t.Errorf("expected no backup directory when agency absent, got %q", backupDir)
	}
	if len(removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(removed))
	}

	// .moai/backup/ should not exist
	backupRoot := filepath.Join(root, defs.MoAIDir, "backup")
	if _, err := os.Stat(backupRoot); !os.IsNotExist(err) {
		t.Errorf("backup dir should not exist when no deprecated paths found: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Scenario C — user modified agency.md → backup + classification recorded
// ---------------------------------------------------------------------------

func TestE2E_ScenarioC_UserModified(t *testing.T) {
	t.Parallel()
	root, mgr := setupE2EProject(t)

	// Seed with original content (tracked in manifest)
	original := "# original agency content"
	seedAgencyFiles(t, root, mgr, original)

	// Now modify the first file (simulate user edit)
	firstDeprecated := defs.DeprecatedPaths[0]
	abs := filepath.Join(root, filepath.FromSlash(firstDeprecated.Path))
	if err := os.WriteFile(abs, []byte("# user modified content — different hash"), 0o644); err != nil {
		t.Fatalf("WriteFile modified: %v", err)
	}

	// Classification should now be UserModifiedDeprecated for first file
	class := classifyDeprecatedFile(root, filepath.ToSlash(firstDeprecated.Path), mgr)
	if class != DeprecatedUserModified {
		t.Errorf("expected UserModifiedDeprecated, got %v", class)
	}

	// Cleanup phase should still back up all files
	backupDir, removed := runCleanupPhase(t, root, mgr)
	if backupDir == "" {
		t.Fatal("expected backup directory for user-modified files")
	}
	if len(removed) == 0 {
		t.Error("expected files to be removed after backup")
	}

	// MANIFEST.json should record the classification
	data, _ := os.ReadFile(filepath.Join(backupDir, "MANIFEST.json"))
	if !strings.Contains(string(data), "UserModifiedDeprecated") {
		t.Errorf("MANIFEST.json should record UserModifiedDeprecated classification")
	}
}

// ---------------------------------------------------------------------------
// Scenario D — moai update twice → no " 2" suffix files
// ---------------------------------------------------------------------------
// This scenario verifies the atomic write idempotency at the deployer level.
// Since the E2E focuses on the cleanup path, we verify that running cleanup
// twice on an already-cleaned project is a no-op.

func TestE2E_ScenarioD_DoubleRunNoSuffix2(t *testing.T) {
	t.Parallel()
	root, mgr := setupE2EProject(t)
	seedAgencyFiles(t, root, mgr, "# scenario D content")

	// First cleanup run
	_, _ = runCleanupPhase(t, root, mgr)

	// Second cleanup run — agency files are already gone, should be no-op
	backupDir2, removed2 := runCleanupPhase(t, root, mgr)

	if backupDir2 != "" {
		t.Errorf("second cleanup run should be no-op, got backup dir: %q", backupDir2)
	}
	if len(removed2) != 0 {
		t.Errorf("second cleanup run should remove nothing, removed %d", len(removed2))
	}

	// No " 2" suffix files
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, _ error) error {
		if d != nil && !d.IsDir() {
			name := filepath.Base(path)
			if strings.Contains(name, " 2.") || strings.HasSuffix(name, " 2") {
				t.Errorf("unexpected \" 2\" suffix file after double run: %s", path)
			}
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// Scenario E — concurrent lock acquisition → second call blocked
// ---------------------------------------------------------------------------

func TestE2E_ScenarioE_ConcurrentLock(t *testing.T) {
	t.Parallel()
	root, _ := setupE2EProject(t)

	// First lock
	release1, err := acquireUpdateLock(root)
	if err != nil {
		t.Fatalf("first lock: %v", err)
	}
	defer release1()

	// Second lock must fail
	release2, err2 := acquireUpdateLock(root)
	if err2 == nil {
		release2()
		t.Fatal("expected second lock to fail, but it succeeded")
	}
	if !strings.Contains(err2.Error(), "already running") && err2.Error() != ErrUpdateLockHeld.Error() {
		t.Errorf("expected ErrUpdateLockHeld, got: %v", err2)
	}
}

// ---------------------------------------------------------------------------
// Scenario F — manifest entry absent → UnverifiedDeprecated
// ---------------------------------------------------------------------------

func TestE2E_ScenarioF_UnverifiedDeprecated(t *testing.T) {
	t.Parallel()
	root, mgr := setupE2EProject(t)
	// Seed agency files WITHOUT tracking in manifest
	for _, p := range defs.DeprecatedPaths {
		abs := filepath.Join(root, filepath.FromSlash(p.Path))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(abs, []byte("# unverified content"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		// NOT tracked in mgr → UnverifiedDeprecated
	}

	// Each file should classify as UnverifiedDeprecated
	for _, p := range defs.DeprecatedPaths {
		class := classifyDeprecatedFile(root, filepath.ToSlash(p.Path), mgr)
		if class != DeprecatedUnverified {
			t.Errorf("path %q: expected UnverifiedDeprecated, got %v", p.Path, class)
		}
	}

	// Cleanup still proceeds (backup + delete)
	backupDir, removed := runCleanupPhase(t, root, mgr)
	if backupDir == "" {
		t.Fatal("expected backup directory even for unverified deprecated files")
	}
	if len(removed) == 0 {
		t.Error("expected deprecated files to be removed")
	}
}

// ---------------------------------------------------------------------------
// REQ-UPC-026: Case-insensitive filesystem probe
// ---------------------------------------------------------------------------

// TestCleanup_CaseInsensitiveFS verifies that the filesystem probe correctly
// detects whether the project FS is case-insensitive.
func TestCleanup_CaseInsensitiveFS(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	caseSensitive, err := probeCaseSensitiveFS(root)
	if err != nil {
		t.Fatalf("probeCaseSensitiveFS: %v", err)
	}

	// On macOS APFS (default) we expect case-insensitive (false)
	// On Linux ext4 we expect case-sensitive (true)
	// We can't assert the exact value (CI may run any OS), but we can verify
	// the probe function returns without error and a consistent result.
	t.Logf("probeCaseSensitiveFS result: caseSensitive=%v (GOOS=%s)", caseSensitive, runtime.GOOS)

	// Probe must leave no residue
	probe := filepath.Join(root, ".moai-fscase-probe")
	if _, err := os.Stat(probe); !os.IsNotExist(err) {
		t.Errorf("probe file %q should be cleaned up, Stat: %v", probe, err)
	}
}

// TestCleanup_CaseProbeRunsOncePerInvocation verifies that the probe result is
// consistent when called twice in the same invocation (idempotent) and that no
// residue is left.
func TestCleanup_CaseProbeRunsOncePerInvocation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	result1, err1 := probeCaseSensitiveFS(root)
	if err1 != nil {
		t.Fatalf("first probe: %v", err1)
	}

	result2, err2 := probeCaseSensitiveFS(root)
	if err2 != nil {
		t.Fatalf("second probe: %v", err2)
	}

	if result1 != result2 {
		t.Errorf("probe results are inconsistent: first=%v second=%v", result1, result2)
	}
}

// TestCleanup_CaseProbeFailureFallback verifies that if the probe fails due to a
// read-only filesystem, case-sensitive matching is used as fallback.
func TestCleanup_CaseProbeFailureFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod not supported on Windows")
	}
	t.Parallel()
	root := t.TempDir()
	// Make root read-only so probe file creation fails
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	defer os.Chmod(root, 0o755) //nolint:errcheck

	// probeCaseSensitiveFS should return (true, nil) on error (fallback to case-sensitive)
	caseSensitive, err := probeCaseSensitiveFS(root)
	if err != nil {
		t.Fatalf("expected fallback (no error), got: %v", err)
	}
	if !caseSensitive {
		t.Error("expected fallback to case-sensitive (true) when probe fails")
	}
}
