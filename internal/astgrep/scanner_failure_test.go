package astgrep_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// stubSGOnPath installs an executable named "sg" that runs script and puts its
// directory first on PATH for the test. The scanner resolves the bare name "sg"
// through PATH, so the stub stands in for the real binary.
func stubSGOnPath(t *testing.T, script string) {
	t.Helper()
	binDir := t.TempDir()
	stub := filepath.Join(binDir, "sg")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatalf("failed to write stub sg: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
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
	stubSGOnPath(t, `echo "error: unexpected argument '--x' found" >&2
exit 2`)
	s := astgrep.NewScanner(&astgrep.ScannerConfig{RulesDir: sgConfigRulesDir(t), SGBinary: "sg"})

	_, err := s.Scan(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("Scan() error = nil; a usage error from sg must fail loud, not read as a clean scan")
	}
	if errors.Is(err, astgrep.ErrRulesetLoad) {
		t.Errorf("Scan() error = %q; a usage error must not be reported as a ruleset load failure", err)
	}
}

// TestScanner_SilentNonZeroExitIsAnError: a non-zero exit with no output at all
// (a killed or crashed child) is not a clean scan.
func TestScanner_SilentNonZeroExitIsAnError(t *testing.T) {
	stubSGOnPath(t, `exit 3`)
	s := astgrep.NewScanner(&astgrep.ScannerConfig{RulesDir: sgConfigRulesDir(t), SGBinary: "sg"})

	findings, err := s.Scan(context.Background(), t.TempDir())
	if err == nil {
		t.Fatalf("Scan() error = nil with %d finding(s); a non-zero exit with no output must fail loud", len(findings))
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
