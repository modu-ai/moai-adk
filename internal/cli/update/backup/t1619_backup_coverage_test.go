package backup

// t1619_backup_coverage_test.go — card t1619 bundle A: changed-line closure for
// the run-marker, run-grouping, and retention-ordering paths of
// internal/cli/update/backup. Each test builds real backup directories and
// marker files, runs the production function, and asserts what is kept and
// what is removed. No production code is changed.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// t1619ConfigRoot is the config backup root of a project.
func t1619ConfigRoot(projectRoot string) string {
	return filepath.Join(projectRoot, defs.BackupsDir)
}

// t1619NamespaceRoot is the namespace backup root of a project.
func t1619NamespaceRoot(projectRoot string) string {
	return filepath.Join(projectRoot, defs.MoAIDir, defs.NamespaceBackupsSubdir)
}

// t1619MkBackupDir creates a backup directory and, when runID is not empty,
// writes its run marker.
func t1619MkBackupDir(t *testing.T, dir, runID string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if runID != "" {
		if err := WriteRunMarker(dir, runID); err != nil {
			t.Fatalf("write run marker in %s: %v", dir, err)
		}
	}
}

// t1619Exists reports whether path exists, without following a symlink.
func t1619Exists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Lstat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	t.Fatalf("lstat %s: %v", path, err)
	return false
}

// TestCleanupLeavesUnmarkedNonBackupNamesAlone covers the unmarked-name guard:
// a directory no backup writer names is neither counted nor deleted, while a
// stamped unmarked directory is one run.
func TestCleanupLeavesUnmarkedNonBackupNamesAlone(t *testing.T) {
	proj := t.TempDir()
	stamped := filepath.Join(t1619ConfigRoot(proj), "20260101_000000")
	notes := filepath.Join(t1619ConfigRoot(proj), "hand-made-notes")
	t1619MkBackupDir(t, stamped, "")
	t1619MkBackupDir(t, notes, "")

	if got := CleanupOldBackups(proj, 0); got != 1 {
		t.Fatalf("CleanupOldBackups(keep 0) removed %d directories, want 1 (the stamped run)", got)
	}
	if t1619Exists(t, stamped) {
		t.Fatal("the stamped unmarked run survived keep 0")
	}
	if !t1619Exists(t, notes) {
		t.Fatal("a directory no backup writer names was removed")
	}
}

// TestCleanupKeepsAndRemovesWholeMarkedRuns covers run grouping by run id. Two
// directories that carry one run id, across both backup roots, form one run:
// keeping two runs keeps both, and keeping one run removes both together.
func TestCleanupKeepsAndRemovesWholeMarkedRuns(t *testing.T) {
	proj := t.TempDir()
	const runA = "2026-01-01T00:00:00.000000000Z"
	const runB = "2026-01-02T00:00:00.000000000Z"
	aConfig := filepath.Join(t1619ConfigRoot(proj), "20260101_000000")
	aNamespace := filepath.Join(t1619NamespaceRoot(proj), "update-2026-01-01T00-00-00Z")
	bConfig := filepath.Join(t1619ConfigRoot(proj), "20260102_000000")
	t1619MkBackupDir(t, aConfig, runA)
	t1619MkBackupDir(t, aNamespace, runA)
	t1619MkBackupDir(t, bConfig, runB)

	if got := CleanupOldBackups(proj, 2); got != 0 {
		t.Fatalf("CleanupOldBackups(keep 2) removed %d directories, want 0 (two runs, both kept)", got)
	}
	if got := CleanupOldBackups(proj, 1); got != 2 {
		t.Fatalf("CleanupOldBackups(keep 1) removed %d directories, want 2 (every directory of the older run)", got)
	}
	if t1619Exists(t, aConfig) || t1619Exists(t, aNamespace) {
		t.Fatal("a directory of the older run survived keep 1")
	}
	if !t1619Exists(t, bConfig) {
		t.Fatal("the newest run was removed by keep 1")
	}
}

// TestCleanupBreaksEqualRunTimesByPath covers the tie-break between runs that
// share a modification time: the lexically first path is the newer run, so it
// survives keep 1.
func TestCleanupBreaksEqualRunTimesByPath(t *testing.T) {
	proj := t.TempDir()
	first := filepath.Join(t1619ConfigRoot(proj), "20260103_000000")
	second := filepath.Join(t1619ConfigRoot(proj), "20260104_000000")
	t1619MkBackupDir(t, first, "")
	t1619MkBackupDir(t, second, "")
	when := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	for _, d := range []string{first, second} {
		if err := os.Chtimes(d, when, when); err != nil {
			t.Fatalf("set mtime of %s: %v", d, err)
		}
	}

	if got := CleanupOldBackups(proj, 1); got != 1 {
		t.Fatalf("CleanupOldBackups(keep 1) removed %d directories, want 1", got)
	}
	if !t1619Exists(t, first) {
		t.Fatal("the lexically first run, which the tie-break ranks newest, was removed")
	}
	if t1619Exists(t, second) {
		t.Fatal("the lexically later run survived keep 1")
	}
}

// TestCleanupIgnoresFilesDirectlyUnderBackupRoot covers the non-directory guard
// of the root listing: a regular file under a backup root is not a run and is
// never removed.
func TestCleanupIgnoresFilesDirectlyUnderBackupRoot(t *testing.T) {
	proj := t.TempDir()
	stray := filepath.Join(t1619ConfigRoot(proj), "stray.txt")
	t1619MkBackupDir(t, filepath.Join(t1619ConfigRoot(proj), "20260101_000000"), "")
	if err := os.WriteFile(stray, []byte("not a run\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := CleanupOldBackups(proj, 0); got != 1 {
		t.Fatalf("CleanupOldBackups(keep 0) removed %d entries, want 1 (the stamped directory only)", got)
	}
	if !t1619Exists(t, stray) {
		t.Fatal("a regular file under the backup root was removed")
	}
}

// TestNewRunIDIsCurrentUTCInRunLayout covers the run-id constructor: the id is
// the current UTC instant in the fixed-width run layout.
func TestNewRunIDIsCurrentUTCInRunLayout(t *testing.T) {
	before := time.Now().UTC()
	id := NewRunID()
	after := time.Now().UTC()

	got, err := time.Parse(runIDLayout, id)
	if err != nil {
		t.Fatalf("NewRunID %q does not parse in the run layout: %v", id, err)
	}
	if got.Before(before) || got.After(after) {
		t.Fatalf("NewRunID %q = %v, want an instant between %v and %v", id, got, before, after)
	}
}

// TestWriteRunMarkerRecordsRunIDLine covers the marker writer: the marker holds
// the run id on one line, and the marker is the file the reader consumes.
func TestWriteRunMarkerRecordsRunIDLine(t *testing.T) {
	dir := t.TempDir()
	const id = "2026-10-10T12:00:00.000000000Z"
	if err := WriteRunMarker(dir, id); err != nil {
		t.Fatalf("WriteRunMarker: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, RunMarkerName))
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if string(data) != id+"\n" {
		t.Fatalf("marker content = %q, want %q", data, id+"\n")
	}
}

// TestReadRunIDParsesValidMarkerAndRejectsMalformed covers the marker reader on
// both sides of the parse: a marker that holds a run id yields its instant, and
// a marker whose first line is not a run id yields no run id.
func TestReadRunIDParsesValidMarkerAndRejectsMalformed(t *testing.T) {
	valid := t.TempDir()
	if err := WriteRunMarker(valid, "2026-10-10T12:00:00.000000000Z"); err != nil {
		t.Fatalf("WriteRunMarker: %v", err)
	}
	at, ok := readRunID(valid)
	if !ok || !at.Equal(time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("readRunID(valid) = (%v, %v), want (2026-10-10T12:00:00Z, true)", at, ok)
	}

	malformed := t.TempDir()
	if err := os.WriteFile(filepath.Join(malformed, RunMarkerName), []byte("not-a-run-id\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := readRunID(malformed); ok {
		t.Fatal("readRunID accepted a marker whose first line is not a run id")
	}
}

// TestCleanupTreatsMalformedMarkerDirectoryAsUnmarked covers the fallback of
// run grouping: a backup-named directory whose marker does not parse is an
// unmarked run, so the name rule applies and it is removed by keep 0.
func TestCleanupTreatsMalformedMarkerDirectoryAsUnmarked(t *testing.T) {
	proj := t.TempDir()
	dir := filepath.Join(t1619ConfigRoot(proj), "20260106_000000")
	t1619MkBackupDir(t, dir, "")
	if err := os.WriteFile(filepath.Join(dir, RunMarkerName), []byte("garbage\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := CleanupOldBackups(proj, 0); got != 1 {
		t.Fatalf("CleanupOldBackups(keep 0) removed %d directories, want 1", got)
	}
	if t1619Exists(t, dir) {
		t.Fatal("the malformed-marker run survived keep 0")
	}
}

// TestRestoreRetainedSkipsRunMarkerInsideSections covers the retained restore's
// marker skip. A run marker inside the backup's sections tree is bookkeeping,
// never a user file: the real section restores, and the marker does not land
// in the config tree.
func TestRestoreRetainedSkipsRunMarkerInsideSections(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(root, defs.BackupsDir, "test")
	sections := filepath.Join(backupDir, "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "kept.yaml"), []byte("kept: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, RunMarkerName), []byte("2026-10-10T12:00:00.000000000Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RestoreMoaiConfigRetained(root, backupDir, nil); err != nil {
		t.Fatalf("RestoreMoaiConfigRetained: %v", err)
	}

	target := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir, "sections")
	if data, err := os.ReadFile(filepath.Join(target, "kept.yaml")); err != nil || string(data) != "kept: true\n" {
		t.Fatalf("the real section was not restored: %q (%v)", data, err)
	}
	if t1619Exists(t, filepath.Join(target, RunMarkerName)) {
		t.Fatal("the run marker was copied into the config tree")
	}
}
