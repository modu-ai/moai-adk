package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// seedBackup writes one timestamped backup entry under .moai-backups.
func seedBackup(t *testing.T, root, stamp string, files map[string]string) {
	t.Helper()
	seedUnder(t, filepath.Join(root, defs.BackupsDir), stamp, files)
}

// seedUnder writes one backup entry named stamp under base holding the given files.
func seedUnder(t *testing.T, base, stamp string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(base, stamp, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// The oldest backup holds one file whose content no retained backup has.
// Removing it would destroy the only copy of that content, so it must stay.
func TestCleanupOldBackups_KeepsOldestWhenItHoldsUniqueContent(t *testing.T) {
	root := t.TempDir()
	oldest := filepath.Join(root, defs.BackupsDir, "20260101_000000")
	seedBackup(t, root, "20260101_000000", map[string]string{
		"a.yml": "shared",
		"b.yml": "user-only-version",
	})
	templates := []string{"template-1", "template-2", "template-3", "template-4", "template-5"}
	for i, stamp := range []string{"20260102_000000", "20260103_000000", "20260104_000000", "20260105_000000", "20260106_000000"} {
		seedBackup(t, root, stamp, map[string]string{
			"a.yml": "shared",
			"b.yml": templates[i],
		})
	}

	if got := CleanupOldBackups(root, 5); got != 0 {
		t.Fatalf("CleanupOldBackups deleted %d backup(s); the oldest holds unique content and must be kept", got)
	}
	data, err := os.ReadFile(filepath.Join(oldest, "b.yml"))
	if err != nil || string(data) != "user-only-version" {
		t.Fatalf("unique copy destroyed: err=%v content=%q", err, data)
	}
	// Idempotent: a second run over the same tree changes nothing.
	if got := CleanupOldBackups(root, 5); got != 0 {
		t.Fatalf("second CleanupOldBackups deleted %d backup(s); want 0", got)
	}
}

// The oldest backup's only file is also present in a retained backup, so
// nothing is lost by removing it: the excess deletion still runs.
func TestCleanupOldBackups_RemovesOldestWhenEveryFileIsRetained(t *testing.T) {
	root := t.TempDir()
	oldest := filepath.Join(root, defs.BackupsDir, "20260101_000000")
	seedBackup(t, root, "20260101_000000", map[string]string{"a.yml": "shared"})
	for _, stamp := range []string{"20260102_000000", "20260103_000000", "20260104_000000", "20260105_000000", "20260106_000000"} {
		seedBackup(t, root, stamp, map[string]string{"a.yml": "shared"})
	}

	if got := CleanupOldBackups(root, 5); got != 1 {
		t.Fatalf("CleanupOldBackups deleted %d backup(s); want 1", got)
	}
	if _, err := os.Stat(oldest); !os.IsNotExist(err) {
		t.Fatalf("redundant oldest backup still present: err=%v", err)
	}
}

// The namespace backup root never had retention. The oldest update-* entry
// holds a user-only file, so the namespace cleanup must keep it, the same
// guard as the .moai-backups root.
func TestCleanupNamespaceBackups_KeepsOldestWhenItHoldsUniqueContent(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	oldest := filepath.Join(base, "update-2026-01-01T00-00-00Z")
	seedUnder(t, base, "update-2026-01-01T00-00-00Z", map[string]string{
		"a.yml": "shared",
		"b.yml": "user-only-version",
	})
	templates := []string{"template-1", "template-2", "template-3", "template-4", "template-5"}
	for i, stamp := range []string{
		"update-2026-01-02T00-00-00Z", "update-2026-01-03T00-00-00Z", "update-2026-01-04T00-00-00Z",
		"update-2026-01-05T00-00-00Z", "update-2026-01-06T00-00-00Z",
	} {
		seedUnder(t, base, stamp, map[string]string{
			"a.yml": "shared",
			"b.yml": templates[i],
		})
	}

	if got := CleanupNamespaceBackups(root, 5); got != 0 {
		t.Fatalf("CleanupNamespaceBackups deleted %d entr(ies); the oldest holds unique content and must be kept", got)
	}
	data, err := os.ReadFile(filepath.Join(oldest, "b.yml"))
	if err != nil || string(data) != "user-only-version" {
		t.Fatalf("unique copy destroyed: err=%v content=%q", err, data)
	}
}

// Every file of the oldest update-* entry is also held at the same path in a
// retained entry, so the namespace cleanup removes it: the cap is enforced.
func TestCleanupNamespaceBackups_RemovesOldestWhenEveryFileIsRetained(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	oldest := filepath.Join(base, "update-2026-01-01T00-00-00Z")
	seedUnder(t, base, "update-2026-01-01T00-00-00Z", map[string]string{"a.yml": "shared"})
	for _, stamp := range []string{
		"update-2026-01-02T00-00-00Z", "update-2026-01-03T00-00-00Z", "update-2026-01-04T00-00-00Z",
		"update-2026-01-05T00-00-00Z", "update-2026-01-06T00-00-00Z",
	} {
		seedUnder(t, base, stamp, map[string]string{"a.yml": "shared"})
	}

	if got := CleanupNamespaceBackups(root, 5); got != 1 {
		t.Fatalf("CleanupNamespaceBackups deleted %d entr(ies); want 1", got)
	}
	if _, err := os.Stat(oldest); !os.IsNotExist(err) {
		t.Fatalf("redundant oldest namespace entry still present: err=%v", err)
	}
}
