package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// Two config backups taken from the same tree hold the same files and differ only
// in backup_metadata.json, which records when each was taken. The older one holds
// nothing the newer one lacks, so keep=1 removes it.
func TestCleanupOldBackups_RemovesOlderConfigBackupWithSameContent(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "quality.yaml"), []byte("constitution: strict\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, err := BackupMoaiConfig(root)
	if err != nil || first == "" {
		t.Fatalf("first backup: dir=%q err=%v", first, err)
	}
	// Backup names carry a one-second stamp; wait so the two names differ.
	time.Sleep(1100 * time.Millisecond)
	second, err := BackupMoaiConfig(root)
	if err != nil || second == "" {
		t.Fatalf("second backup: dir=%q err=%v", second, err)
	}

	if got := CleanupOldBackups(root, 1); got != 1 {
		t.Fatalf("CleanupOldBackups deleted %d backup(s); the older one only repeats the newer, want 1", got)
	}
	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Fatalf("redundant older backup still present: err=%v", err)
	}
	if _, err := os.Stat(second); err != nil {
		t.Fatalf("newer backup removed: %v", err)
	}
}

// Two namespace entries hold the same files and differ only in .complete, the
// completion marker that records each entry's stamp and time. The older one holds
// nothing the newer one lacks, so keep=1 removes it.
func TestCleanupNamespaceBackups_RemovesOlderEntryWhenOnlyCompleteMarkerDiffers(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	older := filepath.Join(base, "update-2026-01-01T00-00-00Z")
	seedUnder(t, base, "update-2026-01-01T00-00-00Z", map[string]string{
		"a.yml":     "shared",
		".complete": "stamp=update-2026-01-01T00-00-00Z\nfiles=1\ntimestamp=2026-01-01T00:00:00Z\n",
	})
	seedUnder(t, base, "update-2026-01-02T00-00-00Z", map[string]string{
		"a.yml":     "shared",
		".complete": "stamp=update-2026-01-02T00-00-00Z\nfiles=1\ntimestamp=2026-01-02T00:00:00Z\n",
	})

	if got := CleanupNamespaceBackups(root, 1); got != 1 {
		t.Fatalf("CleanupNamespaceBackups deleted %d entr(ies); the older one only repeats the newer, want 1", got)
	}
	if _, err := os.Stat(older); !os.IsNotExist(err) {
		t.Fatalf("redundant older namespace entry still present: err=%v", err)
	}
}
