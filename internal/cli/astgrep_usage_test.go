package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli"
)

// TestAstGrepCmd_UsageErrorIsNotRulesetHeadline: a usage error from sg must exit
// non-zero, and its stderr must not carry the ruleset-load headline. Reporting a
// usage error as a load failure is the misreport this test guards against.
func TestAstGrepCmd_UsageErrorIsNotRulesetHeadline(t *testing.T) {
	binDir := t.TempDir()
	stub := filepath.Join(binDir, "sg")
	script := "#!/bin/sh\necho \"error: unexpected argument '--x' found\" >&2\nexit 2\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write stub sg: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	rulesDir := filepath.Join(t.TempDir(), "rules")
	if err := os.MkdirAll(filepath.Join(rulesDir, "go"), 0o755); err != nil {
		t.Fatalf("failed to create rules directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs:\n  - go\n"), 0o644); err != nil {
		t.Fatalf("failed to write sgconfig.yml: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	cmd := cli.NewAstGrepCmd()
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	if err := cmd.Flags().Set("rules-dir", rulesDir); err != nil {
		t.Fatalf("failed to set --rules-dir: %v", err)
	}
	cmd.SetArgs([]string{t.TempDir()})
	err := cmd.Execute()

	var coded interface{ ExitCode() int }
	if err == nil || !errors.As(err, &coded) || coded.ExitCode() == 0 {
		t.Fatalf("Execute() error = %v; want a non-zero exit for a usage error (stdout=%q)", err, outBuf.String())
	}
	if strings.Contains(errBuf.String(), "ruleset failed to load") {
		t.Errorf("stderr = %q; a usage error must not carry the ruleset-load headline", errBuf.String())
	}
}
