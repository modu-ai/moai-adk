package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// updateStatusGolden controls golden snapshot regeneration. Set via UPDATE_GOLDEN=1.
var updateStatusGolden = os.Getenv("UPDATE_GOLDEN") == "1"

// captureStatusCmdWithPkgDir executes statusCmd in a deterministic temp-dir
// environment and returns (stdout, pkgDir) as strings.
//
// The function must change cwd so that runStatus() reads the right project structure
// via os.Getwd(). Unlike doctor_golden_test.go (which does not Chdir), status
// depends on os.Getwd() for the project name and .moai/ path.
//
// pkgDir is the internal/cli package directory recorded BEFORE the Chdir, so that
// callers can write golden files to the correct testdata/ location.
func captureStatusCmdWithPkgDir(t *testing.T) (string, string) {
	t.Helper()

	// Record the package test root BEFORE any Chdir.
	// go test sets cwd to the package directory, so this is always internal/cli/.
	pkgDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	// Create a deterministic temp project with stable project name.
	// We use a well-known directory name so that the "Project" KV row is stable.
	tmpBase := t.TempDir()
	projectDir := filepath.Join(tmpBase, "my-test-project")
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	// Create .moai/ structure with 2 SPECs and 2 config section files.
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "specs", "SPEC-001"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "specs", "SPEC-002"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".moai", "config", "sections", "user.yaml"),
		[]byte("user:\n  name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".moai", "config", "sections", "quality.yaml"),
		[]byte("constitution:\n  development_mode: ddd\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Chdir into the project dir so runStatus reads the right structure.
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if chErr := os.Chdir(pkgDir); chErr != nil {
			t.Logf("failed to restore working directory: %v", chErr)
		}
	})

	// Suppress BODP off-protocol reminder for stable golden output.
	t.Setenv("MOAI_NO_BODP_REMINDER", "1")

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)
	if err := statusCmd.RunE(statusCmd, []string{}); err != nil {
		t.Fatalf("statusCmd.RunE: %v", err)
	}
	// Reset cobra writers to avoid test pollution.
	statusCmd.SetOut(nil)
	statusCmd.SetErr(nil)
	return buf.String(), pkgDir
}

// statusGoldenPathFromPkg returns the absolute testdata path rooted at pkgDir.
// Used by golden tests that call captureStatusCmd (which Chdirs away from pkgDir).
func statusGoldenPathFromPkg(pkgDir, name string) string {
	return filepath.Join(pkgDir, "testdata", name+".golden")
}

// checkStatusGoldenAbs is like checkStatusGolden but uses an absolute path for
// the golden file, so it works correctly after a Chdir inside captureStatusCmd.
func checkStatusGoldenAbs(t *testing.T, pkgDir, name, got string) {
	t.Helper()
	path := statusGoldenPathFromPkg(pkgDir, name)
	if updateStatusGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		t.Logf("updated golden: %s", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with UPDATE_GOLDEN=1 to generate)", path, err)
	}
	if got != string(want) {
		t.Errorf("status output mismatch for %s\ngot:\n%s\nwant:\n%s", name, got, string(want))
	}
}

// --- DDD PRESERVE: Characterization tests for status command output ---
//
// These tests capture the AFTER state of statusCmd output (tui.Section + tui.KV + tui.Pill).
// They serve as the regression baseline for future DDD cycles.
//
// Deterministic env var pinning:
//   - MOAI_NO_BODP_REMINDER=1  → suppresses git-dependent BODP reminder
//   - MOAI_THEME env            → selects light / dark theme
//   - NO_COLOR env              → selects monochrome mode
//
// lipgloss AdaptiveColor behaviour under cmd.SetOut(buf):
//   - lipgloss detects non-TTY writer and disables ANSI colour output.
//   - Each env combination receives its own golden file for clarity.
//
// To regenerate snapshots:
//
//	UPDATE_GOLDEN=1 go test ./internal/cli/ -run "TestStatus_Current" -count=1

// TestStatus_Current_Light captures statusCmd output with light-theme env.
// 특징: tui.Section + tui.KV + tui.Box + tui.Pill 요약.
func TestStatus_Current_Light(t *testing.T) {
	t.Setenv("NO_COLOR", "0")
	t.Setenv("MOAI_THEME", "light")

	got, pkgDir := captureStatusCmdWithPkgDir(t)
	if len(got) == 0 {
		t.Fatal("statusCmd produced no output")
	}
	checkStatusGoldenAbs(t, pkgDir, "status-light", got)
}

// TestStatus_Current_Dark captures statusCmd output with dark-theme env.
// 특징: tui.DarkTheme() 적용, Section + KV + Box + Pill 요약.
func TestStatus_Current_Dark(t *testing.T) {
	t.Setenv("NO_COLOR", "0")
	t.Setenv("MOAI_THEME", "dark")

	got, pkgDir := captureStatusCmdWithPkgDir(t)
	if len(got) == 0 {
		t.Fatal("statusCmd produced no output")
	}
	checkStatusGoldenAbs(t, pkgDir, "status-dark", got)
}

// TestStatus_NoColor captures statusCmd output with NO_COLOR=1 (plain text mode).
// tui.MonochromeTheme() 적용: 모든 ANSI 색상 제거, Pill은 [label] 형식으로 degraded.
func TestStatus_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got, pkgDir := captureStatusCmdWithPkgDir(t)
	if len(got) == 0 {
		t.Fatal("statusCmd produced no output")
	}
	checkStatusGoldenAbs(t, pkgDir, "status-nocolor", got)
}
