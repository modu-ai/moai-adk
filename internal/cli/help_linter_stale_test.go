package cli

// SPEC-CLIFIX-LINTER-STALE-001 REQ-LINT-001-005 regression tests: the root
// --help output must not advertise the phantom "moai brain" command, and every
// advertised help row must resolve to a registered cobra subcommand.

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

// runRootHelpCapture executes `moai --help` against the package rootCmd and
// returns the captured output. Reused by both checks below.
func runRootHelpCapture(t *testing.T) string {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		// --help is non-fatal; tolerate either nil or the help-induced path.
		_ = err
	}
	return buf.String()
}

// TestNoPhantomBrain proves REQ-LINT-001-005: the phantom "moai brain" entry
// no longer appears in root --help output.
func TestNoPhantomBrain(t *testing.T) {
	out := runRootHelpCapture(t)
	if strings.Contains(out, "brain") {
		t.Errorf("root --help must not advertise the phantom 'moai brain' command; output contained 'brain':\n%s", out)
	}
}

// TestHelpRegisteredCommands proves REQ-LINT-001-005 data-driven gate: every
// "moai <name>" row advertised in --help resolves to a registered cobra
// subcommand. The row pattern matches the rendered help (padded command column
// followed by the description); we extract the leading "moai <name>" token.
func TestHelpRegisteredCommands(t *testing.T) {
	out := runRootHelpCapture(t)

	registered := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		if c.Name() != "" {
			registered["moai "+c.Name()] = true
		}
	}

	// Match rendered help rows: lines beginning with whitespace + "moai <word>".
	// The tui render pads the command column; the leading "moai <name>" prefix
	// is stable. ANSI styling is stripped before matching.
	stripAnsi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	rowRe := regexp.MustCompile(`(?m)^\s+(moai\s+[A-Za-z][A-Za-z0-9-]*)\s`)

	for _, m := range rowRe.FindAllStringSubmatch(stripAnsi.ReplaceAllString(out, ""), -1) {
		row := m[1]
		if !registered[row] {
			t.Errorf("help advertises %q but no cobra subcommand is registered for it (REQ-LINT-001-005 gate)", row)
		}
	}
}
