package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// A user file that changes over time leaves one superseded version per backup.
// The cap keeps the newest superseded versions of each file (a budget of three),
// so an entry whose only unique content is an older version beyond that budget is
// pruned, and an entry holding a version inside the budget stays.
func TestCleanupOldBackups_PrunesVersionsBeyondPerFileBudget(t *testing.T) {
	root := t.TempDir()
	seedBackup(t, root, "20260101_000000", map[string]string{"a.yml": "v1"})
	seedBackup(t, root, "20260102_000000", map[string]string{"a.yml": "v2"})
	seedBackup(t, root, "20260103_000000", map[string]string{"a.yml": "v3"})
	seedBackup(t, root, "20260104_000000", map[string]string{"a.yml": "v4"})
	seedBackup(t, root, "20260105_000000", map[string]string{"a.yml": "v5"})

	// keep=1 retains the newest entry. v1 to v4 are superseded versions; the budget of
	// three keeps v2, v3, and v4, so only the entry holding v1 is pruned.
	if got := CleanupOldBackups(root, 1); got != 1 {
		t.Fatalf("CleanupOldBackups deleted %d backup(s); want 1 (only the entry beyond the per-file budget)", got)
	}
	if _, err := os.Stat(filepath.Join(root, defs.BackupsDir, "20260101_000000")); !os.IsNotExist(err) {
		t.Fatalf("entry beyond the per-file budget still present: err=%v", err)
	}
	for _, stamp := range []string{"20260102_000000", "20260103_000000", "20260104_000000", "20260105_000000"} {
		if _, err := os.Stat(filepath.Join(root, defs.BackupsDir, stamp)); err != nil {
			t.Fatalf("entry %s inside the per-file budget removed: %v", stamp, err)
		}
	}
}

// Template renders and their pre-clean copies are rebuilt from the template, so they
// are not user content. An entry whose only unique file is one of them is pruned.
func TestCleanupOldBackups_TemplateOwnedContentIsNotProtected(t *testing.T) {
	root := t.TempDir()
	seedBackup(t, root, "20260101_000000", map[string]string{
		"a.yml": "shared",
		".template-defaults/sections/quality.yaml": "template-old",
		"pre-clean/sections/quality.yaml":          "pre-clean-old",
	})
	seedBackup(t, root, "20260102_000000", map[string]string{
		"a.yml": "shared",
		".template-defaults/sections/quality.yaml": "template-new",
		"pre-clean/sections/quality.yaml":          "pre-clean-new",
	})

	if got := CleanupOldBackups(root, 1); got != 1 {
		t.Fatalf("CleanupOldBackups deleted %d backup(s); the older one differs only in template-owned content, want 1", got)
	}
}

// The namespace root applies the same per-file budget as the timestamped root.
func TestCleanupNamespaceBackups_PrunesVersionsBeyondPerFileBudget(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	for i, content := range []string{"v1", "v2", "v3", "v4", "v5"} {
		seedUnder(t, base, fmt.Sprintf("update-2026-01-%02dT00-00-00Z", i+1), map[string]string{"a.yml": content})
	}

	if got := CleanupNamespaceBackups(root, 1); got != 1 {
		t.Fatalf("CleanupNamespaceBackups deleted %d entr(ies); want 1 (only the entry beyond the per-file budget)", got)
	}
	if _, err := os.Stat(filepath.Join(base, "update-2026-01-01T00-00-00Z")); !os.IsNotExist(err) {
		t.Fatalf("namespace entry beyond the per-file budget still present: err=%v", err)
	}
}
