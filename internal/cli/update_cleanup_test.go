// SPEC-V3R3-UPDATE-CLEANUP-001 — Unit tests for M1~M3 cleanup functionality.
// Coverage: atomic write lock (M1), deprecated path detection/backup (M2),
// user confirmation + provenance + telemetry + symlink safety (M3).

package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// ---------------------------------------------------------------------------
// M1: Lock file (concurrent execution prevention)
// ---------------------------------------------------------------------------

// TestUpdate_ConcurrentLock verifies that a second acquireUpdateLock call fails
// while the lock is held by the first call.
func TestUpdate_ConcurrentLock(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	lockPath := filepath.Join(root, defs.MoAIDir, updateLockFile)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// First lock acquisition should succeed.
	release, err := acquireUpdateLock(root)
	if err != nil {
		t.Fatalf("first acquireUpdateLock: %v", err)
	}
	t.Cleanup(release) // ensure released even if test panics

	// Second acquisition should fail with ErrUpdateLockHeld.
	_, err2 := acquireUpdateLock(root)
	if !errors.Is(err2, ErrUpdateLockHeld) {
		t.Errorf("expected ErrUpdateLockHeld, got: %v", err2)
	}

	// After release, a third acquisition should succeed.
	release()
	release3, err3 := acquireUpdateLock(root)
	if err3 != nil {
		t.Fatalf("third acquireUpdateLock after release: %v", err3)
	}
	release3()
}

// TestUpdate_LockFileJSON verifies that the lock file contains valid JSON
// with pid, started_at, and hostname fields (OQ4 spec).
func TestUpdate_LockFileJSON(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, defs.MoAIDir), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	release, err := acquireUpdateLock(root)
	if err != nil {
		t.Fatalf("acquireUpdateLock: %v", err)
	}
	defer release()

	lockPath := filepath.Join(root, defs.MoAIDir, updateLockFile)
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read lock file: %v", err)
	}

	var lockInfo map[string]interface{}
	if err := json.Unmarshal(data, &lockInfo); err != nil {
		t.Fatalf("lock file is not valid JSON: %v\ncontent: %s", err, data)
	}

	for _, field := range []string{"pid", "started_at", "hostname"} {
		if _, ok := lockInfo[field]; !ok {
			t.Errorf("lock file missing field %q", field)
		}
	}
}

// TestUpdate_LockCleanedUpOnRelease verifies the lock file is removed after release.
func TestUpdate_LockCleanedUpOnRelease(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, defs.MoAIDir), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	release, err := acquireUpdateLock(root)
	if err != nil {
		t.Fatalf("acquireUpdateLock: %v", err)
	}
	release()

	lockPath := filepath.Join(root, defs.MoAIDir, updateLockFile)
	if _, statErr := os.Stat(lockPath); !os.IsNotExist(statErr) {
		t.Errorf("lock file should be removed after release, but Stat returned: %v", statErr)
	}
}

// ---------------------------------------------------------------------------
// M2: Deprecated path detection + backup + skip marker
// ---------------------------------------------------------------------------

// setupProjectWithAgency creates a temp project containing the agency deprecated
// files with the given content for each file.
func setupProjectWithAgency(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	for _, p := range defs.DeprecatedPaths {
		fullPath := filepath.Join(root, filepath.FromSlash(p.Path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("MkdirAll %q: %v", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile %q: %v", fullPath, err)
		}
	}
	return root
}

// TestCleanup_AgencyPathsDetected verifies that agency files are detected by
// scanDeprecatedPaths.
func TestCleanup_AgencyPathsDetected(t *testing.T) {
	t.Parallel()
	root := setupProjectWithAgency(t, "# agency content")

	found, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths: %v", err)
	}
	if len(found) == 0 {
		t.Error("expected to find deprecated paths, got none")
	}
	if len(found) != len(defs.DeprecatedPaths) {
		t.Errorf("found %d deprecated paths, want %d", len(found), len(defs.DeprecatedPaths))
	}
}

// TestCleanup_NoOpWhenAbsent verifies that scanDeprecatedPaths returns nothing
// when no agency files are present.
func TestCleanup_NoOpWhenAbsent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	found, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths: %v", err)
	}
	if len(found) != 0 {
		t.Errorf("expected 0 deprecated paths, got %d", len(found))
	}
}

// TestCleanup_DeletedFileSkipped verifies that a path registered in DeprecatedPaths
// but absent from the project is silently skipped (REQ-UPC-025).
func TestCleanup_DeletedFileSkipped(t *testing.T) {
	t.Parallel()
	root := t.TempDir() // empty — no agency files

	found, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths: %v", err)
	}
	// No file = no results (silent skip)
	if len(found) != 0 {
		t.Errorf("expected 0 for absent deprecated paths, got %d: %v", len(found), found)
	}
}

// TestCleanup_BackupCreated verifies that backupDeprecatedPaths creates a backup
// directory with a MANIFEST.json and the original files.
func TestCleanup_BackupCreated(t *testing.T) {
	t.Parallel()
	root := setupProjectWithAgency(t, "# backup test")

	found, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths: %v", err)
	}

	mgr := newManifestManagerForTest(t, root)
	backupDir, err := backupDeprecatedPaths(root, found, mgr)
	if err != nil {
		t.Fatalf("backupDeprecatedPaths: %v", err)
	}

	// MANIFEST.json must exist
	manifestPath := filepath.Join(backupDir, "MANIFEST.json")
	if _, err := os.Stat(manifestPath); err != nil {
		t.Errorf("MANIFEST.json missing in backup dir %q: %v", backupDir, err)
	}

	// At least one backed-up file must exist
	var fileCount int
	_ = filepath.WalkDir(backupDir, func(p string, d os.DirEntry, _ error) error {
		if !d.IsDir() && filepath.Base(p) != "MANIFEST.json" {
			fileCount++
		}
		return nil
	})
	if fileCount == 0 {
		t.Error("backup directory contains no files")
	}
}

// TestCleanup_BackupManifestFields verifies that MANIFEST.json contains the
// required fields per REQ-UPC-011 and REQ-UPC-017.
func TestCleanup_BackupManifestFields(t *testing.T) {
	t.Parallel()
	root := setupProjectWithAgency(t, "# manifest fields test")

	found, _ := scanDeprecatedPaths(root)
	mgr := newManifestManagerForTest(t, root)
	backupDir, err := backupDeprecatedPaths(root, found, mgr)
	if err != nil {
		t.Fatalf("backupDeprecatedPaths: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(backupDir, "MANIFEST.json"))
	if err != nil {
		t.Fatalf("read MANIFEST.json: %v", err)
	}

	var m backupManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal MANIFEST.json: %v", err)
	}

	if m.SpecID == "" {
		t.Error("MANIFEST.json missing spec_id")
	}
	if m.DeletedAt.IsZero() {
		t.Error("MANIFEST.json missing deleted_at")
	}
	if len(m.Files) == 0 {
		t.Error("MANIFEST.json has no file entries")
	}
}

// TestCleanup_SkipMarkerHonored verifies that a directory containing
// .moai-skip-cleanup is skipped (REQ-UPC-018).
func TestCleanup_SkipMarkerHonored(t *testing.T) {
	t.Parallel()
	root := setupProjectWithAgency(t, "# skip marker test")

	// Place .moai-skip-cleanup in the agency commands directory.
	skipDir := filepath.Join(root, ".claude", "commands", "agency")
	markerPath := filepath.Join(skipDir, ".moai-skip-cleanup")
	if err := os.WriteFile(markerPath, []byte{}, 0o644); err != nil {
		t.Fatalf("WriteFile marker: %v", err)
	}

	found, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths: %v", err)
	}

	var out bytes.Buffer
	filtered, skipped := filterSkipMarkerPaths(root, found, &out)

	if len(skipped) == 0 {
		t.Error("expected at least one path skipped due to .moai-skip-cleanup marker")
	}
	// Paths in the skipped directory must not appear in filtered.
	for _, p := range filtered {
		if strings.HasPrefix(filepath.ToSlash(p), ".claude/commands/agency/") {
			t.Errorf("path %q should be filtered out (skip marker present)", p)
		}
	}
	// INFO log must mention the skipped path.
	outStr := out.String()
	if !strings.Contains(outStr, ".moai-skip-cleanup") {
		t.Errorf("expected [INFO] message about skip marker in output, got: %q", outStr)
	}
}

// TestCleanup_SkipMarkerAbsent verifies normal behaviour when no skip marker
// exists (REQ-UPC-018 complement).
func TestCleanup_SkipMarkerAbsent(t *testing.T) {
	t.Parallel()
	root := setupProjectWithAgency(t, "# no skip marker")

	found, _ := scanDeprecatedPaths(root)
	var out bytes.Buffer
	filtered, skipped := filterSkipMarkerPaths(root, found, &out)

	if len(skipped) != 0 {
		t.Errorf("expected 0 skipped paths when no marker, got %d", len(skipped))
	}
	if len(filtered) != len(found) {
		t.Errorf("filtered=%d want %d", len(filtered), len(found))
	}
}

// ---------------------------------------------------------------------------
// M3: Provenance classification
// ---------------------------------------------------------------------------

// TestCleanup_PristineDeprecated verifies that a file matching its deployed hash
// is classified as PristineDeprecated.
func TestCleanup_PristineDeprecated(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	content := []byte("# pristine content")
	relPath := filepath.ToSlash(defs.DeprecatedPaths[0].Path)
	fullPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	mgr := newManifestManagerForTest(t, root)
	hash := manifest.HashBytes(content)
	// Track the file as template-managed (deployed hash matches current content)
	_ = mgr.Track(relPath, manifest.TemplateManaged, hash)

	class := classifyDeprecatedFile(root, relPath, mgr)
	if class != DeprecatedPristine {
		t.Errorf("classifyDeprecatedFile = %v, want DeprecatedPristine", class)
	}
}

// TestCleanup_UserModifiedDeprecated verifies that a file with a different hash
// than its deployed hash is classified as UserModifiedDeprecated.
func TestCleanup_UserModifiedDeprecated(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relPath := filepath.ToSlash(defs.DeprecatedPaths[0].Path)
	fullPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Write "original" content and track it
	original := []byte("# original deployed content")
	if err := os.WriteFile(fullPath, original, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	mgr := newManifestManagerForTest(t, root)
	_ = mgr.Track(relPath, manifest.TemplateManaged, manifest.HashBytes(original))

	// Now modify the file (simulating user edit)
	modified := []byte("# user modified content — different hash")
	if err := os.WriteFile(fullPath, modified, 0o644); err != nil {
		t.Fatalf("WriteFile modified: %v", err)
	}

	class := classifyDeprecatedFile(root, relPath, mgr)
	if class != DeprecatedUserModified {
		t.Errorf("classifyDeprecatedFile = %v, want DeprecatedUserModified", class)
	}
}

// TestCleanup_UnverifiedDeprecated verifies that a file with no manifest entry
// is classified as UnverifiedDeprecated.
func TestCleanup_UnverifiedDeprecated(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relPath := filepath.ToSlash(defs.DeprecatedPaths[0].Path)
	fullPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("# unverified"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Empty manifest (no entries)
	mgr := newManifestManagerForTest(t, root)

	class := classifyDeprecatedFile(root, relPath, mgr)
	if class != DeprecatedUnverified {
		t.Errorf("classifyDeprecatedFile = %v, want DeprecatedUnverified", class)
	}
}

// ---------------------------------------------------------------------------
// M3: Telemetry
// ---------------------------------------------------------------------------

// TestCleanup_TelemetryEmitted verifies that a .moai/logs/update-cleanup-*.jsonl
// file is created with the required fields (REQ-UPC-022).
func TestCleanup_TelemetryEmitted(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	var stderr bytes.Buffer
	event := CleanupEvent{
		AtomicWriteUsed:      true,
		PreUpdateSuffix2Files: []string{},
		BackupOutcome:        "skipped",
		BackupPath:           "",
		CleanupOutcome:       "completed",
		UserOptOutPaths:      []string{},
	}
	if err := emitCleanupTelemetry(root, event, &stderr); err != nil {
		t.Fatalf("emitCleanupTelemetry: %v", err)
	}

	logsDir := filepath.Join(root, defs.MoAIDir, defs.LogsSubdir)
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		t.Fatalf("ReadDir logs: %v", err)
	}

	var jsonlFile string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "update-cleanup-") && strings.HasSuffix(e.Name(), ".jsonl") {
			jsonlFile = filepath.Join(logsDir, e.Name())
			break
		}
	}
	if jsonlFile == "" {
		t.Fatal("expected update-cleanup-*.jsonl file not found in .moai/logs/")
	}

	data, _ := os.ReadFile(jsonlFile)
	var ev map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(data), &ev); err != nil {
		t.Fatalf("jsonl not valid JSON: %v\ncontent: %s", err, data)
	}

	for _, field := range []string{"atomic_write_used", "pre_update_suffix2_files",
		"backup_outcome", "backup_path", "cleanup_outcome", "user_opt_out_paths"} {
		if _, ok := ev[field]; !ok {
			t.Errorf("telemetry missing field %q", field)
		}
	}

	// stderr should also contain the JSON line (mirror)
	if stderr.Len() == 0 {
		t.Error("expected JSON line mirrored to stderr, got empty")
	}

	// user_opt_out_paths field present
	if uop, ok := ev["user_opt_out_paths"]; ok {
		if uop == nil {
			t.Error("user_opt_out_paths should not be null")
		}
	}
}

// TestCleanup_TelemetryPermissionDenied verifies that if the logs directory
// cannot be written, the telemetry is skipped with a WARN but the function
// does not return an error (REQ-UPC-022 permission denied handling).
func TestCleanup_TelemetryPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod 0500 not supported on Windows")
	}
	t.Parallel()
	root := t.TempDir()
	logsDir := filepath.Join(root, defs.MoAIDir, defs.LogsSubdir)
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Make logs dir read-only so file creation fails
	if err := os.Chmod(logsDir, 0o500); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	defer os.Chmod(logsDir, 0o755) //nolint:errcheck // cleanup only

	var stderr bytes.Buffer
	event := CleanupEvent{
		AtomicWriteUsed: true,
		BackupOutcome:   "skipped",
		CleanupOutcome:  "completed",
	}
	// Must NOT return error
	err := emitCleanupTelemetry(root, event, &stderr)
	if err != nil {
		t.Errorf("emitCleanupTelemetry should not fail on permission denied, got: %v", err)
	}

	// WARN must appear in stderr
	if !strings.Contains(stderr.String(), "[WARN]") || !strings.Contains(stderr.String(), "telemetry persistence skipped") {
		t.Errorf("expected [WARN] telemetry persistence skipped in stderr, got: %q", stderr.String())
	}

	// No .jsonl file should exist (skip)
	entries, _ := os.ReadDir(logsDir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jsonl") {
			t.Errorf("unexpected .jsonl file despite permission denied: %s", e.Name())
		}
	}
}

// ---------------------------------------------------------------------------
// M3: Symlink safety
// ---------------------------------------------------------------------------

// TestCleanup_SymlinkNotFollowed verifies that a symlink among deprecated paths
// is not followed — only the link itself is removed and the target is preserved
// (REQ-UPC-023).
func TestCleanup_SymlinkNotFollowed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require elevated privileges on Windows")
	}
	t.Parallel()
	root := t.TempDir()

	// Create an external target file (outside the deprecated tree)
	targetDir := t.TempDir()
	targetFile := filepath.Join(targetDir, "external.md")
	if err := os.WriteFile(targetFile, []byte("external content"), 0o644); err != nil {
		t.Fatalf("WriteFile target: %v", err)
	}

	// Create a symlink at the deprecated path location pointing to the external file
	relLink := filepath.ToSlash(defs.DeprecatedPaths[0].Path)
	linkPath := filepath.Join(root, filepath.FromSlash(relLink))
	if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Symlink(targetFile, linkPath); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	// removeDeprecatedFile should remove only the link, not the target
	if err := removeDeprecatedFile(linkPath); err != nil {
		t.Fatalf("removeDeprecatedFile: %v", err)
	}

	// Link is gone
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Errorf("expected symlink to be removed, Lstat returned: %v", err)
	}

	// Target still exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Errorf("target file should still exist after symlink removal, Stat returned: %v", err)
	}
}

// TestCleanup_SymlinkBroken verifies that a broken symlink (target missing) is
// removed without error and recorded as symlink_target_status: "broken"
// (REQ-UPC-023 broken symlink, v0.2.1).
func TestCleanup_SymlinkBroken(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require elevated privileges on Windows")
	}
	t.Parallel()
	root := t.TempDir()

	// Create a symlink pointing to a non-existent target
	relLink := filepath.ToSlash(defs.DeprecatedPaths[0].Path)
	linkPath := filepath.Join(root, filepath.FromSlash(relLink))
	if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Symlink("/nonexistent/target.md", linkPath); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	info := inspectDeprecatedPath(linkPath)
	if !info.IsSymlink {
		t.Fatal("expected IsSymlink=true for a symlink")
	}
	if info.SymlinkTargetStatus != "broken" {
		t.Errorf("expected symlink_target_status=broken, got %q", info.SymlinkTargetStatus)
	}

	// Removal should succeed
	if err := removeDeprecatedFile(linkPath); err != nil {
		t.Fatalf("removeDeprecatedFile broken symlink: %v", err)
	}
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Errorf("broken symlink should be removed, Lstat: %v", err)
	}
}

// ---------------------------------------------------------------------------
// M3: Backup self-reference and logs self-reference protection
// ---------------------------------------------------------------------------

// TestCleanup_BackupSelfReferenceSkipped verifies that paths under .moai/backup/
// are excluded from deprecated path scanning (REQ-UPC-024).
func TestCleanup_BackupSelfReferenceSkipped(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create a file under .moai/backup/ that has the same name as a deprecated path
	backupPath := filepath.Join(root, defs.MoAIDir, "backup", "agency-test",
		".claude", "commands", "agency", "agency.md")
	if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
		t.Fatalf("MkdirAll backup: %v", err)
	}
	if err := os.WriteFile(backupPath, []byte("# backup content"), 0o644); err != nil {
		t.Fatalf("WriteFile backup: %v", err)
	}

	// Also create the real deprecated path (to prove it IS detected normally)
	relPath := defs.DeprecatedPaths[0].Path
	realPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(realPath), 0o755); err != nil {
		t.Fatalf("MkdirAll real: %v", err)
	}
	if err := os.WriteFile(realPath, []byte("# real agency"), 0o644); err != nil {
		t.Fatalf("WriteFile real: %v", err)
	}

	found, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths: %v", err)
	}

	// None of the found paths should start with .moai/backup/
	for _, p := range found {
		if strings.HasPrefix(filepath.ToSlash(p), ".moai/backup/") {
			t.Errorf("scan should not include .moai/backup/ path: %s", p)
		}
	}
}

// TestCleanup_LogsSelfReferenceSkipped verifies that paths under .moai/logs/
// are excluded from deprecated path scanning (REQ-UPC-024 extended, v0.2.1).
func TestCleanup_LogsSelfReferenceSkipped(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create a file under .moai/logs/ (telemetry output area)
	logsPath := filepath.Join(root, defs.MoAIDir, "logs", "update-cleanup-2026.jsonl")
	if err := os.MkdirAll(filepath.Dir(logsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll logs: %v", err)
	}
	if err := os.WriteFile(logsPath, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("WriteFile logs: %v", err)
	}

	found, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths: %v", err)
	}

	for _, p := range found {
		if strings.HasPrefix(filepath.ToSlash(p), ".moai/logs/") {
			t.Errorf("scan should not include .moai/logs/ path: %s", p)
		}
	}
}

// ---------------------------------------------------------------------------
// Helper: newManifestManagerForTest creates and loads a manifest manager in root.
// ---------------------------------------------------------------------------
func newManifestManagerForTest(t *testing.T, root string) manifest.Manager {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, defs.MoAIDir), 0o755); err != nil {
		t.Fatalf("MkdirAll .moai: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("manifest Load: %v", err)
	}
	return mgr
}

// ---------------------------------------------------------------------------
// Helper: backupManifest mirrors the JSON schema used by backupDeprecatedPaths.
// ---------------------------------------------------------------------------
type backupManifest struct {
	SpecID    string           `json:"spec_id"`
	DeletedAt time.Time        `json:"deleted_at"`
	Files     []backupFileEntry `json:"files"`
}

type backupFileEntry struct {
	Path           string `json:"path"`
	Hash           string `json:"hash"`
	Classification string `json:"classification"`
	SymlinkTarget  string `json:"symlink_target,omitempty"`
}
