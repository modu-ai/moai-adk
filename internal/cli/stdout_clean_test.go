package cli

// M4e stdout-cleanliness tests (SPEC-CLI-TUX-V3-004 REQ-TUX4-010,
// AC-TUX4-011): the polished surfaces (doctor / status / spec view / banner)
// leak no warning/status strings to stdout — stdout carries data only.
// Test names match the AC run pattern 'StdoutClean|NoWarnOnStdout'.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// assertNoWarnStatus fails when the captured stdout carries the Printer
// warning/error prefixes or the BODP reminder marker — the status vocabulary
// that must stay on stderr (internal/cli/CLAUDE.md output-streams contract).
func assertNoWarnStatus(t *testing.T, surface, stdout string) {
	t.Helper()
	for _, marker := range []string{"Warning:", "Error:", "[!]"} {
		if strings.Contains(stdout, marker) {
			t.Errorf("%s stdout must not carry status marker %q (stderr-only), got:\n%s", surface, marker, stdout)
		}
	}
}

// TestStdoutClean_Doctor verifies the doctor stdout result surface carries no
// warning/status strings (the per-check "warn" status CELL is table data, not
// a status message; the "Warning:" prefix vocabulary is what must not leak).
func TestStdoutClean_Doctor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.99.99")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GIT_VERSION_OVERRIDE", "git version 9.99.99")
	t.Setenv("MOAI_GH_VERSION_OVERRIDE", "gh version 9.99.99 (2099-12-31)")
	t.Setenv("MOAI_GOOS_OVERRIDE", "testos")
	t.Setenv("MOAI_GOARCH_OVERRIDE", "testarch")

	stdout, _ := captureDoctorCmd(t)
	assertNoWarnStatus(t, "doctor", stdout)
}

// TestStdoutClean_Status verifies the status stdout markdown surface carries
// no warning/status strings (the BODP off-protocol reminder rides stderr).
func TestStdoutClean_Status(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	pkgDir := chdirTempProject(t, "clean-project", true)
	_ = pkgDir
	statusCmd.SetOut(outBuf)
	statusCmd.SetErr(errBuf)
	if err := statusCmd.RunE(statusCmd, []string{}); err != nil {
		t.Fatalf("statusCmd.RunE: %v", err)
	}
	statusCmd.SetOut(nil)
	statusCmd.SetErr(nil)

	assertNoWarnStatus(t, "status", outBuf.String())
}

// TestStdoutClean_Banner verifies the compact banner stdout carries no
// warning/status strings.
func TestStdoutClean_Banner(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.26.0")

	// Capture package-level stdout: PrintBanner routes through printer.Data
	// which writes to os.Stdout.
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	uikit.PrintBanner("1.2.3")
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()

	assertNoWarnStatus(t, "banner", buf.String())
}

// TestNoWarnOnStdout_SpecView verifies the spec view command path keeps
// stdout data-only end-to-end: parse warnings are wired to stderr
// (spec_view.go), and a warning-free run emits pure data on stdout.
func TestNoWarnOnStdout_SpecView(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	// Temp project with a minimal SPEC (no acceptance criteria → the
	// "No acceptance criteria found" data line on stdout).
	root := t.TempDir()
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-CLEAN-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	specMD := "---\nid: SPEC-CLEAN-001\n---\n\n# SPEC-CLEAN-001\n\n## Acceptance Criteria\n\n(none yet)\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatal(err)
	}

	origFind := findProjectRootFn
	findProjectRootFn = func() (string, error) { return root, nil }
	defer func() { findProjectRootFn = origFind }()

	cmd := newSpecViewCmd()
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"SPEC-CLEAN-001"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("spec view: %v", err)
	}

	assertNoWarnStatus(t, "spec view", outBuf.String())
	if !strings.Contains(outBuf.String(), "SPEC-CLEAN-001") {
		t.Errorf("spec view stdout should carry the data payload, got:\n%s", outBuf.String())
	}
}
