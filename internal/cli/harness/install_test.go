// Package harness CLI install-command tests.
// SPEC: SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001 REQ-HAW-001..005.
//
// These tests exercise the `moai harness install` subcommand that wires the
// previously-orphaned InjectMarker (layer3) + ScaffoldHarnessDir (layer5)
// installers into a live CLI call path. The command:
//   - scaffolds .moai/harness/ (emitting main.md), and
//   - injects the CLAUDE.md harness marker block,
// using the generating SPEC ID + project domain. It never prompts the user
// (subagent boundary, covered by the package-level TestPropose_NoAskUserQuestion
// guard in propose_boundary_test.go).
package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeHarnessFile is a test helper that creates a file with parent dirs.
func writeHarnessFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// TestRunInstall_ScaffoldsAndInjectsMarker is the GREEN-path test for
// REQ-HAW-001 (marker install) + REQ-HAW-005 (main.md scaffold). After
// install, .moai/harness/main.md must exist and CLAUDE.md must carry exactly
// one paired marker block.
func TestRunInstall_ScaffoldsAndInjectsMarker(t *testing.T) {
	root := t.TempDir()
	// CLAUDE.md must exist for InjectMarker (it reads the file first).
	writeHarnessFile(t, filepath.Join(root, "CLAUDE.md"), "# Project\n")

	opts := InstallOptions{
		ProjectRoot: root,
		SpecID:      "SPEC-PROJ-INIT-001",
		Domain:      "ios-mobile",
	}
	if err := RunInstall(opts); err != nil {
		t.Fatalf("RunInstall: %v", err)
	}

	// REQ-HAW-005: main.md scaffolded.
	mainMDPath := filepath.Join(root, ".moai", "harness", "main.md")
	if _, err := os.Stat(mainMDPath); err != nil {
		t.Errorf("main.md not scaffolded: %v", err)
	}

	// REQ-HAW-001/002: CLAUDE.md has exactly one paired marker block.
	data, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(data)
	if c := strings.Count(content, "<!-- moai:harness-start"); c != 1 {
		t.Errorf("start marker count = %d, want 1", c)
	}
	if c := strings.Count(content, "<!-- moai:harness-end -->"); c != 1 {
		t.Errorf("end marker count = %d, want 1", c)
	}
	if !strings.Contains(content, "ios-mobile") {
		t.Errorf("domain not present in marker block")
	}
	// The marker block @import path should reference the scaffolded main.md.
	if !strings.Contains(content, "@.moai/harness/main.md") {
		t.Errorf("marker block missing @.moai/harness/main.md import")
	}
}

// TestRunInstall_Idempotent verifies REQ-HAW-002: re-running install replaces
// the existing marker block rather than appending a duplicate.
func TestRunInstall_Idempotent(t *testing.T) {
	root := t.TempDir()
	writeHarnessFile(t, filepath.Join(root, "CLAUDE.md"), "# Project\n")

	opts := InstallOptions{ProjectRoot: root, SpecID: "SPEC-A", Domain: "d1"}
	if err := RunInstall(opts); err != nil {
		t.Fatalf("first install: %v", err)
	}
	opts2 := InstallOptions{ProjectRoot: root, SpecID: "SPEC-B", Domain: "d2"}
	if err := RunInstall(opts2); err != nil {
		t.Fatalf("second install: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	content := string(data)
	if c := strings.Count(content, "<!-- moai:harness-start"); c != 1 {
		t.Errorf("start marker count after double install = %d, want 1", c)
	}
	if c := strings.Count(content, "## Project-Specific Configuration (Harness-Generated)"); c != 1 {
		t.Errorf("heading count after double install = %d, want 1", c)
	}
	// Second install's domain should win.
	if !strings.Contains(content, "d2") {
		t.Errorf("second-install domain d2 not reflected")
	}
}

// TestRunInstall_MissingClaudeMd verifies REQ-HAW-004: when CLAUDE.md cannot be
// read, install returns a non-nil wrapped error and does NOT report success.
func TestRunInstall_MissingClaudeMd(t *testing.T) {
	root := t.TempDir()
	// No CLAUDE.md created — InjectMarker's os.ReadFile must fail.
	opts := InstallOptions{ProjectRoot: root, SpecID: "SPEC-A", Domain: "d1"}
	err := RunInstall(opts)
	if err == nil {
		t.Fatal("expected error when CLAUDE.md is absent, got nil")
	}
	if !strings.Contains(err.Error(), "CLAUDE.md") {
		t.Errorf("error should mention CLAUDE.md, got: %v", err)
	}
}

// TestRunInstall_EmptySpecID verifies the install path surfaces InjectMarker's
// empty-specID guard as a structured error.
func TestRunInstall_EmptySpecID(t *testing.T) {
	root := t.TempDir()
	writeHarnessFile(t, filepath.Join(root, "CLAUDE.md"), "# Project\n")
	opts := InstallOptions{ProjectRoot: root, SpecID: "", Domain: "d1"}
	if err := RunInstall(opts); err == nil {
		t.Fatal("expected error for empty specID, got nil")
	}
}

// TestRunInstall_EmptyProjectRoot verifies REQ-HCC-013: RunInstall with an empty
// ProjectRoot returns the empty-root error (install.go:61-63) before touching the
// filesystem. This closes the RunInstall empty-root guard branch.
func TestRunInstall_EmptyProjectRoot(t *testing.T) {
	err := RunInstall(InstallOptions{ProjectRoot: "", SpecID: "SPEC-A", Domain: "d1"})
	if err == nil {
		t.Fatal("RunInstall must return an error for an empty project root")
	}
	if !strings.Contains(err.Error(), "empty project root") {
		t.Errorf("error should mention empty project root, got: %v", err)
	}
}

// TestRunInstall_ScaffoldFails_PreexistingFile verifies REQ-HCC-014: when the
// .moai/harness directory path collides with a pre-existing regular file,
// ScaffoldHarnessDir's MkdirAll fails and RunInstall surfaces the wrapped error
// (install.go:74-76). The wrapped-error prefix is asserted rather than the OS
// message body (EC-2 — message text varies by platform).
func TestRunInstall_ScaffoldFails_PreexistingFile(t *testing.T) {
	root := t.TempDir()
	// Create .moai/harness as a regular FILE so MkdirAll on that path fails.
	writeHarnessFile(t, filepath.Join(root, ".moai", "harness"), "not a directory\n")

	err := RunInstall(InstallOptions{ProjectRoot: root, SpecID: "SPEC-A", Domain: "d1"})
	if err == nil {
		t.Fatal("RunInstall must return an error when .moai/harness is a regular file")
	}
	if !strings.Contains(err.Error(), "scaffold .moai/harness") {
		t.Errorf("error should carry the scaffold wrapping prefix, got: %v", err)
	}
}

// TestNewInstallCmd_FlagsRegistered verifies the cobra factory wires the
// required flags (--spec-id, --domain, --project-root).
func TestNewInstallCmd_FlagsRegistered(t *testing.T) {
	cmd := NewInstallCmd()
	if cmd.Use != "install" {
		t.Errorf("Use = %q, want install", cmd.Use)
	}
	for _, name := range []string{"spec-id", "domain", "project-root"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag --%s not registered", name)
		}
	}
}
