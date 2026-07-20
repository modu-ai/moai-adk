package cli

// SPEC-CLI-TUX-V3-001 REQ-CTX-016 / AC-CTX-016: migrated init.go warning
// paths write to stderr, never stdout.

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
)

// TestInitWarningChannelStderr verifies the init warning-channel contract
// at three levels:
//
//  1. Wiring: the printer constructed the way runInit constructs it routes
//     Warn to the command's stderr stream and never stdout.
//  2. Full run: a real non-interactive init with separated out/err buffers
//     leaks no "Warning:" (and no success card) onto stdout.
//  3. Source: init.go no longer contains the legacy
//     fmt.Fprintf(cmd.OutOrStdout(), "Warning: ...") pattern.
func TestInitWarningChannelStderr(t *testing.T) {
	// --- Level 1: printer wiring (positive stderr assertion) ---
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	p := printer.New(printer.WithWriters(outBuf, errBuf))
	p.Warn("binary update check failed: %v", os.ErrNotExist)

	if !strings.Contains(errBuf.String(), "Warning: binary update check failed") {
		t.Errorf("stderr capture missing warning, got %q", errBuf.String())
	}
	if strings.Contains(outBuf.String(), "Warning:") {
		t.Errorf("stdout capture must not carry warnings, got %q", outBuf.String())
	}

	// --- Level 2: full non-interactive init run with split streams ---
	root := t.TempDir()
	runOut := new(bytes.Buffer)
	runErr := new(bytes.Buffer)
	initCmd.SetOut(runOut)
	initCmd.SetErr(runErr)

	for flag, value := range map[string]string{
		"root":            root,
		"name":            "channel-test",
		"language":        "go",
		"mode":            "tdd",
		"non-interactive": "true",
	} {
		if err := initCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set %s flag: %v", flag, err)
		}
	}
	defer func() {
		for _, flag := range []string{"root", "name", "language", "mode"} {
			if err := initCmd.Flags().Set(flag, ""); err != nil {
				t.Logf("reset %s: %v", flag, err)
			}
		}
		if err := initCmd.Flags().Set("non-interactive", "false"); err != nil {
			t.Logf("reset non-interactive: %v", err)
		}
	}()

	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init RunE error: %v", err)
	}

	if strings.Contains(runOut.String(), "Warning:") {
		t.Errorf("stdout must not carry warnings; stdout = %q", runOut.String())
	}
	if !strings.Contains(runErr.String(), "MoAI project initialized") {
		t.Errorf("success card must render on stderr; stderr = %q", runErr.String())
	}
	if strings.Contains(runOut.String(), "MoAI project initialized") {
		t.Errorf("success card must not render on stdout; stdout = %q", runOut.String())
	}

	// --- Level 3: source-level migration guard (REQ-CTX-016) ---
	src, err := os.ReadFile("init.go")
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	for _, forbidden := range []string{
		`cmd.OutOrStdout(), "Warning:`,
		`cmd.OutOrStdout(), "  Warning:`,
	} {
		if strings.Contains(string(src), forbidden) {
			t.Errorf("init.go still routes warnings to stdout: found %q", forbidden)
		}
	}
}
