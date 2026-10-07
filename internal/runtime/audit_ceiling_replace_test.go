//go:build darwin || linux

package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAppendProgressRecordWriteDeniedKeepsFile (sync-audit-4 F8) — a
// write-restricted progress.md is never silently rewritten: the replace
// verifies the original is writable first and a denial returns a clean
// error with the file unchanged and no temp left behind — the pre-repair
// os.WriteFile failure mode.
func TestAppendProgressRecordWriteDeniedKeepsFile(t *testing.T) {
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	original := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(original), 0o444); err != nil {
		t.Fatal(err)
	}
	err := appendProgressRecord(specDir, "- new record")
	if err == nil {
		t.Fatal("a write-restricted progress.md was silently rewritten — the pre-repair os.WriteFile returned permission denied")
	}
	raw, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(raw) != original {
		t.Fatalf("the read-only original was modified:\n%s", raw)
	}
	entries, derr := os.ReadDir(specDir)
	if derr != nil {
		t.Fatal(derr)
	}
	if len(entries) != 1 || entries[0].Name() != "progress.md" {
		t.Fatalf("temp files left behind: %v", entries)
	}
}

// TestAppendProgressRecordSeedFailureAborts (sync-audit-4 F8) — a metadata
// seeding failure ABORTS the replace: the original stays untouched, no
// temp survives, and the caller gets an error — never a mode-only fallback
// rewrite that drops the ACL.
func TestAppendProgressRecordSeedFailureAborts(t *testing.T) {
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	original := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	restore := stubSeedFailure(t)
	defer restore()
	err := appendProgressRecord(specDir, "- new record")
	if err == nil {
		t.Fatal("a seeding failure fell back to a mode-only replace instead of aborting")
	}
	raw, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(raw) != original {
		t.Fatalf("the original was modified by an aborted replace:\n%s", raw)
	}
	entries, derr := os.ReadDir(specDir)
	if derr != nil {
		t.Fatal(derr)
	}
	if len(entries) != 1 || entries[0].Name() != "progress.md" {
		t.Fatalf("temp files left behind: %v", entries)
	}
}

// stubSeedFailure swaps the metadata-seeding function for one that always
// fails, restoring it at test end.
func stubSeedFailure(t *testing.T) func() {
	t.Helper()
	orig := seedFileMetadataFn
	seedFileMetadataFn = func(tmp, originalPath string) error {
		return errSeedInjected
	}
	return func() { seedFileMetadataFn = orig }
}

var errSeedInjected = errSeedFailure{}

type errSeedFailure struct{}

func (errSeedFailure) Error() string { return "injected seeding failure" }

var _ = strings.TrimSpace // keep strings linked for sibling tests in this file
