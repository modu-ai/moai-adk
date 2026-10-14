package backup

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// The kept entries of a retention pass are the excess entries it left in place,
// because they hold a superseded version within the per-file budget. PruneOldBackups
// and PruneNamespaceBackups report them, oldest first.
func TestPruneBackups_ReportsKeptExcess(t *testing.T) {
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

	if got := PruneOldBackups(root, 3, DefaultPastVersionsKeep).Kept; len(got) != 2 || got[0] != "20260101_000000" {
		t.Fatalf("PruneOldBackups kept %v, want two entries with the oldest 20260101_000000", got)
	}
	if got := PruneNamespaceBackups(root, 3, DefaultPastVersionsKeep).Kept; len(got) != 1 || got[0] != "update-2026-01-01T00-00-00Z" {
		t.Fatalf("PruneNamespaceBackups kept %v, want [update-2026-01-01T00-00-00Z]", got)
	}
}
