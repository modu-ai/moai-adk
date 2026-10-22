package astgrep_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// TestMain also supplies a native sg subprocess. Shell scripts are not native
// executables on Windows and can accidentally resolve a real sg later on PATH.
func TestMain(m *testing.M) {
	if mode := os.Getenv("MOAI_ASTGREP_FAILURE_HELPER"); mode != "" {
		if len(os.Args) < 2 || os.Args[1] != "scan" {
			os.Exit(99)
		}
		if mode == "usage" {
			fmt.Fprintln(os.Stderr, "error: unexpected argument '--x' found")
			os.Exit(2)
		}
		os.Exit(3)
	}
	os.Exit(m.Run())
}

// stubSGOnPath copies this test executable under the scanner's allowed bare
// name, so PATH resolution, native process execution and exit status are real.
func stubSGOnPath(t *testing.T, mode string) {
	t.Helper()
	binDir := t.TempDir()
	name := "sg"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.Open(exe)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := src.Close(); err != nil {
			t.Errorf("close native sg source: %v", err)
		}
	}()
	dst, err := os.OpenFile(filepath.Join(binDir, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil || closeErr != nil {
		t.Fatalf("copy native sg: %v; close: %v", copyErr, closeErr)
	}
	t.Setenv("MOAI_ASTGREP_FAILURE_HELPER", mode)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	resolved, err := exec.LookPath("sg")
	if err != nil || resolved != filepath.Join(binDir, name) {
		t.Fatalf("native sg fixture resolved to %q: %v", resolved, err)
	}
}

// sgConfigRulesDir writes a rules directory with an sgconfig.yml, so the scanner
// takes the config-based path that runs sg scan --config.
func sgConfigRulesDir(t *testing.T) string {
	t.Helper()
	rulesDir := filepath.Join(t.TempDir(), "rules")
	if err := os.MkdirAll(filepath.Join(rulesDir, "go"), 0o755); err != nil {
		t.Fatalf("failed to create rules directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs:\n  - go\n"), 0o644); err != nil {
		t.Fatalf("failed to write sgconfig.yml: %v", err)
	}
	return rulesDir
}

// TestScanner_UsageErrorIsNotARulesetLoad: sg exits non-zero with empty stdout
// and a usage error on stderr. That must fail loud, and it must not be reported
// as a ruleset load failure.
func TestScanner_UsageErrorIsNotARulesetLoad(t *testing.T) {
	stubSGOnPath(t, "usage")
	s := astgrep.NewScanner(&astgrep.ScannerConfig{RulesDir: sgConfigRulesDir(t), SGBinary: "sg"})

	_, err := s.Scan(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("Scan() error = nil; a usage error from sg must fail loud, not read as a clean scan")
	}
	if errors.Is(err, astgrep.ErrRulesetLoad) {
		t.Errorf("Scan() error = %q; a usage error must not be reported as a ruleset load failure", err)
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 {
		t.Fatalf("usage helper exit = %v, want native exit 2", err)
	}
}

// TestScanner_SilentNonZeroExitIsAnError: a non-zero exit with no output at all
// (a killed or crashed child) is not a clean scan.
func TestScanner_SilentNonZeroExitIsAnError(t *testing.T) {
	stubSGOnPath(t, "silent")
	s := astgrep.NewScanner(&astgrep.ScannerConfig{RulesDir: sgConfigRulesDir(t), SGBinary: "sg"})

	findings, err := s.Scan(context.Background(), t.TempDir())
	if err == nil {
		t.Fatalf("Scan() error = nil with %d finding(s); a non-zero exit with no output must fail loud", len(findings))
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("silent helper exit = %v, want native exit 3", err)
	}
}

// TestScanner_MissingDashPrefixedTargetFailsLoud: a missing scan target whose
// name starts with "-" must fail loud. Without a target check, sg reads the name
// as a flag, fails to parse its arguments, and the failure surfaces as a ruleset
// load failure.
func TestScanner_MissingDashPrefixedTargetFailsLoud(t *testing.T) {
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg CLI is not in PATH — the usage-error path needs a real sg")
	}
	s := astgrep.NewScanner(&astgrep.ScannerConfig{RulesDir: sgConfigRulesDir(t), SGBinary: "sg"})

	_, err := s.Scan(context.Background(), "--no-such-target.go")
	if err == nil {
		t.Fatal("Scan() error = nil for a missing dash-prefixed target; it must fail loud")
	}
	if errors.Is(err, astgrep.ErrRulesetLoad) {
		t.Errorf("Scan() error = %q; a missing target must not be reported as a ruleset load failure", err)
	}
}
