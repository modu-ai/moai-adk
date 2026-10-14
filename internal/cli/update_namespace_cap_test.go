package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// Two namespace backups taken from the same user-owned file differ only in the
// .complete marker, which records each backup's stamp and time. The cap treats the
// older one as redundant, so keep=1 removes it. This runs the real producer, not a
// seeded copy of its output.
func TestNamespaceBackupProducer_CapRemovesRedundantEntry(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "agents", "harness")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hns-test-specialist.md"), []byte("user body\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, err := backupUserOwnedNamespace(root)
	if err != nil || first == "" {
		t.Fatalf("first namespace backup: dir=%q err=%v", first, err)
	}
	// Stamps have one-second resolution; wait so the two entries get distinct names.
	time.Sleep(1100 * time.Millisecond)
	second, err := backupUserOwnedNamespace(root)
	if err != nil || second == "" {
		t.Fatalf("second namespace backup: dir=%q err=%v", second, err)
	}

	if got := backup.CleanupNamespaceBackups(root, 1); got != 1 {
		t.Fatalf("CleanupNamespaceBackups removed %d entr(ies); the older one only repeats the newer, want 1", got)
	}
	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Fatalf("redundant older namespace backup still present: err=%v", err)
	}
	if _, err := os.Stat(second); err != nil {
		t.Fatalf("newer namespace backup removed: %v", err)
	}
}

// Housekeeping reports the namespace entries the cap kept as one ledger row, naming
// the count and the oldest kept entry, instead of one line per entry.
func TestRunUpdateHousekeeping_ReportsKeptNamespaceBackupsAsOneRow(t *testing.T) {
	root := t.TempDir()
	markProject(t, root)
	writeNamespaceBackup(t, root, "update-2026-01-01T00-00-00Z", map[string]string{"b.yml": "user-only"})
	writeNamespaceBackup(t, root, "update-2026-01-02T00-00-00Z", map[string]string{"b.yml": "user-only-2"})
	for i := 3; i <= 7; i++ {
		writeNamespaceBackup(t, root, fmt.Sprintf("update-2026-01-%02dT00-00-00Z", i), map[string]string{"b.yml": fmt.Sprintf("template-%d", i)})
	}

	updateLedger.reset()
	runUpdateHousekeeping(root)

	var rows []string
	for _, row := range updateLedger.reference {
		if strings.Contains(row.msg, "namespace backup") {
			rows = append(rows, row.msg)
		}
	}
	want := "kept 2 older namespace backup(s) holding a superseded version within the per-file budget (oldest: update-2026-01-01T00-00-00Z)"
	if len(rows) != 1 || rows[0] != want {
		t.Fatalf("namespace backup rows = %q, want exactly one row %q", rows, want)
	}
}

// Every namespace backup the pass removes gets its own report row, named, so the
// update shows what was deleted rather than a count.
func TestRunUpdateHousekeeping_ReportsEachRemovedNamespaceBackup(t *testing.T) {
	root := t.TempDir()
	markProject(t, root)
	for i := 1; i <= 7; i++ {
		writeNamespaceBackup(t, root, fmt.Sprintf("update-2026-01-%02dT00-00-00Z", i), map[string]string{"b.yml": "shared"})
	}

	updateLedger.reset()
	runUpdateHousekeeping(root)

	var removed []string
	for _, row := range updateLedger.reference {
		if strings.HasPrefix(row.msg, "removed namespace backup ") {
			removed = append(removed, row.msg)
		}
	}
	if len(removed) != 2 {
		t.Fatalf("removed-namespace-backup rows = %q, want one row for each of the two redundant entries", removed)
	}
	for _, name := range []string{"update-2026-01-01T00-00-00Z", "update-2026-01-02T00-00-00Z"} {
		found := false
		for _, msg := range removed {
			if strings.Contains(msg, name) {
				found = true
			}
		}
		if !found {
			t.Fatalf("no removal row names %s: %q", name, removed)
		}
	}
}
