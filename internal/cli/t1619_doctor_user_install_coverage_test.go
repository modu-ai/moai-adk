package cli

// t1619_doctor_user_install_coverage_test.go — card t1619, bundle B2: the unexecuted
// changed lines of doctor_user_install.go (checkProjectVsLock). Family: TestDoctor.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// TestDoctorUserInstallProjectManifestUnreadableIsWarned — a project manifest that exists
// but cannot be read (a directory where the file belongs) is a warning that names the
// unreadable manifest, never the "nothing to compare" disguise.
func TestDoctorUserInstallProjectManifestUnreadableIsWarned(t *testing.T) {
	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".moai", "manifest.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	check := checkProjectVsLock(projectRoot, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("an unreadable project manifest must report a warning, got status %v (%q)", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "project manifest unreadable") {
		t.Fatalf("the warning must name the unreadable manifest, got %q", check.Message)
	}
}

// TestDoctorUserInstallCorruptProjectManifestDetailShowsOnlyVerbose — a corrupt project
// manifest fails the row and is left at its path; the read-only explanation appears in
// verbose output only.
func TestDoctorUserInstallCorruptProjectManifestDetailShowsOnlyVerbose(t *testing.T) {
	projectRoot := t.TempDir()
	manifestPath := filepath.Join(projectRoot, ".moai", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	const corrupt = `{"files":123`
	if err := os.WriteFile(manifestPath, []byte(corrupt), 0o644); err != nil {
		t.Fatal(err)
	}

	quiet := checkProjectVsLock(projectRoot, false)
	if quiet.Status != uikit.CheckFail {
		t.Fatalf("a corrupt project manifest must fail the row, got status %v (%q)", quiet.Status, quiet.Message)
	}
	if quiet.Detail != "" {
		t.Fatalf("the quiet row must carry no detail, got %q", quiet.Detail)
	}

	verbose := checkProjectVsLock(projectRoot, true)
	if verbose.Status != uikit.CheckFail {
		t.Fatalf("the verbose row must fail the same way, got status %v", verbose.Status)
	}
	if !strings.Contains(verbose.Detail, "read-only") {
		t.Fatalf("the verbose detail must state the read-only handling, got %q", verbose.Detail)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil || string(data) != corrupt {
		t.Fatalf("the corrupt manifest must be left untouched, got %q (%v)", data, err)
	}
}

// TestDoctorUserInstallTrackedCommonAssetFileMatchesLock — a file under a template-recorded
// common-asset root that the lock lists is lock-comparable, and a tree whose only
// comparable file is listed matches the lock file.
func TestDoctorUserInstallTrackedCommonAssetFileMatchesLock(t *testing.T) {
	projectRoot := t.TempDir()
	rel := projectCommonAssetRels[0] + "tracked.md"
	abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte("tracked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(projectRoot, ".moai", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, []byte(`{"files":{"`+rel+`":{}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkProjectVsLock(projectRoot, false)
	if check.Status != uikit.CheckOK || check.Message != "project tree matches the lock file" {
		t.Fatalf("a listed common-asset file must match the lock, got status %v (%q)", check.Status, check.Message)
	}
}
