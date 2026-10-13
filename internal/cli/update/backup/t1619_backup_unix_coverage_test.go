//go:build unix

// t1619_backup_unix_coverage_test.go — card t1619 bundle A: the unix-only arm of
// the changed-line closure for internal/cli/update/backup. A backup root that
// is readable but not searchable lists its entries, yet stat of each entry
// fails, which is the listing path that skips an entry whose info cannot be
// read. The test runs only as a non-root user, because root ignores the mode
// bits this injection relies on.
package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupSkipsEntriesWhoseInfoCannotBeRead(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores mode bits — the chmod injection cannot fail")
	}
	proj := t.TempDir()
	root := t1619ConfigRoot(proj)
	dir := filepath.Join(root, "20260101_000000")
	t1619MkBackupDir(t, dir, "")

	if err := os.Chmod(root, 0o444); err != nil {
		t.Fatalf("make the backup root readable but not searchable: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })

	if got := CleanupOldBackups(proj, 0); got != 0 {
		t.Fatalf("CleanupOldBackups(keep 0) removed %d directories, want 0 (no entry's info can be read)", got)
	}
	if err := os.Chmod(root, 0o755); err != nil {
		t.Fatalf("restore the backup root mode: %v", err)
	}
	if !t1619Exists(t, dir) {
		t.Fatal("the directory whose info could not be read was removed")
	}
}
