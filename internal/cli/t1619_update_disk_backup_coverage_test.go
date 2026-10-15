package cli

// t1619_update_disk_backup_coverage_test.go — card t1619, bundle B2: the unexecuted changed
// lines of update_disk_backup.go (runMarkerWarning's language fallback). Family: TestUpdate.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestUpdateRunMarkerWarningUsesEnglishWhenNoLanguageRecorded — a project that records no
// conversation language gets the English warning naming the marker failure.
func TestUpdateRunMarkerWarningUsesEnglishWhenNoLanguageRecorded(t *testing.T) {
	projectRoot := t.TempDir()
	cause := errors.New("marker path blocked")
	got := runMarkerWarning(projectRoot, cause)
	want := fmt.Sprintf(tui.Translate("update.backup.run_marker_failed", "en"), cause)
	if got != want {
		t.Fatalf("a project without a recorded language must get the English warning, got %q want %q", got, want)
	}
}

// TestUpdateRunMarkerWarningUsesRecordedLanguage — a recorded conversation language selects
// the warning in that language.
func TestUpdateRunMarkerWarningUsesRecordedLanguage(t *testing.T) {
	projectRoot := t.TempDir()
	sections := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "language.yaml"), []byte("language:\n  conversation_language: ko\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cause := errors.New("marker path blocked")
	got := runMarkerWarning(projectRoot, cause)
	want := fmt.Sprintf(tui.Translate("update.backup.run_marker_failed", "ko"), cause)
	if got != want {
		t.Fatalf("a recorded language must select the warning in that language, got %q want %q", got, want)
	}
}
