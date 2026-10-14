package backup

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// After a retention pass, the excess entries still present are the ones it kept
// because they hold content no retained entry has. OldBackupsKept and
// NamespaceBackupsKept report them as the count and the oldest name, which is the
// one ledger row the update prints in place of one stderr line per kept entry.
func TestBackupsKept_ReportsExcessLeftByTheCap(t *testing.T) {
	root := t.TempDir()
	seedBackup(t, root, "20260101_000000", map[string]string{"a.yml": "shared", "b.yml": "user-only"})
	seedBackup(t, root, "20260102_000000", map[string]string{"a.yml": "shared", "b.yml": "user-only-2"})
	for _, stamp := range []string{"20260103_000000", "20260104_000000", "20260105_000000"} {
		seedBackup(t, root, stamp, map[string]string{"a.yml": "shared", "b.yml": "template-" + stamp})
	}
	base := filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	seedUnder(t, base, "update-2026-01-01T00-00-00Z", map[string]string{"a.yml": "shared", "b.yml": "user-only"})
	for _, stamp := range []string{"update-2026-01-02T00-00-00Z", "update-2026-01-03T00-00-00Z", "update-2026-01-04T00-00-00Z"} {
		seedUnder(t, base, stamp, map[string]string{"a.yml": "shared", "b.yml": "template-" + stamp})
	}

	// keep=3 removes nothing from either root: the oldest entries hold unique content.
	if got := CleanupOldBackups(root, 3); got != 0 {
		t.Fatalf("CleanupOldBackups deleted %d backup(s); want 0", got)
	}
	if got := CleanupNamespaceBackups(root, 3); got != 0 {
		t.Fatalf("CleanupNamespaceBackups deleted %d entr(ies); want 0", got)
	}

	if kept, oldest := OldBackupsKept(root, 3); kept != 2 || oldest != "20260101_000000" {
		t.Fatalf("OldBackupsKept = (%d, %q), want (2, %q)", kept, oldest, "20260101_000000")
	}
	if kept, oldest := NamespaceBackupsKept(root, 3); kept != 1 || oldest != "update-2026-01-01T00-00-00Z" {
		t.Fatalf("NamespaceBackupsKept = (%d, %q), want (1, %q)", kept, oldest, "update-2026-01-01T00-00-00Z")
	}
}
