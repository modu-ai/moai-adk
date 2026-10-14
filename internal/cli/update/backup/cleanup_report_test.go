package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// The report names each removed entry and each kept entry, oldest first, so the
// update can show what a pass did instead of a bare count.
func TestPruneOldBackups_ReportsEachRemovedAndKeptEntry(t *testing.T) {
	root := t.TempDir()
	for i, stamp := range []string{"20260101_000000", "20260102_000000", "20260103_000000", "20260104_000000"} {
		seedBackup(t, root, stamp, map[string]string{"a.yml": fmt.Sprintf("v%d", i+1)})
	}
	seedBackup(t, root, "20260105_000000", map[string]string{"a.yml": "v5"})

	got := PruneOldBackups(root, 1, DefaultPastVersionsKeep)
	if !equalNames(got.Removed, []string{"20260101_000000"}) {
		t.Fatalf("Removed = %v, want [20260101_000000]", got.Removed)
	}
	if !equalNames(got.Kept, []string{"20260102_000000", "20260103_000000", "20260104_000000"}) {
		t.Fatalf("Kept = %v, want the three entries inside the per-file budget", got.Kept)
	}
}

// The budget is the caller's. With a budget of one only the newest superseded
// version survives, so the three older entries go.
func TestPruneOldBackups_BudgetParameterIsHonored(t *testing.T) {
	root := t.TempDir()
	for i, stamp := range []string{"20260101_000000", "20260102_000000", "20260103_000000", "20260104_000000"} {
		seedBackup(t, root, stamp, map[string]string{"a.yml": fmt.Sprintf("v%d", i+1)})
	}
	seedBackup(t, root, "20260105_000000", map[string]string{"a.yml": "v5"})

	got := PruneOldBackups(root, 1, 1)
	if !equalNames(got.Removed, []string{"20260101_000000", "20260102_000000", "20260103_000000"}) {
		t.Fatalf("Removed = %v, want the three entries beyond a budget of one", got.Removed)
	}
	if !equalNames(got.Kept, []string{"20260104_000000"}) {
		t.Fatalf("Kept = %v, want [20260104_000000]", got.Kept)
	}
}

// Entries under .moai-backups whose names lack the timestamp shape are never pruned,
// and the update report counts them with their size.
func TestUnmanagedEntries_CountsAndSizesTimestamplessEntries(t *testing.T) {
	root := t.TempDir()
	seedBackup(t, root, "20260101_000000", map[string]string{"a.yml": "x"})
	seedBackup(t, root, "sync-2026-10-01-000000", map[string]string{"f.txt": strings.Repeat("a", 100)})
	seedBackup(t, root, "settings-backup", map[string]string{"s.json": strings.Repeat("b", 50)})
	if err := os.WriteFile(filepath.Join(root, defs.BackupsDir, "SPEC-X.md.bak"), []byte(strings.Repeat("c", 30)), 0o644); err != nil {
		t.Fatal(err)
	}

	count, size := UnmanagedEntries(root)
	if count != 3 || size != 180 {
		t.Fatalf("UnmanagedEntries = (%d, %d), want (3, 180)", count, size)
	}
	if got := PruneOldBackups(root, 1, DefaultPastVersionsKeep); len(got.Removed) != 0 {
		t.Fatalf("PruneOldBackups removed %v; timestamp-less entries must never be pruned", got.Removed)
	}
	for _, name := range []string{"sync-2026-10-01-000000", "settings-backup", "SPEC-X.md.bak"} {
		if _, err := os.Stat(filepath.Join(root, defs.BackupsDir, name)); err != nil {
			t.Fatalf("timestamp-less entry %s removed: %v", name, err)
		}
	}
}

// The namespace root reports its removals entry by entry as well.
func TestPruneNamespaceBackups_ReportsRemovedEntries(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	for i, content := range []string{"v1", "v2", "v3", "v4", "v5"} {
		seedUnder(t, base, fmt.Sprintf("update-2026-01-%02dT00-00-00Z", i+1), map[string]string{"a.yml": content})
	}

	got := PruneNamespaceBackups(root, 1, DefaultPastVersionsKeep)
	if !equalNames(got.Removed, []string{"update-2026-01-01T00-00-00Z"}) {
		t.Fatalf("Removed = %v, want [update-2026-01-01T00-00-00Z]", got.Removed)
	}
}

// A symbolic-linked timestamped root reports nothing and removes nothing.
func TestPruneOldBackups_RefusedRootReportsNothing(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "backups")
	for i := 1; i <= 4; i++ {
		seedUnder(t, outside, fmt.Sprintf("2026010%d_000000", i), map[string]string{"a.yml": "v"})
	}
	if err := os.Symlink(outside, filepath.Join(root, defs.BackupsDir)); err != nil {
		t.Skip(err)
	}

	got := PruneOldBackups(root, 1, DefaultPastVersionsKeep)
	if len(got.Removed) != 0 || len(got.Kept) != 0 {
		t.Fatalf("refused root reported %+v; want nothing", got)
	}
	if _, err := os.Stat(filepath.Join(outside, "20260101_000000")); err != nil {
		t.Fatalf("entry outside the project removed: %v", err)
	}
}

func equalNames(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
