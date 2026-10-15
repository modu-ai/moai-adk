// recordless_archive_test.go — U-06 (AC-RC30-07, REQ-RC30-008): a record-less
// project that holds a migration archive is stamped local with a visible
// warning, never silently. A record-less project without an archive is stamped
// without one.
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/config"
)

// recordlessProjectRoot seeds a project whose llm.yaml carries no
// deployment_mode record.
func recordlessProjectRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("llm:\n  harness: claude\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestRecordlessArchiveWarning: the stamp of a record-less project warns when a
// migration archive exists, and the record it writes reads local.
func TestRecordlessArchiveWarning(t *testing.T) {
	t.Run("archive_present_warns_and_stamps_local", func(t *testing.T) {
		root := recordlessProjectRoot(t)
		archived := filepath.Join(root, update.ArchiveFilesRoot(), "commands", "old.md")
		if err := os.MkdirAll(filepath.Dir(archived), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(archived, []byte("archived\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		var errOut bytes.Buffer
		stampMigrationDeployMode(root, migrateOptedOut, &errOut)

		if !strings.Contains(errOut.String(), "recorded as local") {
			t.Errorf("record-less stamp over a migration archive printed no visible warning:\n%q", errOut.String())
		}
		if got := config.ReadDeployMode(root); got != "local" {
			t.Errorf("record after the stamp = %q, want local", got)
		}
	})

	t.Run("no_archive_stamps_without_warning", func(t *testing.T) {
		root := recordlessProjectRoot(t)

		var errOut bytes.Buffer
		stampMigrationDeployMode(root, migrateOptedOut, &errOut)

		if strings.Contains(errOut.String(), "recorded as local") {
			t.Errorf("warning printed without a migration archive:\n%q", errOut.String())
		}
		if got := config.ReadDeployMode(root); got != "local" {
			t.Errorf("record after the stamp = %q, want local", got)
		}
	})
}
